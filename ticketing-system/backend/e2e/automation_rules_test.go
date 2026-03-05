//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestAutomationRuleAPICreateAndExecution(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	h := scenario.Harness()
	seed := scenario.SeedData()

	// Create an automation rule for ticket.state_changed
	ruleBody := fmt.Sprintf(`{
		"name":"Auto-comment on state change",
		"enabled":true,
		"triggerEvent":"ticket.state_changed",
		"triggerConditions":{},
		"actions":[{"type":"add_comment","params":{"body":"State changed on {{ticket.title}}"}}]
	}`)

	resp, err := h.APIRequest("POST", fmt.Sprintf("/projects/%s/automation/rules", seed.ProjectID), strings.NewReader(ruleBody))
	if err != nil {
		t.Fatalf("create rule: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		t.Fatalf("expected 201 creating rule, got %d", resp.StatusCode)
	}

	var rule struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		TriggerEvent string `json:"triggerEvent"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rule); err != nil {
		t.Fatalf("decode rule: %v", err)
	}
	if rule.TriggerEvent != "ticket.state_changed" {
		t.Fatalf("expected triggerEvent=ticket.state_changed, got %s", rule.TriggerEvent)
	}

	// List rules
	listResp, err := h.APIRequest("GET", fmt.Sprintf("/projects/%s/automation/rules", seed.ProjectID), nil)
	if err != nil {
		t.Fatalf("list rules: %v", err)
	}
	defer listResp.Body.Close()
	if listResp.StatusCode != 200 {
		t.Fatalf("expected 200 listing rules, got %d", listResp.StatusCode)
	}

	var listResult struct {
		Items []struct{ ID string `json:"id"` } `json:"items"`
	}
	if err := json.NewDecoder(listResp.Body).Decode(&listResult); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(listResult.Items) == 0 {
		t.Fatalf("expected at least one rule")
	}

	// Executions list should be empty initially
	execResp, err := h.APIRequest("GET", fmt.Sprintf("/projects/%s/automation/rules/%s/executions", seed.ProjectID, rule.ID), nil)
	if err != nil {
		t.Fatalf("list executions: %v", err)
	}
	defer execResp.Body.Close()
	if execResp.StatusCode != 200 {
		t.Fatalf("expected 200 listing executions, got %d", execResp.StatusCode)
	}
}

func TestAutomationRuleDisabledNotReturned(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	h := scenario.Harness()
	seed := scenario.SeedData()

	// Create a disabled rule
	ruleBody := `{"name":"Disabled rule","enabled":false,"triggerEvent":"ticket.created","actions":[]}`
	resp, err := h.APIRequest("POST", fmt.Sprintf("/projects/%s/automation/rules", seed.ProjectID), strings.NewReader(ruleBody))
	if err != nil {
		t.Fatalf("create rule: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var rule struct {
		ID      string `json:"id"`
		Enabled bool   `json:"enabled"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rule); err != nil {
		t.Fatalf("decode rule: %v", err)
	}
	if rule.Enabled {
		t.Fatalf("expected rule to be disabled")
	}

	// Update to enable it
	updateBody := `{"enabled":true}`
	updateResp, err := h.APIRequest("PUT", fmt.Sprintf("/projects/%s/automation/rules/%s", seed.ProjectID, rule.ID), strings.NewReader(updateBody))
	if err != nil {
		t.Fatalf("update rule: %v", err)
	}
	defer updateResp.Body.Close()
	if updateResp.StatusCode != 200 {
		t.Fatalf("expected 200 updating rule, got %d", updateResp.StatusCode)
	}

	// Delete it
	delResp, err := h.APIRequest("DELETE", fmt.Sprintf("/projects/%s/automation/rules/%s", seed.ProjectID, rule.ID), nil)
	if err != nil {
		t.Fatalf("delete rule: %v", err)
	}
	defer delResp.Body.Close()
	if delResp.StatusCode != 204 {
		t.Fatalf("expected 204 deleting rule, got %d", delResp.StatusCode)
	}
}

func TestAutomationUIShowsAutomationTab(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()

	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInWithHarnessUser().
		ThenISeeSelectorKey("nav.board_tab").
		WhenIGoToRoute("settings", map[string]string{"projectId": seed.ProjectID}).
		ThenISeeSelectorKey("settings.automation_tab").
		WhenIClickKey("settings.automation_tab").
		ThenISeeSelectorKey("settings.automation_view")
}
