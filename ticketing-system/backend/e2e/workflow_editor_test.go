//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestWorkflowEditorAddAndSave(t *testing.T) {
	t.Parallel()

	sc := NewScenario(t)
	defer sc.Close()

	seed := sc.SeedData()
	id, pw := sc.Harness().LoginCredentials()

	sc.
		GivenAppIsRunning().
		WhenIGoTo("/").
		WhenILogInAs(id, pw).
		ThenISeeSelectorKey("nav.board_tab").
		WhenISelectProjectByID(seed.ProjectID).
		WhenIClickRefresh().
		ThenISeeSelectorKey("nav.settings_tab").
		WhenINavigateToSettings().
		WhenIOpenWorkflowTab().
		ThenISeeAtLeastWorkflowStateRows(3).
		ThenWorkflowIncludesStates("Backlog", "In Progress", "Done").
		WhenIClickKey("workflow.add_state_button").
		ThenISeeAtLeastWorkflowStateRows(4).
		WhenINameLastWorkflowState("QA Review").
		WhenIClickKey("workflow.save_button").
		ThenISeeWorkflowSavedNotice().
		Then("the API returns 4 states including QA Review", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest(http.MethodGet, fmt.Sprintf("/projects/%s/workflow", seed.ProjectID), nil)
			if err != nil {
				return fmt.Errorf("GET workflow: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("expected 200, got %d", resp.StatusCode)
			}

			var result struct {
				States []struct {
					Name string `json:"name"`
				} `json:"states"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return fmt.Errorf("decode workflow: %w", err)
			}
			if len(result.States) != 4 {
				return fmt.Errorf("expected 4 states, got %d", len(result.States))
			}

			for _, st := range result.States {
				if st.Name == "QA Review" {
					return nil
				}
			}
			return fmt.Errorf("QA Review state not found in API response")
		})
}

func TestWorkflowEditorRenameAndToggle(t *testing.T) {
	t.Parallel()

	sc := NewScenario(t)
	defer sc.Close()

	id, pw := sc.Harness().LoginCredentials()
	seed := sc.SeedData()

	sc.
		GivenAppIsRunning().
		WhenIGoTo("/").
		WhenILogInAs(id, pw).
		ThenISeeSelectorKey("nav.board_tab").
		WhenISelectProjectByID(seed.ProjectID).
		WhenIClickRefresh().
		ThenISeeSelectorKey("nav.settings_tab").
		WhenINavigateToSettings().
		WhenIOpenWorkflowTab().
		ThenISeeAtLeastWorkflowStateRows(3).
		WhenIRenameWorkflowState(1, "Active").
		WhenIToggleWorkflowStateClosed(1).
		WhenIClickKey("workflow.save_button").
		ThenISeeWorkflowSavedNotice().
		Then("the API shows renamed state 'Active' with isClosed", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest(http.MethodGet, fmt.Sprintf("/projects/%s/workflow", seed.ProjectID), nil)
			if err != nil {
				return fmt.Errorf("GET workflow: %w", err)
			}
			defer resp.Body.Close()

			var result struct {
				States []struct {
					Name     string `json:"name"`
					IsClosed bool   `json:"isClosed"`
				} `json:"states"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return fmt.Errorf("decode workflow: %w", err)
			}

			for _, st := range result.States {
				if st.Name == "Active" {
					if !st.IsClosed {
						return fmt.Errorf("expected Active to be closed")
					}
					return nil
				}
			}
			return fmt.Errorf("state 'Active' not found in API response")
		})
}

func TestWorkflowEditorValidation(t *testing.T) {
	t.Parallel()

	sc := NewScenario(t)
	defer sc.Close()

	id, pw := sc.Harness().LoginCredentials()
	seed := sc.SeedData()

	sc.
		GivenAppIsRunning().
		WhenIGoTo("/").
		WhenILogInAs(id, pw).
		ThenISeeSelectorKey("nav.board_tab").
		WhenISelectProjectByID(seed.ProjectID).
		WhenIClickRefresh().
		ThenISeeSelectorKey("nav.settings_tab").
		WhenINavigateToSettings().
		WhenIOpenWorkflowTab().
		ThenISeeAtLeastWorkflowStateRows(3).
		WhenIClearWorkflowStateName(0).
		WhenIClickKey("workflow.save_button").
		ThenISeeWorkflowError().
		AndISeeText("All states must have a name.")
}

func TestWorkflowEditorReorderViaAPI(t *testing.T) {
	t.Parallel()

	sc := NewScenario(t)
	defer sc.Close()

	seed := sc.SeedData()
	id, pw := sc.Harness().LoginCredentials()

	reorderBody, _ := json.Marshal(map[string]any{
		"states": []map[string]any{
			{"id": seed.DoneID, "name": "Done", "order": 1, "isDefault": false, "isClosed": true},
			{"id": seed.BacklogID, "name": "Backlog", "order": 2, "isDefault": true, "isClosed": false},
			{"id": seed.InProgressID, "name": "In Progress", "order": 3, "isDefault": false, "isClosed": false},
		},
	})

	sc.
		GivenAppIsRunning().
		When("I reorder states via PUT API", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest(
				http.MethodPut,
				fmt.Sprintf("/projects/%s/workflow", seed.ProjectID),
				bytes.NewReader(reorderBody),
			)
			if err != nil {
				return fmt.Errorf("PUT workflow: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("expected 200, got %d", resp.StatusCode)
			}
			return nil
		}).
		WhenIGoTo("/").
		WhenILogInAs(id, pw).
		ThenISeeSelectorKey("nav.board_tab").
		WhenISelectProjectByID(seed.ProjectID).
		WhenIClickRefresh().
		ThenISeeSelectorKey("nav.settings_tab").
		WhenINavigateToSettings().
		WhenIOpenWorkflowTab().
		ThenISeeAtLeastWorkflowStateRows(3).
		ThenISeeWorkflowStateOrder("Done", "Backlog", "In Progress")
}
