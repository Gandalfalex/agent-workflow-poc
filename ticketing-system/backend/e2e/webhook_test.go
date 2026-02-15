//go:build e2e

package e2e

import (
	"fmt"
	"testing"
	"time"
)

func TestWebhookFiresOnTicketCreation(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t, WithWebhookCapture())
	defer scenario.Close()

	seed := scenario.SeedData()
	capture := scenario.Harness().WebhookCapture()

	title := fmt.Sprintf("Webhook Create Ticket %d", time.Now().UnixNano())

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
		WhenIClickKey("new_ticket.create_button").
		ThenISeeText(title).
		Then("webhook captured ticket.created event", func(s *Scenario) error {
			if !capture.WaitForEvent("ticket.created", 5*time.Second) {
				return fmt.Errorf("expected ticket.created webhook event, got: %v", capture.Events())
			}
			return nil
		})
}

func TestWebhookFiresOnTicketStateChange(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t, WithWebhookCapture())
	defer scenario.Close()

	seed := scenario.SeedData()
	capture := scenario.Harness().WebhookCapture()

	title := fmt.Sprintf("Webhook State Ticket %d", time.Now().UnixNano())

	// Create ticket via UI first
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
		WhenIClickKey("new_ticket.create_button").
		ThenISeeText(title)

	// Reset capture to only track state change
	capture.Reset()

	// Open ticket and change state
	scenario.
		WhenIClickTicketByText(title).
		ThenISeeSelectorKey("ticket.modal").
		WhenISelectOptionByValueKey("ticket.state_select", seed.InProgressID).
		WhenIClickKey("ticket.save_button").
		ThenISeeText(title).
		Then("webhook captured ticket.state_changed event", func(s *Scenario) error {
			if !capture.WaitForEvent("ticket.state_changed", 5*time.Second) {
				events := capture.Events()
				return fmt.Errorf("expected ticket.state_changed webhook event, got: %v", events)
			}
			return nil
		})
}
