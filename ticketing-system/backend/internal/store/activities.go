package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Activity struct {
	ID        uuid.UUID
	TicketID  uuid.UUID
	ActorID   uuid.UUID
	ActorName string
	Action    string
	Field     *string
	OldValue  *string
	NewValue  *string
	CreatedAt time.Time
}

type ActivityCreateInput struct {
	ActorID   uuid.UUID
	ActorName string
	Action    string
	Field     *string
	OldValue  *string
	NewValue  *string
}

func (s *Store) ListActivities(ctx context.Context, ticketID uuid.UUID) ([]Activity, error) {
	query := mustSQL("activities_list", nil)
	return queryMany(ctx, s.db, query, scanActivity, ticketID)
}

func (s *Store) CreateActivity(ctx context.Context, ticketID uuid.UUID, input ActivityCreateInput) error {
	query := mustSQL("activities_insert", nil)
	var id uuid.UUID
	return s.db.QueryRow(ctx, query,
		ticketID,
		input.ActorID,
		input.ActorName,
		input.Action,
		input.Field,
		input.OldValue,
		input.NewValue,
	).Scan(&id)
}

func scanActivity(row pgx.Row) (Activity, error) {
	var a Activity
	err := row.Scan(
		&a.ID,
		&a.TicketID,
		&a.ActorID,
		&a.ActorName,
		&a.Action,
		&a.Field,
		&a.OldValue,
		&a.NewValue,
		&a.CreatedAt,
	)
	return a, err
}
