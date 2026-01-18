package store

import (
	"context"
	"errors"
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
	querySQL := mustSQL("users_list.sql", map[string]any{
		"Where": "",
	})

	rows, err := s.db.Query(ctx, querySQL)
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

	// If no query, return all users sorted by name
	if strings.TrimSpace(query) == "" {
		return users, nil
	}

	// Client-side fuzzy filtering for better matching
	q := strings.ToLower(strings.TrimSpace(query))
	filtered := []UserSummary{}
	for _, user := range users {
		if fuzzyMatch(q, user.Name) || fuzzyMatch(q, user.Email) {
			filtered = append(filtered, user)
		}
	}

	return filtered, nil
}

// fuzzyMatch performs a simple fuzzy match where all characters in query
// must appear in text in order (but not necessarily consecutive)
func fuzzyMatch(query, text string) bool {
	text = strings.ToLower(text)
	queryIdx := 0
	for i := 0; i < len(text) && queryIdx < len(query); i++ {
		if text[i] == query[queryIdx] {
			queryIdx++
		}
	}
	return queryIdx == len(query)
}
