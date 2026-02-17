package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type WebhookDelivery struct {
	ID           uuid.UUID
	WebhookID    uuid.UUID
	Event        string
	Attempt      int
	StatusCode   *int
	ResponseBody *string
	Error        *string
	Delivered    bool
	DurationMs   int
	CreatedAt    time.Time
}

type WebhookDeliveryCreateInput struct {
	WebhookID    uuid.UUID
	Event        string
	Attempt      int
	StatusCode   *int
	ResponseBody *string
	Error        *string
	Delivered    bool
	DurationMs   int
}

func (s *Store) ListWebhookDeliveries(ctx context.Context, webhookID uuid.UUID) ([]WebhookDelivery, error) {
	query := mustSQL("webhook_deliveries_list", nil)
	return queryMany(ctx, s.db, query, scanWebhookDelivery, webhookID)
}

func (s *Store) CreateWebhookDelivery(ctx context.Context, input WebhookDeliveryCreateInput) (WebhookDelivery, error) {
	query := mustSQL("webhook_deliveries_insert", nil)
	var id uuid.UUID
	if err := s.db.QueryRow(ctx, query,
		input.WebhookID,
		input.Event,
		input.Attempt,
		input.StatusCode,
		input.ResponseBody,
		input.Error,
		input.Delivered,
		input.DurationMs,
	).Scan(&id); err != nil {
		return WebhookDelivery{}, err
	}

	return WebhookDelivery{
		ID:           id,
		WebhookID:    input.WebhookID,
		Event:        input.Event,
		Attempt:      input.Attempt,
		StatusCode:   input.StatusCode,
		ResponseBody: input.ResponseBody,
		Error:        input.Error,
		Delivered:    input.Delivered,
		DurationMs:   input.DurationMs,
		CreatedAt:    time.Now().UTC(),
	}, nil
}

func scanWebhookDelivery(row pgx.Row) (WebhookDelivery, error) {
	var d WebhookDelivery
	err := row.Scan(
		&d.ID,
		&d.WebhookID,
		&d.Event,
		&d.Attempt,
		&d.StatusCode,
		&d.ResponseBody,
		&d.Error,
		&d.Delivered,
		&d.DurationMs,
		&d.CreatedAt,
	)
	return d, err
}
