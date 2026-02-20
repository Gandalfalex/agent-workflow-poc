package httpapi

import (
	"net/http"
	"strings"

	"ticketing-system/backend/internal/store"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h *API) ListBoardFilterPresets(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}
	ownerID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	items, err := h.store.ListBoardFilterPresets(r.Context(), projectUUID, ownerID)
	if handleListError(w, r, err, "board filter presets", "board_filter_preset_list") {
		return
	}

	writeJSON(w, http.StatusOK, boardFilterPresetListResponse{Items: mapSlice(items, mapBoardFilterPreset)})
}

func (h *API) CreateBoardFilterPreset(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}
	ownerID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	req, ok := decodeJSON[boardFilterPresetCreateRequest](w, r, "board_filter_preset_create")
	if !ok {
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeError(w, http.StatusBadRequest, "invalid_board_filter_preset", "name is required")
		return
	}

	preset, err := h.store.CreateBoardFilterPreset(r.Context(), projectUUID, ownerID, store.BoardFilterPresetCreateInput{
		Name:               name,
		Filters:            mapStoreBoardFilter(req.Filters),
		GenerateShareToken: derefBool(req.GenerateShareToken, false),
	})
	if handleDBErrorWithCode(w, r, err, "board filter preset", "board_filter_preset_create", "board_filter_preset_create_failed") {
		return
	}

	writeJSON(w, http.StatusCreated, mapBoardFilterPreset(preset))
}

func (h *API) UpdateBoardFilterPreset(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, presetId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}
	ownerID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	req, ok := decodeJSON[boardFilterPresetUpdateRequest](w, r, "board_filter_preset_update")
	if !ok {
		return
	}

	input := store.BoardFilterPresetUpdateInput{
		GenerateShareToken: req.GenerateShareToken,
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		input.Name = &name
	}
	if req.Filters != nil {
		mapped := mapStoreBoardFilter(*req.Filters)
		input.Filters = &mapped
	}

	preset, err := h.store.UpdateBoardFilterPreset(r.Context(), projectUUID, ownerID, uuid.UUID(presetId), input)
	if handleDBErrorWithCode(w, r, err, "board filter preset", "board_filter_preset_update", "board_filter_preset_update_failed") {
		return
	}

	writeJSON(w, http.StatusOK, mapBoardFilterPreset(preset))
}

func (h *API) DeleteBoardFilterPreset(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, presetId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}
	ownerID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	if err := h.store.DeleteBoardFilterPreset(r.Context(), projectUUID, ownerID, uuid.UUID(presetId)); handleDeleteError(w, r, err, "board filter preset", "board_filter_preset_delete") {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *API) GetSharedBoardFilterPreset(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, token string) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}

	preset, err := h.store.GetSharedBoardFilterPreset(r.Context(), projectUUID, token)
	if handleDBError(w, r, err, "board filter preset", "board_filter_preset_shared") {
		return
	}

	writeJSON(w, http.StatusOK, mapBoardFilterPreset(preset))
}

func currentUserID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	user, ok := authUser(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing session")
		return uuid.Nil, false
	}
	id, err := uuid.Parse(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "invalid_user", "invalid user id")
		return uuid.Nil, false
	}
	return id, true
}

func derefBool(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}
