package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type TicketWebhookEvent struct {
	ID        uuid.UUID
	TicketID  uuid.UUID
	WebhookID *uuid.UUID
	Event     string
	Payload   map[string]any
	Delivered bool
	CreatedAt time.Time
}

type TicketWebhookEventCreateInput struct {
	TicketID  uuid.UUID
	WebhookID *uuid.UUID
	Event     string
	Payload   map[string]any
	Delivered bool
}

func (s *Store) CreateTicketWebhookEvent(ctx context.Context, input TicketWebhookEventCreateInput) error {
	query := mustSQL("ticket_webhook_events_insert", nil)
	payload := input.Payload
	if payload == nil {
		payload = map[string]any{}
	}
	_, err := s.db.Exec(ctx, query, input.TicketID, input.WebhookID, input.Event, payload, input.Delivered)
	return err
}

func (s *Store) ListTicketWebhookEvents(ctx context.Context, ticketID uuid.UUID) ([]TicketWebhookEvent, error) {
	query := mustSQL("ticket_webhook_events_list", nil)
	return queryMany(ctx, s.db, query, scanTicketWebhookEvent, ticketID)
}

func scanTicketWebhookEvent(row pgx.Row) (TicketWebhookEvent, error) {
	var (
		event       TicketWebhookEvent
		payloadJSON []byte
	)
	err := row.Scan(
		&event.ID,
		&event.TicketID,
		&event.WebhookID,
		&event.Event,
		&payloadJSON,
		&event.Delivered,
		&event.CreatedAt,
	)
	if err != nil {
		return TicketWebhookEvent{}, err
	}
	if len(payloadJSON) > 0 {
		_ = json.Unmarshal(payloadJSON, &event.Payload)
	}
	if event.Payload == nil {
		event.Payload = map[string]any{}
	}
	return event, nil
}
