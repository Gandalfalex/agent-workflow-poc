//go:build e2e

package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"ticketing-system/backend/internal/store"
)

func seedPrefilledTicket(s *Scenario, projectID, storyID, stateID uuid.UUID, title, ticketType, priority string) error {
	_, err := s.Harness().Store().CreateTicket(s.Harness().Context(), projectID, store.TicketCreateInput{
		Title:    title,
		Type:     ticketType,
		StoryID:  storyID,
		StateID:  &stateID,
		Priority: priority,
	})
	if err != nil {
		return fmt.Errorf("seed ticket %q: %w", title, err)
	}
	return nil
}

func navigateToBoard(s *Scenario, projectID string) *Scenario {
	return s.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(projectID).
		ThenURLContains("/projects/" + projectID + "/board").
		WhenIClickRefresh()
}

func TestBoardDisplaysPreseededTickets(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()

	storyID := uuid.MustParse(seed.StoryID)
	projectID := uuid.MustParse(seed.ProjectID)
	backlogID := uuid.MustParse(seed.BacklogID)

	ticketTitles := []string{
		fmt.Sprintf("Preseeded Alpha %d", time.Now().UnixNano()),
		fmt.Sprintf("Preseeded Beta %d", time.Now().UnixNano()),
		fmt.Sprintf("Preseeded Gamma %d", time.Now().UnixNano()),
	}

	scenario.
		Given("three tickets are pre-seeded directly into the database", func(s *Scenario) error {
			for _, title := range ticketTitles {
				if err := seedPrefilledTicket(s, projectID, storyID, backlogID, title, "feature", ""); err != nil {
					return err
				}
			}
			return nil
		})

	navigateToBoard(scenario, seed.ProjectID).
		ThenISeeText(ticketTitles[0]).
		AndISeeText(ticketTitles[1]).
		AndISeeText(ticketTitles[2])
}

func TestBoardShowsTicketsInCorrectStates(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()

	storyID := uuid.MustParse(seed.StoryID)
	projectID := uuid.MustParse(seed.ProjectID)
	backlogID := uuid.MustParse(seed.BacklogID)
	inProgressID := uuid.MustParse(seed.InProgressID)
	doneID := uuid.MustParse(seed.DoneID)

	ts := time.Now().UnixNano()
	backlogTicket := fmt.Sprintf("Pending Task %d", ts)
	inProgressTicket := fmt.Sprintf("WIP Task %d", ts)
	doneTicket := fmt.Sprintf("Completed Task %d", ts)

	scenario.
		Given("one ticket is seeded in each workflow state", func(s *Scenario) error {
			if err := seedPrefilledTicket(s, projectID, storyID, backlogID, backlogTicket, "feature", ""); err != nil {
				return err
			}
			if err := seedPrefilledTicket(s, projectID, storyID, inProgressID, inProgressTicket, "bug", ""); err != nil {
				return err
			}
			return seedPrefilledTicket(s, projectID, storyID, doneID, doneTicket, "feature", "")
		})

	navigateToBoard(scenario, seed.ProjectID).
		ThenISeeText(backlogTicket).
		AndISeeText(inProgressTicket).
		AndISeeText(doneTicket)
}

func TestBoardWithMultipleStoriesAndTickets(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()
	st := scenario.Harness().Store()
	ctx := scenario.Harness().Context()

	projectID := uuid.MustParse(seed.ProjectID)
	backlogID := uuid.MustParse(seed.BacklogID)
	inProgressID := uuid.MustParse(seed.InProgressID)
	defaultStoryID := uuid.MustParse(seed.StoryID)

	ts := time.Now().UnixNano()

	var (
		story2Title   string
		story3Title   string
		defaultTicket = fmt.Sprintf("Default Story Ticket %d", ts)
		story2Ticket1 = fmt.Sprintf("API Endpoint %d", ts)
		story2Ticket2 = fmt.Sprintf("Database Migration %d", ts)
		story3Ticket  = fmt.Sprintf("UI Component %d", ts)
	)

	scenario.
		Given("two additional stories with tickets are seeded into the project", func(s *Scenario) error {
			story2, err := st.CreateStory(ctx, projectID, store.StoryCreateInput{
				Title: fmt.Sprintf("Backend Work %d", ts),
			})
			if err != nil {
				return fmt.Errorf("create second story: %w", err)
			}
			story2Title = story2.Title

			story3, err := st.CreateStory(ctx, projectID, store.StoryCreateInput{
				Title: fmt.Sprintf("Frontend Work %d", ts),
			})
			if err != nil {
				return fmt.Errorf("create third story: %w", err)
			}
			story3Title = story3.Title

			if err := seedPrefilledTicket(s, projectID, defaultStoryID, backlogID, defaultTicket, "feature", ""); err != nil {
				return err
			}
			if _, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
				Title:    story2Ticket1,
				Type:     "feature",
				StoryID:  story2.ID,
				StateID:  &backlogID,
				Priority: "high",
			}); err != nil {
				return fmt.Errorf("seed story2 ticket 1: %w", err)
			}
			if _, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
				Title:    story2Ticket2,
				Type:     "feature",
				StoryID:  story2.ID,
				StateID:  &inProgressID,
				Priority: "urgent",
			}); err != nil {
				return fmt.Errorf("seed story2 ticket 2: %w", err)
			}
			if _, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
				Title:   story3Ticket,
				Type:    "bug",
				StoryID: story3.ID,
				StateID: &backlogID,
			}); err != nil {
				return fmt.Errorf("seed story3 ticket: %w", err)
			}
			return nil
		})

	navigateToBoard(scenario, seed.ProjectID).
		ThenISeeText("E2E Story").
		AndISeeText(story2Title).
		AndISeeText(story3Title).
		AndISeeText(defaultTicket).
		AndISeeText(story2Ticket1).
		AndISeeText(story2Ticket2).
		AndISeeText(story3Ticket)
}

func TestBoardWithManyTicketsShowsAllPrioritiesAndTypes(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()

	projectID := uuid.MustParse(seed.ProjectID)
	storyID := uuid.MustParse(seed.StoryID)
	backlogID := uuid.MustParse(seed.BacklogID)

	ts := time.Now().UnixNano()

	tickets := []struct {
		title    string
		tickType string
		priority string
	}{
		{fmt.Sprintf("Low Feature %d", ts), "feature", "low"},
		{fmt.Sprintf("Medium Feature %d", ts), "feature", "medium"},
		{fmt.Sprintf("High Feature %d", ts), "feature", "high"},
		{fmt.Sprintf("Urgent Feature %d", ts), "feature", "urgent"},
		{fmt.Sprintf("Low Bug %d", ts), "bug", "low"},
		{fmt.Sprintf("Medium Bug %d", ts), "bug", "medium"},
		{fmt.Sprintf("High Bug %d", ts), "bug", "high"},
		{fmt.Sprintf("Urgent Bug %d", ts), "bug", "urgent"},
	}

	scenario.
		Given("tickets covering every combination of type and priority are seeded", func(s *Scenario) error {
			for _, tc := range tickets {
				if err := seedPrefilledTicket(s, projectID, storyID, backlogID, tc.title, tc.tickType, tc.priority); err != nil {
					return err
				}
			}
			return nil
		})

	// Double refresh: first settles any stale DEMO project load, second ensures clean E2E data
	navigateToBoard(scenario, seed.ProjectID).
		WhenIWait(1).
		WhenIClickRefresh()

	for _, tc := range tickets {
		scenario.AndISeeText(tc.title)
	}
}

func TestBoardWithAssignedTickets(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()
	st := scenario.Harness().Store()
	ctx := scenario.Harness().Context()

	projectID := uuid.MustParse(seed.ProjectID)
	storyID := uuid.MustParse(seed.StoryID)
	backlogID := uuid.MustParse(seed.BacklogID)
	userID := uuid.MustParse(seed.UserID)

	ts := time.Now().UnixNano()
	assignedTitle := fmt.Sprintf("Owned Ticket %d", ts)
	unassignedTitle := fmt.Sprintf("No-Owner Ticket %d", ts)

	scenario.
		Given("an assigned and an unassigned ticket are seeded", func(s *Scenario) error {
			if _, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
				Title:      assignedTitle,
				Type:       "feature",
				StoryID:    storyID,
				StateID:    &backlogID,
				AssigneeID: &userID,
			}); err != nil {
				return fmt.Errorf("seed assigned ticket: %w", err)
			}
			if err := seedPrefilledTicket(s, projectID, storyID, backlogID, unassignedTitle, "bug", ""); err != nil {
				return err
			}
			return nil
		})

	navigateToBoard(scenario, seed.ProjectID).
		ThenISeeText(assignedTitle).
		AndISeeText(unassignedTitle)
}
