package httpapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"ticketing-system/backend/internal/store"
)

// recordAudit fires-and-forgets an audit log entry. Errors are silently dropped.
func (h *API) recordAudit(ctx context.Context, actorID uuid.UUID, actorName, action, resourceType, resourceID, resourceName string, projectID *uuid.UUID, details map[string]any) {
	if details == nil {
		details = map[string]any{}
	}
	_ = h.store.CreateAuditLog(ctx, store.AuditLogCreateInput{
		ActorID:      actorID,
		ActorName:    actorName,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		ResourceName: resourceName,
		ProjectID:    projectID,
		Details:      details,
	})
}

// ListAuditLog handles GET /admin/audit-log
func (h *API) ListAuditLog(w http.ResponseWriter, r *http.Request, params ListAuditLogParams) {
	if !requireAdmin(w, r) {
		return
	}

	filter := store.AuditLogFilter{Limit: 50}
	if params.Limit != nil {
		filter.Limit = *params.Limit
	}
	if params.Offset != nil {
		filter.Offset = *params.Offset
	}
	if params.ProjectId != nil {
		id := uuid.UUID(*params.ProjectId)
		filter.ProjectID = &id
	}
	if params.ResourceType != nil {
		filter.ResourceType = params.ResourceType
	}
	if params.ActorId != nil {
		id := uuid.UUID(*params.ActorId)
		filter.ActorID = &id
	}

	entries, total, err := h.store.ListAuditLog(r.Context(), filter)
	if handleListError(w, r, err, "audit log", "audit_log_list") {
		return
	}

	writeJSON(w, http.StatusOK, auditLogListResponse{
		Items: mapSlice(entries, mapAuditLogEntry),
		Total: total,
	})
}
