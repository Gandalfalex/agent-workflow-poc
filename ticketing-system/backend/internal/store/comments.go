package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Comment struct {
	ID         uuid.UUID
	TicketID   uuid.UUID
	AuthorID   uuid.UUID
	AuthorName string
	Message    string
	CreatedAt  time.Time
}

type CommentCreateInput struct {
	AuthorID   uuid.UUID
	AuthorName string
	Message    string
}

func (s *Store) ListComments(ctx context.Context, ticketID uuid.UUID) ([]Comment, error) {
	query := mustSQL("comments_list", nil)
	return queryMany(ctx, s.db, query, scanComment, ticketID)
}

func (s *Store) CreateComment(ctx context.Context, ticketID uuid.UUID, input CommentCreateInput) (Comment, error) {
	message := strings.TrimSpace(input.Message)
	if message == "" {
		return Comment{}, errors.New("message required")
	}
	if input.AuthorName == "" {
		return Comment{}, errors.New("author name required")
	}

	query := mustSQL("comments_insert", nil)
	var id uuid.UUID
	if err := s.db.QueryRow(ctx, query, ticketID, input.AuthorID, input.AuthorName, message).Scan(&id); err != nil {
		return Comment{}, err
	}

	comments, err := s.ListComments(ctx, ticketID)
	if err != nil {
		return Comment{}, err
	}
	for _, comment := range comments {
		if comment.ID == id {
			return comment, nil
		}
	}
	return Comment{}, pgx.ErrNoRows
}

func (s *Store) DeleteComment(ctx context.Context, id uuid.UUID) error {
	query := mustSQL("comments_delete", nil)
	return execOne(ctx, s.db, query, pgx.ErrNoRows, id)
}

func scanComment(row pgx.Row) (Comment, error) {
	var comment Comment
	err := row.Scan(
		&comment.ID,
		&comment.TicketID,
		&comment.AuthorID,
		&comment.AuthorName,
		&comment.Message,
		&comment.CreatedAt,
	)
	return comment, err
}
