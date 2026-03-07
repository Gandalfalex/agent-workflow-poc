//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestWIPLimitSavedAndReturned(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	h := scenario.Harness()
	seed := scenario.SeedData()

	// Fetch current workflow states.
	wfResp, err := h.APIRequest("GET", fmt.Sprintf("/projects/%s/workflow", seed.ProjectID), nil)
	if err != nil {
		t.Fatalf("get workflow: %v", err)
	}
	defer wfResp.Body.Close()
	if wfResp.StatusCode != 200 {
		t.Fatalf("expected 200 getting workflow, got %d", wfResp.StatusCode)
	}

	var wf struct {
		States []struct {
			ID             string `json:"id"`
			Name           string `json:"name"`
			Order          int    `json:"order"`
			IsDefault      bool   `json:"isDefault"`
			IsClosed       bool   `json:"isClosed"`
			WipLimit       *int   `json:"wipLimit"`
			WipEnforcement bool   `json:"wipEnforcement"`
		} `json:"states"`
	}
	if err := json.NewDecoder(wfResp.Body).Decode(&wf); err != nil {
		t.Fatalf("decode workflow: %v", err)
	}
	if len(wf.States) == 0 {
		t.Skip("no workflow states; skipping")
	}

	// Update workflow with wipLimit=3, wipEnforcement=true on first non-closed state.
	var updatedStates []map[string]any
	for _, s := range wf.States {
		entry := map[string]any{
			"id":             s.ID,
			"name":           s.Name,
			"order":          s.Order,
			"isDefault":      s.IsDefault,
			"isClosed":       s.IsClosed,
			"wipEnforcement": false,
		}
		if !s.IsClosed {
			lim := 3
			entry["wipLimit"] = lim
			entry["wipEnforcement"] = true
		}
		updatedStates = append(updatedStates, entry)
	}
	body, _ := json.Marshal(map[string]any{"states": updatedStates})

	putResp, err := h.APIRequest("PUT", fmt.Sprintf("/projects/%s/workflow", seed.ProjectID), strings.NewReader(string(body)))
	if err != nil {
		t.Fatalf("update workflow: %v", err)
	}
	defer putResp.Body.Close()
	if putResp.StatusCode != 200 {
		t.Fatalf("expected 200 updating workflow, got %d", putResp.StatusCode)
	}

	var updated struct {
		States []struct {
			WipLimit       *int `json:"wipLimit"`
			WipEnforcement bool `json:"wipEnforcement"`
			IsClosed       bool `json:"isClosed"`
		} `json:"states"`
	}
	if err := json.NewDecoder(putResp.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated workflow: %v", err)
	}

	for _, s := range updated.States {
		if !s.IsClosed {
			if s.WipLimit == nil || *s.WipLimit != 3 {
				t.Fatalf("expected wipLimit=3 on non-closed state, got %v", s.WipLimit)
			}
			if !s.WipEnforcement {
				t.Fatal("expected wipEnforcement=true on non-closed state")
			}
		}
	}
}

func TestFlowHealthAPIReturnsStats(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	h := scenario.Harness()
	seed := scenario.SeedData()

	resp, err := h.APIRequest("GET", fmt.Sprintf("/projects/%s/flow-health", seed.ProjectID), nil)
	if err != nil {
		t.Fatalf("get flow health: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 getting flow health, got %d", resp.StatusCode)
	}

	var stats struct {
		WipStats   []any `json:"wipStats"`
		Throughput []any `json:"throughput"`
		CycleTime  struct {
			SampleCount int `json:"sampleCount"`
		} `json:"cycleTime"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		t.Fatalf("decode flow health: %v", err)
	}
	if stats.WipStats == nil {
		t.Fatal("expected wipStats array in response")
	}
	if stats.Throughput == nil {
		t.Fatal("expected throughput array in response")
	}
}

func TestWIPEnforcementBlocksMove(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	h := scenario.Harness()
	seed := scenario.SeedData()

	// Get workflow states.
	wfResp, err := h.APIRequest("GET", fmt.Sprintf("/projects/%s/workflow", seed.ProjectID), nil)
	if err != nil {
		t.Fatalf("get workflow: %v", err)
	}
	defer wfResp.Body.Close()
	var wf struct {
		States []struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			Order     int    `json:"order"`
			IsDefault bool   `json:"isDefault"`
			IsClosed  bool   `json:"isClosed"`
		} `json:"states"`
	}
	if err := json.NewDecoder(wfResp.Body).Decode(&wf); err != nil {
		t.Fatalf("decode workflow: %v", err)
	}
	if len(wf.States) < 2 {
		t.Skip("need at least 2 states for enforcement test")
	}

	// Find a non-closed, non-default state to enforce (or second state).
	var enforceStateID string
	for _, s := range wf.States {
		if !s.IsDefault && !s.IsClosed {
			enforceStateID = s.ID
			break
		}
	}
	if enforceStateID == "" {
		t.Skip("no suitable non-default non-closed state found")
	}

	// Set wipLimit=1 enforcement=true on that state.
	var updatedStates []map[string]any
	for _, s := range wf.States {
		entry := map[string]any{
			"id":             s.ID,
			"name":           s.Name,
			"order":          s.Order,
			"isDefault":      s.IsDefault,
			"isClosed":       s.IsClosed,
			"wipEnforcement": false,
		}
		if s.ID == enforceStateID {
			lim := 1
			entry["wipLimit"] = lim
			entry["wipEnforcement"] = true
		}
		updatedStates = append(updatedStates, entry)
	}
	body, _ := json.Marshal(map[string]any{"states": updatedStates})
	putResp, err := h.APIRequest("PUT", fmt.Sprintf("/projects/%s/workflow", seed.ProjectID), strings.NewReader(string(body)))
	if err != nil {
		t.Fatalf("update workflow: %v", err)
	}
	putResp.Body.Close()
	if putResp.StatusCode != 200 {
		t.Fatalf("expected 200 setting up WIP, got %d", putResp.StatusCode)
	}

	// Create two tickets in default state.
	var ticketIDs []string
	for i := 0; i < 2; i++ {
		ticketBody := fmt.Sprintf(`{"title":"WIP test ticket %d","type":"feature","priority":"medium","storyId":"%s"}`, i, seed.StoryID)
		cr, err := h.APIRequest("POST", fmt.Sprintf("/projects/%s/tickets", seed.ProjectID), strings.NewReader(ticketBody))
		if err != nil {
			t.Fatalf("create ticket %d: %v", i, err)
		}
		defer cr.Body.Close()
		if cr.StatusCode != 201 {
			t.Fatalf("expected 201 creating ticket, got %d", cr.StatusCode)
		}
		var ticket struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(cr.Body).Decode(&ticket); err != nil {
			t.Fatalf("decode ticket: %v", err)
		}
		ticketIDs = append(ticketIDs, ticket.ID)
	}

	// Move first ticket to enforced state — should succeed (0 → 1).
	mv1Body := fmt.Sprintf(`{"stateId":"%s"}`, enforceStateID)
	mv1, err := h.APIRequest("PATCH", fmt.Sprintf("/tickets/%s", ticketIDs[0]), strings.NewReader(mv1Body))
	if err != nil {
		t.Fatalf("move ticket 1: %v", err)
	}
	mv1.Body.Close()
	if mv1.StatusCode != 200 {
		t.Fatalf("expected 200 moving first ticket, got %d", mv1.StatusCode)
	}

	// Move second ticket to enforced state — should be blocked (1 ≥ 1).
	mv2Body := fmt.Sprintf(`{"stateId":"%s"}`, enforceStateID)
	mv2, err := h.APIRequest("PATCH", fmt.Sprintf("/tickets/%s", ticketIDs[1]), strings.NewReader(mv2Body))
	if err != nil {
		t.Fatalf("move ticket 2: %v", err)
	}
	mv2.Body.Close()
	if mv2.StatusCode != 409 {
		t.Fatalf("expected 409 (wip_limit_exceeded) when enforcement blocks move, got %d", mv2.StatusCode)
	}
}

func TestFlowHealthViewerCanAccess(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t, WithViewerUser())
	defer scenario.Close()

	h := scenario.Harness()
	seed := scenario.SeedData()

	resp, err := h.APIRequest("GET", fmt.Sprintf("/projects/%s/flow-health", seed.ProjectID), nil)
	if err != nil {
		t.Fatalf("viewer get flow health: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 for viewer, got %d", resp.StatusCode)
	}
}
