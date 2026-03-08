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
		ThenURLContains("/projects/"+seed.ProjectID+"/board").
		WhenIClickRefresh().
		WhenIClickKey("board.add_ticket_button").
		ThenISeeSelectorKey("new_ticket.modal").
		ThenButtonIsDisabledKey("new_ticket.create_button").
		WhenIFillKey("new_ticket.title_input", "   ").
		ThenButtonIsDisabledKey("new_ticket.create_button").
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
		ThenURLContains("/projects/"+seed.ProjectID+"/board").
		WhenIClickRefresh().
		ThenISeeSelectorKey("board.add_ticket_button").
		When("I press N to open new ticket modal", func(s *Scenario) error {
			return s.Harness().page.Keyboard().Press("n")
		}).
		ThenISeeSelectorKey("new_ticket.modal").
		WhenIFillKey("new_ticket.title_input", "Ticket without story").
		ThenButtonIsDisabledKey("new_ticket.create_button")
}

func TestEmptyCommentPostButtonDisabled(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()
	var title string

	scenario.
		Given("a ticket exists in backlog", func(s *Scenario) error {
			st := s.Harness().Store()
			ctx := s.Harness().Context()
			projectID := uuid.MustParse(seed.ProjectID)
			storyID := uuid.MustParse(seed.StoryID)
			backlogID := uuid.MustParse(seed.BacklogID)

			title = fmt.Sprintf("Comment Test Ticket %d", time.Now().UnixNano())
			_, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
				Title:   title,
				Type:    "feature",
				StoryID: storyID,
				StateID: &backlogID,
			})
			if err != nil {
				return fmt.Errorf("seed ticket: %w", err)
			}
			return nil
		}).
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/"+seed.ProjectID+"/board").
		WhenIClickRefresh().
		ThenISeeText(title).
		WhenIClickTicketByText(title).
		ThenISeeSelectorKey("ticket.modal").
		ThenButtonIsDisabledKey("ticket.post_comment_button").
		WhenIFillKey("ticket.comment_input", "   ").
		ThenButtonIsDisabledKey("ticket.post_comment_button").
		WhenIFillKey("ticket.comment_input", "A real comment").
		ThenButtonIsEnabledKey("ticket.post_comment_button")
}

func TestDeleteTicketAndVerifyRemoval(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()
	var keepTitle, deleteTitle string

	scenario.
		Given("two tickets exist — one to keep and one to delete", func(s *Scenario) error {
			st := s.Harness().Store()
			ctx := s.Harness().Context()
			projectID := uuid.MustParse(seed.ProjectID)
			storyID := uuid.MustParse(seed.StoryID)
			backlogID := uuid.MustParse(seed.BacklogID)

			ts := time.Now().UnixNano()
			keepTitle = fmt.Sprintf("Keep Ticket %d", ts)
			deleteTitle = fmt.Sprintf("Delete Ticket %d", ts)

			if _, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
				Title:   keepTitle,
				Type:    "feature",
				StoryID: storyID,
				StateID: &backlogID,
			}); err != nil {
				return fmt.Errorf("seed keep ticket: %w", err)
			}
			if _, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
				Title:   deleteTitle,
				Type:    "bug",
				StoryID: storyID,
				StateID: &backlogID,
			}); err != nil {
				return fmt.Errorf("seed delete ticket: %w", err)
			}
			return nil
		}).
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/"+seed.ProjectID+"/board").
		WhenIClickRefresh().
		ThenISeeText(keepTitle).
		ThenISeeText(deleteTitle).
		WhenIClickTicketByText(deleteTitle).
		ThenISeeSelectorKey("ticket.modal").
		When("I click the kebab menu", func(s *Scenario) error {
			return s.Harness().Click("button[aria-label='Ticket actions']")
		}).
		WhenIAcceptNextDialog().
		WhenIClickKey("ticket.delete_button").
		ThenIDoNotSeeText(deleteTitle).
		ThenISeeText(keepTitle)
}

func TestDeleteStoryAndVerifyTicketsRemoved(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()

	var deleteStoryTitle, deletedTicketTitle, keepTicketTitle string

	scenario.
		Given("a story with one ticket and a separate survivor ticket exist", func(s *Scenario) error {
			st := s.Harness().Store()
			ctx := s.Harness().Context()
			projectID := uuid.MustParse(seed.ProjectID)
			storyID := uuid.MustParse(seed.StoryID)
			backlogID := uuid.MustParse(seed.BacklogID)

			ts := time.Now().UnixNano()

			deleteStory, err := st.CreateStory(ctx, projectID, store.StoryCreateInput{
				Title: fmt.Sprintf("Delete Story %d", ts),
			})
			if err != nil {
				return fmt.Errorf("create story: %w", err)
			}
			deleteStoryTitle = deleteStory.Title

			deletedTicketTitle = fmt.Sprintf("Doomed Ticket %d", ts)
			if _, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
				Title:   deletedTicketTitle,
				Type:    "feature",
				StoryID: deleteStory.ID,
				StateID: &backlogID,
			}); err != nil {
				return fmt.Errorf("seed doomed ticket: %w", err)
			}

			keepTicketTitle = fmt.Sprintf("Survivor Ticket %d", ts)
			if _, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
				Title:   keepTicketTitle,
				Type:    "feature",
				StoryID: storyID,
				StateID: &backlogID,
			}); err != nil {
				return fmt.Errorf("seed survivor ticket: %w", err)
			}
			return nil
		}).
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/"+seed.ProjectID+"/board").
		WhenIClickRefresh().
		ThenISeeText(deleteStoryTitle).
		ThenISeeText(deletedTicketTitle).
		ThenISeeText(keepTicketTitle).
		WhenIAcceptNextDialog().
		When("I click the story kebab menu and delete", func(s *Scenario) error {
			storyCell := s.Harness().page.Locator("div").Filter(playwright.LocatorFilterOptions{
				HasText: deleteStoryTitle,
			}).Locator("button[aria-label='Story actions']")
			if err := storyCell.First().Click(); err != nil {
				return fmt.Errorf("click story kebab for %q: %w", deleteStoryTitle, err)
			}
			return s.Harness().Click("button:has-text('Delete story')")
		}).
		ThenIDoNotSeeText(deleteStoryTitle).
		ThenIDoNotSeeText(deletedTicketTitle).
		ThenISeeText(keepTicketTitle).
		ThenISeeText("E2E Story")
}
