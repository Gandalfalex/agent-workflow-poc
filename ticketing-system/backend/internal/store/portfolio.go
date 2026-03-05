package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ProjectPortfolioEntry struct {
	ID                    uuid.UUID
	ProjectKey            string
	ProjectName           string
	TotalOpen             int
	TotalClosed           int
	UrgentOpen            int
	HighOpen              int
	BlockedOpen           int
	WeeklyThroughput      int
	AvgCycleTimeHours     float64
	ActiveSprintName      *string
	ActiveSprintEndDate   *time.Time
	ActiveSprintCommitted int
	ActiveSprintRemaining int
}

type PortfolioFilter struct {
	OwnerID *uuid.UUID
	GroupID *uuid.UUID
}

type PortfolioTotals struct {
	TotalProjects int
	TotalOpen     int
	TotalClosed   int
	TotalBlocked  int
	TotalUrgent   int
}

type PortfolioStats struct {
	Totals   PortfolioTotals
	Projects []ProjectPortfolioEntry
}

func (s *Store) GetPortfolioStats(ctx context.Context, userID uuid.UUID, filter PortfolioFilter) (PortfolioStats, error) {
	var projects []ProjectPortfolioEntry
	var err error

	switch {
	case filter.GroupID != nil:
		projects, err = queryMany(ctx, s.db, mustSQL("portfolio_stats_by_group", nil), scanPortfolioEntry, userID, *filter.GroupID)
	case filter.OwnerID != nil:
		projects, err = queryMany(ctx, s.db, mustSQL("portfolio_stats_by_owner", nil), scanPortfolioEntry, userID, *filter.OwnerID)
	default:
		projects, err = queryMany(ctx, s.db, mustSQL("portfolio_stats_by_user", nil), scanPortfolioEntry, userID)
	}
	if err != nil {
		return PortfolioStats{}, err
	}

	totals := PortfolioTotals{TotalProjects: len(projects)}
	for _, p := range projects {
		totals.TotalOpen += p.TotalOpen
		totals.TotalClosed += p.TotalClosed
		totals.TotalBlocked += p.BlockedOpen
		totals.TotalUrgent += p.UrgentOpen
	}

	return PortfolioStats{Totals: totals, Projects: projects}, nil
}

func scanPortfolioEntry(row pgx.Row) (ProjectPortfolioEntry, error) {
	var out ProjectPortfolioEntry
	err := row.Scan(
		&out.ID,
		&out.ProjectName,
		&out.ProjectKey,
		&out.TotalOpen,
		&out.TotalClosed,
		&out.UrgentOpen,
		&out.HighOpen,
		&out.BlockedOpen,
		&out.WeeklyThroughput,
		&out.AvgCycleTimeHours,
		&out.ActiveSprintName,
		&out.ActiveSprintEndDate,
		&out.ActiveSprintCommitted,
		&out.ActiveSprintRemaining,
	)
	return out, err
}
