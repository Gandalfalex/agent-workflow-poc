package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Story struct {
	ID          uuid.UUID
	ProjectID   uuid.UUID
	Title       string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type StoryCreateInput struct {
	Title       string
	Description *string
}

type StoryUpdateInput struct {
	Title       *string
	Description *string
}

func (s *Store) ListStories(ctx context.Context, projectID uuid.UUID) ([]Story, error) {
	query := mustSQL("stories_list", nil)
	return queryMany(ctx, s.db, query, scanStory, projectID)
}

func (s *Store) GetStory(ctx context.Context, id uuid.UUID) (Story, error) {
	query := mustSQL("stories_get", nil)
	return queryOne(ctx, s.db, query, scanStory, id)
}

func (s *Store) CreateStory(ctx context.Context, projectID uuid.UUID, input StoryCreateInput) (Story, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return Story{}, errors.New("title required")
	}

	query := mustSQL("stories_insert", nil)
	var id uuid.UUID
	if err := s.db.QueryRow(ctx, query, projectID, title, input.Description).Scan(&id); err != nil {
		return Story{}, err
	}
	return s.GetStory(ctx, id)
}

func (s *Store) UpdateStory(ctx context.Context, id uuid.UUID, input StoryUpdateInput) (Story, error) {
	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			return Story{}, errors.New("title required")
		}
		input.Title = &title
	}

	query := mustSQL("stories_update", nil)
	var updatedID uuid.UUID
	if err := s.db.QueryRow(ctx, query, id, input.Title, input.Description).Scan(&updatedID); err != nil {
		return Story{}, err
	}
	return s.GetStory(ctx, updatedID)
}

func (s *Store) DeleteStory(ctx context.Context, id uuid.UUID) error {
	query := mustSQL("stories_delete", nil)
	return execOne(ctx, s.db, query, pgx.ErrNoRows, id)
}

func scanStory(row pgx.Row) (Story, error) {
	var story Story
	err := row.Scan(
		&story.ID,
		&story.ProjectID,
		&story.Title,
		&story.Description,
		&story.CreatedAt,
		&story.UpdatedAt,
	)
	return story, err
}
