package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"ticketing-system/backend/internal/store"
)

func (h *API) GetApprovalPolicies(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectID := uuid.UUID(projectId)
	if !h.requireProjectRole(w, r, projectID, roleViewer) {
		return
	}
	policies, err := h.store.GetApprovalPolicies(r.Context(), projectID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "approval_policies_failed", "failed to load approval policies")
		return
	}
	writeJSON(w, http.StatusOK, approvalPolicyListResponse{Items: mapSlice(policies, mapApprovalPolicy)})
}

func (h *API) ReplaceApprovalPolicies(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectID := uuid.UUID(projectId)
	if !h.requireProjectRole(w, r, projectID, roleAdmin) {
		return
	}
	req, ok := decodeJSON[approvalPoliciesReplaceRequest](w, r, "approval_policies_replace")
	if !ok {
		return
	}
	inputs := make([]store.ApprovalPolicyInput, 0, len(req.Items))
	for _, item := range req.Items {
		fromID, err := parseOpenapiUUID(item.FromStateId)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_from_state_id", "fromStateId must be a UUID")
			return
		}
		toID, err := parseOpenapiUUID(item.ToStateId)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_to_state_id", "toStateId must be a UUID")
			return
		}
		rc := 1
		if item.RequiredCount != nil {
			rc = *item.RequiredCount
		}
		inp := store.ApprovalPolicyInput{
			FromStateID:   fromID,
			ToStateID:     toID,
			RequiredCount: rc,
			Scope:         string(item.Scope),
		}
		if item.GroupId != nil {
			gid := uuid.UUID(*item.GroupId)
			inp.GroupID = &gid
		}
		if item.Role != nil {
			inp.Role = item.Role
		}
		inputs = append(inputs, inp)
	}
	policies, err := h.store.ReplaceApprovalPolicies(r.Context(), projectID, inputs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "approval_policies_replace_failed", "failed to replace approval policies")
		return
	}
	writeJSON(w, http.StatusOK, approvalPolicyListResponse{Items: mapSlice(policies, mapApprovalPolicy)})
}

func (h *API) ListApprovalRequests(w http.ResponseWriter, r *http.Request, ticketId openapi_types.UUID) {
	ticketID := uuid.UUID(ticketId)
	ticket, err := h.store.GetTicket(r.Context(), ticketID)
	if handleDBError(w, r, err, "ticket", "ticket_get") {
		return
	}
	if !h.requireProjectRole(w, r, ticket.ProjectID, roleViewer) {
		return
	}
	requests, err := h.store.ListApprovalRequests(r.Context(), ticketID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "approval_requests_failed", "failed to load approval requests")
		return
	}
	writeJSON(w, http.StatusOK, approvalRequestListResponse{Items: mapSlice(requests, mapApprovalRequest)})
}

func (h *API) ApproveRequest(w http.ResponseWriter, r *http.Request, ticketId, requestId openapi_types.UUID) {
	h.handleApprovalDecision(w, r, ticketId, requestId, "approved")
}

func (h *API) RejectRequest(w http.ResponseWriter, r *http.Request, ticketId, requestId openapi_types.UUID) {
	h.handleApprovalDecision(w, r, ticketId, requestId, "rejected")
}

func (h *API) handleApprovalDecision(w http.ResponseWriter, r *http.Request, ticketId, requestId openapi_types.UUID, decision string) {
	ticketID := uuid.UUID(ticketId)
	requestID := uuid.UUID(requestId)

	ticket, err := h.store.GetTicket(r.Context(), ticketID)
	if handleDBError(w, r, err, "ticket", "ticket_get") {
		return
	}
	if !h.requireProjectRole(w, r, ticket.ProjectID, roleContributor) {
		return
	}

	req, err := h.store.GetApprovalRequestByID(r.Context(), requestID, ticketID)
	if err != nil {
		writeError(w, http.StatusNotFound, "approval_request_not_found", "approval request not found")
		return
	}
	if req.Status != "pending" {
		writeError(w, http.StatusConflict, "approval_request_resolved", "this request is no longer pending")
		return
	}

	var body *approvalDecisionRequest
	var bodyVal approvalDecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&bodyVal); err == nil {
		body = &bodyVal
	}

	user, _ := authUser(r.Context())
	approverID, _ := uuid.Parse(user.ID)

	// Prevent duplicate decisions from the same user.
	isDuplicate, _ := h.store.HasDuplicateApprovalDecision(r.Context(), requestID, approverID)
	if isDuplicate {
		writeError(w, http.StatusConflict, "already_decided", "you have already submitted a decision on this request")
		return
	}

	var reason *string
	if body != nil && body.Reason != nil && *body.Reason != "" {
		reason = body.Reason
	}

	if _, err := h.store.AddApprovalDecision(r.Context(), requestID, approverID, user.Name, decision, reason); err != nil {
		writeError(w, http.StatusInternalServerError, "approval_decision_failed", "failed to record decision")
		return
	}

	go h.recordAudit(r.Context(), approverID, user.Name, "approval_decision."+decision, "ticket", ticketID.String(), ticket.Title, &ticket.ProjectID, map[string]any{
		"requestId": requestID.String(),
		"decision":  decision,
		"reason":    reason,
	})

	newApprovedCount := req.ApprovedCount
	newStatus := req.Status
	if decision == "approved" {
		newApprovedCount++
		if newApprovedCount >= req.RequiredCount {
			newStatus = "approved"
		}
	} else {
		newStatus = "rejected"
	}

	updated, err := h.store.UpdateApprovalRequestStatus(r.Context(), requestID, newStatus, newApprovedCount)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "approval_update_failed", "failed to update approval request")
		return
	}
	writeJSON(w, http.StatusOK, mapApprovalRequest(*updated))
}
