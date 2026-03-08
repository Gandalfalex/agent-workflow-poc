//go:build e2e

package e2e

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"ticketing-system/backend/internal/store"
)

func TestDependencyGraphOverlayNodeClickOpensTicket(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()

	var ticketATitle, ticketAID string
	var ticketBTitle, ticketBID string

	scenario.
		Given("two tickets A and B are seeded in backlog", func(s *Scenario) error {
			st := s.Harness().Store()
			ctx := s.Harness().Context()
			projectID := uuid.MustParse(seed.ProjectID)
			storyID := uuid.MustParse(seed.StoryID)
			backlogID := uuid.MustParse(seed.BacklogID)

			ts := time.Now().UnixNano()
			ticketA, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
				Title:   fmt.Sprintf("Graph Source %d", ts),
				Type:    "feature",
				StoryID: storyID,
				StateID: &backlogID,
			})
			if err != nil {
				return fmt.Errorf("seed ticket A: %w", err)
			}
			ticketATitle = ticketA.Title
			ticketAID = ticketA.ID.String()

			ticketB, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
				Title:   fmt.Sprintf("Graph Target %d", ts),
				Type:    "feature",
				StoryID: storyID,
				StateID: &backlogID,
			})
			if err != nil {
				return fmt.Errorf("seed ticket B: %w", err)
			}
			ticketBTitle = ticketB.Title
			ticketBID = ticketB.ID.String()
			return nil
		}).
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/"+seed.ProjectID+"/board").
		Then("dependency A→B is created via the API", func(s *Scenario) error {
			return apiCreateDependencyExpect(s.Harness(), ticketAID, ticketBID, "blocks", http.StatusCreated, "")
		}).
		WhenIClickRefresh().
		ThenISeeText(ticketATitle).
		WhenIClickTicketByText(ticketATitle).
		ThenISeeSelectorKey("ticket.modal").
		ThenISeeSelectorKey("ticket.dependency_graph_open_button").
		WhenIClickKey("ticket.dependency_graph_open_button").
		ThenISeeSelectorKey("ticket.dependency_graph_overlay").
		When("I click the target ticket node in the dependency graph", func(s *Scenario) error {
			selector := fmt.Sprintf("[data-testid=\"ticket.dependency-graph-node-%s\"]", ticketBID)
			result, err := s.Harness().page.Evaluate(`(sel) => {
				const node = document.querySelector(sel);
				if (!node) return false;
				node.dispatchEvent(new MouseEvent("click", { bubbles: true, cancelable: true }));
				return true;
			}`, selector)
			if err != nil {
				return err
			}
			ok, _ := result.(bool)
			if !ok {
				return fmt.Errorf("dependency graph node not found: %s", selector)
			}
			return nil
		}).
		Then("the ticket modal switches to show the dependency ticket", func(s *Scenario) error {
			input := s.Harness().page.Locator("[data-testid=\"ticket.title-input\"]")
			if err := input.WaitFor(); err != nil {
				return err
			}
			value, err := input.InputValue()
			if err != nil {
				return err
			}
			if value != ticketBTitle {
				return fmt.Errorf("expected title input %q, got %q", ticketBTitle, value)
			}
			return nil
		})
}
