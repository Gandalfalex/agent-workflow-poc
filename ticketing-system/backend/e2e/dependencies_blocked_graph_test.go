//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"ticketing-system/backend/internal/store"
)

func TestDependenciesBlockedFilterAndGraph(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()
	st := scenario.Harness().Store()
	ctx := scenario.Harness().Context()
	projectID := uuid.MustParse(seed.ProjectID)
	storyID := uuid.MustParse(seed.StoryID)
	backlogID := uuid.MustParse(seed.BacklogID)

	ticketA, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
		Title:   fmt.Sprintf("Dep A %d", time.Now().UnixNano()),
		Type:    "feature",
		StoryID: storyID,
		StateID: &backlogID,
	})
	if err != nil {
		t.Fatalf("seed ticket A: %v", err)
	}
	ticketB, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
		Title:   fmt.Sprintf("Dep B %d", time.Now().UnixNano()),
		Type:    "feature",
		StoryID: storyID,
		StateID: &backlogID,
	})
	if err != nil {
		t.Fatalf("seed ticket B: %v", err)
	}
	ticketC, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
		Title:   fmt.Sprintf("Dep C %d", time.Now().UnixNano()),
		Type:    "feature",
		StoryID: storyID,
		StateID: &backlogID,
	})
	if err != nil {
		t.Fatalf("seed ticket C: %v", err)
	}

	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/"+seed.ProjectID+"/board").
		Then("dependencies are created and cycles are rejected by API", func(s *Scenario) error {
			if err := apiCreateDependencyExpect(s.Harness(), ticketA.ID.String(), ticketB.ID.String(), "blocks", http.StatusCreated, ""); err != nil {
				return err
			}
			if err := apiCreateDependencyExpect(s.Harness(), ticketB.ID.String(), ticketC.ID.String(), "blocks", http.StatusCreated, ""); err != nil {
				return err
			}
			return apiCreateDependencyExpect(s.Harness(), ticketC.ID.String(), ticketA.ID.String(), "blocks", http.StatusConflict, "dependency_cycle")
		}).
		WhenIClickRefresh().
		ThenISeeSelectorKey("board.ticket_blocked_badge").
		WhenIClickKey("board.filter_toggle_button").
		WhenIClickKey("board.filter_blocked_checkbox").
		Then("blocked filter hides unblocked ticket and keeps blocked ticket visible", func(s *Scenario) error {
			aVisible, err := ticketTitleVisible(s.Harness(), ticketA.Key)
			if err != nil {
				return err
			}
			if aVisible {
				return fmt.Errorf("expected unblocked ticket %q to be hidden", ticketA.Key)
			}
			bVisible, err := ticketTitleVisible(s.Harness(), ticketB.Key)
			if err != nil {
				return err
			}
			if !bVisible {
				return fmt.Errorf("expected blocked ticket %q to be visible", ticketB.Key)
			}
			return nil
		}).
		When("I open a blocked ticket modal", func(s *Scenario) error {
			return s.Harness().page.GetByText(ticketB.Title).First().Click()
		}).
		ThenISeeSelectorKey("ticket.modal").
		ThenISeeSelectorKey("ticket.dependencies_section").
		AndISeeSelectorKey("ticket.dependency_graph").
		When("I close ticket modal", func(s *Scenario) error {
			return s.Harness().Click("[data-testid=\"ticket.close-button\"]")
		}).
		WhenIClickKey("nav.dashboard_tab").
		ThenURLContains("/projects/" + seed.ProjectID + "/dashboard").
		ThenISeeSelectorKey("dashboard.dependency_graph")
}

func apiCreateDependencyExpect(h *Harness, ticketID, relatedTicketID, relationType string, expectStatus int, expectBody string) error {
	payload := map[string]any{
		"relatedTicketId": relatedTicketID,
		"relationType":    relationType,
	}
	raw, _ := json.Marshal(payload)
	resp, err := h.APIRequest(http.MethodPost, "/tickets/"+ticketID+"/dependencies", bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("create dependency request: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	bodyText := strings.TrimSpace(string(body))
	if resp.StatusCode != expectStatus {
		return fmt.Errorf("create dependency status %d: %s", resp.StatusCode, bodyText)
	}
	if expectBody != "" && !strings.Contains(bodyText, expectBody) {
		return fmt.Errorf("expected response body to contain %q, got %s", expectBody, bodyText)
	}
	return nil
}

func ticketTitleVisible(h *Harness, title string) (bool, error) {
	return h.page.GetByText(title).First().IsVisible()
}
