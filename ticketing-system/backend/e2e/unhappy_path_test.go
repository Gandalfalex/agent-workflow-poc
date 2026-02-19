//go:build e2e

package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/playwright-community/playwright-go"
	"ticketing-system/backend/internal/store"
)

func TestInvalidLoginShowsError(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		ThenISeeSelectorKey("login.view").
		WhenILogInAs("AdminUser", "wrongpassword").
		ThenISeeText("Invalid credentials.").
		ThenISeeSelectorKey("login.view")
}

func TestCreateTicketWithEmptyTitleButtonDisabled(t *testing.T) {
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
		// Empty title → create button disabled
		ThenButtonIsDisabledKey("new_ticket.create_button").
		// Fill whitespace only → still disabled
		WhenIFillKey("new_ticket.title_input", "   ").
		ThenButtonIsDisabledKey("new_ticket.create_button").
		// Fill a real title → enabled
		WhenIFillKey("new_ticket.title_input", "Real title").
		ThenButtonIsEnabledKey("new_ticket.create_button")
}

func TestCreateTicketWithoutStoryButtonDisabled(t *testing.T) {
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
		ThenISeeSelectorKey("board.add_ticket_button").
		// Press N to open new ticket modal (no pre-selected story)
		When("I press N to open new ticket modal", func(s *Scenario) error {
			return s.Harness().page.Keyboard().Press("n")
		}).
		ThenISeeSelectorKey("new_ticket.modal").
		WhenIFillKey("new_ticket.title_input", "Ticket without story").
		// No story selected → button disabled
		ThenButtonIsDisabledKey("new_ticket.create_button")
}

func TestEmptyCommentPostButtonDisabled(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()
	st := scenario.Harness().Store()
	ctx := scenario.Harness().Context()

	projectID := uuid.MustParse(seed.ProjectID)
	storyID := uuid.MustParse(seed.StoryID)
	backlogID := uuid.MustParse(seed.BacklogID)

	title := fmt.Sprintf("Comment Test Ticket %d", time.Now().UnixNano())
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
		// Empty comment → post button disabled
		ThenButtonIsDisabledKey("ticket.post_comment_button").
		// Whitespace only → still disabled
		WhenIFillKey("ticket.comment_input", "   ").
		ThenButtonIsDisabledKey("ticket.post_comment_button").
		// Real text → enabled
		WhenIFillKey("ticket.comment_input", "A real comment").
		ThenButtonIsEnabledKey("ticket.post_comment_button")
}

func TestDeleteTicketAndVerifyRemoval(t *testing.T) {
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
	keepTitle := fmt.Sprintf("Keep Ticket %d", ts)
	deleteTitle := fmt.Sprintf("Delete Ticket %d", ts)

	_, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
		Title:   keepTitle,
		Type:    "feature",
		StoryID: storyID,
		StateID: &backlogID,
	})
	if err != nil {
		t.Fatalf("seed keep ticket: %v", err)
	}
	_, err = st.CreateTicket(ctx, projectID, store.TicketCreateInput{
		Title:   deleteTitle,
		Type:    "bug",
		StoryID: storyID,
		StateID: &backlogID,
	})
	if err != nil {
		t.Fatalf("seed delete ticket: %v", err)
	}

	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/" + seed.ProjectID + "/board").
		WhenIClickRefresh().
		ThenISeeText(keepTitle).
		ThenISeeText(deleteTitle).
		// Open the ticket to delete
		WhenIClickTicketByText(deleteTitle).
		ThenISeeSelectorKey("ticket.modal").
		// Open kebab menu and click delete
		When("I click the kebab menu", func(s *Scenario) error {
			return s.Harness().Click("button[aria-label='Ticket actions']")
		}).
		WhenIAcceptNextDialog().
		WhenIClickKey("ticket.delete_button").
		// Deleted ticket gone, other remains
		ThenIDoNotSeeText(deleteTitle).
		ThenISeeText(keepTitle)
}

func TestDeleteStoryAndVerifyTicketsRemoved(t *testing.T) {
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

	// Create a story that will be deleted
	deleteStory, err := st.CreateStory(ctx, projectID, store.StoryCreateInput{
		Title: fmt.Sprintf("Delete Story %d", ts),
	})
	if err != nil {
		t.Fatalf("create story: %v", err)
	}

	// Ticket in the story to be deleted
	deletedTicketTitle := fmt.Sprintf("Doomed Ticket %d", ts)
	_, err = st.CreateTicket(ctx, projectID, store.TicketCreateInput{
		Title:   deletedTicketTitle,
		Type:    "feature",
		StoryID: deleteStory.ID,
		StateID: &backlogID,
	})
	if err != nil {
		t.Fatalf("seed doomed ticket: %v", err)
	}

	// Ticket in the default story (should survive)
	keepTicketTitle := fmt.Sprintf("Survivor Ticket %d", ts)
	_, err = st.CreateTicket(ctx, projectID, store.TicketCreateInput{
		Title:   keepTicketTitle,
		Type:    "feature",
		StoryID: storyID,
		StateID: &backlogID,
	})
	if err != nil {
		t.Fatalf("seed survivor ticket: %v", err)
	}

	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/" + seed.ProjectID + "/board").
		WhenIClickRefresh().
		ThenISeeText(deleteStory.Title).
		ThenISeeText(deletedTicketTitle).
		ThenISeeText(keepTicketTitle).
		// Click the kebab menu for the story to delete
		WhenIAcceptNextDialog().
		When("I click the story kebab menu and delete", func(s *Scenario) error {
			// Find the story cell that contains the story title, then find its kebab button
			storyCell := s.Harness().page.Locator("div").Filter(playwright.LocatorFilterOptions{
				HasText: deleteStory.Title,
			}).Locator("button[aria-label='Story actions']")
			if err := storyCell.First().Click(); err != nil {
				return fmt.Errorf("click story kebab for %q: %w", deleteStory.Title, err)
			}
			// Click "Delete story" button that appears
			return s.Harness().Click("button:has-text('Delete story')")
		}).
		// Story and its ticket gone, survivor ticket remains
		ThenIDoNotSeeText(deleteStory.Title).
		ThenIDoNotSeeText(deletedTicketTitle).
		ThenISeeText(keepTicketTitle).
		ThenISeeText("E2E Story")
}
