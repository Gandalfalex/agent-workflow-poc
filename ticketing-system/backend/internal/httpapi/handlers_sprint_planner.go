package httpapi

import (
	"net/http"
	"strings"

	"ticketing-system/backend/internal/store"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h *API) ListProjectSprints(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}

	items, err := h.store.ListSprints(r.Context(), projectUUID)
	if handleListError(w, r, err, "sprints", "sprint_list") {
		return
	}
	writeJSON(w, http.StatusOK, sprintListResponse{Items: mapSlice(items, mapSprint)})
}

func (h *API) CreateProjectSprint(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}
	if !h.requireProjectRole(w, r, projectUUID, roleContributor) {
		return
	}

	req, ok := decodeJSON[sprintCreateRequest](w, r, "sprint_create")
	if !ok {
		return
	}

	sourceTicketIDs := []openapi_types.UUID{}
	if req.TicketIds != nil {
		sourceTicketIDs = *req.TicketIds
	}
	ticketIDs := make([]uuid.UUID, 0, len(sourceTicketIDs))
	for _, id := range sourceTicketIDs {
		ticketIDs = append(ticketIDs, uuid.UUID(id))
	}
	var goal *string
	if req.Goal != nil {
		value := strings.TrimSpace(*req.Goal)
		if value != "" {
			goal = &value
		}
	}

	sprint, err := h.store.CreateSprint(r.Context(), projectUUID, store.SprintCreateInput{
		Name:      strings.TrimSpace(req.Name),
		Goal:      goal,
		StartDate: req.StartDate.Time,
		EndDate:   req.EndDate.Time,
		TicketIDs: ticketIDs,
	})
	if handleDBErrorWithCode(w, r, err, "sprint", "sprint_create", "sprint_create_failed") {
		return
	}
	writeJSON(w, http.StatusCreated, mapSprint(sprint))
}

func (h *API) ListProjectCapacitySettings(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}

	items, err := h.store.ListCapacitySettings(r.Context(), projectUUID)
	if handleListError(w, r, err, "capacity settings", "capacity_settings_list") {
		return
	}
	writeJSON(w, http.StatusOK, capacitySettingsResponse{Items: mapSlice(items, mapCapacitySetting)})
}

func (h *API) ReplaceProjectCapacitySettings(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}
	if !h.requireProjectRole(w, r, projectUUID, roleContributor) {
		return
	}

	req, ok := decodeJSON[capacitySettingsReplaceRequest](w, r, "capacity_settings_replace")
	if !ok {
		return
	}
	inputs := make([]store.CapacitySettingInput, 0, len(req.Items))
	for _, item := range req.Items {
		input := store.CapacitySettingInput{
			Scope:    string(item.Scope),
			Label:    strings.TrimSpace(item.Label),
			Capacity: item.Capacity,
		}
		if item.UserId != nil {
			value := uuid.UUID(*item.UserId)
			input.UserID = &value
		}
		inputs = append(inputs, input)
	}
	items, err := h.store.ReplaceCapacitySettings(r.Context(), projectUUID, inputs)
	if handleDBErrorWithCode(w, r, err, "capacity settings", "capacity_settings_replace", "capacity_settings_replace_failed") {
		return
	}
	writeJSON(w, http.StatusOK, capacitySettingsResponse{Items: mapSlice(items, mapCapacitySetting)})
}

func (h *API) GetProjectSprintForecast(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, params GetProjectSprintForecastParams) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}

	var sprintID *uuid.UUID
	if params.SprintId != nil {
		value := uuid.UUID(*params.SprintId)
		sprintID = &value
	}
	iterations := 250
	if params.Iterations != nil {
		iterations = *params.Iterations
	}

	summary, err := h.store.GetSprintForecastSummary(r.Context(), projectUUID, sprintID, iterations)
	if handleListError(w, r, err, "sprint forecast", "sprint_forecast") {
		return
	}
	writeJSON(w, http.StatusOK, mapSprintForecastSummary(summary))
}
