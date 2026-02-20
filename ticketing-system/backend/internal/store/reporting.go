package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type DateValuePoint struct {
	Date  time.Time
	Value int
}

type StateOpenSeriesPoint struct {
	Date   time.Time
	Counts []StatCount
}

type ProjectReportingSummary struct {
	From                  time.Time
	To                    time.Time
	ThroughputByDay       []DateValuePoint
	AverageCycleTimeHours float64
	OpenByState           []StateOpenSeriesPoint
}

func (s *Store) GetProjectReportingSummary(ctx context.Context, projectID uuid.UUID, from, to time.Time) (ProjectReportingSummary, error) {
	from = normalizeDateUTC(from)
	to = normalizeDateUTC(to)
	if to.Before(from) {
		from, to = to, from
	}

	summary := ProjectReportingSummary{
		From:            from,
		To:              to,
		ThroughputByDay: make([]DateValuePoint, 0),
		OpenByState:     make([]StateOpenSeriesPoint, 0),
	}

	throughput, err := queryMany(ctx, s.db, mustSQL("reporting_throughput_by_day", nil), scanDateValuePoint, projectID, from, to)
	if err != nil {
		return summary, err
	}
	summary.ThroughputByDay = throughput

	if err := s.db.QueryRow(ctx, mustSQL("reporting_average_cycle_time_hours", nil), projectID, from, to).Scan(&summary.AverageCycleTimeHours); err != nil {
		return summary, err
	}

	states, err := s.ListWorkflowStates(ctx, projectID)
	if err != nil {
		return summary, err
	}
	openStates := make([]string, 0, len(states))
	for _, state := range states {
		if !state.IsClosed {
			openStates = append(openStates, state.Name)
		}
	}

	type openStateRow struct {
		Day   time.Time
		State string
		Value int
	}
	scanOpenStateRow := func(row pgx.Row) (openStateRow, error) {
		var item openStateRow
		err := row.Scan(&item.Day, &item.State, &item.Value)
		return item, err
	}
	rows, err := queryMany(ctx, s.db, mustSQL("reporting_open_by_state_series", nil), scanOpenStateRow, projectID, from, to)
	if err != nil {
		return summary, err
	}

	countsByDay := make(map[string]map[string]int)
	for _, row := range rows {
		day := normalizeDateUTC(row.Day).Format("2006-01-02")
		if _, ok := countsByDay[day]; !ok {
			countsByDay[day] = make(map[string]int)
		}
		countsByDay[day][row.State] = row.Value
	}

	day := from
	for !day.After(to) {
		dayKey := day.Format("2006-01-02")
		stateCounts := make([]StatCount, 0, len(openStates))
		for _, stateName := range openStates {
			stateCounts = append(stateCounts, StatCount{
				Label: stateName,
				Value: countsByDay[dayKey][stateName],
			})
		}
		summary.OpenByState = append(summary.OpenByState, StateOpenSeriesPoint{
			Date:   day,
			Counts: stateCounts,
		})
		day = day.AddDate(0, 0, 1)
	}

	return summary, nil
}

func normalizeDateUTC(value time.Time) time.Time {
	return time.Date(value.UTC().Year(), value.UTC().Month(), value.UTC().Day(), 0, 0, 0, 0, time.UTC)
}

func scanDateValuePoint(row pgx.Row) (DateValuePoint, error) {
	var out DateValuePoint
	err := row.Scan(&out.Date, &out.Value)
	return out, err
}
