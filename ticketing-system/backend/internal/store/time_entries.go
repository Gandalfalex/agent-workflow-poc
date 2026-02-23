package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type TimeEntry struct {
	ID          uuid.UUID
	TicketID    uuid.UUID
	UserID      uuid.UUID
	UserName    string
	Minutes     int
	Description *string
	LoggedAt    time.Time
	CreatedAt   time.Time
}

type TimeEntryCreateInput struct {
	UserID      uuid.UUID
	UserName    string
	Minutes     int
	Description *string
	LoggedAt    *time.Time
}

func (s *Store) ListTimeEntries(ctx context.Context, ticketID uuid.UUID) ([]TimeEntry, int, error) {
	listQuery := mustSQL("time_entries_list", nil)
	entries, err := queryMany(ctx, s.db, listQuery, scanTimeEntry, ticketID)
	if err != nil {
		return nil, 0, err
	}

	totalQuery := mustSQL("time_entries_total", nil)
	var total int
	if err := s.db.QueryRow(ctx, totalQuery, ticketID).Scan(&total); err != nil {
		return nil, 0, err
	}

	return entries, total, nil
}

func (s *Store) CreateTimeEntry(ctx context.Context, ticketID uuid.UUID, input TimeEntryCreateInput) (TimeEntry, error) {
	if input.Minutes <= 0 {
		return TimeEntry{}, errors.New("minutes must be positive")
	}

	loggedAt := time.Now()
	if input.LoggedAt != nil {
		loggedAt = *input.LoggedAt
	}

	query := mustSQL("time_entries_insert", nil)
	var id uuid.UUID
	var createdAt time.Time
	if err := s.db.QueryRow(ctx, query,
		ticketID,
		input.UserID,
		input.UserName,
		input.Minutes,
		input.Description,
		loggedAt,
	).Scan(&id, &createdAt); err != nil {
		return TimeEntry{}, err
	}

	return TimeEntry{
		ID:          id,
		TicketID:    ticketID,
		UserID:      input.UserID,
		UserName:    input.UserName,
		Minutes:     input.Minutes,
		Description: input.Description,
		LoggedAt:    loggedAt,
		CreatedAt:   createdAt,
	}, nil
}

func (s *Store) DeleteTimeEntry(ctx context.Context, id uuid.UUID) error {
	query := mustSQL("time_entries_delete", nil)
	return execOne(ctx, s.db, query, pgx.ErrNoRows, id)
}

func scanTimeEntry(row pgx.Row) (TimeEntry, error) {
	var entry TimeEntry
	err := row.Scan(
		&entry.ID,
		&entry.TicketID,
		&entry.UserID,
		&entry.UserName,
		&entry.Minutes,
		&entry.Description,
		&entry.LoggedAt,
		&entry.CreatedAt,
	)
	return entry, err
}
