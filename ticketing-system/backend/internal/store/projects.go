package store

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var projectKeyPattern = regexp.MustCompile(`^[A-Z0-9]{4}$`)

type Project struct {
	ID          uuid.UUID
	Key         string
	Name        string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ProjectCreateInput struct {
	Key         string
	Name        string
	Description *string
}

type ProjectUpdateInput struct {
	Name        *string
	Description *string
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

	query := mustSQL("projects_update", nil)
	var updatedID uuid.UUID
	if err := s.db.QueryRow(ctx, query, id, input.Name, input.Description).Scan(&updatedID); err != nil {
		return Project{}, err
	}
	return s.GetProject(ctx, updatedID)
}

func (s *Store) DeleteProject(ctx context.Context, id uuid.UUID) error {
	query := mustSQL("projects_delete", nil)
	return execOne(ctx, s.db, query, pgx.ErrNoRows, id)
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
