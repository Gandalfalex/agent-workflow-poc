package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

var ErrReleaseNotFound = errors.New("release not found")

type Release struct {
	ID               uuid.UUID
	ProjectID        uuid.UUID
	Name             string
	Version          *string
	Status           string
	TargetDate       *openapi_types.Date
	Notes            *string
	TicketCount      int
	ClosedTicketCount int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type ReleaseCreateInput struct {
	Name       string
	Version    *string
	Status     string
	TargetDate *openapi_types.Date
	Notes      *string
}

type ReleaseUpdateInput struct {
	Name       *string
	Version    *string
	Status     *string
	TargetDate *openapi_types.Date
	Notes      *string
	ClearTargetDate bool
	ClearVersion    bool
	ClearNotes      bool
}

type ReleaseTicket struct {
	ID        uuid.UUID
	Key       string
	Title     string
	Type      string
	IsClosed  bool
	StateName string
}

func (s *Store) ListReleases(ctx context.Context, projectID uuid.UUID) ([]Release, error) {
	return queryMany(ctx, s.db, mustSQL("releases_list", nil), scanRelease, projectID)
}

func (s *Store) GetRelease(ctx context.Context, id, projectID uuid.UUID) (Release, error) {
	r, err := queryOne(ctx, s.db, mustSQL("release_by_id", nil), scanRelease, id, projectID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Release{}, ErrReleaseNotFound
	}
	return r, err
}

func (s *Store) CreateRelease(ctx context.Context, projectID uuid.UUID, input ReleaseCreateInput) (Release, error) {
	id := uuid.New()
	now := time.Now().UTC()
	status := input.Status
	if status == "" {
		status = "planned"
	}
	_, err := s.db.Exec(ctx, mustSQL("release_insert", nil),
		id, projectID, input.Name, input.Version, status,
		targetDateVal(input.TargetDate), input.Notes, now, now,
	)
	if err != nil {
		return Release{}, err
	}
	return s.GetRelease(ctx, id, projectID)
}

func (s *Store) UpdateRelease(ctx context.Context, id, projectID uuid.UUID, input ReleaseUpdateInput) (Release, error) {
	current, err := s.GetRelease(ctx, id, projectID)
	if err != nil {
		return Release{}, err
	}

	name := current.Name
	if input.Name != nil {
		name = *input.Name
	}
	version := current.Version
	if input.ClearVersion {
		version = nil
	} else if input.Version != nil {
		version = input.Version
	}
	status := current.Status
	if input.Status != nil {
		status = *input.Status
	}
	targetDate := current.TargetDate
	if input.ClearTargetDate {
		targetDate = nil
	} else if input.TargetDate != nil {
		targetDate = input.TargetDate
	}
	notes := current.Notes
	if input.ClearNotes {
		notes = nil
	} else if input.Notes != nil {
		notes = input.Notes
	}

	now := time.Now().UTC()
	_, err = s.db.Exec(ctx, mustSQL("release_update", nil),
		name, version, status, targetDateVal(targetDate), notes, now, id, projectID,
	)
	if err != nil {
		return Release{}, err
	}
	return s.GetRelease(ctx, id, projectID)
}

func (s *Store) DeleteRelease(ctx context.Context, id, projectID uuid.UUID) error {
	_, err := s.db.Exec(ctx, mustSQL("release_delete", nil), id, projectID)
	return err
}

func (s *Store) GetReleaseTickets(ctx context.Context, releaseID uuid.UUID) ([]ReleaseTicket, error) {
	return queryMany(ctx, s.db, mustSQL("release_tickets", nil), scanReleaseTicket, releaseID)
}

func scanRelease(row pgx.Row) (Release, error) {
	var r Release
	err := row.Scan(
		&r.ID, &r.ProjectID, &r.Name, &r.Version, &r.Status,
		&r.TargetDate, &r.Notes, &r.TicketCount, &r.ClosedTicketCount,
		&r.CreatedAt, &r.UpdatedAt,
	)
	return r, err
}

func scanReleaseTicket(row pgx.Row) (ReleaseTicket, error) {
	var t ReleaseTicket
	err := row.Scan(&t.ID, &t.Key, &t.Title, &t.Type, &t.IsClosed, &t.StateName)
	return t, err
}

func targetDateVal(d *openapi_types.Date) any {
	if d == nil {
		return nil
	}
	return d.Time
}
