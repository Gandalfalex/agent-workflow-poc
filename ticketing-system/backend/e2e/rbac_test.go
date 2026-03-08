//go:build e2e

package e2e

import (
	"fmt"
	"strings"
	"testing"
)

func TestRBACViewerCannotCreateTicketUI(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t, WithViewerUser())
	defer scenario.Close()

	seed := scenario.SeedData()

	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInWithHarnessUser().
		ThenISeeSelectorKey("nav.board_tab").
		WhenIGoToRoute("board", map[string]string{"projectId": seed.ProjectID}).
		ThenISeeText("E2E Project").
		ThenIDoNotSeeSelectorKey("board.add_ticket_button").
		AndIDoNotSeeSelectorKey("board.create_story_button")
}

func TestRBACViewerCannotSeeSettingsTab(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t, WithViewerUser())
	defer scenario.Close()

	seed := scenario.SeedData()

	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInWithHarnessUser().
		ThenISeeSelectorKey("nav.board_tab").
		WhenIGoToRoute("board", map[string]string{"projectId": seed.ProjectID}).
		ThenISeeText("E2E Project").
		ThenIDoNotSeeSelectorKey("nav.settings_tab")
}

func TestRBACViewerAPICreateTicket403(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t, WithViewerUser())
	defer scenario.Close()

	seed := scenario.SeedData()

	scenario.
		When("a viewer POSTs a new ticket via the API", func(s *Scenario) error {
			body := `{"title":"Should Fail"}`
			resp, err := s.Harness().APIRequest("POST", fmt.Sprintf("/projects/%s/tickets", seed.ProjectID), strings.NewReader(body))
			if err != nil {
				return fmt.Errorf("API request failed: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 403 {
				return fmt.Errorf("expected 403, got %d", resp.StatusCode)
			}
			return nil
		}).
		Then("the request is rejected with 403 Forbidden", func(s *Scenario) error {
			return nil
		})
}

func TestRBACViewerAPIUpdateWorkflow403(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t, WithViewerUser())
	defer scenario.Close()

	seed := scenario.SeedData()

	scenario.
		When("a viewer PUTs a workflow update via the API", func(s *Scenario) error {
			body := `{"states":[{"name":"Done","order":1,"isDefault":true,"isClosed":true}]}`
			resp, err := s.Harness().APIRequest("PUT", fmt.Sprintf("/projects/%s/workflow", seed.ProjectID), strings.NewReader(body))
			if err != nil {
				return fmt.Errorf("API request failed: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 403 {
				return fmt.Errorf("expected 403, got %d", resp.StatusCode)
			}
			return nil
		}).
		Then("the request is rejected with 403 Forbidden", func(s *Scenario) error {
			return nil
		})
}
