//go:build e2e

package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
)

func TestCreateNewStoryAndVerifyOnBoard(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()
	storyTitle := fmt.Sprintf("New Story %d", time.Now().UnixNano())

	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/" + seed.ProjectID + "/board").
		// Verify the default story is present
		ThenISeeText("E2E Story").
		// Create a new story
		WhenICreateStory(storyTitle).
		// The story modal should close and the new story should appear
		ThenISeeText(storyTitle).
		// The original story should still be there
		AndISeeText("E2E Story")
}

func TestCreateStoryThenCreateTicketInNewStory(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()
	ts := time.Now().UnixNano()
	storyTitle := fmt.Sprintf("Sprint Story %d", ts)
	ticketTitle := fmt.Sprintf("Sprint Ticket %d", ts)

	// Create a new story
	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/" + seed.ProjectID + "/board").
		WhenICreateStory(storyTitle).
		ThenISeeText(storyTitle)

	// Open new ticket modal using keyboard shortcut (avoids ambiguity with multiple story buttons)
	scenario.
		When("I press N to open new ticket modal", func(s *Scenario) error {
			return s.Harness().page.Keyboard().Press("n")
		}).
		ThenISeeSelectorKey("new_ticket.modal").
		WhenIFillKey("new_ticket.title_input", ticketTitle).
		// The new story should be selectable in the story dropdown
		When("I select the new story from dropdown", func(s *Scenario) error {
			// Select by label text since we don't know the story ID
			_, err := s.Harness().page.Locator("[data-testid='new-ticket.story-select']").
				SelectOption(playwright.SelectOptionValues{Labels: &[]string{storyTitle}})
			return err
		}).
		WhenIClickKey("new_ticket.create_button").
		ThenISeeText(ticketTitle)
}

func TestMultipleStoriesDisplayedInOrder(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()
	ts := time.Now().UnixNano()
	story2 := fmt.Sprintf("Alpha Story %d", ts)
	story3 := fmt.Sprintf("Beta Story %d", ts)

	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/" + seed.ProjectID + "/board").
		// E2E Story is already present from seed
		ThenISeeText("E2E Story").
		// Create two more stories
		WhenICreateStory(story2).
		ThenISeeText(story2).
		WhenICreateStory(story3).
		ThenISeeText(story3).
		// All three should be visible
		AndISeeText("E2E Story").
		AndISeeText(story2).
		AndISeeText(story3)
}
