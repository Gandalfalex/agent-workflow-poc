//go:build e2e

package e2e

import (
	"fmt"
	"testing"
	"time"
)

func TestLoginSelectProjectAndCreateTicket(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()
	ticketTitle := fmt.Sprintf("E2E Ticket %d", time.Now().UnixNano())

	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		ThenISeeSelectorKey("login.view").
		WhenILogInAs("AdminUser", "admin123").
		ThenISeeSelectorKey("nav.project_select").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/"+seed.ProjectID+"/board").
		AndISeeSelectorKey("board.add_ticket_button").
		WhenIClickKey("board.add_ticket_button").
		ThenISeeSelectorKey("new_ticket.modal").
		WhenIFillKey("new_ticket.title_input", ticketTitle).
		WhenIClickKey("new_ticket.create_button").
		ThenISeeText(ticketTitle)
}
