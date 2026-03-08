package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"ticketing-system/backend/internal/automation"
	"ticketing-system/backend/internal/store"
)

func (h *API) SimulateAutomationRule(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectID := uuid.UUID(projectId)
	if !h.requireProjectRole(w, r, projectID, roleContributor) {
		return
	}

	var req SimulationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}

	lookbackDays := int(req.LookbackDays)

	// Build a store.AutomationRule from the request (no ID — draft)
	rule := store.AutomationRule{
		TriggerEvent: req.Rule.TriggerEvent,
	}
	if req.Rule.TriggerConditions != nil {
		rule.TriggerConditions = *req.Rule.TriggerConditions
	}
	for _, a := range req.Rule.Actions {
		rule.Actions = append(rule.Actions, store.AutomationAction{
			Type:   string(a.Type),
			Params: a.Params,
		})
	}

	// Fetch historical events for this trigger type
	events, err := h.store.GetSimulationEvents(r.Context(), projectID, rule.TriggerEvent, lookbackDays)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "simulation_events_failed", "failed to load simulation events")
		return
	}

	// Evaluate each event
	matchedCount := 0
	failureCount := 0
	results := make([]store.SimulationResultCreateInput, 0, len(events))
	for _, ev := range events {
		evalResult := automation.EvaluateRule(rule, ev.EventType, ev.Extra)
		ri := store.SimulationResultCreateInput{
			TicketID:       ev.TicketID,
			TicketKey:      ev.TicketKey,
			EventType:      ev.EventType,
			EventTimestamp: ev.Timestamp,
			Matched:        evalResult.Matched,
			FailureReason:  evalResult.FailureReason,
		}
		if evalResult.Matched {
			matchedCount++
			ri.PredictedActions = evalResult.PredictedActions
			for _, a := range evalResult.PredictedActions {
				if !a.WouldSucceed {
					failureCount++
					break
				}
			}
		} else {
			ri.PredictedActions = []store.SimulatedActionOutcome{}
		}
		results = append(results, ri)
	}

	// Build rule snapshot for storage
	snapshot := map[string]any{
		"triggerEvent":      rule.TriggerEvent,
		"triggerConditions": rule.TriggerConditions,
		"actions":           rule.Actions,
	}

	// Store the run
	run, err := h.store.CreateSimulationRun(r.Context(), store.SimulationRunCreateInput{
		ProjectID:    projectID,
		RuleSnapshot: snapshot,
		LookbackDays: lookbackDays,
		TotalEvents:  len(events),
		MatchedCount: matchedCount,
		FailureCount: failureCount,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "simulation_run_failed", "failed to store simulation run")
		return
	}

	// Store results
	for _, ri := range results {
		ri.RunID = run.ID
		_ = h.store.CreateSimulationResult(r.Context(), ri)
	}

	// Return paginated results (first page)
	limit := 50
	resultItems, total, err := h.store.ListSimulationResults(r.Context(), run.ID, limit, 0)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "simulation_results_failed", "failed to load results")
		return
	}

	writeJSON(w, http.StatusOK, simulationRunResponse{
		Run:   mapSimulationRun(run),
		Items: mapSlice(resultItems, mapSimulationResult),
		Total: total,
	})
}

func (h *API) ListSimulationRuns(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectID := uuid.UUID(projectId)
	if !h.requireProjectRole(w, r, projectID, roleViewer) {
		return
	}
	runs, err := h.store.ListSimulationRuns(r.Context(), projectID)
	if handleListError(w, r, err, "simulation_runs", "simulation_runs_list") {
		return
	}
	writeJSON(w, http.StatusOK, simulationRunListResponse{Items: mapSlice(runs, mapSimulationRun)})
}

func (h *API) GetSimulationRun(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, runId openapi_types.UUID, params GetSimulationRunParams) {
	projectID := uuid.UUID(projectId)
	runID := uuid.UUID(runId)
	if !h.requireProjectRole(w, r, projectID, roleViewer) {
		return
	}

	run, err := h.store.GetSimulationRun(r.Context(), runID, projectID)
	if handleDBError(w, r, err, "simulation_run", "simulation_run_get") {
		return
	}

	limit := 50
	offset := 0
	if params.Limit != nil {
		limit = *params.Limit
	}
	if params.Offset != nil {
		offset = *params.Offset
	}

	results, total, err := h.store.ListSimulationResults(r.Context(), runID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "simulation_results_failed", "failed to load results")
		return
	}

	writeJSON(w, http.StatusOK, simulationRunResponse{
		Run:   mapSimulationRun(run),
		Items: mapSlice(results, mapSimulationResult),
		Total: total,
	})
}
