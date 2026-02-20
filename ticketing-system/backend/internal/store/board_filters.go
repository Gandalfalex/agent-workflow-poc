package store

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type BoardFilter struct {
	AssigneeID *uuid.UUID `json:"assigneeId,omitempty"`
	StateID    *uuid.UUID `json:"stateId,omitempty"`
	Priority   *string    `json:"priority,omitempty"`
	Type       *string    `json:"type,omitempty"`
	Query      *string    `json:"q,omitempty"`
	Blocked    *bool      `json:"blocked,omitempty"`
}

type BoardFilterPreset struct {
	ID         uuid.UUID
	ProjectID  uuid.UUID
	OwnerID    uuid.UUID
	Name       string
	Filters    BoardFilter
	ShareToken *string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type BoardFilterPresetCreateInput struct {
	Name               string
	Filters            BoardFilter
	GenerateShareToken bool
}

type BoardFilterPresetUpdateInput struct {
	Name               *string
	Filters            *BoardFilter
	GenerateShareToken *bool
}

func (s *Store) ListBoardFilterPresets(ctx context.Context, projectID, ownerID uuid.UUID) ([]BoardFilterPreset, error) {
	query := mustSQL("board_filter_presets_list", nil)
	return queryMany(ctx, s.db, query, scanBoardFilterPreset, projectID, ownerID)
}

func (s *Store) CreateBoardFilterPreset(ctx context.Context, projectID, ownerID uuid.UUID, input BoardFilterPresetCreateInput) (BoardFilterPreset, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return BoardFilterPreset{}, errors.New("name required")
	}
	if err := validateBoardFilter(input.Filters); err != nil {
		return BoardFilterPreset{}, err
	}

	payload, err := json.Marshal(input.Filters)
	if err != nil {
		return BoardFilterPreset{}, err
	}

	var shareToken *string
	if input.GenerateShareToken {
		token, err := generateShareToken()
		if err != nil {
			return BoardFilterPreset{}, err
		}
		shareToken = &token
	}

	var id uuid.UUID
	query := mustSQL("board_filter_presets_insert", nil)
	if err := s.db.QueryRow(ctx, query, projectID, ownerID, name, payload, shareToken).Scan(&id); err != nil {
		return BoardFilterPreset{}, err
	}
	return s.GetBoardFilterPreset(ctx, projectID, ownerID, id)
}

func (s *Store) UpdateBoardFilterPreset(ctx context.Context, projectID, ownerID, presetID uuid.UUID, input BoardFilterPresetUpdateInput) (BoardFilterPreset, error) {
	updates := []string{"updated_at = now()"}
	args := []any{}
	arg := func(value any) string {
		args = append(args, value)
		return "$" + strconv.Itoa(len(args))
	}

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return BoardFilterPreset{}, errors.New("name required")
		}
		updates = append(updates, "name = "+arg(name))
	}

	if input.Filters != nil {
		if err := validateBoardFilter(*input.Filters); err != nil {
			return BoardFilterPreset{}, err
		}
		payload, err := json.Marshal(*input.Filters)
		if err != nil {
			return BoardFilterPreset{}, err
		}
		updates = append(updates, "filters = "+arg(payload))
	}

	if input.GenerateShareToken != nil {
		if *input.GenerateShareToken {
			token, err := generateShareToken()
			if err != nil {
				return BoardFilterPreset{}, err
			}
			updates = append(updates, "share_token = "+arg(token))
		} else {
			updates = append(updates, "share_token = NULL")
		}
	}

	if len(updates) == 1 {
		return BoardFilterPreset{}, errors.New("no updates")
	}

	args = append(args, projectID, ownerID, presetID)
	query := mustSQL("board_filter_presets_update", map[string]any{
		"Updates":    strings.Join(updates, ", "),
		"ProjectArg": len(args) - 2,
		"OwnerArg":   len(args) - 1,
		"IDArg":      len(args),
	})
	if err := execOne(ctx, s.db, query, pgx.ErrNoRows, args...); err != nil {
		return BoardFilterPreset{}, err
	}
	return s.GetBoardFilterPreset(ctx, projectID, ownerID, presetID)
}

func (s *Store) DeleteBoardFilterPreset(ctx context.Context, projectID, ownerID, presetID uuid.UUID) error {
	query := mustSQL("board_filter_presets_delete", nil)
	return execOne(ctx, s.db, query, pgx.ErrNoRows, projectID, ownerID, presetID)
}

func (s *Store) GetBoardFilterPreset(ctx context.Context, projectID, ownerID, presetID uuid.UUID) (BoardFilterPreset, error) {
	query := mustSQL("board_filter_presets_get", nil)
	return queryOne(ctx, s.db, query, scanBoardFilterPreset, projectID, ownerID, presetID)
}

func (s *Store) GetSharedBoardFilterPreset(ctx context.Context, projectID uuid.UUID, token string) (BoardFilterPreset, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return BoardFilterPreset{}, pgx.ErrNoRows
	}
	query := mustSQL("board_filter_presets_get_shared", nil)
	return queryOne(ctx, s.db, query, scanBoardFilterPreset, projectID, token)
}

func scanBoardFilterPreset(row pgx.Row) (BoardFilterPreset, error) {
	var preset BoardFilterPreset
	var filtersRaw []byte
	if err := row.Scan(
		&preset.ID,
		&preset.ProjectID,
		&preset.OwnerID,
		&preset.Name,
		&filtersRaw,
		&preset.ShareToken,
		&preset.CreatedAt,
		&preset.UpdatedAt,
	); err != nil {
		return BoardFilterPreset{}, err
	}
	if len(filtersRaw) > 0 {
		if err := json.Unmarshal(filtersRaw, &preset.Filters); err != nil {
			return BoardFilterPreset{}, err
		}
	}
	return preset, nil
}

func validateBoardFilter(filter BoardFilter) error {
	if filter.Priority != nil {
		switch strings.ToLower(strings.TrimSpace(*filter.Priority)) {
		case "low", "medium", "high", "urgent":
		default:
			return errors.New("invalid priority")
		}
	}
	if filter.Type != nil {
		switch strings.ToLower(strings.TrimSpace(*filter.Type)) {
		case "feature", "bug":
		default:
			return errors.New("invalid ticket type")
		}
	}
	if filter.Query != nil {
		q := strings.TrimSpace(*filter.Query)
		filter.Query = &q
	}
	return nil
}

func generateShareToken() (string, error) {
	buf := make([]byte, 18)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
