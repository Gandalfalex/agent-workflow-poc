package store

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var projectKeyPattern = regexp.MustCompile(`^[A-Z0-9]{4}$`)

type Project struct {
	ID                       uuid.UUID
	Key                      string
	Name                     string
	Description              *string
	DefaultSprintDurationDays *int
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

type ProjectCreateInput struct {
	Key         string
	Name        string
	Description *string
}

type ProjectUpdateInput struct {
	Name                      *string
	Description               *string
	DefaultSprintDurationDays *int
}

func (s *Store) ListProjects(ctx context.Context) ([]Project, error) {
	query := mustSQL("projects_list", nil)
	return queryMany(ctx, s.db, query, scanProject)
}

func (s *Store) ListProjectsForUser(ctx context.Context, userID uuid.UUID) ([]Project, error) {
	query := mustSQL("projects_list_for_user", nil)
	return queryMany(ctx, s.db, query, scanProject, userID)
}

func (s *Store) ListProjectIDsForUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	query := mustSQL("project_ids_for_user", nil)
	return queryMany(ctx, s.db, query, scanProjectID, userID)
}

func (s *Store) GetProject(ctx context.Context, id uuid.UUID) (Project, error) {
	query := mustSQL("projects_get", nil)
	return queryOne(ctx, s.db, query, scanProject, id)
}

func (s *Store) CreateProject(ctx context.Context, input ProjectCreateInput) (Project, error) {
	key, err := normalizeProjectKey(input.Key)
	if err != nil {
		return Project{}, err
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Project{}, errors.New("name required")
	}

	query := mustSQL("projects_insert", nil)
	var id uuid.UUID
	if err := s.db.QueryRow(ctx, query, key, name, input.Description).Scan(&id); err != nil {
		return Project{}, err
	}
	return s.GetProject(ctx, id)
}

func (s *Store) UpdateProject(ctx context.Context, id uuid.UUID, input ProjectUpdateInput) (Project, error) {
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return Project{}, errors.New("name required")
		}
		input.Name = &name
	}

	updates := []string{"updated_at = now()"}
	args := []any{}
	arg := func(value any) string {
		args = append(args, value)
		return fmt.Sprintf("$%d", len(args))
	}

	if input.Name != nil {
		updates = append(updates, fmt.Sprintf("name = %s", arg(*input.Name)))
	}
	if input.Description != nil {
		updates = append(updates, fmt.Sprintf("description = %s", arg(*input.Description)))
	}
	if input.DefaultSprintDurationDays != nil {
		updates = append(updates, fmt.Sprintf("default_sprint_duration_days = %s", arg(*input.DefaultSprintDurationDays)))
	}

	if len(updates) == 1 {
		return s.GetProject(ctx, id)
	}

	args = append(args, id)
	query := mustSQL("projects_update", map[string]any{
		"Updates": strings.Join(updates, ", "),
		"IDArg":   len(args),
	})

	if _, err := s.db.Exec(ctx, query, args...); err != nil {
		return Project{}, err
	}
	return s.GetProject(ctx, id)
}

func (s *Store) DeleteProject(ctx context.Context, id uuid.UUID) error {
	query := mustSQL("projects_delete", nil)
	return execOne(ctx, s.db, query, pgx.ErrNoRows, id)
}

func (s *Store) GetProjectRoleForUser(ctx context.Context, projectID, userID uuid.UUID) (string, error) {
	query := mustSQL("project_role_for_user", nil)
	var role string
	err := s.db.QueryRow(ctx, query, projectID, userID).Scan(&role)
	if err == pgx.ErrNoRows {
		return "", nil
	}
	return role, err
}

func normalizeProjectKey(value string) (string, error) {
	key := strings.ToUpper(strings.TrimSpace(value))
	if !projectKeyPattern.MatchString(key) {
		return "", errors.New("invalid project key")
	}
	return key, nil
}

func scanProject(row pgx.Row) (Project, error) {
	var project Project
	err := row.Scan(
		&project.ID,
		&project.Key,
		&project.Name,
		&project.Description,
		&project.DefaultSprintDurationDays,
		&project.CreatedAt,
		&project.UpdatedAt,
	)
	return project, err
}

func scanProjectID(row pgx.Row) (uuid.UUID, error) {
	var id uuid.UUID
	err := row.Scan(&id)
	return id, err
}
