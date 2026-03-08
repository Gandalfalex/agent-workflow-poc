//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

	var title string

	scenario.
		Given("a ticket exists and has been moved to In Progress via the store", func(s *Scenario) error {
			st := s.Harness().Store()
			ctx := s.Harness().Context()
			projectID := uuid.MustParse(seed.ProjectID)
			storyID := uuid.MustParse(seed.StoryID)
			backlogID := uuid.MustParse(seed.BacklogID)
			inProgressID := uuid.MustParse(seed.InProgressID)

			title = fmt.Sprintf("API State Change %d", time.Now().UnixNano())
			ticket, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
				Title:   title,
				Type:    "feature",
				StoryID: storyID,
				StateID: &backlogID,
			})
			if err != nil {
				return fmt.Errorf("seed ticket: %w", err)
			}
			_, err = st.UpdateTicket(ctx, ticket.ID, store.TicketUpdateInput{
				StateID: &inProgressID,
			})
			if err != nil {
				return fmt.Errorf("update ticket state: %w", err)
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

func TestBoardAutoRefreshesOnLiveEventWithoutManualRefresh(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()
	var createdTitle string

	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/"+seed.ProjectID+"/board").
		When("a ticket is created via API while the board is open", func(s *Scenario) error {
			title, err := apiCreateTicketWithTitle(s.Harness(), seed.ProjectID, seed.StoryID)
			if err != nil {
				return err
			}
			createdTitle = title
			return nil
		}).
		Then("the new ticket appears on the board without clicking refresh", func(s *Scenario) error {
			if strings.TrimSpace(createdTitle) == "" {
				return fmt.Errorf("created title is empty")
			}
			return s.Harness().ExpectTextVisible(createdTitle)
		})
}

func apiCreateTicketWithTitle(h *Harness, projectID, storyID string) (string, error) {
	title := fmt.Sprintf("Live Event Ticket %d", time.Now().UnixNano())
	payload := map[string]any{
		"title":   title,
		"storyId": storyID,
		"type":    "feature",
	}
	raw, _ := json.Marshal(payload)
	resp, err := h.APIRequest(http.MethodPost, fmt.Sprintf("/projects/%s/tickets", projectID), bytes.NewReader(raw))
	if err != nil {
		return "", fmt.Errorf("create ticket request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("create ticket status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return title, nil
}

func TestStoryWithZeroTicketsDisplaysCorrectly(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()
	var emptyStoryTitle string

	scenario.
		Given("an empty story with no tickets is created via the store", func(s *Scenario) error {
			st := s.Harness().Store()
			ctx := s.Harness().Context()
			projectID := uuid.MustParse(seed.ProjectID)

			emptyStoryTitle = fmt.Sprintf("Empty Story %d", time.Now().UnixNano())
			_, err := st.CreateStory(ctx, projectID, store.StoryCreateInput{
				Title: emptyStoryTitle,
			})
			if err != nil {
				return fmt.Errorf("create empty story: %w", err)
			}
			return nil
		}).
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/"+seed.ProjectID+"/board").
		WhenIClickRefresh().
		ThenISeeText(emptyStoryTitle).
		AndISeeSelectorKey("board.add_ticket_button")
}

func TestBoardHandlesLargeNumberOfTickets(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()
	var sampleTitles []string

	scenario.
		Given("25 tickets are pre-seeded across 3 states", func(s *Scenario) error {
			st := s.Harness().Store()
			ctx := s.Harness().Context()
			projectID := uuid.MustParse(seed.ProjectID)
			storyID := uuid.MustParse(seed.StoryID)
			backlogID := uuid.MustParse(seed.BacklogID)
			inProgressID := uuid.MustParse(seed.InProgressID)
			doneID := uuid.MustParse(seed.DoneID)

			ts := time.Now().UnixNano()
			states := []*uuid.UUID{&backlogID, &inProgressID, &doneID}

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
					return fmt.Errorf("seed ticket %d: %w", i, err)
				}
				if i == 0 || i == 12 || i == 24 {
					sampleTitles = append(sampleTitles, title)
				}
			}
			return nil
		}).
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/"+seed.ProjectID+"/board").
		WhenIClickRefresh().
		WhenIClickRefresh().
		Then("all sample tickets are visible on the board", func(s *Scenario) error {
			for _, title := range sampleTitles {
				if err := s.Harness().ExpectTextVisible(title); err != nil {
					return fmt.Errorf("expected ticket %q to be visible: %w", title, err)
				}
			}
			return nil
		}).
		AndISeeSelectorKey("board.add_ticket_button")
}

func TestTicketWithLongTitleAndDescription(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()
	var titlePrefix string

	scenario.
		Given("a ticket with a 200-character title and long description is created via the store", func(s *Scenario) error {
			st := s.Harness().Store()
			ctx := s.Harness().Context()
			projectID := uuid.MustParse(seed.ProjectID)
			storyID := uuid.MustParse(seed.StoryID)
			backlogID := uuid.MustParse(seed.BacklogID)

			longTitle := fmt.Sprintf("LongTitle-%d-%s", time.Now().UnixNano(), strings.Repeat("x", 180))
			if len(longTitle) > 200 {
				longTitle = longTitle[:200]
			}
			titlePrefix = longTitle[:30]

			_, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
				Title:       longTitle,
				Description: strings.Repeat("This is a very long description. ", 20),
				Type:        "feature",
				StoryID:     storyID,
				StateID:     &backlogID,
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
		ThenISeeText(titlePrefix).
		AndISeeSelectorKey("board.add_ticket_button")
}
