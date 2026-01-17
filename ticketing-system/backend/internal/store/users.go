package store

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type UserSummary struct {
	ID    uuid.UUID
	Name  string
	Email string
}

type UserUpsertInput struct {
	ID    uuid.UUID
	Name  string
	Email string
}

func (s *Store) UpsertUser(ctx context.Context, input UserUpsertInput) error {
	name := strings.TrimSpace(input.Name)
	email := strings.TrimSpace(input.Email)
	if name == "" {
		return errors.New("name required")
	}
	if email == "" {
		return errors.New("email required")
	}
	querySQL := mustSQL("users_upsert.sql", nil)
	_, err := s.db.Exec(ctx, querySQL, input.ID, name, email)
	return err
}

func (s *Store) ListUsers(ctx context.Context, query string) ([]UserSummary, error) {
	conditions := []string{}
	args := []any{}
	arg := func(value any) string {
		args = append(args, value)
		return fmt.Sprintf("$%d", len(args))
	}

	if strings.TrimSpace(query) != "" {
		q := "%%" + strings.TrimSpace(query) + "%%"
		conditions = append(conditions, fmt.Sprintf("(name ILIKE %s OR email ILIKE %s)", arg(q), arg(q)))
	}

	where := strings.Join(conditions, " AND ")
	querySQL := mustSQL("users_list.sql", map[string]any{
		"Where": where,
	})

	rows, err := s.db.Query(ctx, querySQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []UserSummary{}
	for rows.Next() {
		var user UserSummary
		if err := rows.Scan(&user.ID, &user.Name, &user.Email); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
