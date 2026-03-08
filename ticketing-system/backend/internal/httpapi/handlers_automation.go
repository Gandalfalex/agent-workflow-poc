package httpapi

import (
	"net/http"

	"ticketing-system/backend/internal/store"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h *API) ListAutomationRules(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}

	rules, err := h.store.ListAutomationRules(r.Context(), projectUUID)
	if handleListError(w, r, err, "automation rules", "automation_rules_list") {
		return
	}
	writeJSON(w, http.StatusOK, automationRuleListResponse{Items: mapSlice(rules, mapAutomationRule)})
}

func (h *API) CreateAutomationRule(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}
	if !h.requireProjectRole(w, r, projectUUID, roleContributor) {
		return
	}

	req, ok := decodeJSON[automationRuleCreateRequest](w, r, "automation_rule_create")
	if !ok {
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	conds := map[string]string{}
	if req.TriggerConditions != nil {
		conds = *req.TriggerConditions
	}

	rule, err := h.store.CreateAutomationRule(r.Context(), projectUUID, store.AutomationRuleCreateInput{
		Name:              req.Name,
		Enabled:           enabled,
		TriggerEvent:      req.TriggerEvent,
		TriggerConditions: conds,
		Actions:           mapSlice(req.Actions, mapActionFromAPI),
		Graph:             mapAutomationGraphToStore(req.Graph),
	})
	if handleDBErrorWithCode(w, r, err, "automation rule", "automation_rule_create", "automation_rule_create_failed") {
		return
	}
	writeJSON(w, http.StatusCreated, mapAutomationRule(rule))
}

func (h *API) GetAutomationRule(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, ruleId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	ruleUUID := uuid.UUID(ruleId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}

	rule, err := h.store.GetAutomationRule(r.Context(), ruleUUID, projectUUID)
	if handleDBError(w, r, err, "automation rule", "automation_rule_get") {
		return
	}
	writeJSON(w, http.StatusOK, mapAutomationRule(rule))
}

func (h *API) UpdateAutomationRule(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, ruleId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	ruleUUID := uuid.UUID(ruleId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}
	if !h.requireProjectRole(w, r, projectUUID, roleContributor) {
		return
	}

	req, ok := decodeJSON[automationRuleUpdateRequest](w, r, "automation_rule_update")
	if !ok {
		return
	}

	var conds map[string]string
	if req.TriggerConditions != nil {
		conds = *req.TriggerConditions
	}

	input := store.AutomationRuleUpdateInput{
		Name:              req.Name,
		Enabled:           req.Enabled,
		TriggerEvent:      req.TriggerEvent,
		TriggerConditions: conds,
	}
	if req.Actions != nil {
		input.Actions = mapSlice(*req.Actions, mapActionFromAPI)
		input.HasActions = true
	}
	if req.Graph != nil {
		input.Graph = mapAutomationGraphToStore(req.Graph)
		input.HasGraph = true
	}

	rule, err := h.store.UpdateAutomationRule(r.Context(), ruleUUID, projectUUID, input)
	if handleDBError(w, r, err, "automation rule", "automation_rule_update") {
		return
	}
	writeJSON(w, http.StatusOK, mapAutomationRule(rule))
}

func (h *API) DeleteAutomationRule(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, ruleId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	ruleUUID := uuid.UUID(ruleId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}
	if !h.requireProjectRole(w, r, projectUUID, roleContributor) {
		return
	}

	if err := h.store.DeleteAutomationRule(r.Context(), ruleUUID, projectUUID); handleDeleteError(w, r, err, "automation rule", "automation_rule_delete") {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *API) ListAutomationExecutions(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, ruleId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	ruleUUID := uuid.UUID(ruleId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}

	// Verify rule belongs to project
	if _, err := h.store.GetAutomationRule(r.Context(), ruleUUID, projectUUID); handleDBError(w, r, err, "automation rule", "automation_rule_get") {
		return
	}

	executions, err := h.store.ListAutomationExecutions(r.Context(), ruleUUID)
	if handleListError(w, r, err, "automation executions", "automation_executions_list") {
		return
	}
	writeJSON(w, http.StatusOK, automationExecutionListResponse{Items: mapSlice(executions, mapAutomationExecution)})
}

func mapActionFromAPI(a AutomationAction) store.AutomationAction {
	params := map[string]string{}
	if a.Params != nil {
		params = a.Params
	}
	return store.AutomationAction{
		Type:   string(a.Type),
		Params: params,
	}
}
