package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type FlowHealthWipStat struct {
	StateID        uuid.UUID
	StateName      string
	WipLimit       *int
	WipEnforcement bool
	TicketCount    int
	MaxAgeHours    float64
	AvgAgeHours    float64
}

type CycleTimePercentiles struct {
	P50Hours    float64
	P75Hours    float64
	P95Hours    float64
	SampleCount int
}

type ThroughputDay struct {
	Day   time.Time
	Count int
}

type FlowHealthStats struct {
	WipStats    []FlowHealthWipStat
	CycleTime   CycleTimePercentiles
	Throughput  []ThroughputDay
}

func (s *Store) GetFlowHealthStats(ctx context.Context, projectID uuid.UUID) (FlowHealthStats, error) {
	wipStats, err := queryMany(ctx, s.db, mustSQL("flow_health_wip_stats", nil), scanFlowHealthWipStat, projectID)
	if err != nil {
		return FlowHealthStats{}, err
	}

	cycleTime, err := queryOne(ctx, s.db, mustSQL("flow_health_cycle_time", nil), scanCycleTimePercentiles, projectID)
	if err != nil {
		return FlowHealthStats{}, err
	}

	throughput, err := queryMany(ctx, s.db, mustSQL("flow_health_throughput", nil), scanThroughputDay, projectID)
	if err != nil {
		return FlowHealthStats{}, err
	}

	return FlowHealthStats{
		WipStats:   wipStats,
		CycleTime:  cycleTime,
		Throughput: throughput,
	}, nil
}

// GetStateTicketCount returns the number of tickets currently in a state.
func (s *Store) GetStateTicketCount(ctx context.Context, stateID uuid.UUID) (int, error) {
	var count int
	err := s.db.QueryRow(ctx, mustSQL("workflow_state_ticket_count", nil), stateID).Scan(&count)
	return count, err
}

func scanFlowHealthWipStat(row pgx.Row) (FlowHealthWipStat, error) {
	var s FlowHealthWipStat
	err := row.Scan(
		&s.StateID, &s.StateName, &s.WipLimit, &s.WipEnforcement,
		&s.TicketCount, &s.MaxAgeHours, &s.AvgAgeHours,
	)
	return s, err
}

func scanCycleTimePercentiles(row pgx.Row) (CycleTimePercentiles, error) {
	var c CycleTimePercentiles
	err := row.Scan(&c.P50Hours, &c.P75Hours, &c.P95Hours, &c.SampleCount)
	return c, err
}

func scanThroughputDay(row pgx.Row) (ThroughputDay, error) {
	var d ThroughputDay
	err := row.Scan(&d.Day, &d.Count)
	return d, err
}
