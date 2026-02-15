//go:build e2e

package e2e

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"ticketing-system/backend/internal/store"
)

func TestTicketStateChangeViaAPIReflectsOnBoard(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()
	st := scenario.Harness().Store()
	ctx := scenario.Harness().Context()

	projectID := uuid.MustParse(seed.ProjectID)
	storyID := uuid.MustParse(seed.StoryID)
	backlogID := uuid.MustParse(seed.BacklogID)
	inProgressID := uuid.MustParse(seed.InProgressID)

	title := fmt.Sprintf("API State Change %d", time.Now().UnixNano())

	ticket, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
		Title:   title,
		Type:    "feature",
		StoryID: storyID,
		StateID: &backlogID,
	})
	if err != nil {
		t.Fatalf("seed ticket: %v", err)
	}

	// Move ticket via store API (not UI)
	_, err = st.UpdateTicket(ctx, ticket.ID, store.TicketUpdateInput{
		StateID: &inProgressID,
	})
	if err != nil {
		t.Fatalf("update ticket state: %v", err)
	}

	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/" + seed.ProjectID + "/board").
		WhenIClickRefresh().
		ThenISeeText(title).
		// Open the ticket and verify state shows In Progress
		WhenIClickTicketByText(title).
		ThenISeeSelectorKey("ticket.modal").
		Then("ticket state select shows In Progress", func(s *Scenario) error {
			sel, err := s.Harness().Selector("ticket.state_select")
			if err != nil {
				return err
			}
			value, err := s.Harness().page.Locator(sel).InputValue()
			if err != nil {
				return fmt.Errorf("get state select value: %w", err)
			}
			if value != seed.InProgressID {
				return fmt.Errorf("expected state %q, got %q", seed.InProgressID, value)
			}
			return nil
		})
}

func TestStoryWithZeroTicketsDisplaysCorrectly(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()
	st := scenario.Harness().Store()
	ctx := scenario.Harness().Context()

	projectID := uuid.MustParse(seed.ProjectID)

	emptyStoryTitle := fmt.Sprintf("Empty Story %d", time.Now().UnixNano())

	_, err := st.CreateStory(ctx, projectID, store.StoryCreateInput{
		Title: emptyStoryTitle,
	})
	if err != nil {
		t.Fatalf("create empty story: %v", err)
	}

	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/" + seed.ProjectID + "/board").
		WhenIClickRefresh().
		ThenISeeText(emptyStoryTitle).
		// Board is still functional - add ticket button visible
		AndISeeSelectorKey("board.add_ticket_button")
}

func TestBoardHandlesLargeNumberOfTickets(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()
	st := scenario.Harness().Store()
	ctx := scenario.Harness().Context()

	projectID := uuid.MustParse(seed.ProjectID)
	storyID := uuid.MustParse(seed.StoryID)
	backlogID := uuid.MustParse(seed.BacklogID)
	inProgressID := uuid.MustParse(seed.InProgressID)
	doneID := uuid.MustParse(seed.DoneID)

	ts := time.Now().UnixNano()
	states := []*uuid.UUID{&backlogID, &inProgressID, &doneID}

	// Pre-seed 25 tickets across 3 states
	var sampleTitles []string
	for i := 0; i < 25; i++ {
		title := fmt.Sprintf("Bulk Ticket %d-%d", ts, i)
		stateIdx := i % 3
		_, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
			Title:   title,
			Type:    "feature",
			StoryID: storyID,
			StateID: states[stateIdx],
		})
		if err != nil {
			t.Fatalf("seed ticket %d: %v", i, err)
		}
		// Save a few sample titles to verify
		if i == 0 || i == 12 || i == 24 {
			sampleTitles = append(sampleTitles, title)
		}
	}

	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/" + seed.ProjectID + "/board").
		WhenIClickRefresh().
		WhenIClickRefresh()

	// Verify sample tickets visible
	for _, title := range sampleTitles {
		scenario.ThenISeeText(title)
	}

	// Board still functional
	scenario.AndISeeSelectorKey("board.add_ticket_button")
}

func TestTicketWithLongTitleAndDescription(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()
	st := scenario.Harness().Store()
	ctx := scenario.Harness().Context()

	projectID := uuid.MustParse(seed.ProjectID)
	storyID := uuid.MustParse(seed.StoryID)
	backlogID := uuid.MustParse(seed.BacklogID)

	// 200-character title
	longTitle := fmt.Sprintf("LongTitle-%d-%s", time.Now().UnixNano(), strings.Repeat("x", 180))
	if len(longTitle) > 200 {
		longTitle = longTitle[:200]
	}

	_, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
		Title:       longTitle,
		Description: strings.Repeat("This is a very long description. ", 20),
		Type:        "feature",
		StoryID:     storyID,
		StateID:     &backlogID,
	})
	if err != nil {
		t.Fatalf("seed ticket: %v", err)
	}

	// Use a prefix of the title for text matching (board may truncate)
	titlePrefix := longTitle[:30]

	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/" + seed.ProjectID + "/board").
		WhenIClickRefresh().
		ThenISeeText(titlePrefix).
		// Board still functional
		AndISeeSelectorKey("board.add_ticket_button")
}
