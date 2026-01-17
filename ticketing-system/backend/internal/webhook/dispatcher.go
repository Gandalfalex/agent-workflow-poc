package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"ticketing-system/backend/internal/store"

	"github.com/google/uuid"
)

type Store interface {
	ListWebhooksForEvent(ctx context.Context, projectID uuid.UUID, event string) ([]store.Webhook, error)
}

type Dispatcher struct {
	store  Store
	client *http.Client
}

type Envelope struct {
	Event  string    `json:"event"`
	SentAt time.Time `json:"sentAt"`
	Data   any       `json:"data"`
}

type Result struct {
	Delivered    bool
	StatusCode   int
	ResponseBody string
	Error        error
}

func New(store Store) *Dispatcher {
	return &Dispatcher{
		store: store,
		client: &http.Client{
			Timeout: 6 * time.Second,
		},
	}
}

func (d *Dispatcher) Dispatch(ctx context.Context, projectID uuid.UUID, event string, data any) {
	webhooks, err := d.store.ListWebhooksForEvent(ctx, projectID, event)
	if err != nil {
		return
	}

	for _, hook := range webhooks {
		hook := hook
		go func() {
			_, _ = d.deliver(context.Background(), hook, Envelope{
				Event:  event,
				SentAt: time.Now().UTC(),
				Data:   data,
			})
		}()
	}
}

func (d *Dispatcher) Test(ctx context.Context, hook store.Webhook, event string, data any) (Result, error) {
	envelope := Envelope{
		Event:  event,
		SentAt: time.Now().UTC(),
		Data:   data,
	}
	return d.deliver(ctx, hook, envelope)
}

func (d *Dispatcher) deliver(ctx context.Context, hook store.Webhook, envelope Envelope) (Result, error) {
	body, err := json.Marshal(envelope)
	if err != nil {
		return Result{Error: err}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hook.URL, bytes.NewReader(body))
	if err != nil {
		return Result{Error: err}, err
	}
	if hook.Secret != nil && *hook.Secret != "" {
		signature := sign(*hook.Secret, body)
		req.Header.Set("X-Ticketing-Signature", signature)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Ticketing-Event", envelope.Event)

	resp, err := d.client.Do(req)
	if err != nil {
		return Result{Delivered: false, Error: err}, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	return Result{
		Delivered:    resp.StatusCode >= 200 && resp.StatusCode < 300,
		StatusCode:   resp.StatusCode,
		ResponseBody: string(respBody),
		Error:        nil,
	}, nil
}

func sign(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
