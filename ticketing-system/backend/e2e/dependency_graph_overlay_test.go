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
	st := scenario.Harness().Store()
	ctx := scenario.Harness().Context()
	projectID := uuid.MustParse(seed.ProjectID)
	storyID := uuid.MustParse(seed.StoryID)
	backlogID := uuid.MustParse(seed.BacklogID)

	ticketA, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
		Title:   fmt.Sprintf("Graph Source %d", time.Now().UnixNano()),
		Type:    "feature",
		StoryID: storyID,
		StateID: &backlogID,
	})
	if err != nil {
		t.Fatalf("seed ticket A: %v", err)
	}
	ticketB, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
		Title:   fmt.Sprintf("Graph Target %d", time.Now().UnixNano()),
		Type:    "feature",
		StoryID: storyID,
		StateID: &backlogID,
	})
	if err != nil {
		t.Fatalf("seed ticket B: %v", err)
	}

	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/"+seed.ProjectID+"/board").
		Then("dependency is created", func(s *Scenario) error {
			return apiCreateDependencyExpect(s.Harness(), ticketA.ID.String(), ticketB.ID.String(), "blocks", http.StatusCreated, "")
		}).
		WhenIClickRefresh().
		ThenISeeText(ticketA.Title).
		WhenIClickTicketByText(ticketA.Title).
		ThenISeeSelectorKey("ticket.modal").
		ThenISeeSelectorKey("ticket.dependency_graph_open_button").
		WhenIClickKey("ticket.dependency_graph_open_button").
		ThenISeeSelectorKey("ticket.dependency_graph_overlay").
		When("I click target ticket node in graph", func(s *Scenario) error {
			selector := fmt.Sprintf("[data-testid=\"ticket.dependency-graph-node-%s\"]", ticketB.ID.String())
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
		Then("ticket modal switches to dependency ticket", func(s *Scenario) error {
			input := s.Harness().page.Locator("[data-testid=\"ticket.title-input\"]")
			if err := input.WaitFor(); err != nil {
				return err
			}
			value, err := input.InputValue()
			if err != nil {
				return err
			}
			if value != ticketB.Title {
				return fmt.Errorf("expected title input %q, got %q", ticketB.Title, value)
			}
			return nil
		})
}
