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
	seed := scenario.SeedData()

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
	var updated struct {
		States []struct {
			WipLimit       *int `json:"wipLimit"`
			WipEnforcement bool `json:"wipEnforcement"`
			IsClosed       bool `json:"isClosed"`
		} `json:"states"`
	}

	scenario.
		Given("the project workflow states are fetched", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("GET", fmt.Sprintf("/projects/%s/workflow", seed.ProjectID), nil)
			if err != nil {
				return fmt.Errorf("get workflow: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200 getting workflow, got %d", resp.StatusCode)
			}
			if err := json.NewDecoder(resp.Body).Decode(&wf); err != nil {
				return fmt.Errorf("decode workflow: %w", err)
			}
			if len(wf.States) == 0 {
				t.Skip("no workflow states; skipping")
			}
			return nil
		}).
		When("I update all non-closed states with wipLimit=3 and wipEnforcement=true", func(s *Scenario) error {
			var updatedStates []map[string]any
			for _, st := range wf.States {
				entry := map[string]any{
					"id":             st.ID,
					"name":           st.Name,
					"order":          st.Order,
					"isDefault":      st.IsDefault,
					"isClosed":       st.IsClosed,
					"wipEnforcement": false,
				}
				if !st.IsClosed {
					lim := 3
					entry["wipLimit"] = lim
					entry["wipEnforcement"] = true
				}
				updatedStates = append(updatedStates, entry)
			}
			body, _ := json.Marshal(map[string]any{"states": updatedStates})
			resp, err := s.Harness().APIRequest("PUT", fmt.Sprintf("/projects/%s/workflow", seed.ProjectID), strings.NewReader(string(body)))
			if err != nil {
				return fmt.Errorf("update workflow: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200 updating workflow, got %d", resp.StatusCode)
			}
			return json.NewDecoder(resp.Body).Decode(&updated)
		}).
		Then("every non-closed state has wipLimit=3 and wipEnforcement=true", func(s *Scenario) error {
			for _, st := range updated.States {
				if !st.IsClosed {
					if st.WipLimit == nil || *st.WipLimit != 3 {
						return fmt.Errorf("expected wipLimit=3 on non-closed state, got %v", st.WipLimit)
					}
					if !st.WipEnforcement {
						return fmt.Errorf("expected wipEnforcement=true on non-closed state")
					}
				}
			}
			return nil
		})
}

func TestFlowHealthAPIReturnsStats(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()
	seed := scenario.SeedData()

	var stats struct {
		WipStats   []any `json:"wipStats"`
		Throughput []any `json:"throughput"`
		CycleTime  struct {
			SampleCount int `json:"sampleCount"`
		} `json:"cycleTime"`
	}

	scenario.
		When("I request the flow-health endpoint", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("GET", fmt.Sprintf("/projects/%s/flow-health", seed.ProjectID), nil)
			if err != nil {
				return fmt.Errorf("get flow health: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200 getting flow health, got %d", resp.StatusCode)
			}
			return json.NewDecoder(resp.Body).Decode(&stats)
		}).
		Then("the response includes a wipStats array", func(s *Scenario) error {
			if stats.WipStats == nil {
				return fmt.Errorf("expected wipStats array in response")
			}
			return nil
		}).
		And("the response includes a throughput array", func(s *Scenario) error {
			if stats.Throughput == nil {
				return fmt.Errorf("expected throughput array in response")
			}
			return nil
		})
}

func TestWIPEnforcementBlocksMove(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()
	seed := scenario.SeedData()

	var enforceStateID string
	var ticketIDs []string
	var wf struct {
		States []struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			Order     int    `json:"order"`
			IsDefault bool   `json:"isDefault"`
			IsClosed  bool   `json:"isClosed"`
		} `json:"states"`
	}

	scenario.
		Given("a non-default non-closed workflow state has wipLimit=1 and wipEnforcement=true", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("GET", fmt.Sprintf("/projects/%s/workflow", seed.ProjectID), nil)
			if err != nil {
				return fmt.Errorf("get workflow: %w", err)
			}
			defer resp.Body.Close()
			if err := json.NewDecoder(resp.Body).Decode(&wf); err != nil {
				return fmt.Errorf("decode workflow: %w", err)
			}
			if len(wf.States) < 2 {
				t.Skip("need at least 2 states for enforcement test")
			}
			for _, st := range wf.States {
				if !st.IsDefault && !st.IsClosed {
					enforceStateID = st.ID
					break
				}
			}
			if enforceStateID == "" {
				t.Skip("no suitable non-default non-closed state found")
			}

			var updatedStates []map[string]any
			for _, st := range wf.States {
				entry := map[string]any{
					"id":             st.ID,
					"name":           st.Name,
					"order":          st.Order,
					"isDefault":      st.IsDefault,
					"isClosed":       st.IsClosed,
					"wipEnforcement": false,
				}
				if st.ID == enforceStateID {
					lim := 1
					entry["wipLimit"] = lim
					entry["wipEnforcement"] = true
				}
				updatedStates = append(updatedStates, entry)
			}
			body, _ := json.Marshal(map[string]any{"states": updatedStates})
			putResp, err := s.Harness().APIRequest("PUT", fmt.Sprintf("/projects/%s/workflow", seed.ProjectID), strings.NewReader(string(body)))
			if err != nil {
				return fmt.Errorf("update workflow: %w", err)
			}
			putResp.Body.Close()
			if putResp.StatusCode != 200 {
				return fmt.Errorf("expected 200 setting up WIP, got %d", putResp.StatusCode)
			}
			return nil
		}).
		And("two tickets exist in the default state", func(s *Scenario) error {
			for i := 0; i < 2; i++ {
				body := fmt.Sprintf(`{"title":"WIP test ticket %d","type":"feature","priority":"medium","storyId":"%s"}`, i, seed.StoryID)
				resp, err := s.Harness().APIRequest("POST", fmt.Sprintf("/projects/%s/tickets", seed.ProjectID), strings.NewReader(body))
				if err != nil {
					return fmt.Errorf("create ticket %d: %w", i, err)
				}
				defer resp.Body.Close()
				if resp.StatusCode != 201 {
					return fmt.Errorf("expected 201 creating ticket, got %d", resp.StatusCode)
				}
				var ticket struct {
					ID string `json:"id"`
				}
				if err := json.NewDecoder(resp.Body).Decode(&ticket); err != nil {
					return fmt.Errorf("decode ticket: %w", err)
				}
				ticketIDs = append(ticketIDs, ticket.ID)
			}
			return nil
		}).
		When("I move the first ticket into the enforced state", func(s *Scenario) error {
			body := fmt.Sprintf(`{"stateId":"%s"}`, enforceStateID)
			resp, err := s.Harness().APIRequest("PATCH", fmt.Sprintf("/tickets/%s", ticketIDs[0]), strings.NewReader(body))
			if err != nil {
				return fmt.Errorf("move ticket 1: %w", err)
			}
			resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200 moving first ticket, got %d", resp.StatusCode)
			}
			return nil
		}).
		Then("moving the second ticket into the enforced state is blocked with 409", func(s *Scenario) error {
			body := fmt.Sprintf(`{"stateId":"%s"}`, enforceStateID)
			resp, err := s.Harness().APIRequest("PATCH", fmt.Sprintf("/tickets/%s", ticketIDs[1]), strings.NewReader(body))
			if err != nil {
				return fmt.Errorf("move ticket 2: %w", err)
			}
			resp.Body.Close()
			if resp.StatusCode != 409 {
				return fmt.Errorf("expected 409 (wip_limit_exceeded) when enforcement blocks move, got %d", resp.StatusCode)
			}
			return nil
		})
}

func TestFlowHealthViewerCanAccess(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t, WithViewerUser())
	defer scenario.Close()
	seed := scenario.SeedData()

	scenario.
		When("a viewer requests the flow-health endpoint", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("GET", fmt.Sprintf("/projects/%s/flow-health", seed.ProjectID), nil)
			if err != nil {
				return fmt.Errorf("viewer get flow health: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200 for viewer, got %d", resp.StatusCode)
			}
			return nil
		}).
		Then("the request succeeds with 200", func(s *Scenario) error {
			return nil
		})
}
