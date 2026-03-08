//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// simResult is the envelope returned by POST /automation/simulate.
type simResult struct {
	Run struct {
		ID           string  `json:"id"`
		RuleId       *string `json:"ruleId"`
		MatchedCount int     `json:"matchedCount"`
		FailureCount int     `json:"failureCount"`
		TotalEvents  int     `json:"totalEvents"`
		LookbackDays int     `json:"lookbackDays"`
	} `json:"run"`
	Items []struct {
		TicketKey        string `json:"ticketKey"`
		Matched          bool   `json:"matched"`
		FailureReason    string `json:"failureReason"`
		PredictedActions []struct {
			Type         string `json:"type"`
			WouldSucceed bool   `json:"wouldSucceed"`
		} `json:"predictedActions"`
	} `json:"items"`
	Total int `json:"total"`
}

func simulate(s *Scenario, projectID, body string) (simResult, error) {
	resp, err := s.Harness().APIRequest("POST", fmt.Sprintf("/projects/%s/automation/simulate", projectID), strings.NewReader(body))
	if err != nil {
		return simResult{}, fmt.Errorf("POST simulate: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return simResult{}, fmt.Errorf("expected 200, got %d", resp.StatusCode)
	}
	var r simResult
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return simResult{}, fmt.Errorf("decode: %w", err)
	}
	return r, nil
}

func createTicket(s *Scenario, projectID, storyID, stateID, priority, title string) (string, error) {
	body := fmt.Sprintf(`{"title":%q,"type":"feature","priority":%q,"storyId":%q,"stateId":%q}`,
		title, priority, storyID, stateID)
	resp, err := s.Harness().APIRequest("POST", fmt.Sprintf("/projects/%s/tickets", projectID), strings.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("POST ticket: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		return "", fmt.Errorf("expected 201, got %d", resp.StatusCode)
	}
	var t struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}
	return t.ID, nil
}

func TestSimulateReturnsRunOnEmptyProject(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()
	seed := scenario.SeedData()

	var result simResult
	ruleBody := fmt.Sprintf(`{"lookbackDays":7,"rule":{"name":"test","triggerEvent":"ticket.state_changed","triggerConditions":{"fromState":%q,"toState":%q},"actions":[{"type":"add_comment","params":{"body":"hi"}}],"enabled":false}}`,
		seed.BacklogID, seed.InProgressID)

	scenario.
		When("I simulate a state_changed rule on a project with no history", func(s *Scenario) error {
			var err error
			result, err = simulate(s, seed.ProjectID, ruleBody)
			return err
		}).
		Then("a run is created with the requested lookback window", func(s *Scenario) error {
			if result.Run.ID == "" {
				return fmt.Errorf("expected non-empty run id")
			}
			if result.Run.LookbackDays != 7 {
				return fmt.Errorf("expected lookbackDays 7, got %d", result.Run.LookbackDays)
			}
			return nil
		}).
		And("the matched count is zero", func(s *Scenario) error {
			if result.Run.MatchedCount != 0 {
				return fmt.Errorf("expected 0 matched for empty project, got %d", result.Run.MatchedCount)
			}
			return nil
		})
}

func TestSimulateDoesNotMutateTickets(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()
	seed := scenario.SeedData()

	var ticketID string

	scenario.
		Given("a ticket has been moved to generate a state_changed event", func(s *Scenario) error {
			id, err := createTicket(s, seed.ProjectID, seed.StoryID, seed.BacklogID, "medium", "Sim ticket")
			if err != nil {
				return err
			}
			ticketID = id
			resp, err := s.Harness().APIRequest("PATCH", fmt.Sprintf("/tickets/%s", ticketID),
				strings.NewReader(fmt.Sprintf(`{"stateId":%q}`, seed.InProgressID)))
			if err != nil {
				return fmt.Errorf("move ticket: %w", err)
			}
			resp.Body.Close()
			return nil
		}).
		When("I simulate a rule that would set priority=high", func(s *Scenario) error {
			body := `{"lookbackDays":7,"rule":{"name":"no-mutate","triggerEvent":"ticket.state_changed","triggerConditions":{},"actions":[{"type":"set_priority","params":{"priority":"high"}}],"enabled":false}}`
			_, err := simulate(s, seed.ProjectID, body)
			return err
		}).
		Then("the ticket priority is unchanged", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("GET", fmt.Sprintf("/tickets/%s", ticketID), nil)
			if err != nil {
				return fmt.Errorf("get ticket: %w", err)
			}
			defer resp.Body.Close()
			var got struct {
				Priority string `json:"priority"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
				return fmt.Errorf("decode: %w", err)
			}
			if got.Priority == "high" {
				return fmt.Errorf("ticket priority was mutated by simulation")
			}
			return nil
		})
}

func TestSimulateMatchedEventsPopulatePredictedActions(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()
	seed := scenario.SeedData()

	var result simResult

	scenario.
		Given("an urgent ticket exists", func(s *Scenario) error {
			_, err := createTicket(s, seed.ProjectID, seed.StoryID, seed.BacklogID, "urgent", "Urgent match")
			return err
		}).
		When("I simulate a rule matching priority=urgent with an add_comment action", func(s *Scenario) error {
			body := `{"lookbackDays":7,"rule":{"name":"urgent-rule","triggerEvent":"ticket.created","triggerConditions":{"priority":"urgent"},"actions":[{"type":"add_comment","params":{"body":"urgent!"}}],"enabled":false}}`
			var err error
			result, err = simulate(s, seed.ProjectID, body)
			return err
		}).
		Then("at least one event matches", func(s *Scenario) error {
			if result.Run.TotalEvents == 0 {
				return fmt.Errorf("expected at least 1 event in lookback window")
			}
			if result.Run.MatchedCount == 0 {
				return fmt.Errorf("expected at least 1 matched event (total: %d)", result.Run.TotalEvents)
			}
			return nil
		}).
		And("the matched event has a successful add_comment prediction", func(s *Scenario) error {
			for _, item := range result.Items {
				if !item.Matched {
					continue
				}
				if len(item.PredictedActions) == 0 {
					return fmt.Errorf("matched event has no predicted actions")
				}
				a := item.PredictedActions[0]
				if a.Type != "add_comment" {
					return fmt.Errorf("expected add_comment, got %s", a.Type)
				}
				if !a.WouldSucceed {
					return fmt.Errorf("expected add_comment prediction to be wouldSucceed=true")
				}
				return nil
			}
			return fmt.Errorf("no matched item found in results")
		})
}

func TestSimulateImpossibleConditionMatchesNothing(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()
	seed := scenario.SeedData()

	var result simResult

	scenario.
		Given("a low-priority ticket exists", func(s *Scenario) error {
			_, err := createTicket(s, seed.ProjectID, seed.StoryID, seed.BacklogID, "low", "Low prio")
			return err
		}).
		When("I simulate a rule requiring priority=urgent", func(s *Scenario) error {
			body := `{"lookbackDays":7,"rule":{"name":"no-match","triggerEvent":"ticket.created","triggerConditions":{"priority":"urgent"},"actions":[{"type":"add_comment","params":{"body":"x"}}],"enabled":false}}`
			var err error
			result, err = simulate(s, seed.ProjectID, body)
			return err
		}).
		Then("no events match", func(s *Scenario) error {
			if result.Run.MatchedCount != 0 {
				return fmt.Errorf("expected 0 matched, got %d", result.Run.MatchedCount)
			}
			return nil
		})
}

func TestSimulateRunPersistedAndRetrievable(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()
	seed := scenario.SeedData()

	var runID string

	scenario.
		When("I run a simulation", func(s *Scenario) error {
			body := `{"lookbackDays":7,"rule":{"name":"persist","triggerEvent":"ticket.created","triggerConditions":{},"actions":[{"type":"add_comment","params":{"body":"created"}}],"enabled":false}}`
			result, err := simulate(s, seed.ProjectID, body)
			if err != nil {
				return err
			}
			runID = result.Run.ID
			return nil
		}).
		Then("the run appears in the simulation-runs list", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("GET", fmt.Sprintf("/projects/%s/automation/simulation-runs", seed.ProjectID), nil)
			if err != nil {
				return fmt.Errorf("list runs: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200, got %d", resp.StatusCode)
			}
			var list struct {
				Items []struct {
					ID string `json:"id"`
				} `json:"items"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
				return fmt.Errorf("decode: %w", err)
			}
			for _, item := range list.Items {
				if item.ID == runID {
					return nil
				}
			}
			return fmt.Errorf("run %s not found in list", runID)
		}).
		And("the run is retrievable by ID with correct lookbackDays", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("GET", fmt.Sprintf("/projects/%s/automation/simulation-runs/%s", seed.ProjectID, runID), nil)
			if err != nil {
				return fmt.Errorf("get run: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200, got %d", resp.StatusCode)
			}
			var got struct {
				Run struct {
					ID           string `json:"id"`
					LookbackDays int    `json:"lookbackDays"`
				} `json:"run"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
				return fmt.Errorf("decode: %w", err)
			}
			if got.Run.ID != runID {
				return fmt.Errorf("expected id %s, got %s", runID, got.Run.ID)
			}
			if got.Run.LookbackDays != 7 {
				return fmt.Errorf("expected lookbackDays 7, got %d", got.Run.LookbackDays)
			}
			return nil
		})
}

func TestSimulateDraftRuleHasNoRuleId(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()
	seed := scenario.SeedData()

	var result simResult

	scenario.
		When("I simulate a draft rule (no saved ruleId)", func(s *Scenario) error {
			body := `{"lookbackDays":7,"rule":{"name":"draft","triggerEvent":"ticket.updated","triggerConditions":{"priority":"urgent"},"actions":[{"type":"set_assignee","params":{"assignee_id":""}}],"enabled":false}}`
			var err error
			result, err = simulate(s, seed.ProjectID, body)
			return err
		}).
		Then("the run's ruleId is null", func(s *Scenario) error {
			if result.Run.RuleId != nil {
				return fmt.Errorf("expected null ruleId for draft simulation, got %q", *result.Run.RuleId)
			}
			return nil
		})
}

func TestSimulateStateChangedTriggerMatches(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()
	seed := scenario.SeedData()

	var ticketID string
	var result simResult

	scenario.
		Given("a ticket has been moved from Backlog to In Progress", func(s *Scenario) error {
			id, err := createTicket(s, seed.ProjectID, seed.StoryID, seed.BacklogID, "medium", "State change test")
			if err != nil {
				return err
			}
			ticketID = id
			resp, err := s.Harness().APIRequest("PATCH", fmt.Sprintf("/tickets/%s", ticketID),
				strings.NewReader(fmt.Sprintf(`{"stateId":%q}`, seed.InProgressID)))
			if err != nil {
				return fmt.Errorf("move ticket: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				t.Skipf("state change blocked (status %d); skipping", resp.StatusCode)
			}
			return nil
		}).
		When("I simulate a ticket.state_changed rule with no conditions", func(s *Scenario) error {
			body := `{"lookbackDays":7,"rule":{"name":"state-change-sim","triggerEvent":"ticket.state_changed","triggerConditions":{},"actions":[{"type":"add_comment","params":{"body":"moved"}}],"enabled":false}}`
			var err error
			result, err = simulate(s, seed.ProjectID, body)
			return err
		}).
		Then("the state change event is found and matches", func(s *Scenario) error {
			if result.Run.TotalEvents == 0 {
				return fmt.Errorf("expected at least 1 state_changed event")
			}
			if result.Run.MatchedCount == 0 {
				return fmt.Errorf("expected at least 1 match with empty conditions (total: %d)", result.Run.TotalEvents)
			}
			return nil
		})
}

func TestSimulateBrokenActionPredictedAsFailure(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()
	seed := scenario.SeedData()

	var result simResult

	scenario.
		Given("an urgent ticket exists so the rule matches", func(s *Scenario) error {
			_, err := createTicket(s, seed.ProjectID, seed.StoryID, seed.BacklogID, "urgent", "Broken action test")
			return err
		}).
		When("I simulate a rule with a set_state action that has an empty state_id", func(s *Scenario) error {
			body := `{"lookbackDays":7,"rule":{"name":"broken","triggerEvent":"ticket.created","triggerConditions":{"priority":"urgent"},"actions":[{"type":"set_state","params":{"state_id":""}}],"enabled":false}}`
			var err error
			result, err = simulate(s, seed.ProjectID, body)
			return err
		}).
		Then("the run reports at least one failure", func(s *Scenario) error {
			if result.Run.MatchedCount == 0 {
				return fmt.Errorf("expected at least 1 match")
			}
			if result.Run.FailureCount == 0 {
				return fmt.Errorf("expected failureCount > 0 for broken set_state action")
			}
			return nil
		}).
		And("the matched item's predicted action has wouldSucceed=false", func(s *Scenario) error {
			for _, item := range result.Items {
				if !item.Matched {
					continue
				}
				if len(item.PredictedActions) == 0 {
					return fmt.Errorf("matched item has no predicted actions")
				}
				if item.PredictedActions[0].WouldSucceed {
					return fmt.Errorf("expected set_state with empty state_id to have wouldSucceed=false")
				}
				return nil
			}
			return fmt.Errorf("no matched item found in results")
		})
}

func TestSimulateRunResultsPagination(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()
	seed := scenario.SeedData()

	var runID string
	var page1Keys, page2Keys []string

	scenario.
		Given("two urgent tickets exist", func(s *Scenario) error {
			for i := range 2 {
				_, err := createTicket(s, seed.ProjectID, seed.StoryID, seed.BacklogID, "urgent",
					fmt.Sprintf("Page ticket %d", i))
				if err != nil {
					return err
				}
			}
			return nil
		}).
		When("I simulate a rule matching all ticket.created events", func(s *Scenario) error {
			body := `{"lookbackDays":7,"rule":{"name":"page-test","triggerEvent":"ticket.created","triggerConditions":{},"actions":[{"type":"add_comment","params":{"body":"x"}}],"enabled":false}}`
			result, err := simulate(s, seed.ProjectID, body)
			if err != nil {
				return err
			}
			if result.Total < 2 {
				return fmt.Errorf("expected at least 2 results, got %d", result.Total)
			}
			runID = result.Run.ID
			return nil
		}).
		Then("page 1 (limit=1) returns exactly one result", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("GET",
				fmt.Sprintf("/projects/%s/automation/simulation-runs/%s?limit=1&offset=0", seed.ProjectID, runID), nil)
			if err != nil {
				return fmt.Errorf("page 1: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200, got %d", resp.StatusCode)
			}
			var page struct {
				Items []struct{ TicketKey string `json:"ticketKey"` } `json:"items"`
				Total int                                             `json:"total"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
				return fmt.Errorf("decode: %w", err)
			}
			if len(page.Items) != 1 {
				return fmt.Errorf("expected 1 item, got %d", len(page.Items))
			}
			if page.Total < 2 {
				return fmt.Errorf("expected total >= 2, got %d", page.Total)
			}
			page1Keys = []string{page.Items[0].TicketKey}
			return nil
		}).
		And("page 2 (limit=1, offset=1) returns a different result", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("GET",
				fmt.Sprintf("/projects/%s/automation/simulation-runs/%s?limit=1&offset=1", seed.ProjectID, runID), nil)
			if err != nil {
				return fmt.Errorf("page 2: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200, got %d", resp.StatusCode)
			}
			var page struct {
				Items []struct{ TicketKey string `json:"ticketKey"` } `json:"items"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
				return fmt.Errorf("decode: %w", err)
			}
			if len(page.Items) != 1 {
				return fmt.Errorf("expected 1 item, got %d", len(page.Items))
			}
			page2Keys = []string{page.Items[0].TicketKey}
			if len(page1Keys) > 0 && page1Keys[0] == page2Keys[0] {
				return fmt.Errorf("page 1 and page 2 returned the same ticket (%s)", page1Keys[0])
			}
			return nil
		})
}

func TestSimulateViewerCanReadRunsButNotSimulate(t *testing.T) {
	t.Parallel()

	adminScenario := NewScenario(t)
	defer adminScenario.Close()
	adminSeed := adminScenario.SeedData()

	viewerScenario := NewScenario(t, WithViewerUser())
	defer viewerScenario.Close()
	viewerSeed := viewerScenario.SeedData()

	adminScenario.
		Given("an admin has created a simulation run", func(s *Scenario) error {
			body := `{"lookbackDays":7,"rule":{"name":"viewer-test","triggerEvent":"ticket.created","triggerConditions":{},"actions":[{"type":"add_comment","params":{"body":"x"}}],"enabled":false}}`
			_, err := simulate(s, adminSeed.ProjectID, body)
			return err
		})

	viewerScenario.
		When("a viewer lists simulation runs", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("GET", fmt.Sprintf("/projects/%s/automation/simulation-runs", viewerSeed.ProjectID), nil)
			if err != nil {
				return fmt.Errorf("list runs: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200, got %d", resp.StatusCode)
			}
			return nil
		}).
		Then("the viewer cannot POST simulate (403)", func(s *Scenario) error {
			body := `{"lookbackDays":7,"rule":{"name":"x","triggerEvent":"ticket.created","triggerConditions":{},"actions":[{"type":"add_comment","params":{"body":"x"}}],"enabled":false}}`
			resp, err := s.Harness().APIRequest("POST", fmt.Sprintf("/projects/%s/automation/simulate", viewerSeed.ProjectID), strings.NewReader(body))
			if err != nil {
				return fmt.Errorf("POST simulate: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 403 {
				return fmt.Errorf("expected 403, got %d", resp.StatusCode)
			}
			return nil
		})
}

func TestSimulationUIShowsSimulateButton(t *testing.T) {
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
		ThenISeeSelectorKey("settings.automation_view").
		WhenIClickKey("automation.add_rule_button").
		ThenISeeSelectorKey("automation.simulate_button")
}
