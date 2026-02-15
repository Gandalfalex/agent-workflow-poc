//go:build e2e

package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"ticketing-system/backend/internal/store"
)

func TestOpenTicketAndEditTitle(t *testing.T) {
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
	originalTitle := fmt.Sprintf("Original Title %d", ts)
	updatedTitle := fmt.Sprintf("Updated Title %d", ts)

	// Pre-seed a ticket
	_, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
		Title:   originalTitle,
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
		ThenURLContains("/projects/" + seed.ProjectID + "/board").
		WhenIClickRefresh().
		ThenISeeText(originalTitle).
		// Click the ticket to open the edit modal
		WhenIClickTicketByText(originalTitle).
		ThenISeeSelectorKey("ticket.modal").
		// Clear and update the title
		When("I clear and type new title", func(s *Scenario) error {
			sel, err := s.Harness().Selector("ticket.title_input")
			if err != nil {
				return err
			}
			if err := s.Harness().WaitVisible(sel); err != nil {
				return err
			}
			if err := s.Harness().page.Locator(sel).Fill(updatedTitle); err != nil {
				return fmt.Errorf("fill new title: %w", err)
			}
			return nil
		}).
		WhenIClickKey("ticket.save_button").
		// The modal should close and the updated title should be visible
		ThenISeeText(updatedTitle)
}

func TestOpenTicketAndChangeState(t *testing.T) {
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
	title := fmt.Sprintf("State Change Ticket %d", ts)

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
		ThenURLContains("/projects/" + seed.ProjectID + "/board").
		WhenIClickRefresh().
		ThenISeeText(title).
		WhenIClickTicketByText(title).
		ThenISeeSelectorKey("ticket.modal").
		// Change state to "In Progress"
		WhenISelectOptionByValueKey("ticket.state_select", seed.InProgressID).
		WhenIClickKey("ticket.save_button").
		// Ticket should still be visible on the board
		ThenISeeText(title)
}

func TestOpenTicketAndChangePriority(t *testing.T) {
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
	title := fmt.Sprintf("Priority Change Ticket %d", ts)

	_, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
		Title:    title,
		Type:     "feature",
		StoryID:  storyID,
		StateID:  &backlogID,
		Priority: "low",
	})
	if err != nil {
		t.Fatalf("seed ticket: %v", err)
	}

	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/" + seed.ProjectID + "/board").
		WhenIClickRefresh().
		ThenISeeText(title).
		WhenIClickTicketByText(title).
		ThenISeeSelectorKey("ticket.modal").
		// Change priority to urgent
		WhenISelectOptionByValueKey("ticket.priority_select", "urgent").
		WhenIClickKey("ticket.save_button").
		// Ticket should still be visible
		ThenISeeText(title)
}

func TestOpenTicketAndChangeType(t *testing.T) {
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
	title := fmt.Sprintf("Type Change Ticket %d", ts)

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
		ThenURLContains("/projects/" + seed.ProjectID + "/board").
		WhenIClickRefresh().
		ThenISeeText(title).
		WhenIClickTicketByText(title).
		ThenISeeSelectorKey("ticket.modal").
		// Change type to bug
		WhenISelectOptionByValueKey("ticket.type_select", "bug").
		WhenIClickKey("ticket.save_button").
		ThenISeeText(title)
}

func TestOpenTicketAndAddComment(t *testing.T) {
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
	title := fmt.Sprintf("Commentable Ticket %d", ts)
	commentText := fmt.Sprintf("This is a test comment %d", ts)

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
		ThenURLContains("/projects/" + seed.ProjectID + "/board").
		WhenIClickRefresh().
		ThenISeeText(title).
		WhenIClickTicketByText(title).
		ThenISeeSelectorKey("ticket.modal").
		// Type a comment
		WhenIFillKey("ticket.comment_input", commentText).
		WhenIClickKey("ticket.post_comment_button").
		// The comment should appear in the modal
		ThenISeeText(commentText)
}

func TestOpenTicketEditMultipleFieldsAndSave(t *testing.T) {
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
	originalTitle := fmt.Sprintf("Multi Edit Original %d", ts)
	updatedTitle := fmt.Sprintf("Multi Edit Updated %d", ts)

	_, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
		Title:    originalTitle,
		Type:     "feature",
		StoryID:  storyID,
		StateID:  &backlogID,
		Priority: "low",
	})
	if err != nil {
		t.Fatalf("seed ticket: %v", err)
	}

	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/" + seed.ProjectID + "/board").
		WhenIClickRefresh().
		ThenISeeText(originalTitle).
		WhenIClickTicketByText(originalTitle).
		ThenISeeSelectorKey("ticket.modal").
		// Edit title
		When("I update the title", func(s *Scenario) error {
			sel, err := s.Harness().Selector("ticket.title_input")
			if err != nil {
				return err
			}
			return s.Harness().page.Locator(sel).Fill(updatedTitle)
		}).
		// Change type to bug
		WhenISelectOptionByValueKey("ticket.type_select", "bug").
		// Change priority to urgent
		WhenISelectOptionByValueKey("ticket.priority_select", "urgent").
		// Change state to In Progress
		WhenISelectOptionByValueKey("ticket.state_select", seed.InProgressID).
		// Save all changes
		WhenIClickKey("ticket.save_button").
		ThenISeeText(updatedTitle)
}
