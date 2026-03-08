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

	var ticketAID, ticketAKey string
	var ticketBTitle, ticketBKey string

	scenario.
		Given("three tickets A, B, C are seeded in backlog", func(s *Scenario) error {
			st := s.Harness().Store()
			ctx := s.Harness().Context()
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
				return fmt.Errorf("seed ticket A: %w", err)
			}
			ticketAID = ticketA.ID.String()
			ticketAKey = ticketA.Key

			ticketB, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
				Title:   fmt.Sprintf("Dep B %d", time.Now().UnixNano()),
				Type:    "feature",
				StoryID: storyID,
				StateID: &backlogID,
			})
			if err != nil {
				return fmt.Errorf("seed ticket B: %w", err)
			}
			ticketBTitle = ticketB.Title
			ticketBKey = ticketB.Key

			ticketC, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
				Title:   fmt.Sprintf("Dep C %d", time.Now().UnixNano()),
				Type:    "feature",
				StoryID: storyID,
				StateID: &backlogID,
			})
			if err != nil {
				return fmt.Errorf("seed ticket C: %w", err)
			}

			_ = ticketC // used below via ID captured in Then step
			return nil
		}).
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/"+seed.ProjectID+"/board").
		Then("dependencies A→B and B→C are created and cycle C→A is rejected", func(s *Scenario) error {
			// Re-fetch ticket C ID by listing tickets for the project, but since we already
			// captured ticketAID and ticketBID implicitly, we just need ticketC. Re-seed
			// approach: capture all three inside the Given using the same closure variables.
			// ticketC was created in Given; capture its ID via the API list endpoint.
			resp, err := s.Harness().APIRequest(http.MethodGet,
				fmt.Sprintf("/projects/%s/tickets?limit=100", seed.ProjectID), nil)
			if err != nil {
				return fmt.Errorf("list tickets: %w", err)
			}
			defer resp.Body.Close()
			var list struct {
				Items []struct {
					ID    string `json:"id"`
					Title string `json:"title"`
				} `json:"items"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
				return fmt.Errorf("decode ticket list: %w", err)
			}
			var ticketBID, ticketCID string
			for _, item := range list.Items {
				if item.Title == ticketBTitle {
					ticketBID = item.ID
				}
				if strings.HasPrefix(item.Title, "Dep C ") {
					ticketCID = item.ID
				}
			}
			if ticketBID == "" {
				return fmt.Errorf("could not find ticket B by title")
			}
			if ticketCID == "" {
				return fmt.Errorf("could not find ticket C by title prefix")
			}

			if err := apiCreateDependencyExpect(s.Harness(), ticketAID, ticketBID, "blocks", http.StatusCreated, ""); err != nil {
				return err
			}
			if err := apiCreateDependencyExpect(s.Harness(), ticketBID, ticketCID, "blocks", http.StatusCreated, ""); err != nil {
				return err
			}
			return apiCreateDependencyExpect(s.Harness(), ticketCID, ticketAID, "blocks", http.StatusConflict, "dependency_cycle")
		}).
		WhenIClickRefresh().
		ThenISeeSelectorKey("board.ticket_blocked_badge").
		WhenIClickKey("board.filter_toggle_button").
		WhenIClickKey("board.filter_blocked_checkbox").
		Then("blocked filter hides unblocked ticket A and keeps blocked ticket B visible", func(s *Scenario) error {
			if err := s.Harness().ExpectTextHidden(ticketAKey); err != nil {
				return fmt.Errorf("expected unblocked ticket %q to be hidden: %w", ticketAKey, err)
			}
			if err := s.Harness().ExpectTextVisible(ticketBKey); err != nil {
				return fmt.Errorf("expected blocked ticket %q to be visible: %w", ticketBKey, err)
			}
			return nil
		}).
		When("I open a blocked ticket modal", func(s *Scenario) error {
			return s.Harness().page.GetByText(ticketBTitle).First().Click()
		}).
		ThenISeeSelectorKey("ticket.modal").
		ThenISeeSelectorKey("ticket.dependencies_section").
		AndISeeSelectorKey("ticket.dependency_graph").
		When("I close ticket modal", func(s *Scenario) error {
			return s.Harness().Click("[data-testid=\"ticket.close-button\"]")
		}).
		WhenIClickKey("nav.dashboard_tab").
		ThenURLContains("/projects/"+seed.ProjectID+"/dashboard").
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
