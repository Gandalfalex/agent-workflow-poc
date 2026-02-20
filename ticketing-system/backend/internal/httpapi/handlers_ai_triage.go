package httpapi

import (
	"errors"
	"net/http"
	"sort"
	"strings"

	"ticketing-system/backend/internal/store"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

const (
	aiTriagePromptVersion = "triage-v1"
	aiTriageModel         = "heuristic-local-v1"
)

func (h *API) GetProjectAiTriageSettings(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}
	settings, err := h.store.GetAiTriageSettings(r.Context(), projectUUID)
	if handleListError(w, r, err, "ai triage settings", "ai_triage_settings_get") {
		return
	}
	writeJSON(w, http.StatusOK, mapAiTriageSettings(settings))
}

func (h *API) UpdateProjectAiTriageSettings(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectRole(w, r, projectUUID, roleAdmin) {
		return
	}
	req, ok := decodeJSON[aiTriageSettingsUpdateRequest](w, r, "ai_triage_settings_update")
	if !ok {
		return
	}
	settings, err := h.store.UpdateAiTriageSettings(r.Context(), projectUUID, req.Enabled)
	if handleDBErrorWithCode(w, r, err, "ai triage settings", "ai_triage_settings_update", "ai_triage_settings_update_failed") {
		return
	}
	writeJSON(w, http.StatusOK, mapAiTriageSettings(settings))
}

func (h *API) CreateAiTriageSuggestion(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectRole(w, r, projectUUID, roleContributor) {
		return
	}
	settings, err := h.store.GetAiTriageSettings(r.Context(), projectUUID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ai_triage_settings_error", "unable to load ai triage settings")
		return
	}
	if !settings.Enabled {
		writeError(w, http.StatusForbidden, "ai_triage_disabled", "ai triage is disabled for this project")
		return
	}

	req, ok := decodeJSON[aiTriageSuggestionCreateRequest](w, r, "ai_triage_suggestion_create")
	if !ok {
		return
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		writeError(w, http.StatusBadRequest, "invalid_ai_triage_request", "title is required")
		return
	}

	actorID, _, hasActor := currentActor(r)
	if !hasActor {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing session")
		return
	}

	stateID, stateConfidence, err := h.pickSuggestedState(r, projectUUID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ai_triage_state_error", "unable to resolve workflow state")
		return
	}
	priority, priorityConfidence := suggestPriority(title, req.Description, req.Type)
	summary, summaryConfidence := suggestSummary(title, req.Description)
	assigneeID, assigneeConfidence := h.suggestAssignee(r, projectUUID, title, req.Description)

	var inputDescription *string
	if req.Description != nil {
		trimmed := strings.TrimSpace(*req.Description)
		if trimmed != "" {
			inputDescription = &trimmed
		}
	}
	var inputType *string
	if req.Type != nil {
		value := string(*req.Type)
		inputType = &value
	}

	suggestion, err := h.store.CreateAiTriageSuggestion(r.Context(), projectUUID, store.AiTriageSuggestionCreateInput{
		ActorID:            actorID,
		InputTitle:         title,
		InputDescription:   inputDescription,
		InputType:          inputType,
		Summary:            summary,
		Priority:           priority,
		StateID:            stateID,
		AssigneeID:         assigneeID,
		ConfidenceSummary:  summaryConfidence,
		ConfidencePriority: priorityConfidence,
		ConfidenceState:    stateConfidence,
		ConfidenceAssignee: assigneeConfidence,
		PromptVersion:      aiTriagePromptVersion,
		Model:              aiTriageModel,
	})
	if handleDBErrorWithCode(w, r, err, "ai triage suggestion", "ai_triage_suggestion_create", "ai_triage_suggestion_create_failed") {
		return
	}

	writeJSON(w, http.StatusCreated, mapAiTriageSuggestion(suggestion))
}

func (h *API) RecordAiTriageSuggestionDecision(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, suggestionId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectRole(w, r, projectUUID, roleContributor) {
		return
	}
	settings, err := h.store.GetAiTriageSettings(r.Context(), projectUUID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ai_triage_settings_error", "unable to load ai triage settings")
		return
	}
	if !settings.Enabled {
		writeError(w, http.StatusForbidden, "ai_triage_disabled", "ai triage is disabled for this project")
		return
	}

	req, ok := decodeJSON[aiTriageSuggestionDecisionRequest](w, r, "ai_triage_suggestion_decision")
	if !ok {
		return
	}
	accepted, err := normalizeAiFields(req.AcceptedFields)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_ai_triage_fields", err.Error())
		return
	}
	rejected, err := normalizeAiFields(req.RejectedFields)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_ai_triage_fields", err.Error())
		return
	}

	if _, err := h.store.GetAiTriageSuggestion(r.Context(), projectUUID, uuid.UUID(suggestionId)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "ai_triage_suggestion_not_found", "suggestion not found")
			return
		}
		if handleDBError(w, r, err, "ai triage suggestion", "ai_triage_suggestion_load") {
			return
		}
	}

	actorID, _, hasActor := currentActor(r)
	if !hasActor {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing session")
		return
	}

	decision, err := h.store.CreateAiTriageSuggestionDecision(r.Context(), projectUUID, uuid.UUID(suggestionId), store.AiTriageSuggestionDecisionCreateInput{
		ActorID:        actorID,
		AcceptedFields: accepted,
		RejectedFields: rejected,
	})
	if handleDBErrorWithCode(w, r, err, "ai triage decision", "ai_triage_decision_create", "ai_triage_decision_create_failed") {
		return
	}

	writeJSON(w, http.StatusCreated, mapAiTriageSuggestionDecision(decision))
}

func (h *API) pickSuggestedState(r *http.Request, projectID uuid.UUID) (uuid.UUID, float32, error) {
	states, err := h.store.ListWorkflowStates(r.Context(), projectID)
	if err != nil {
		return uuid.Nil, 0, err
	}
	if len(states) == 0 {
		return uuid.Nil, 0, errors.New("no workflow states")
	}
	for _, state := range states {
		if state.IsDefault {
			return state.ID, 0.85, nil
		}
	}
	return states[0].ID, 0.55, nil
}

func suggestPriority(title string, description *string, ticketType *TicketType) (string, float32) {
	text := strings.ToLower(strings.TrimSpace(title))
	if description != nil {
		text += " " + strings.ToLower(strings.TrimSpace(*description))
	}
	if strings.Contains(text, "sev1") || strings.Contains(text, "critical") || strings.Contains(text, "blocker") || strings.Contains(text, "outage") {
		return "urgent", 0.92
	}
	if strings.Contains(text, "fail") || strings.Contains(text, "bug") || strings.Contains(text, "error") || strings.Contains(text, "incident") {
		return "high", 0.78
	}
	if ticketType != nil && *ticketType == "bug" {
		return "high", 0.65
	}
	return "medium", 0.58
}

func suggestSummary(title string, description *string) (string, float32) {
	title = strings.TrimSpace(title)
	if description == nil {
		return title, 0.62
	}
	text := strings.TrimSpace(*description)
	if text == "" {
		return title, 0.62
	}
	for _, sep := range []string{"\n", ".", "!", "?"} {
		if idx := strings.Index(text, sep); idx > 0 {
			text = text[:idx]
			break
		}
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return title, 0.62
	}
	if len(text) > 140 {
		text = strings.TrimSpace(text[:140])
	}
	return text, 0.74
}

func (h *API) suggestAssignee(r *http.Request, projectID uuid.UUID, title string, description *string) (*uuid.UUID, float32) {
	text := title
	if description != nil {
		text += "\n" + *description
	}
	mentions := extractMentions(text)
	if len(mentions) == 0 {
		return nil, 0.15
	}

	users, err := h.store.ListUsers(r.Context(), "")
	if err != nil {
		return nil, 0.15
	}
	byAlias := map[string]uuid.UUID{}
	for _, user := range users {
		byAlias[strings.ToLower(user.Name)] = user.ID
		if parts := strings.Split(strings.ToLower(user.Email), "@"); len(parts) > 0 && parts[0] != "" {
			byAlias[parts[0]] = user.ID
		}
	}
	for _, mention := range mentions {
		candidate, ok := byAlias[strings.ToLower(mention)]
		if !ok {
			continue
		}
		role, roleErr := h.store.GetProjectRoleForUser(r.Context(), projectID, candidate)
		if roleErr != nil || role == "" {
			continue
		}
		return &candidate, 0.88
	}
	return nil, 0.25
}

func normalizeAiFields(fields []AiTriageField) ([]string, error) {
	if len(fields) == 0 {
		return []string{}, nil
	}
	out := make([]string, 0, len(fields))
	seen := map[string]struct{}{}
	for _, field := range fields {
		value := string(field)
		switch value {
		case "summary", "priority", "state", "assignee":
		default:
			return nil, errors.New("invalid ai triage field")
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out, nil
}
