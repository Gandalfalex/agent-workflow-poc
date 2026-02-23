package httpapi

import (
	"net/http"

	"ticketing-system/backend/internal/store"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h *API) ListTicketTimeEntries(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, ticketId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}

	ticketUUID := uuid.UUID(ticketId)
	entries, totalMinutes, err := h.store.ListTimeEntries(r.Context(), ticketUUID)
	if handleListError(w, r, err, "time_entries", "time_entry_list") {
		return
	}

	writeJSON(w, http.StatusOK, timeEntryListResponse{
		Items:        mapSlice(entries, mapTimeEntry),
		TotalMinutes: totalMinutes,
	})
}

func (h *API) CreateTicketTimeEntry(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, ticketId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectRole(w, r, projectUUID, roleContributor) {
		return
	}

	ticketUUID := uuid.UUID(ticketId)
	req, ok := decodeJSON[timeEntryCreateRequest](w, r, "time_entry_create")
	if !ok {
		return
	}

	if req.Minutes < 1 {
		writeError(w, http.StatusBadRequest, "invalid_time_entry", "minutes must be at least 1")
		return
	}

	user, ok := authUser(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing session")
		return
	}
	userID, err := uuid.Parse(user.ID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_user_id", "user id must be a UUID")
		return
	}

	input := store.TimeEntryCreateInput{
		UserID:      userID,
		UserName:    user.Name,
		Minutes:     req.Minutes,
		Description: req.Description,
	}
	if req.LoggedAt != nil {
		t := req.LoggedAt.Time
		input.LoggedAt = &t
	}

	entry, err := h.store.CreateTimeEntry(r.Context(), ticketUUID, input)
	if handleDBErrorWithCode(w, r, err, "time_entry", "time_entry_create", "time_entry_create_failed") {
		return
	}

	writeJSON(w, http.StatusCreated, mapTimeEntry(entry))
}

func (h *API) DeleteTicketTimeEntry(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, ticketId openapi_types.UUID, timeEntryId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectRole(w, r, projectUUID, roleContributor) {
		return
	}

	entryUUID := uuid.UUID(timeEntryId)
	if err := h.store.DeleteTimeEntry(r.Context(), entryUUID); handleDeleteError(w, r, err, "time_entry", "time_entry_delete") {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
