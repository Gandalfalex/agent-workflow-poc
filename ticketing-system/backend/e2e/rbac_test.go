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
	h := scenario.Harness()

	body := `{"title":"Should Fail"}`
	resp, err := h.APIRequest("POST", fmt.Sprintf("/api/projects/%s/tickets", seed.ProjectID), strings.NewReader(body))
	if err != nil {
		t.Fatalf("API request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 403 {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
}

func TestRBACViewerAPIUpdateWorkflow403(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t, WithViewerUser())
	defer scenario.Close()

	seed := scenario.SeedData()
	h := scenario.Harness()

	body := `{"states":[{"name":"Done","order":1,"isDefault":true,"isClosed":true}]}`
	resp, err := h.APIRequest("PUT", fmt.Sprintf("/api/projects/%s/workflow", seed.ProjectID), strings.NewReader(body))
	if err != nil {
		t.Fatalf("API request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 403 {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
}
