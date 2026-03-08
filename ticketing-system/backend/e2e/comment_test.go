//go:build e2e

package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"ticketing-system/backend/internal/store"
)

// seedCommentTicket creates a ticket directly via the store for comment tests.
func seedCommentTicket(s *Scenario, projectID, storyID, backlogID uuid.UUID, title string) error {
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

func TestMultipleCommentsOnTicketDisplayInOrder(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()

	projectID := uuid.MustParse(seed.ProjectID)
	storyID := uuid.MustParse(seed.StoryID)
	backlogID := uuid.MustParse(seed.BacklogID)

	ts := time.Now().UnixNano()
	title := fmt.Sprintf("Multi Comment Ticket %d", ts)
	comment1 := fmt.Sprintf("First comment %d", ts)
	comment2 := fmt.Sprintf("Second comment %d", ts)
	comment3 := fmt.Sprintf("Third comment %d", ts)

	scenario.
		Given("a ticket exists for multi-comment testing", func(s *Scenario) error {
			return seedCommentTicket(s, projectID, storyID, backlogID, title)
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
		WhenIFillKey("ticket.comment_input", comment1).
		WhenIClickKey("ticket.post_comment_button").
		ThenISeeText(comment1).
		WhenIFillKey("ticket.comment_input", comment2).
		WhenIClickKey("ticket.post_comment_button").
		ThenISeeText(comment2).
		WhenIFillKey("ticket.comment_input", comment3).
		WhenIClickKey("ticket.post_comment_button").
		ThenISeeText(comment3).
		AndISeeText(comment1).
		AndISeeText(comment2).
		AndISeeText(comment3)
}

func TestCommentWithSpecialCharacters(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()

	projectID := uuid.MustParse(seed.ProjectID)
	storyID := uuid.MustParse(seed.StoryID)
	backlogID := uuid.MustParse(seed.BacklogID)

	ts := time.Now().UnixNano()
	title := fmt.Sprintf("Special Chars Ticket %d", ts)
	markdownComment := "**bold text** and `inline code` and <script>alert('xss')</script>"

	scenario.
		Given("a ticket exists for special-character comment testing", func(s *Scenario) error {
			return seedCommentTicket(s, projectID, storyID, backlogID, title)
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
		WhenIFillKey("ticket.comment_input", markdownComment).
		WhenIClickKey("ticket.post_comment_button").
		ThenISeeText("bold text").
		AndISeeText("inline code").
		AndISeeSelectorKey("ticket.modal")
}
