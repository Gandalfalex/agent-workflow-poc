package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
	query := mustSQL("workflow_list", nil)
	return queryMany(ctx, s.db, query, scanWorkflowState, projectID)
}

func (s *Store) ReplaceWorkflowStates(ctx context.Context, projectID uuid.UUID, inputs []WorkflowStateInput) ([]WorkflowState, error) {
	return withTx(ctx, s.db, func(tx pgx.Tx) ([]WorkflowState, error) {
		deleteQuery := mustSQL("workflow_delete", nil)
		if _, err := tx.Exec(ctx, deleteQuery, projectID); err != nil {
			return nil, err
		}

		states := make([]WorkflowState, 0, len(inputs))
		now := time.Now().UTC()
		insertQuery := mustSQL("workflow_insert", nil)

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

			_, err := tx.Exec(ctx, insertQuery, state.ID, projectID, state.Name, state.Order, state.IsDefault, state.IsClosed, state.CreatedAt, state.UpdatedAt)
			if err != nil {
				return nil, err
			}

			states = append(states, state)
		}

		return states, nil
	})
}

func scanWorkflowState(row pgx.Row) (WorkflowState, error) {
	var state WorkflowState
	err := row.Scan(
		&state.ID,
		&state.Name,
		&state.Order,
		&state.IsDefault,
		&state.IsClosed,
		&state.CreatedAt,
		&state.UpdatedAt,
	)
	return state, err
}
