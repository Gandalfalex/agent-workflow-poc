package httpapi

import (
	"net/http"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"ticketing-system/backend/internal/store"
)

func (h *API) GetSlaPolicies(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}
	policies, err := h.store.GetSlaPolicies(r.Context(), projectUUID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "sla_load_error", "Failed to load SLA policies")
		return
	}
	writeJSON(w, http.StatusOK, slaPolicyListResponse{Items: mapSlaEntries(policies)})
}

func (h *API) UpdateSlaPolicies(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectRole(w, r, projectUUID, roleAdmin) {
		return
	}

	req, ok := decodeJSON[slaPolicyUpdateRequest](w, r, "sla_update")
	if !ok {
		return
	}

	entries := make([]store.SlaPolicyEntry, 0, len(req.Items))
	for _, item := range req.Items {
		e := store.SlaPolicyEntry{
			Priority:           string(item.Priority),
			TicketType:         string(item.TicketType),
			FirstResponseHours: item.FirstResponseHours,
			CompletionHours:    item.CompletionHours,
			AtRiskThresholdPct: item.AtRiskThresholdPct,
		}
		entries = append(entries, e)
	}

	updated, err := h.store.UpdateSlaPolicies(r.Context(), projectUUID, entries)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "sla_update_error", "Failed to update SLA policies")
		return
	}
	writeJSON(w, http.StatusOK, slaPolicyListResponse{Items: mapSlaEntries(updated)})
}

func mapSlaEntries(entries []store.SlaPolicyEntry) []slaPolicyEntryResponse {
	out := make([]slaPolicyEntryResponse, 0, len(entries))
	for _, e := range entries {
		out = append(out, slaPolicyEntryResponse{
			Priority:           TicketPriority(e.Priority),
			TicketType:         TicketType(e.TicketType),
			FirstResponseHours: e.FirstResponseHours,
			CompletionHours:    e.CompletionHours,
			AtRiskThresholdPct: e.AtRiskThresholdPct,
		})
	}
	return out
}
