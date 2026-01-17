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
	query := mustSQL("projects_list.sql", nil)
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := []Project{}
	for rows.Next() {
		project, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return projects, nil
}

func (s *Store) GetProject(ctx context.Context, id uuid.UUID) (Project, error) {
	query := mustSQL("projects_get.sql", nil)
	row := s.db.QueryRow(ctx, query, id)
	return scanProject(row)
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

	query := mustSQL("projects_insert.sql", nil)
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

	query := mustSQL("projects_update.sql", nil)
	var updatedID uuid.UUID
	if err := s.db.QueryRow(ctx, query, id, input.Name, input.Description).Scan(&updatedID); err != nil {
		return Project{}, err
	}
	return s.GetProject(ctx, updatedID)
}

func (s *Store) DeleteProject(ctx context.Context, id uuid.UUID) error {
	query := mustSQL("projects_delete.sql", nil)
	tag, err := s.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
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
