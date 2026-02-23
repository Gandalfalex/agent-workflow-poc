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
	ID                    uuid.UUID
	ProjectID             uuid.UUID
	ProjectKey            string
	Key                   string
	Number                int
	Type                  string
	StoryID               uuid.UUID
	StoryTitle            string
	StorySummary          *string
	StoryStoryPoints      *int
	StoryCreated          time.Time
	StoryUpdated          time.Time
	Title                 string
	Description           string
	StateID               uuid.UUID
	StateName             string
	StateOrder            int
	StateDefault          bool
	StateClosed           bool
	BlockedByCount        int
	IsBlocked             bool
	AssigneeID            *uuid.UUID
	AssigneeName          *string
	Priority              string
	IncidentEnabled       bool
	IncidentSeverity      *string
	IncidentImpact        *string
	IncidentCommanderID   *uuid.UUID
	IncidentCommanderName *string
	Position              float64
	StoryPoints           *int
	TimeEstimate          *int
	TimeLogged            int
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type TicketFilter struct {
	ProjectID  uuid.UUID
	StateID    *uuid.UUID
	AssigneeID *uuid.UUID
	Blocked    *bool
	Query      string
	Limit      int
	Offset     int
}

type TicketCreateInput struct {
	Title               string
	Description         string
	Type                string
	StoryID             uuid.UUID
	StateID             *uuid.UUID
	AssigneeID          *uuid.UUID
	Priority            string
	IncidentEnabled     bool
	IncidentSeverity    *string
	IncidentImpact      *string
	IncidentCommanderID *uuid.UUID
	StoryPoints         *int
	TimeEstimate        *int
}

type TicketUpdateInput struct {
	Title               *string
	Description         *string
	Type                *string
	StoryID             *uuid.UUID
	StateID             *uuid.UUID
	AssigneeID          *uuid.UUID
	Priority            *string
	IncidentEnabled     *bool
	IncidentSeverity    *string
	IncidentImpact      *string
	IncidentCommanderID *uuid.UUID
	Position            *float64
	StoryPoints         *int
	TimeEstimate        *int
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
	if filter.Blocked != nil {
		if *filter.Blocked {
			conditions = append(conditions, "EXISTS (SELECT 1 FROM ticket_dependencies td WHERE td.relation_type = 'blocks' AND td.to_ticket_id = t.id)")
		} else {
			conditions = append(conditions, "NOT EXISTS (SELECT 1 FROM ticket_dependencies td WHERE td.relation_type = 'blocks' AND td.to_ticket_id = t.id)")
		}
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
	incidentSeverity, err := normalizeIncidentSeverity(input.IncidentSeverity)
	if err != nil {
		return Ticket{}, err
	}
	incidentImpact := normalizeIncidentImpact(input.IncidentImpact)
	var incidentCommanderID *uuid.UUID
	if input.IncidentEnabled {
		incidentCommanderID = input.IncidentCommanderID
	}
	if !input.IncidentEnabled {
		incidentSeverity = nil
		incidentImpact = nil
	}

	position, err := s.nextPosition(ctx, stateID)
	if err != nil {
		return Ticket{}, err
	}

	var ticketID uuid.UUID
	query := mustSQL("tickets_insert", nil)
	row := s.db.QueryRow(
		ctx,
		query,
		projectID,
		title,
		input.Description,
		ticketType,
		input.StoryID,
		stateID,
		input.AssigneeID,
		priority,
		input.IncidentEnabled,
		incidentSeverity,
		incidentImpact,
		incidentCommanderID,
		position,
		input.StoryPoints,
		input.TimeEstimate,
	)

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
		if input.IncidentEnabled != nil {
			updates = append(updates, fmt.Sprintf("incident_enabled = %s", arg(*input.IncidentEnabled)))
			if !*input.IncidentEnabled {
				updates = append(updates, "incident_severity = NULL")
				updates = append(updates, "incident_impact = NULL")
				updates = append(updates, "incident_commander_id = NULL")
			}
		}
		if input.IncidentSeverity != nil {
			sev, err := normalizeIncidentSeverity(input.IncidentSeverity)
			if err != nil {
				return struct{}{}, err
			}
			if sev == nil {
				updates = append(updates, "incident_severity = NULL")
			} else {
				updates = append(updates, fmt.Sprintf("incident_severity = %s", arg(*sev)))
			}
		}
		if input.IncidentImpact != nil {
			impact := normalizeIncidentImpact(input.IncidentImpact)
			if impact == nil {
				updates = append(updates, "incident_impact = NULL")
			} else {
				updates = append(updates, fmt.Sprintf("incident_impact = %s", arg(*impact)))
			}
		}
		if input.IncidentCommanderID != nil {
			updates = append(updates, fmt.Sprintf("incident_commander_id = %s", arg(*input.IncidentCommanderID)))
		}
		if input.StoryPoints != nil {
			updates = append(updates, fmt.Sprintf("story_points = %s", arg(*input.StoryPoints)))
		}
		if input.TimeEstimate != nil {
			updates = append(updates, fmt.Sprintf("time_estimate = %s", arg(*input.TimeEstimate)))
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
		&ticket.StoryStoryPoints,
		&ticket.StoryCreated,
		&ticket.StoryUpdated,
		&ticket.Title,
		&ticket.Description,
		&ticket.StateID,
		&ticket.AssigneeID,
		&ticket.Priority,
		&ticket.IncidentEnabled,
		&ticket.IncidentSeverity,
		&ticket.IncidentImpact,
		&ticket.IncidentCommanderID,
		&ticket.Position,
		&ticket.StoryPoints,
		&ticket.TimeEstimate,
		&ticket.TimeLogged,
		&ticket.CreatedAt,
		&ticket.UpdatedAt,
		&ticket.StateName,
		&ticket.StateOrder,
		&ticket.StateDefault,
		&ticket.StateClosed,
		&ticket.BlockedByCount,
		&ticket.IsBlocked,
		&ticket.AssigneeName,
		&ticket.IncidentCommanderName,
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

func normalizeIncidentSeverity(input *string) (*string, error) {
	if input == nil {
		return nil, nil
	}
	value := strings.ToLower(strings.TrimSpace(*input))
	switch value {
	case "":
		return nil, nil
	case "sev1", "sev2", "sev3", "sev4":
		return &value, nil
	default:
		return nil, errors.New("invalid incident severity")
	}
}

func normalizeIncidentImpact(input *string) *string {
	if input == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*input)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
