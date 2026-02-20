package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type StatCount struct {
	Label string
	Value int
}

type ProjectStats struct {
	TotalOpen   int
	TotalClosed int
	BlockedOpen int
	ByState     []StatCount
	ByPriority  []StatCount
	ByType      []StatCount
	ByAssignee  []StatCount
}

func scanStatCount(row pgx.Row) (StatCount, error) {
	var sc StatCount
	err := row.Scan(&sc.Label, &sc.Value)
	return sc, err
}

func (s *Store) GetProjectStats(ctx context.Context, projectID uuid.UUID) (ProjectStats, error) {
	var stats ProjectStats

	byState, err := queryMany(ctx, s.db, mustSQL("stats_by_state", nil), scanStatCount, projectID)
	if err != nil {
		return stats, err
	}
	stats.ByState = byState

	byPriority, err := queryMany(ctx, s.db, mustSQL("stats_by_priority", nil), scanStatCount, projectID)
	if err != nil {
		return stats, err
	}
	stats.ByPriority = byPriority

	byType, err := queryMany(ctx, s.db, mustSQL("stats_by_type", nil), scanStatCount, projectID)
	if err != nil {
		return stats, err
	}
	stats.ByType = byType

	byAssignee, err := queryMany(ctx, s.db, mustSQL("stats_by_assignee", nil), scanStatCount, projectID)
	if err != nil {
		return stats, err
	}
	stats.ByAssignee = byAssignee

	// Open/closed counts
	type openClosedRow struct {
		IsClosed bool
		Count    int
	}
	rows, err := s.db.Query(ctx, mustSQL("stats_open_closed", nil), projectID)
	if err != nil {
		return stats, err
	}
	defer rows.Close()
	for rows.Next() {
		var isClosed bool
		var count int
		if err := rows.Scan(&isClosed, &count); err != nil {
			return stats, err
		}
		if isClosed {
			stats.TotalClosed = count
		} else {
			stats.TotalOpen = count
		}
	}
	if err := rows.Err(); err != nil {
		return stats, err
	}

	if err := s.db.QueryRow(ctx, mustSQL("stats_blocked_open", nil), projectID).Scan(&stats.BlockedOpen); err != nil {
		return stats, err
	}

	return stats, nil
}
