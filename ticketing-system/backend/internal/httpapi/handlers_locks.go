package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

const lockTTL = 2 * time.Minute

type ticketLockResponse struct {
	TicketID     string `json:"ticketId"`
	LockedBy     string `json:"lockedBy"`
	LockedByName string `json:"lockedByName"`
	AcquiredAt   string `json:"acquiredAt"`
	ExpiresAt    string `json:"expiresAt"`
}

// AcquireTicketLock — POST /rest/v1/tickets/{id}/lock
func (h *API) AcquireTicketLock(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ticketID := uuid.UUID(id)

	ticket, err := h.store.GetTicket(r.Context(), ticketID)
	if handleDBError(w, r, err, "ticket", "ticket_load") {
		return
	}
	if !h.requireProjectRole(w, r, ticket.ProjectID, roleContributor) {
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

	lock, acquired, err := h.store.AcquireTicketLock(r.Context(), ticketID, userID, user.Name, lockTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "lock_error", "failed to acquire lock")
		return
	}

	if acquired {
		h.publishProjectLiveEvent(ticket.ProjectID, projectEventLockAcquired, map[string]any{
			"ticketId":     ticketID.String(),
			"lockedBy":     lock.LockedBy.String(),
			"lockedByName": lock.LockedByName,
			"expiresAt":    lock.ExpiresAt.Format(time.RFC3339),
		})
		writeJSON(w, http.StatusOK, ticketLockResponse{
			TicketID:     lock.TicketID.String(),
			LockedBy:     lock.LockedBy.String(),
			LockedByName: lock.LockedByName,
			AcquiredAt:   lock.AcquiredAt.Format(time.RFC3339),
			ExpiresAt:    lock.ExpiresAt.Format(time.RFC3339),
		})
		return
	}

	// Lock held by someone else — return 409
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusConflict)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error":        "ticket_locked",
		"lockedBy":     lock.LockedBy.String(),
		"lockedByName": lock.LockedByName,
		"expiresAt":    lock.ExpiresAt.Format(time.RFC3339),
	})
}

// RenewTicketLock — PUT /rest/v1/tickets/{id}/lock
func (h *API) RenewTicketLock(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ticketID := uuid.UUID(id)

	ticket, err := h.store.GetTicket(r.Context(), ticketID)
	if handleDBError(w, r, err, "ticket", "ticket_load") {
		return
	}
	if !h.requireProjectRole(w, r, ticket.ProjectID, roleContributor) {
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

	lock, err := h.store.RenewTicketLock(r.Context(), ticketID, userID, lockTTL)
	if err != nil {
		writeError(w, http.StatusForbidden, "lock_not_held", "no active lock held by this user")
		return
	}

	writeJSON(w, http.StatusOK, ticketLockResponse{
		TicketID:     lock.TicketID.String(),
		LockedBy:     lock.LockedBy.String(),
		LockedByName: lock.LockedByName,
		AcquiredAt:   lock.AcquiredAt.Format(time.RFC3339),
		ExpiresAt:    lock.ExpiresAt.Format(time.RFC3339),
	})
}

// ReleaseTicketLock — DELETE /rest/v1/tickets/{id}/lock
func (h *API) ReleaseTicketLock(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ticketID := uuid.UUID(id)

	ticket, err := h.store.GetTicket(r.Context(), ticketID)
	if handleDBError(w, r, err, "ticket", "ticket_load") {
		return
	}
	if !h.requireProjectRole(w, r, ticket.ProjectID, roleContributor) {
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

	if err := h.store.ReleaseTicketLock(r.Context(), ticketID, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "lock_release_error", "failed to release lock")
		return
	}

	h.publishProjectLiveEvent(ticket.ProjectID, projectEventLockReleased, map[string]any{
		"ticketId": ticketID.String(),
	})

	w.WriteHeader(http.StatusNoContent)
}
