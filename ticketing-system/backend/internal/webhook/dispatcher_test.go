package webhook

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ticketing-system/backend/internal/store"

	"github.com/google/uuid"
)

type fakeStore struct {
	webhooks []store.Webhook
	err      error
}

func (f *fakeStore) ListWebhooksForEvent(ctx context.Context, projectID uuid.UUID, event string) ([]store.Webhook, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.webhooks, nil
}

func (f *fakeStore) CreateWebhookDelivery(ctx context.Context, input store.WebhookDeliveryCreateInput) (store.WebhookDelivery, error) {
	return store.WebhookDelivery{}, nil
}

func TestNew(t *testing.T) {
	d := New(&fakeStore{})

	if d.store == nil {
		t.Error("expected store to be set")
	}
	if d.client == nil {
		t.Error("expected http client to be set")
	}
	if d.client.Timeout != 6*time.Second {
		t.Errorf("expected 6s timeout, got %v", d.client.Timeout)
	}
}

func TestSign(t *testing.T) {
	tests := []struct {
		name     string
		secret   string
		payload  []byte
		expected string
	}{
		{
			name:     "basic signing",
			secret:   "mysecret",
			payload:  []byte(`{"event":"test"}`),
			expected: "sha256=",
		},
		{
			name:     "empty payload",
			secret:   "mysecret",
			payload:  []byte{},
			expected: "sha256=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sign(tt.secret, tt.payload)
			if !strings.HasPrefix(result, tt.expected) {
				t.Errorf("expected signature to start with %q, got %q", tt.expected, result)
			}
			// Verify signature is deterministic
			result2 := sign(tt.secret, tt.payload)
			if result != result2 {
				t.Error("signature should be deterministic")
			}
		})
	}

	t.Run("different secrets produce different signatures", func(t *testing.T) {
		payload := []byte(`{"event":"test"}`)
		sig1 := sign("secret1", payload)
		sig2 := sign("secret2", payload)
		if sig1 == sig2 {
			t.Error("different secrets should produce different signatures")
		}
	})

	t.Run("different payloads produce different signatures", func(t *testing.T) {
		secret := "mysecret"
		sig1 := sign(secret, []byte(`{"event":"test1"}`))
		sig2 := sign(secret, []byte(`{"event":"test2"}`))
		if sig1 == sig2 {
			t.Error("different payloads should produce different signatures")
		}
	})
}

func TestDeliver(t *testing.T) {
	t.Run("successful delivery", func(t *testing.T) {
		var receivedBody []byte
		var receivedHeaders http.Header

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedHeaders = r.Header
			receivedBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		}))
		defer server.Close()

		d := New(&fakeStore{})
		hook := store.Webhook{
			ID:      uuid.New(),
			URL:     server.URL,
			Events:  []string{"ticket.created"},
			Enabled: true,
		}

		envelope := Envelope{
			Event:  "ticket.created",
			SentAt: time.Now().UTC(),
			Data:   map[string]string{"id": "123"},
		}

		result, err := d.deliver(context.Background(), hook, envelope)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !result.Delivered {
			t.Error("expected delivered to be true")
		}
		if result.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", result.StatusCode)
		}
		if result.ResponseBody != "OK" {
			t.Errorf("expected response body 'OK', got %q", result.ResponseBody)
		}
		if result.Error != nil {
			t.Errorf("expected no error, got %v", result.Error)
		}

		// Verify headers
		if receivedHeaders.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type 'application/json', got %q", receivedHeaders.Get("Content-Type"))
		}
		if receivedHeaders.Get("X-Ticketing-Event") != "ticket.created" {
			t.Errorf("expected X-Ticketing-Event 'ticket.created', got %q", receivedHeaders.Get("X-Ticketing-Event"))
		}

		// Verify body
		var received Envelope
		if err := json.Unmarshal(receivedBody, &received); err != nil {
			t.Fatalf("failed to unmarshal received body: %v", err)
		}
		if received.Event != "ticket.created" {
			t.Errorf("expected event 'ticket.created', got %q", received.Event)
		}
	})

	t.Run("delivery with signature", func(t *testing.T) {
		var receivedSignature string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedSignature = r.Header.Get("X-Ticketing-Signature")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		d := New(&fakeStore{})
		secret := "my-webhook-secret"
		hook := store.Webhook{
			ID:      uuid.New(),
			URL:     server.URL,
			Events:  []string{"ticket.created"},
			Enabled: true,
			Secret:  &secret,
		}

		envelope := Envelope{
			Event:  "ticket.created",
			SentAt: time.Now().UTC(),
			Data:   map[string]string{"id": "123"},
		}

		_, err := d.deliver(context.Background(), hook, envelope)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if receivedSignature == "" {
			t.Error("expected signature header to be set")
		}
		if !strings.HasPrefix(receivedSignature, "sha256=") {
			t.Errorf("expected signature to start with 'sha256=', got %q", receivedSignature)
		}
	})

	t.Run("no signature when secret is empty", func(t *testing.T) {
		var receivedSignature string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedSignature = r.Header.Get("X-Ticketing-Signature")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		d := New(&fakeStore{})
		emptySecret := ""
		hook := store.Webhook{
			ID:      uuid.New(),
			URL:     server.URL,
			Events:  []string{"ticket.created"},
			Enabled: true,
			Secret:  &emptySecret,
		}

		envelope := Envelope{
			Event:  "ticket.created",
			SentAt: time.Now().UTC(),
			Data:   nil,
		}

		_, err := d.deliver(context.Background(), hook, envelope)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if receivedSignature != "" {
			t.Errorf("expected no signature header when secret is empty, got %q", receivedSignature)
		}
	})

	t.Run("no signature when secret is nil", func(t *testing.T) {
		var receivedSignature string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedSignature = r.Header.Get("X-Ticketing-Signature")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		d := New(&fakeStore{})
		hook := store.Webhook{
			ID:      uuid.New(),
			URL:     server.URL,
			Events:  []string{"ticket.created"},
			Enabled: true,
			Secret:  nil,
		}

		envelope := Envelope{
			Event:  "ticket.created",
			SentAt: time.Now().UTC(),
			Data:   nil,
		}

		_, err := d.deliver(context.Background(), hook, envelope)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if receivedSignature != "" {
			t.Errorf("expected no signature header when secret is nil, got %q", receivedSignature)
		}
	})

	t.Run("server error returns not delivered", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		}))
		defer server.Close()

		d := New(&fakeStore{})
		hook := store.Webhook{
			ID:      uuid.New(),
			URL:     server.URL,
			Events:  []string{"ticket.created"},
			Enabled: true,
		}

		envelope := Envelope{
			Event:  "ticket.created",
			SentAt: time.Now().UTC(),
			Data:   nil,
		}

		result, err := d.deliver(context.Background(), hook, envelope)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Delivered {
			t.Error("expected delivered to be false for 5xx status")
		}
		if result.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", result.StatusCode)
		}
	})

	t.Run("4xx returns not delivered", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		d := New(&fakeStore{})
		hook := store.Webhook{
			ID:      uuid.New(),
			URL:     server.URL,
			Events:  []string{"ticket.created"},
			Enabled: true,
		}

		envelope := Envelope{
			Event:  "ticket.created",
			SentAt: time.Now().UTC(),
			Data:   nil,
		}

		result, err := d.deliver(context.Background(), hook, envelope)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Delivered {
			t.Error("expected delivered to be false for 4xx status")
		}
	})

	t.Run("2xx statuses are delivered", func(t *testing.T) {
		statuses := []int{200, 201, 202, 204, 299}

		for _, status := range statuses {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
			}))

			d := New(&fakeStore{})
			hook := store.Webhook{
				ID:      uuid.New(),
				URL:     server.URL,
				Events:  []string{"ticket.created"},
				Enabled: true,
			}

			envelope := Envelope{
				Event:  "ticket.created",
				SentAt: time.Now().UTC(),
				Data:   nil,
			}

			result, err := d.deliver(context.Background(), hook, envelope)
			server.Close()

			if err != nil {
				t.Fatalf("status %d: unexpected error: %v", status, err)
			}
			if !result.Delivered {
				t.Errorf("status %d: expected delivered to be true", status)
			}
		}
	})
}

func TestTest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))
	defer server.Close()

	d := New(&fakeStore{})
	hook := store.Webhook{
		ID:      uuid.New(),
		URL:     server.URL,
		Events:  []string{"ticket.updated"},
		Enabled: true,
	}

	result, err := d.Test(context.Background(), hook, "ticket.updated", map[string]string{"test": "data"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Delivered {
		t.Error("expected delivered to be true")
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", result.StatusCode)
	}
	if result.ResponseBody != "test response" {
		t.Errorf("expected response body 'test response', got %q", result.ResponseBody)
	}
}

func TestDispatch(t *testing.T) {
	t.Run("dispatches to multiple webhooks", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		projectID := uuid.New()
		webhooks := []store.Webhook{
			{ID: uuid.New(), URL: server.URL, Events: []string{"ticket.created"}, Enabled: true},
			{ID: uuid.New(), URL: server.URL, Events: []string{"ticket.created"}, Enabled: true},
		}

		d := New(&fakeStore{webhooks: webhooks})
		d.Dispatch(context.Background(), projectID, "ticket.created", map[string]string{"id": "123"})

		// Give goroutines time to complete
		time.Sleep(100 * time.Millisecond)

		if callCount != 2 {
			t.Errorf("expected 2 webhook calls, got %d", callCount)
		}
	})

	t.Run("handles store error gracefully", func(t *testing.T) {
		d := New(&fakeStore{err: context.DeadlineExceeded})

		// Should not panic
		d.Dispatch(context.Background(), uuid.New(), "ticket.created", nil)
	})

	t.Run("handles empty webhooks list", func(t *testing.T) {
		d := New(&fakeStore{webhooks: []store.Webhook{}})

		// Should not panic
		d.Dispatch(context.Background(), uuid.New(), "ticket.created", nil)
	})
}

func TestEnvelope_JSONMarshaling(t *testing.T) {
	sentAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	envelope := Envelope{
		Event:  "ticket.created",
		SentAt: sentAt,
		Data: map[string]any{
			"ticketId": "123",
			"title":    "Test ticket",
		},
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("failed to marshal envelope: %v", err)
	}

	var decoded Envelope
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal envelope: %v", err)
	}

	if decoded.Event != "ticket.created" {
		t.Errorf("expected event 'ticket.created', got %q", decoded.Event)
	}
	if !decoded.SentAt.Equal(sentAt) {
		t.Errorf("expected sentAt %v, got %v", sentAt, decoded.SentAt)
	}
}
