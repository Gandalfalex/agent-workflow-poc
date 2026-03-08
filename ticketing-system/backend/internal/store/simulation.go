package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type SimulationRun struct {
	ID           uuid.UUID
	ProjectID    uuid.UUID
	RuleID       *uuid.UUID
	RuleSnapshot map[string]any
	LookbackDays int
	TotalEvents  int
	MatchedCount int
	FailureCount int
	CreatedAt    time.Time
}

type SimulationResult struct {
	ID               uuid.UUID
	RunID            uuid.UUID
	TicketID         uuid.UUID
	TicketKey        string
	EventType        string
	EventTimestamp   time.Time
	Matched          bool
	FailureReason    *string
	PredictedActions []SimulatedActionOutcome
}

type SimulatedActionOutcome struct {
	Type          string            `json:"type"`
	Params        map[string]string `json:"params"`
	WouldSucceed  bool              `json:"wouldSucceed"`
	FailureReason *string           `json:"failureReason,omitempty"`
}

// SimulationEvent is a reconstructed event for dry-run evaluation.
type SimulationEvent struct {
	TicketID  uuid.UUID
	TicketKey string
	EventType string
	Timestamp time.Time
	Extra     map[string]string
}

type SimulationRunCreateInput struct {
	ProjectID    uuid.UUID
	RuleID       *uuid.UUID
	RuleSnapshot map[string]any
	LookbackDays int
	TotalEvents  int
	MatchedCount int
	FailureCount int
}

type SimulationResultCreateInput struct {
	RunID            uuid.UUID
	TicketID         uuid.UUID
	TicketKey        string
	EventType        string
	EventTimestamp   time.Time
	Matched          bool
	FailureReason    *string
	PredictedActions []SimulatedActionOutcome
}

func (s *Store) CreateSimulationRun(ctx context.Context, input SimulationRunCreateInput) (SimulationRun, error) {
	snapshotJSON, _ := json.Marshal(input.RuleSnapshot)
	return queryOne(ctx, s.db, mustSQL("simulation_run_insert", nil), scanSimulationRun,
		input.ProjectID,
		input.RuleID,
		snapshotJSON,
		input.LookbackDays,
		input.TotalEvents,
		input.MatchedCount,
		input.FailureCount,
	)
}

func (s *Store) ListSimulationRuns(ctx context.Context, projectID uuid.UUID) ([]SimulationRun, error) {
	return queryMany(ctx, s.db, mustSQL("simulation_runs_by_project", nil), scanSimulationRun, projectID)
}

func (s *Store) GetSimulationRun(ctx context.Context, id, projectID uuid.UUID) (SimulationRun, error) {
	return queryOne(ctx, s.db, mustSQL("simulation_run_by_id", nil), scanSimulationRun, id, projectID)
}

func (s *Store) CreateSimulationResult(ctx context.Context, input SimulationResultCreateInput) error {
	actionsJSON, _ := json.Marshal(input.PredictedActions)
	_, err := s.db.Exec(ctx, mustSQL("simulation_result_insert", nil),
		input.RunID,
		input.TicketID,
		input.TicketKey,
		input.EventType,
		input.EventTimestamp,
		input.Matched,
		input.FailureReason,
		actionsJSON,
	)
	return err
}

func (s *Store) ListSimulationResults(ctx context.Context, runID uuid.UUID, limit, offset int) ([]SimulationResult, int, error) {
	if limit <= 0 {
		limit = 50
	}
	var total int
	if err := s.db.QueryRow(ctx, mustSQL("simulation_results_count", nil), runID).Scan(&total); err != nil {
		return nil, 0, err
	}
	results, err := queryMany(ctx, s.db, mustSQL("simulation_results_by_run", nil), scanSimulationResult, runID, limit, offset)
	return results, total, err
}

// GetSimulationEvents returns reconstructed events for a given trigger event type and lookback window.
func (s *Store) GetSimulationEvents(ctx context.Context, projectID uuid.UUID, triggerEvent string, lookbackDays int) ([]SimulationEvent, error) {
	switch triggerEvent {
	case "ticket.state_changed":
		return s.getStateChangeEvents(ctx, projectID, lookbackDays)
	case "ticket.created":
		return s.getCreatedEvents(ctx, projectID, lookbackDays)
	case "ticket.updated":
		return s.getUpdatedEvents(ctx, projectID, lookbackDays)
	default:
		return nil, nil
	}
}

func (s *Store) getStateChangeEvents(ctx context.Context, projectID uuid.UUID, lookbackDays int) ([]SimulationEvent, error) {
	rows, err := s.db.Query(ctx, mustSQL("simulation_state_change_events", nil), projectID, lookbackDays)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []SimulationEvent
	for rows.Next() {
		var ticketID uuid.UUID
		var key, priority, fromState, toState string
		var ts time.Time
		if err := rows.Scan(&ticketID, &key, &priority, &ts, &fromState, &toState); err != nil {
			return nil, err
		}
		events = append(events, SimulationEvent{
			TicketID:  ticketID,
			TicketKey: key,
			EventType: "ticket.state_changed",
			Timestamp: ts,
			Extra:     map[string]string{"fromState": fromState, "toState": toState, "priority": priority},
		})
	}
	return events, rows.Err()
}

func (s *Store) getCreatedEvents(ctx context.Context, projectID uuid.UUID, lookbackDays int) ([]SimulationEvent, error) {
	rows, err := s.db.Query(ctx, mustSQL("simulation_created_events", nil), projectID, lookbackDays)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []SimulationEvent
	for rows.Next() {
		var ticketID uuid.UUID
		var key, priority, stateID string
		var ts time.Time
		if err := rows.Scan(&ticketID, &key, &priority, &stateID, &ts); err != nil {
			return nil, err
		}
		events = append(events, SimulationEvent{
			TicketID:  ticketID,
			TicketKey: key,
			EventType: "ticket.created",
			Timestamp: ts,
			Extra:     map[string]string{"priority": priority, "stateId": stateID},
		})
	}
	return events, rows.Err()
}

func (s *Store) getUpdatedEvents(ctx context.Context, projectID uuid.UUID, lookbackDays int) ([]SimulationEvent, error) {
	rows, err := s.db.Query(ctx, mustSQL("simulation_updated_events", nil), projectID, lookbackDays)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []SimulationEvent
	for rows.Next() {
		var ticketID uuid.UUID
		var key, priority, stateID string
		var ts time.Time
		if err := rows.Scan(&ticketID, &key, &priority, &stateID, &ts); err != nil {
			return nil, err
		}
		events = append(events, SimulationEvent{
			TicketID:  ticketID,
			TicketKey: key,
			EventType: "ticket.updated",
			Timestamp: ts,
			Extra:     map[string]string{"priority": priority, "stateId": stateID},
		})
	}
	return events, rows.Err()
}

func scanSimulationRun(row pgx.Row) (SimulationRun, error) {
	var r SimulationRun
	var snapshotRaw []byte
	err := row.Scan(
		&r.ID,
		&r.ProjectID,
		&r.RuleID,
		&snapshotRaw,
		&r.LookbackDays,
		&r.TotalEvents,
		&r.MatchedCount,
		&r.FailureCount,
		&r.CreatedAt,
	)
	if err != nil {
		return r, err
	}
	if len(snapshotRaw) > 0 {
		_ = json.Unmarshal(snapshotRaw, &r.RuleSnapshot)
	}
	return r, nil
}

func scanSimulationResult(row pgx.Row) (SimulationResult, error) {
	var r SimulationResult
	var actionsRaw []byte
	err := row.Scan(
		&r.ID,
		&r.RunID,
		&r.TicketID,
		&r.TicketKey,
		&r.EventType,
		&r.EventTimestamp,
		&r.Matched,
		&r.FailureReason,
		&actionsRaw,
	)
	if err != nil {
		return r, err
	}
	if len(actionsRaw) > 0 {
		_ = json.Unmarshal(actionsRaw, &r.PredictedActions)
	}
	if r.PredictedActions == nil {
		r.PredictedActions = []SimulatedActionOutcome{}
	}
	return r, nil
}
