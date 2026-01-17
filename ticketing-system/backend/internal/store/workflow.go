package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type WorkflowState struct {
	ID        uuid.UUID
	Name      string
	Order     int
	IsDefault bool
	IsClosed  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type WorkflowStateInput struct {
	ID        *uuid.UUID
	Name      string
	Order     int
	IsDefault bool
	IsClosed  bool
}

func (s *Store) ListWorkflowStates(ctx context.Context, projectID uuid.UUID) ([]WorkflowState, error) {
	query := mustSQL("workflow_list.sql", nil)
	rows, err := s.db.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var states []WorkflowState
	for rows.Next() {
		var state WorkflowState
		if err := rows.Scan(
			&state.ID,
			&state.Name,
			&state.Order,
			&state.IsDefault,
			&state.IsClosed,
			&state.CreatedAt,
			&state.UpdatedAt,
		); err != nil {
			return nil, err
		}
		states = append(states, state)
	}

	return states, rows.Err()
}

func (s *Store) ReplaceWorkflowStates(ctx context.Context, projectID uuid.UUID, inputs []WorkflowStateInput) ([]WorkflowState, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	deleteQuery := mustSQL("workflow_delete.sql", nil)
	if _, err := tx.Exec(ctx, deleteQuery, projectID); err != nil {
		return nil, err
	}

	states := make([]WorkflowState, 0, len(inputs))
	now := time.Now().UTC()

	for _, input := range inputs {
		id := uuid.New()
		if input.ID != nil {
			id = *input.ID
		}

		state := WorkflowState{
			ID:        id,
			Name:      input.Name,
			Order:     input.Order,
			IsDefault: input.IsDefault,
			IsClosed:  input.IsClosed,
			CreatedAt: now,
			UpdatedAt: now,
		}

		insertQuery := mustSQL("workflow_insert.sql", nil)
		_, err := tx.Exec(ctx, insertQuery, state.ID, projectID, state.Name, state.Order, state.IsDefault, state.IsClosed, state.CreatedAt, state.UpdatedAt)
		if err != nil {
			return nil, err
		}

		states = append(states, state)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return states, nil
}
