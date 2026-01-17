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
	query := mustSQL("comments_list.sql", nil)
	rows, err := s.db.Query(ctx, query, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := []Comment{}
	for rows.Next() {
		var comment Comment
		if err := rows.Scan(
			&comment.ID,
			&comment.TicketID,
			&comment.AuthorID,
			&comment.AuthorName,
			&comment.Message,
			&comment.CreatedAt,
		); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

func (s *Store) CreateComment(ctx context.Context, ticketID uuid.UUID, input CommentCreateInput) (Comment, error) {
	message := strings.TrimSpace(input.Message)
	if message == "" {
		return Comment{}, errors.New("message required")
	}
	if input.AuthorName == "" {
		return Comment{}, errors.New("author name required")
	}

	query := mustSQL("comments_insert.sql", nil)
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
	query := mustSQL("comments_delete.sql", nil)
	tag, err := s.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
