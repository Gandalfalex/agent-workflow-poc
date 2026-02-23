//go:build e2e

package e2e

import (
	"fmt"
	"testing"
	"time"
)

func TestCreateBugTicketWithUrgentPriority(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()
	title := fmt.Sprintf("Urgent Bug %d", time.Now().UnixNano())

	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/" + seed.ProjectID + "/board").
		WhenIClickRefresh().
		WhenIClickKey("board.add_ticket_button").
		ThenISeeSelectorKey("new_ticket.modal").
		WhenIFillKey("new_ticket.title_input", title).
		WhenISelectOptionByValueKey("new_ticket.type_select", "bug").
		WhenISelectOptionByValueKey("new_ticket.priority_select", "urgent").
		WhenIClickKey("new_ticket.create_button").
		ThenISeeText(title)
}

func TestCreateTicketWithDescription(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()
	title := fmt.Sprintf("Described Ticket %d", time.Now().UnixNano())
	description := "This ticket has a detailed description for testing purposes."

	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/" + seed.ProjectID + "/board").
		WhenIClickRefresh().
		WhenIClickKey("board.add_ticket_button").
		ThenISeeSelectorKey("new_ticket.modal").
		WhenIFillKey("new_ticket.title_input", title).
		WhenIFillKey("new_ticket.description_input", description).
		WhenIClickKey("new_ticket.create_button").
		ThenISeeText(title)
}

func TestCreateMultipleTicketsSequentially(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()
	ts := time.Now().UnixNano()
	ticket1 := fmt.Sprintf("First Ticket %d", ts)
	ticket2 := fmt.Sprintf("Second Ticket %d", ts)
	ticket3 := fmt.Sprintf("Third Ticket %d", ts)

	// Create ticket 1
	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/" + seed.ProjectID + "/board").
		WhenIClickRefresh().
		WhenIClickKey("board.add_ticket_button").
		ThenISeeSelectorKey("new_ticket.modal").
		WhenIFillKey("new_ticket.title_input", ticket1).
		WhenIClickKey("new_ticket.create_button").
		ThenISeeText(ticket1)

	// Create ticket 2
	scenario.
		WhenIClickKey("board.add_ticket_button").
		ThenISeeSelectorKey("new_ticket.modal").
		WhenIFillKey("new_ticket.title_input", ticket2).
		WhenISelectOptionByValueKey("new_ticket.priority_select", "high").
		WhenIClickKey("new_ticket.create_button").
		ThenISeeText(ticket2)

	// Create ticket 3
	scenario.
		WhenIClickKey("board.add_ticket_button").
		ThenISeeSelectorKey("new_ticket.modal").
		WhenIFillKey("new_ticket.title_input", ticket3).
		WhenISelectOptionByValueKey("new_ticket.type_select", "bug").
		WhenIClickKey("new_ticket.create_button").
		ThenISeeText(ticket3)

	// All three should be visible on the board
	scenario.
		AndISeeText(ticket1).
		AndISeeText(ticket2).
		AndISeeText(ticket3)
}

func TestCreateTicketInSpecificState(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()
	title := fmt.Sprintf("In Progress Ticket %d", time.Now().UnixNano())

	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/" + seed.ProjectID + "/board").
		WhenIClickRefresh().
		WhenIClickKey("board.add_ticket_button").
		ThenISeeSelectorKey("new_ticket.modal").
		WhenIFillKey("new_ticket.title_input", title).
		WhenISelectOptionByValueKey("new_ticket.state_select", seed.InProgressID).
		WhenIClickKey("new_ticket.create_button").
		ThenISeeText(title)
}

func TestCreateTicketModalClosesOnCancel(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()

	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/" + seed.ProjectID + "/board").
		WhenIClickRefresh().
		WhenIClickKey("board.add_ticket_button").
		ThenISeeSelectorKey("new_ticket.modal").
		When("I click Cancel in new ticket modal", func(s *Scenario) error {
			return s.Harness().Click("[data-testid=\"new-ticket.cancel-button\"]")
		}).
		ThenIDoNotSeeSelectorKey("new_ticket.modal")
}
