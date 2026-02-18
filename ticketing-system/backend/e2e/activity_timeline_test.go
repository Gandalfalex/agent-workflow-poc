//go:build e2e

package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"ticketing-system/backend/internal/store"
)

func TestActivityTimelineShowsStateChange(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()
	st := scenario.Harness().Store()
	ctx := scenario.Harness().Context()

	projectID := uuid.MustParse(seed.ProjectID)
	storyID := uuid.MustParse(seed.StoryID)
	backlogID := uuid.MustParse(seed.BacklogID)

	ts := time.Now().UnixNano()
	title := fmt.Sprintf("Activity Timeline Test %d", ts)

	_, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
		Title:   title,
		Type:    "feature",
		StoryID: storyID,
		StateID: &backlogID,
	})
	if err != nil {
		t.Fatalf("seed ticket: %v", err)
	}

	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/"+seed.ProjectID+"/board").
		WhenIClickRefresh().
		ThenISeeText(title).
		WhenIClickTicketByText(title).
		ThenISeeSelectorKey("ticket.modal").
		// Change state from Backlog to In Progress
		WhenISelectOptionByValueKey("ticket.state_select", seed.InProgressID).
		WhenIClickKey("ticket.save_button").
		// Reopen the ticket to see the activity timeline
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
	st := scenario.Harness().Store()
	ctx := scenario.Harness().Context()

	projectID := uuid.MustParse(seed.ProjectID)
	storyID := uuid.MustParse(seed.StoryID)
	backlogID := uuid.MustParse(seed.BacklogID)

	ts := time.Now().UnixNano()
	title := fmt.Sprintf("Priority Activity Test %d", ts)

	_, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
		Title:   title,
		Type:    "feature",
		StoryID: storyID,
		StateID: &backlogID,
	})
	if err != nil {
		t.Fatalf("seed ticket: %v", err)
	}

	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/"+seed.ProjectID+"/board").
		WhenIClickRefresh().
		ThenISeeText(title).
		WhenIClickTicketByText(title).
		ThenISeeSelectorKey("ticket.modal").
		// Change priority to urgent
		WhenISelectOptionByValueKey("ticket.priority_select", "urgent").
		WhenIClickKey("ticket.save_button").
		// Reopen to check activity
		WhenIClickRefresh().
		ThenISeeText(title).
		WhenIClickTicketByText(title).
		ThenISeeSelectorKey("ticket.modal").
		ThenISeeSelectorKey("ticket.activity_timeline").
		ThenISeeText("changed priority from")
}
