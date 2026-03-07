package httpapi

import (
	"net/http"

	"ticketing-system/backend/internal/presence"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type presenceUpsertRequest struct {
	View     string  `json:"view"`
	TicketID *string `json:"ticketId"`
}

type presenceEntryResponse struct {
	UserID   string  `json:"userId"`
	UserName string  `json:"userName"`
	View     string  `json:"view"`
	TicketID *string `json:"ticketId,omitempty"`
}

type presenceResponse struct {
	Users []presenceEntryResponse `json:"users"`
}

// UpsertPresence — POST /rest/v1/projects/{projectId}/presence
func (h *API) UpsertPresence(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	if h.presenceStore == nil {
		writeError(w, http.StatusNotImplemented, "not_implemented", "live collaboration is not enabled")
		return
	}

	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
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

	req, ok := decodeJSON[presenceUpsertRequest](w, r, "presence_upsert")
	if !ok {
		return
	}

	entry := presence.Entry{
		UserID:    userID,
		UserName:  user.Name,
		ProjectID: projectUUID,
		View:      req.View,
	}
	if req.TicketID != nil && *req.TicketID != "" {
		tid, err := uuid.Parse(*req.TicketID)
		if err == nil {
			entry.TicketID = &tid
		}
	}

	h.presenceStore.Upsert(entry)

	users := h.presenceStore.ListByProject(projectUUID)
	resp := buildPresenceResponse(users)

	h.publishProjectLiveEvent(projectUUID, projectEventPresenceUpdate, map[string]any{
		"users": resp.Users,
	})

	writeJSON(w, http.StatusOK, resp)
}

func buildPresenceResponse(entries []presence.Entry) presenceResponse {
	users := make([]presenceEntryResponse, 0, len(entries))
	for _, e := range entries {
		pe := presenceEntryResponse{
			UserID:   e.UserID.String(),
			UserName: e.UserName,
			View:     e.View,
		}
		if e.TicketID != nil {
			s := e.TicketID.String()
			pe.TicketID = &s
		}
		users = append(users, pe)
	}
	return presenceResponse{Users: users}
}
