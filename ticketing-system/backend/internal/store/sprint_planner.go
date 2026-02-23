package store

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Sprint struct {
	ID               uuid.UUID
	ProjectID        uuid.UUID
	Name             string
	Goal             *string
	StartDate        time.Time
	EndDate          time.Time
	TicketIDs        []uuid.UUID
	CommittedTickets int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type SprintCreateInput struct {
	Name      string
	Goal      *string
	StartDate time.Time
	EndDate   time.Time
	TicketIDs []uuid.UUID
}

type CapacitySetting struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Scope     string
	UserID    *uuid.UUID
	Label     string
	Capacity  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CapacitySettingInput struct {
	Scope    string
	UserID   *uuid.UUID
	Label    string
	Capacity int
}

type SprintForecastSummary struct {
	Sprint              *Sprint
	CommittedTickets    int
	Capacity            int
	ProjectedCompletion int
	OverCapacityDelta   int
	Confidence          float32
	Iterations          int
}

func (s *Store) ListSprints(ctx context.Context, projectID uuid.UUID) ([]Sprint, error) {
	rows, err := queryMany(ctx, s.db, mustSQL("sprints_list", nil), scanSprintRow, projectID)
	if err != nil {
		return nil, err
	}
	out := make([]Sprint, 0, len(rows))
	for _, row := range rows {
		ticketIDs, err := queryMany(ctx, s.db, mustSQL("sprint_tickets_list", nil), scanSprintTicketID, row.ID)
		if err != nil {
			return nil, err
		}
		row.TicketIDs = ticketIDs
		out = append(out, row)
	}
	return out, nil
}

func (s *Store) CreateSprint(ctx context.Context, projectID uuid.UUID, input SprintCreateInput) (Sprint, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Sprint{}, errors.New("name required")
	}
	startDate := normalizeDateUTC(input.StartDate)
	endDate := normalizeDateUTC(input.EndDate)
	if endDate.Before(startDate) {
		return Sprint{}, errors.New("end_date must be on or after start_date")
	}

	sprintID, err := withTx(ctx, s.db, func(tx pgx.Tx) (uuid.UUID, error) {
		var createdID uuid.UUID
		insertSQL := mustSQL("sprints_insert", nil)
		if err := tx.QueryRow(ctx, insertSQL, projectID, name, input.Goal, startDate, endDate).Scan(&createdID); err != nil {
			return uuid.Nil, err
		}

		for _, ticketID := range input.TicketIDs {
			var exists bool
			if err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM tickets WHERE id = $1 AND project_id = $2)", ticketID, projectID).Scan(&exists); err != nil {
				return uuid.Nil, err
			}
			if !exists {
				return uuid.Nil, fmt.Errorf("ticket %s not in project", ticketID)
			}
			if _, err := tx.Exec(ctx, mustSQL("sprint_tickets_insert", nil), createdID, ticketID); err != nil {
				return uuid.Nil, err
			}
		}
		return createdID, nil
	})
	if err != nil {
		return Sprint{}, err
	}

	sprints, err := s.ListSprints(ctx, projectID)
	if err != nil {
		return Sprint{}, err
	}
	for _, sprint := range sprints {
		if sprint.ID == sprintID {
			return sprint, nil
		}
	}
	return Sprint{}, pgx.ErrNoRows
}

func (s *Store) GetSprint(ctx context.Context, projectID, sprintID uuid.UUID) (Sprint, error) {
	sprint, err := queryOne(ctx, s.db, mustSQL("sprints_get", nil), scanSprintRow, sprintID, projectID)
	if err != nil {
		return Sprint{}, err
	}
	ticketIDs, err := queryMany(ctx, s.db, mustSQL("sprint_tickets_list", nil), scanSprintTicketID, sprintID)
	if err != nil {
		return Sprint{}, err
	}
	sprint.TicketIDs = ticketIDs
	return sprint, nil
}

func (s *Store) AddSprintTickets(ctx context.Context, projectID, sprintID uuid.UUID, ticketIDs []uuid.UUID) (Sprint, error) {
	_, err := withTx(ctx, s.db, func(tx pgx.Tx) (struct{}, error) {
		// Verify sprint belongs to project
		var exists bool
		if err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM sprints WHERE id = $1 AND project_id = $2)", sprintID, projectID).Scan(&exists); err != nil {
			return struct{}{}, err
		}
		if !exists {
			return struct{}{}, fmt.Errorf("sprint %s not in project", sprintID)
		}
		for _, ticketID := range ticketIDs {
			var ticketExists bool
			if err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM tickets WHERE id = $1 AND project_id = $2)", ticketID, projectID).Scan(&ticketExists); err != nil {
				return struct{}{}, err
			}
			if !ticketExists {
				return struct{}{}, fmt.Errorf("ticket %s not in project", ticketID)
			}
			if _, err := tx.Exec(ctx, mustSQL("sprint_tickets_insert", nil), sprintID, ticketID); err != nil {
				return struct{}{}, err
			}
		}
		return struct{}{}, nil
	})
	if err != nil {
		return Sprint{}, err
	}
	return s.GetSprint(ctx, projectID, sprintID)
}

func (s *Store) RemoveSprintTickets(ctx context.Context, projectID, sprintID uuid.UUID, ticketIDs []uuid.UUID) (Sprint, error) {
	_, err := withTx(ctx, s.db, func(tx pgx.Tx) (struct{}, error) {
		// Verify sprint belongs to project
		var exists bool
		if err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM sprints WHERE id = $1 AND project_id = $2)", sprintID, projectID).Scan(&exists); err != nil {
			return struct{}{}, err
		}
		if !exists {
			return struct{}{}, fmt.Errorf("sprint %s not in project", sprintID)
		}
		for _, ticketID := range ticketIDs {
			if _, err := tx.Exec(ctx, mustSQL("sprint_tickets_delete", nil), sprintID, ticketID); err != nil {
				return struct{}{}, err
			}
		}
		return struct{}{}, nil
	})
	if err != nil {
		return Sprint{}, err
	}
	return s.GetSprint(ctx, projectID, sprintID)
}

func (s *Store) ListCapacitySettings(ctx context.Context, projectID uuid.UUID) ([]CapacitySetting, error) {
	return queryMany(ctx, s.db, mustSQL("capacity_settings_list", nil), scanCapacitySetting, projectID)
}

func (s *Store) ReplaceCapacitySettings(ctx context.Context, projectID uuid.UUID, inputs []CapacitySettingInput) ([]CapacitySetting, error) {
	_, err := withTx(ctx, s.db, func(tx pgx.Tx) (struct{}, error) {
		if _, err := tx.Exec(ctx, mustSQL("capacity_settings_delete_for_project", nil), projectID); err != nil {
			return struct{}{}, err
		}
		for _, input := range inputs {
			scope := strings.ToLower(strings.TrimSpace(input.Scope))
			if scope != "team" && scope != "user" {
				return struct{}{}, errors.New("invalid capacity scope")
			}
			label := strings.TrimSpace(input.Label)
			if label == "" {
				return struct{}{}, errors.New("capacity label required")
			}
			if input.Capacity < 0 {
				return struct{}{}, errors.New("capacity must be >= 0")
			}
			if scope == "user" && input.UserID == nil {
				return struct{}{}, errors.New("user scope requires user id")
			}
			if scope == "team" {
				input.UserID = nil
			}
			if _, err := tx.Exec(ctx, mustSQL("capacity_settings_insert", nil), projectID, scope, input.UserID, label, input.Capacity); err != nil {
				return struct{}{}, err
			}
		}
		return struct{}{}, nil
	})
	if err != nil {
		return nil, err
	}
	return s.ListCapacitySettings(ctx, projectID)
}

func (s *Store) GetSprintForecastSummary(ctx context.Context, projectID uuid.UUID, sprintID *uuid.UUID, iterations int) (SprintForecastSummary, error) {
	if iterations <= 0 {
		iterations = 250
	}
	if iterations < 10 {
		iterations = 10
	}
	if iterations > 5000 {
		iterations = 5000
	}

	summary := SprintForecastSummary{Iterations: iterations}
	sprints, err := s.ListSprints(ctx, projectID)
	if err != nil {
		return summary, err
	}
	var selected *Sprint
	if sprintID != nil {
		for i := range sprints {
			if sprints[i].ID == *sprintID {
				selected = &sprints[i]
				break
			}
		}
	} else if len(sprints) > 0 {
		today := normalizeDateUTC(time.Now())
		for i := range sprints {
			if !today.Before(sprints[i].StartDate) && !today.After(sprints[i].EndDate) {
				selected = &sprints[i]
				break
			}
		}
		if selected == nil {
			selected = &sprints[0]
		}
	}

	settings, err := s.ListCapacitySettings(ctx, projectID)
	if err != nil {
		return summary, err
	}
	for _, item := range settings {
		summary.Capacity += item.Capacity
	}

	if selected == nil {
		return summary, nil
	}
	summary.Sprint = selected
	summary.CommittedTickets = selected.CommittedTickets
	if summary.CommittedTickets > summary.Capacity {
		summary.OverCapacityDelta = summary.CommittedTickets - summary.Capacity
	}

	historyRows, err := queryMany(ctx, s.db, mustSQL("forecast_daily_throughput_history", nil), scanDateValuePoint, projectID)
	if err != nil {
		return summary, err
	}
	history := make([]int, 0, len(historyRows))
	for _, point := range historyRows {
		if point.Value >= 0 {
			history = append(history, point.Value)
		}
	}
	if len(history) == 0 {
		history = []int{0, 1, 1, 2}
	}

	days := int(selected.EndDate.Sub(selected.StartDate).Hours()/24) + 1
	if days < 1 {
		days = 1
	}

	h := fnv.New64a()
	_, _ = h.Write([]byte(projectID.String()))
	_, _ = h.Write([]byte(selected.ID.String()))
	seed := int64(h.Sum64())
	rng := rand.New(rand.NewSource(seed))

	samples := make([]int, 0, iterations)
	hits := 0
	for i := 0; i < iterations; i++ {
		total := 0
		for day := 0; day < days; day++ {
			total += history[rng.Intn(len(history))]
		}
		samples = append(samples, total)
		if total >= summary.CommittedTickets {
			hits++
		}
	}

	sort.Ints(samples)
	summary.ProjectedCompletion = samples[len(samples)/2]
	summary.Confidence = float32(hits) / float32(iterations)
	return summary, nil
}

func scanSprintRow(row pgx.Row) (Sprint, error) {
	var out Sprint
	err := row.Scan(
		&out.ID,
		&out.ProjectID,
		&out.Name,
		&out.Goal,
		&out.StartDate,
		&out.EndDate,
		&out.CreatedAt,
		&out.UpdatedAt,
		&out.CommittedTickets,
	)
	return out, err
}

func scanSprintTicketID(row pgx.Row) (uuid.UUID, error) {
	var id uuid.UUID
	err := row.Scan(&id)
	return id, err
}

func scanCapacitySetting(row pgx.Row) (CapacitySetting, error) {
	var out CapacitySetting
	err := row.Scan(
		&out.ID,
		&out.ProjectID,
		&out.Scope,
		&out.UserID,
		&out.Label,
		&out.Capacity,
		&out.CreatedAt,
		&out.UpdatedAt,
	)
	return out, err
}
