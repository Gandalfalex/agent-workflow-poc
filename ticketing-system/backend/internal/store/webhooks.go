package store

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Webhook struct {
	ID        uuid.UUID
	URL       string
	Events    []string
	Enabled   bool
	Secret    *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type WebhookCreateInput struct {
	URL     string
	Events  []string
	Enabled bool
	Secret  *string
}

type WebhookUpdateInput struct {
	URL     *string
	Events  *[]string
	Enabled *bool
	Secret  *string
}

func (s *Store) ListWebhooks(ctx context.Context, projectID uuid.UUID) ([]Webhook, error) {
	query := mustSQL("webhooks_list", nil)
	return queryMany(ctx, s.db, query, scanWebhook, projectID)
}

func (s *Store) ListWebhooksForEvent(ctx context.Context, projectID uuid.UUID, event string) ([]Webhook, error) {
	query := mustSQL("webhooks_list_by_event", nil)
	return queryMany(ctx, s.db, query, scanWebhook, projectID, event)
}

func (s *Store) GetWebhook(ctx context.Context, projectID uuid.UUID, id uuid.UUID) (Webhook, error) {
	query := mustSQL("webhooks_get", nil)
	return queryOne(ctx, s.db, query, scanWebhook, projectID, id)
}

func (s *Store) CreateWebhook(ctx context.Context, projectID uuid.UUID, input WebhookCreateInput) (Webhook, error) {
	if err := validateWebhookURL(input.URL); err != nil {
		return Webhook{}, err
	}
	if len(input.Events) == 0 {
		return Webhook{}, errors.New("events required")
	}
	if err := validateWebhookEvents(input.Events); err != nil {
		return Webhook{}, err
	}

	payload, err := json.Marshal(input.Events)
	if err != nil {
		return Webhook{}, err
	}

	var id uuid.UUID
	query := mustSQL("webhooks_insert", nil)
	row := s.db.QueryRow(ctx, query, projectID, input.URL, payload, input.Enabled, input.Secret)

	if err := row.Scan(&id); err != nil {
		return Webhook{}, err
	}

	return s.GetWebhook(ctx, projectID, id)
}

func (s *Store) UpdateWebhook(ctx context.Context, projectID uuid.UUID, id uuid.UUID, input WebhookUpdateInput) (Webhook, error) {
	updates := []string{"updated_at = now()"}
	args := []any{}
	arg := func(value any) string {
		args = append(args, value)
		return "$" + strconv.Itoa(len(args))
	}

	if input.URL != nil {
		if err := validateWebhookURL(*input.URL); err != nil {
			return Webhook{}, err
		}
		updates = append(updates, "url = "+arg(*input.URL))
	}

	if input.Events != nil {
		if len(*input.Events) == 0 {
			return Webhook{}, errors.New("events required")
		}
		if err := validateWebhookEvents(*input.Events); err != nil {
			return Webhook{}, err
		}
		payload, err := json.Marshal(*input.Events)
		if err != nil {
			return Webhook{}, err
		}
		updates = append(updates, "events = "+arg(payload))
	}

	if input.Enabled != nil {
		updates = append(updates, "enabled = "+arg(*input.Enabled))
	}

	if input.Secret != nil {
		updates = append(updates, "secret = "+arg(*input.Secret))
	}

	if len(updates) == 1 {
		return Webhook{}, errors.New("no updates")
	}

	args = append(args, projectID, id)
	query := mustSQL("webhooks_update", map[string]any{
		"Updates":    strings.Join(updates, ", "),
		"ProjectArg": len(args) - 1,
		"IDArg":      len(args),
	})
	if err := execOne(ctx, s.db, query, pgx.ErrNoRows, args...); err != nil {
		return Webhook{}, err
	}

	return s.GetWebhook(ctx, projectID, id)
}

func (s *Store) DeleteWebhook(ctx context.Context, projectID uuid.UUID, id uuid.UUID) error {
	query := mustSQL("webhooks_delete", nil)
	return execOne(ctx, s.db, query, pgx.ErrNoRows, projectID, id)
}

func validateWebhookURL(raw string) error {
	value := strings.TrimSpace(raw)
	if value == "" {
		return errors.New("url required")
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return errors.New("invalid url")
	}
	return nil
}

func validateWebhookEvents(events []string) error {
	allowed := map[string]bool{
		"ticket.created":       true,
		"ticket.updated":       true,
		"ticket.deleted":       true,
		"ticket.state_changed": true,
	}
	for _, event := range events {
		if !allowed[event] {
			return errors.New("invalid event")
		}
	}
	return nil
}

func scanWebhook(row pgx.Row) (Webhook, error) {
	var hook Webhook
	var eventsRaw []byte
	if err := row.Scan(&hook.ID, &hook.URL, &eventsRaw, &hook.Enabled, &hook.Secret, &hook.CreatedAt, &hook.UpdatedAt); err != nil {
		return Webhook{}, err
	}
	if err := json.Unmarshal(eventsRaw, &hook.Events); err != nil {
		return Webhook{}, err
	}
	return hook, nil
}
