package store

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type AiTriageSettings struct {
	Enabled bool
}

type AiTriageSuggestion struct {
	ID                 uuid.UUID
	ProjectID          uuid.UUID
	Summary            string
	Priority           string
	StateID            uuid.UUID
	AssigneeID         *uuid.UUID
	ConfidenceSummary  float32
	ConfidencePriority float32
	ConfidenceState    float32
	ConfidenceAssignee float32
	PromptVersion      string
	Model              string
	CreatedAt          time.Time
}

type AiTriageSuggestionCreateInput struct {
	ActorID            uuid.UUID
	InputTitle         string
	InputDescription   *string
	InputType          *string
	Summary            string
	Priority           string
	StateID            uuid.UUID
	AssigneeID         *uuid.UUID
	ConfidenceSummary  float32
	ConfidencePriority float32
	ConfidenceState    float32
	ConfidenceAssignee float32
	PromptVersion      string
	Model              string
}

type AiTriageSuggestionDecision struct {
	ID             uuid.UUID
	SuggestionID   uuid.UUID
	ProjectID      uuid.UUID
	ActorID        uuid.UUID
	AcceptedFields []string
	RejectedFields []string
	CreatedAt      time.Time
}

type AiTriageSuggestionDecisionCreateInput struct {
	ActorID        uuid.UUID
	AcceptedFields []string
	RejectedFields []string
}

func (s *Store) GetAiTriageSettings(ctx context.Context, projectID uuid.UUID) (AiTriageSettings, error) {
	var enabled bool
	err := s.db.QueryRow(ctx, mustSQL("ai_triage_settings_get", nil), projectID).Scan(&enabled)
	if err == pgx.ErrNoRows {
		return AiTriageSettings{Enabled: false}, nil
	}
	if err != nil {
		return AiTriageSettings{}, err
	}
	return AiTriageSettings{Enabled: enabled}, nil
}

func (s *Store) UpdateAiTriageSettings(ctx context.Context, projectID uuid.UUID, enabled bool) (AiTriageSettings, error) {
	var updated bool
	if err := s.db.QueryRow(ctx, mustSQL("ai_triage_settings_upsert", nil), projectID, enabled).Scan(&updated); err != nil {
		return AiTriageSettings{}, err
	}
	return AiTriageSettings{Enabled: updated}, nil
}

func (s *Store) CreateAiTriageSuggestion(ctx context.Context, projectID uuid.UUID, input AiTriageSuggestionCreateInput) (AiTriageSuggestion, error) {
	input.InputTitle = strings.TrimSpace(input.InputTitle)
	input.Summary = strings.TrimSpace(input.Summary)
	var id uuid.UUID
	var createdAt time.Time
	err := s.db.QueryRow(
		ctx,
		mustSQL("ai_triage_suggestion_insert", nil),
		projectID,
		input.ActorID,
		input.InputTitle,
		input.InputDescription,
		input.InputType,
		input.Summary,
		input.Priority,
		input.StateID,
		input.AssigneeID,
		input.ConfidenceSummary,
		input.ConfidencePriority,
		input.ConfidenceState,
		input.ConfidenceAssignee,
		input.PromptVersion,
		input.Model,
	).Scan(&id, &createdAt)
	if err != nil {
		return AiTriageSuggestion{}, err
	}
	return s.GetAiTriageSuggestion(ctx, projectID, id)
}

func (s *Store) GetAiTriageSuggestion(ctx context.Context, projectID, suggestionID uuid.UUID) (AiTriageSuggestion, error) {
	var item AiTriageSuggestion
	err := s.db.QueryRow(ctx, mustSQL("ai_triage_suggestion_get", nil), suggestionID, projectID).Scan(
		&item.ID,
		&item.ProjectID,
		&item.Summary,
		&item.Priority,
		&item.StateID,
		&item.AssigneeID,
		&item.ConfidenceSummary,
		&item.ConfidencePriority,
		&item.ConfidenceState,
		&item.ConfidenceAssignee,
		&item.PromptVersion,
		&item.Model,
		&item.CreatedAt,
	)
	return item, err
}

func (s *Store) CreateAiTriageSuggestionDecision(ctx context.Context, projectID, suggestionID uuid.UUID, input AiTriageSuggestionDecisionCreateInput) (AiTriageSuggestionDecision, error) {
	accepted := dedupeAiFields(input.AcceptedFields)
	rejected := dedupeAiFields(input.RejectedFields)

	var decision AiTriageSuggestionDecision
	decision.SuggestionID = suggestionID
	decision.ProjectID = projectID
	decision.ActorID = input.ActorID
	decision.AcceptedFields = accepted
	decision.RejectedFields = rejected

	err := s.db.QueryRow(
		ctx,
		mustSQL("ai_triage_decision_insert", nil),
		suggestionID,
		projectID,
		input.ActorID,
		accepted,
		rejected,
	).Scan(&decision.ID, &decision.CreatedAt)
	if err != nil {
		return AiTriageSuggestionDecision{}, err
	}
	return decision, nil
}

func dedupeAiFields(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		norm := strings.TrimSpace(strings.ToLower(value))
		if norm == "" {
			continue
		}
		if _, ok := seen[norm]; ok {
			continue
		}
		seen[norm] = struct{}{}
		out = append(out, norm)
	}
	return out
}
