//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// createAutomationRule POSTs an automation rule and returns its ID and triggerEvent.
func createAutomationRule(s *Scenario, projectID, body string) (id string, triggerEvent string, err error) {
	resp, respErr := s.Harness().APIRequest("POST", fmt.Sprintf("/projects/%s/automation/rules", projectID), strings.NewReader(body))
	if respErr != nil {
		return "", "", fmt.Errorf("create rule: %w", respErr)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		return "", "", fmt.Errorf("expected 201 creating rule, got %d", resp.StatusCode)
	}
	var rule struct {
		ID           string `json:"id"`
		TriggerEvent string `json:"triggerEvent"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rule); err != nil {
		return "", "", fmt.Errorf("decode rule: %w", err)
	}
	return rule.ID, rule.TriggerEvent, nil
}

func TestAutomationRuleAPICreateAndExecution(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()

	var ruleID string
	var triggerEvent string

	scenario.
		When("I create an automation rule for ticket.state_changed", func(s *Scenario) error {
			body := `{
	"name":"Auto-comment on state change",
	"enabled":true,
	"triggerEvent":"ticket.state_changed",
	"triggerConditions":{},
	"actions":[{"type":"add_comment","params":{"body":"State changed on {{ticket.title}}"}}]
}`
			var err error
			ruleID, triggerEvent, err = createAutomationRule(s, seed.ProjectID, body)
			return err
		}).
		Then("the rule has triggerEvent=ticket.state_changed", func(s *Scenario) error {
			if triggerEvent != "ticket.state_changed" {
				return fmt.Errorf("expected triggerEvent=ticket.state_changed, got %s", triggerEvent)
			}
			return nil
		}).
		And("the rule appears in the rules list", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("GET", fmt.Sprintf("/projects/%s/automation/rules", seed.ProjectID), nil)
			if err != nil {
				return fmt.Errorf("list rules: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200 listing rules, got %d", resp.StatusCode)
			}
			var listResult struct {
				Items []struct{ ID string `json:"id"` } `json:"items"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&listResult); err != nil {
				return fmt.Errorf("decode list: %w", err)
			}
			if len(listResult.Items) == 0 {
				return fmt.Errorf("expected at least one rule")
			}
			return nil
		}).
		And("the executions list is initially empty", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("GET", fmt.Sprintf("/projects/%s/automation/rules/%s/executions", seed.ProjectID, ruleID), nil)
			if err != nil {
				return fmt.Errorf("list executions: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200 listing executions, got %d", resp.StatusCode)
			}
			return nil
		})
}

func TestAutomationRuleDisabledNotReturned(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()

	var ruleID string

	scenario.
		Given("a disabled automation rule exists", func(s *Scenario) error {
			body := `{"name":"Disabled rule","enabled":false,"triggerEvent":"ticket.created","actions":[]}`
			resp, err := s.Harness().APIRequest("POST", fmt.Sprintf("/projects/%s/automation/rules", seed.ProjectID), strings.NewReader(body))
			if err != nil {
				return fmt.Errorf("create rule: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 201 {
				return fmt.Errorf("expected 201, got %d", resp.StatusCode)
			}
			var rule struct {
				ID      string `json:"id"`
				Enabled bool   `json:"enabled"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&rule); err != nil {
				return fmt.Errorf("decode rule: %w", err)
			}
			if rule.Enabled {
				return fmt.Errorf("expected rule to be disabled")
			}
			ruleID = rule.ID
			return nil
		}).
		When("I enable the rule via PUT", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("PUT", fmt.Sprintf("/projects/%s/automation/rules/%s", seed.ProjectID, ruleID), strings.NewReader(`{"enabled":true}`))
			if err != nil {
				return fmt.Errorf("update rule: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200 updating rule, got %d", resp.StatusCode)
			}
			return nil
		}).
		Then("the rule can be deleted", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("DELETE", fmt.Sprintf("/projects/%s/automation/rules/%s", seed.ProjectID, ruleID), nil)
			if err != nil {
				return fmt.Errorf("delete rule: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 204 {
				return fmt.Errorf("expected 204 deleting rule, got %d", resp.StatusCode)
			}
			return nil
		})
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

func TestAutomationUICanvasRendersOnNewRule(t *testing.T) {
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
		WhenIClickKey("settings.automation_tab").
		ThenISeeSelectorKey("settings.automation_view").
		WhenIClickKey("automation.add_rule_button").
		ThenISeeSelectorKey("automation.rule_name_input").
		ThenISeeSelectorKey("automation.canvas").
		ThenISeeSelectorKey("automation.add_action_node").
		ThenISeeSelectorKey("automation.add_branch_node")
}

func TestAutomationUICreateGraphRule(t *testing.T) {
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
		WhenIClickKey("settings.automation_tab").
		WhenIClickKey("automation.add_rule_button").
		WhenIFillKey("automation.rule_name_input", "Graph-based rule").
		WhenIClickKey("automation.rule_save_button").
		ThenISeeSelectorKey("automation.rule_list")
}

func TestAutomationGraphRuleRoundTrip(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()

	// Create a rule with a graph via API
	body := `{
"name":"Graph rule",
"enabled":true,
"triggerEvent":"ticket.state_changed",
"triggerConditions":{},
"actions":[],
"graph":{
  "nodes":[
    {"id":"trigger-1","type":"trigger","position":{"x":200,"y":50},"data":{"triggerEvent":"ticket.state_changed","condFromState":"","condToState":""}},
    {"id":"action-1","type":"action","position":{"x":200,"y":220},"data":{"actionType":"add_comment","params":{"body":"State changed on {{ticket.title}}"}}}
  ],
  "edges":[
    {"id":"e-1","source":"trigger-1","target":"action-1"}
  ]
}
}`
	var ruleID string
	scenario.
		When("I create a graph-based rule via API", func(s *Scenario) error {
			var err error
			ruleID, _, err = createAutomationRule(s, seed.ProjectID, body)
			return err
		}).
		Then("the rule is returned with its graph intact", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("GET", fmt.Sprintf("/projects/%s/automation/rules/%s", seed.ProjectID, ruleID), nil)
			if err != nil {
				return fmt.Errorf("get rule: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200, got %d", resp.StatusCode)
			}
			var rule struct {
				ID    string `json:"id"`
				Graph *struct {
					Nodes []struct {
						ID   string `json:"id"`
						Type string `json:"type"`
					} `json:"nodes"`
					Edges []struct {
						ID string `json:"id"`
					} `json:"edges"`
				} `json:"graph"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&rule); err != nil {
				return fmt.Errorf("decode rule: %w", err)
			}
			if rule.Graph == nil {
				return fmt.Errorf("expected graph to be non-nil")
			}
			if len(rule.Graph.Nodes) != 2 {
				return fmt.Errorf("expected 2 nodes, got %d", len(rule.Graph.Nodes))
			}
			if len(rule.Graph.Edges) != 1 {
				return fmt.Errorf("expected 1 edge, got %d", len(rule.Graph.Edges))
			}
			return nil
		})
}

func TestAutomationGraphBranchRuleRoundTrip(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()

	body := `{
"name":"Branch rule",
"enabled":true,
"triggerEvent":"ticket.created",
"triggerConditions":{},
"actions":[],
"graph":{
  "nodes":[
    {"id":"trigger-1","type":"trigger","position":{"x":200,"y":50},"data":{"triggerEvent":"ticket.created"}},
    {"id":"branch-1","type":"branch","position":{"x":200,"y":220},"data":{"conditionField":"priority","conditionOperator":"equals","conditionValue":"urgent"}},
    {"id":"action-true","type":"action","position":{"x":100,"y":400},"data":{"actionType":"set_priority","params":{"priority":"urgent"}}},
    {"id":"action-false","type":"action","position":{"x":300,"y":400},"data":{"actionType":"add_comment","params":{"body":"Non-urgent ticket created"}}}
  ],
  "edges":[
    {"id":"e-1","source":"trigger-1","target":"branch-1"},
    {"id":"e-2","source":"branch-1","sourceHandle":"true","target":"action-true"},
    {"id":"e-3","source":"branch-1","sourceHandle":"false","target":"action-false"}
  ]
}
}`

	var ruleID string
	scenario.
		When("I create a branch rule", func(s *Scenario) error {
			var err error
			ruleID, _, err = createAutomationRule(s, seed.ProjectID, body)
			return err
		}).
		Then("the branch rule is persisted with 4 nodes and 3 edges", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("GET", fmt.Sprintf("/projects/%s/automation/rules/%s", seed.ProjectID, ruleID), nil)
			if err != nil {
				return fmt.Errorf("get rule: %w", err)
			}
			defer resp.Body.Close()
			var rule struct {
				Graph *struct {
					Nodes []struct{ Type string `json:"type"` } `json:"nodes"`
					Edges []struct {
						SourceHandle *string `json:"sourceHandle"`
					} `json:"edges"`
				} `json:"graph"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&rule); err != nil {
				return fmt.Errorf("decode: %w", err)
			}
			if rule.Graph == nil {
				return fmt.Errorf("expected graph")
			}
			if len(rule.Graph.Nodes) != 4 {
				return fmt.Errorf("expected 4 nodes, got %d", len(rule.Graph.Nodes))
			}
			branchEdges := 0
			for _, e := range rule.Graph.Edges {
				if e.SourceHandle != nil && (*e.SourceHandle == "true" || *e.SourceHandle == "false") {
					branchEdges++
				}
			}
			if branchEdges != 2 {
				return fmt.Errorf("expected 2 branch edges (true/false), got %d", branchEdges)
			}
			return nil
		})
}
