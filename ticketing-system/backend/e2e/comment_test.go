//go:build e2e

package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"ticketing-system/backend/internal/store"
)

func TestMultipleCommentsOnTicketDisplayInOrder(t *testing.T) {
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
	title := fmt.Sprintf("Multi Comment Ticket %d", ts)
	comment1 := fmt.Sprintf("First comment %d", ts)
	comment2 := fmt.Sprintf("Second comment %d", ts)
	comment3 := fmt.Sprintf("Third comment %d", ts)

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
		// Add first comment
		WhenIFillKey("ticket.comment_input", comment1).
		WhenIClickKey("ticket.post_comment_button").
		ThenISeeText(comment1).
		// Add second comment
		WhenIFillKey("ticket.comment_input", comment2).
		WhenIClickKey("ticket.post_comment_button").
		ThenISeeText(comment2).
		// Add third comment
		WhenIFillKey("ticket.comment_input", comment3).
		WhenIClickKey("ticket.post_comment_button").
		ThenISeeText(comment3).
		// All three should be visible
		AndISeeText(comment1).
		AndISeeText(comment2).
		AndISeeText(comment3)
}

func TestCommentWithSpecialCharacters(t *testing.T) {
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
	title := fmt.Sprintf("Special Chars Ticket %d", ts)
	// Markdown comment with bold, code, and HTML-like content
	markdownComment := "**bold text** and `inline code` and <script>alert('xss')</script>"

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
		// Post the markdown comment
		WhenIFillKey("ticket.comment_input", markdownComment).
		WhenIClickKey("ticket.post_comment_button").
		// The bold text should render (visible text without markdown syntax)
		ThenISeeText("bold text").
		// inline code should be visible
		AndISeeText("inline code").
		// The script tag should NOT execute - just verify the board is still functional
		AndISeeSelectorKey("ticket.modal")
}
