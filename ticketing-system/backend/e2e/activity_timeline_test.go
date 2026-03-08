//go:build e2e

package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"ticketing-system/backend/internal/store"
)

// seedActivityTicket creates a ticket directly via the store for activity timeline tests.
func seedActivityTicket(s *Scenario, projectID, storyID, backlogID uuid.UUID, title string) error {
	_, err := s.Harness().Store().CreateTicket(s.Harness().Context(), projectID, store.TicketCreateInput{
		Title:   title,
		Type:    "feature",
		StoryID: storyID,
		StateID: &backlogID,
	})
	if err != nil {
		return fmt.Errorf("seed ticket: %w", err)
	}
	return nil
}

func TestActivityTimelineShowsStateChange(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()

	projectID := uuid.MustParse(seed.ProjectID)
	storyID := uuid.MustParse(seed.StoryID)
	backlogID := uuid.MustParse(seed.BacklogID)

	ts := time.Now().UnixNano()
	title := fmt.Sprintf("Activity Timeline Test %d", ts)

	scenario.
		Given("a ticket exists for activity timeline testing", func(s *Scenario) error {
			return seedActivityTicket(s, projectID, storyID, backlogID, title)
		}).
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/" + seed.ProjectID + "/board").
		WhenIClickRefresh().
		ThenISeeText(title).
		WhenIClickTicketByText(title).
		ThenISeeSelectorKey("ticket.modal").
		WhenISelectOptionByValueKey("ticket.state_select", seed.InProgressID).
		WhenIClickKey("ticket.save_button").
		WhenIClickRefresh().
		ThenISeeText(title).
		WhenIClickTicketByText(title).
		ThenISeeSelectorKey("ticket.modal").
		ThenISeeSelectorKey("ticket.activity_timeline").
		ThenISeeText("changed state from").
		ThenISeeSelectorKey("ticket.activity_item")
}

func TestActivityTimelineShowsPriorityChange(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()

	projectID := uuid.MustParse(seed.ProjectID)
	storyID := uuid.MustParse(seed.StoryID)
	backlogID := uuid.MustParse(seed.BacklogID)

	ts := time.Now().UnixNano()
	title := fmt.Sprintf("Priority Activity Test %d", ts)

	scenario.
		Given("a ticket exists for priority activity testing", func(s *Scenario) error {
			return seedActivityTicket(s, projectID, storyID, backlogID, title)
		}).
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/" + seed.ProjectID + "/board").
		WhenIClickRefresh().
		ThenISeeText(title).
		WhenIClickTicketByText(title).
		ThenISeeSelectorKey("ticket.modal").
		WhenISelectOptionByValueKey("ticket.priority_select", "urgent").
		WhenIClickKey("ticket.save_button").
		WhenIClickRefresh().
		ThenISeeText(title).
		WhenIClickTicketByText(title).
		ThenISeeSelectorKey("ticket.modal").
		ThenISeeSelectorKey("ticket.activity_timeline").
		ThenISeeText("changed priority from")
}
