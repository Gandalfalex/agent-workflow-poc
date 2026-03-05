package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type AuditLogEntry struct {
	ID           uuid.UUID
	ActorID      uuid.UUID
	ActorName    string
	Action       string
	ResourceType string
	ResourceID   string
	ResourceName string
	ProjectID    *uuid.UUID
	Details      map[string]any
	CreatedAt    time.Time
}

type AuditLogCreateInput struct {
	ActorID      uuid.UUID
	ActorName    string
	Action       string
	ResourceType string
	ResourceID   string
	ResourceName string
	ProjectID    *uuid.UUID
	Details      map[string]any
}

type AuditLogFilter struct {
	ProjectID    *uuid.UUID
	ResourceType *string
	ActorID      *uuid.UUID
	Limit        int
	Offset       int
}

func (s *Store) CreateAuditLog(ctx context.Context, input AuditLogCreateInput) error {
	detailsJSON, _ := json.Marshal(input.Details)
	var id uuid.UUID
	return s.db.QueryRow(ctx, mustSQL("audit_log_insert", nil),
		input.ActorID,
		input.ActorName,
		input.Action,
		input.ResourceType,
		input.ResourceID,
		input.ResourceName,
		input.ProjectID,
		detailsJSON,
	).Scan(&id)
}

func (s *Store) ListAuditLog(ctx context.Context, filter AuditLogFilter) ([]AuditLogEntry, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	var total int
	if err := s.db.QueryRow(ctx, mustSQL("audit_log_count", nil),
		filter.ProjectID,
		filter.ResourceType,
		filter.ActorID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	entries, err := queryMany(ctx, s.db, mustSQL("audit_log_list", nil), scanAuditLogEntry,
		filter.ProjectID,
		filter.ResourceType,
		filter.ActorID,
		filter.Limit,
		filter.Offset,
	)
	return entries, total, err
}

func scanAuditLogEntry(row pgx.Row) (AuditLogEntry, error) {
	var e AuditLogEntry
	var detailsRaw []byte
	err := row.Scan(
		&e.ID,
		&e.ActorID,
		&e.ActorName,
		&e.Action,
		&e.ResourceType,
		&e.ResourceID,
		&e.ResourceName,
		&e.ProjectID,
		&detailsRaw,
		&e.CreatedAt,
	)
	if err != nil {
		return e, err
	}
	if len(detailsRaw) > 0 {
		_ = json.Unmarshal(detailsRaw, &e.Details)
	}
	if e.Details == nil {
		e.Details = map[string]any{}
	}
	return e, nil
}
