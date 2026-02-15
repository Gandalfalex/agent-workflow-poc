package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Ticket struct {
	ID           uuid.UUID
	ProjectID    uuid.UUID
	ProjectKey   string
	Key          string
	Number       int
	Type         string
	StoryID      uuid.UUID
	StoryTitle   string
	StorySummary *string
	StoryCreated time.Time
	StoryUpdated time.Time
	Title        string
	Description  string
	StateID      uuid.UUID
	StateName    string
	StateOrder   int
	StateDefault bool
	StateClosed  bool
	AssigneeID   *uuid.UUID
	AssigneeName *string
	Priority     string
	Position     float64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type TicketFilter struct {
	ProjectID  uuid.UUID
	StateID    *uuid.UUID
	AssigneeID *uuid.UUID
	Query      string
	Limit      int
	Offset     int
}

type TicketCreateInput struct {
	Title       string
	Description string
	Type        string
	StoryID     uuid.UUID
	StateID     *uuid.UUID
	AssigneeID  *uuid.UUID
	Priority    string
}

type TicketUpdateInput struct {
	Title       *string
	Description *string
	Type        *string
	StoryID     *uuid.UUID
	StateID     *uuid.UUID
	AssigneeID  *uuid.UUID
	Priority    *string
	Position    *float64
}

func (s *Store) ListTickets(ctx context.Context, filter TicketFilter) ([]Ticket, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 200 {
		filter.Limit = 200
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	conditions := []string{"t.project_id = $1"}
	args := []any{filter.ProjectID}
	arg := func(value any) string {
		args = append(args, value)
		return fmt.Sprintf("$%d", len(args))
	}

	if filter.StateID != nil {
		conditions = append(conditions, fmt.Sprintf("t.state_id = %s", arg(*filter.StateID)))
	}
	if filter.AssigneeID != nil {
		conditions = append(conditions, fmt.Sprintf("t.assignee_id = %s", arg(*filter.AssigneeID)))
	}
	if strings.TrimSpace(filter.Query) != "" {
		q := "%%" + strings.TrimSpace(filter.Query) + "%%"
		conditions = append(conditions, fmt.Sprintf("(t.title ILIKE %s OR t.description ILIKE %s OR t.key ILIKE %s)", arg(q), arg(q), arg(q)))
	}

	where := strings.Join(conditions, " AND ")

	countSQL := mustSQL("tickets_count", map[string]any{
		"Where": where,
	})
	var total int
	if err := s.db.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, filter.Limit, filter.Offset)
	listSQL := mustSQL("tickets_list", map[string]any{
		"Where":     where,
		"LimitArg":  len(args) - 1,
		"OffsetArg": len(args),
	})

	tickets, err := queryMany(ctx, s.db, listSQL, scanTicket, args...)
	if err != nil {
		return nil, 0, err
	}

	return tickets, total, nil
}

func (s *Store) ListTicketsForBoard(ctx context.Context, projectID uuid.UUID) ([]Ticket, error) {
	query := mustSQL("tickets_board", nil)
	return queryMany(ctx, s.db, query, scanTicket, projectID)
}

func (s *Store) GetTicket(ctx context.Context, id uuid.UUID) (Ticket, error) {
	query := mustSQL("tickets_get", nil)
	return queryOne(ctx, s.db, query, scanTicket, id)
}

func (s *Store) CreateTicket(ctx context.Context, projectID uuid.UUID, input TicketCreateInput) (Ticket, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return Ticket{}, errors.New("title required")
	}

	stateID, err := s.resolveState(ctx, projectID, input.StateID)
	if err != nil {
		return Ticket{}, err
	}

	priority := normalizePriority(input.Priority)
	ticketType, err := normalizeTicketType(input.Type)
	if err != nil {
		return Ticket{}, err
	}

	position, err := s.nextPosition(ctx, stateID)
	if err != nil {
		return Ticket{}, err
	}

	var ticketID uuid.UUID
	query := mustSQL("tickets_insert", nil)
	row := s.db.QueryRow(ctx, query, projectID, title, input.Description, ticketType, input.StoryID, stateID, input.AssigneeID, priority, position)

	if err := row.Scan(&ticketID); err != nil {
		return Ticket{}, err
	}

	return s.GetTicket(ctx, ticketID)
}

func (s *Store) UpdateTicket(ctx context.Context, id uuid.UUID, input TicketUpdateInput) (Ticket, error) {
	_, err := withTx(ctx, s.db, func(tx pgx.Tx) (struct{}, error) {
		var currentState uuid.UUID
		currentStateQuery := mustSQL("tickets_current_state", nil)
		if err := tx.QueryRow(ctx, currentStateQuery, id).Scan(&currentState); err != nil {
			return struct{}{}, err
		}

		newState := currentState
		if input.StateID != nil {
			newState = *input.StateID
		}

		if input.Priority != nil {
			if *input.Priority == "" {
				return struct{}{}, errors.New("priority cannot be empty")
			}
		}

		updates := []string{"updated_at = now()"}
		args := []any{}
		arg := func(value any) string {
			args = append(args, value)
			return fmt.Sprintf("$%d", len(args))
		}

		if input.Title != nil {
			updates = append(updates, fmt.Sprintf("title = %s", arg(strings.TrimSpace(*input.Title))))
		}
		if input.Description != nil {
			updates = append(updates, fmt.Sprintf("description = %s", arg(*input.Description)))
		}
		if input.Type != nil {
			ticketType, err := normalizeTicketType(*input.Type)
			if err != nil {
				return struct{}{}, err
			}
			updates = append(updates, fmt.Sprintf("type = %s", arg(ticketType)))
		}
		if input.StoryID != nil {
			updates = append(updates, fmt.Sprintf("story_id = %s", arg(*input.StoryID)))
		}
		if input.StateID != nil {
			updates = append(updates, fmt.Sprintf("state_id = %s", arg(newState)))
		}
		if input.AssigneeID != nil {
			updates = append(updates, fmt.Sprintf("assignee_id = %s", arg(*input.AssigneeID)))
		}
		if input.Priority != nil {
			updates = append(updates, fmt.Sprintf("priority = %s", arg(normalizePriority(*input.Priority))))
		}

		position := input.Position
		if position == nil && input.StateID != nil && newState != currentState {
			nextPos, err := s.nextPositionTx(ctx, tx, newState)
			if err != nil {
				return struct{}{}, err
			}
			position = &nextPos
		}

		if position != nil {
			updates = append(updates, fmt.Sprintf("position = %s", arg(*position)))
		}

		if len(updates) == 1 {
			return struct{}{}, errors.New("no updates")
		}

		args = append(args, id)
		query := mustSQL("tickets_update", map[string]any{
			"Updates": strings.Join(updates, ", "),
			"IDArg":   len(args),
		})

		if _, err := tx.Exec(ctx, query, args...); err != nil {
			return struct{}{}, err
		}

		return struct{}{}, nil
	})
	if err != nil {
		return Ticket{}, err
	}

	return s.GetTicket(ctx, id)
}

func (s *Store) DeleteTicket(ctx context.Context, id uuid.UUID) error {
	query := mustSQL("tickets_delete", nil)
	return execOne(ctx, s.db, query, pgx.ErrNoRows, id)
}

func (s *Store) resolveState(ctx context.Context, projectID uuid.UUID, provided *uuid.UUID) (uuid.UUID, error) {
	if provided != nil {
		return *provided, nil
	}

	var id uuid.UUID
	query := mustSQL("tickets_state_default", nil)
	err := s.db.QueryRow(ctx, query, projectID).Scan(&id)
	if err == nil {
		return id, nil
	}

	fallbackQuery := mustSQL("tickets_state_any", nil)
	err = s.db.QueryRow(ctx, fallbackQuery, projectID).Scan(&id)
	if err != nil {
		return uuid.Nil, errors.New("no workflow state available")
	}

	return id, nil
}

func (s *Store) nextPosition(ctx context.Context, stateID uuid.UUID) (float64, error) {
	var position float64
	query := mustSQL("tickets_next_position", nil)
	err := s.db.QueryRow(ctx, query, stateID).Scan(&position)
	return position, err
}

func (s *Store) nextPositionTx(ctx context.Context, tx pgx.Tx, stateID uuid.UUID) (float64, error) {
	var position float64
	query := mustSQL("tickets_next_position", nil)
	err := tx.QueryRow(ctx, query, stateID).Scan(&position)
	return position, err
}

func scanTicket(row pgx.Row) (Ticket, error) {
	var ticket Ticket
	err := row.Scan(
		&ticket.ID,
		&ticket.ProjectID,
		&ticket.ProjectKey,
		&ticket.Key,
		&ticket.Number,
		&ticket.Type,
		&ticket.StoryID,
		&ticket.StoryTitle,
		&ticket.StorySummary,
		&ticket.StoryCreated,
		&ticket.StoryUpdated,
		&ticket.Title,
		&ticket.Description,
		&ticket.StateID,
		&ticket.AssigneeID,
		&ticket.Priority,
		&ticket.Position,
		&ticket.CreatedAt,
		&ticket.UpdatedAt,
		&ticket.StateName,
		&ticket.StateOrder,
		&ticket.StateDefault,
		&ticket.StateClosed,
		&ticket.AssigneeName,
	)
	return ticket, err
}

func normalizePriority(input string) string {
	value := strings.ToLower(strings.TrimSpace(input))
	switch value {
	case "low", "medium", "high", "urgent":
		return value
	case "":
		return "medium"
	default:
		return "medium"
	}
}

func normalizeTicketType(input string) (string, error) {
	value := strings.ToLower(strings.TrimSpace(input))
	switch value {
	case "feature", "bug":
		return value, nil
	case "":
		return "feature", nil
	default:
		return "", errors.New("invalid ticket type")
	}
}
