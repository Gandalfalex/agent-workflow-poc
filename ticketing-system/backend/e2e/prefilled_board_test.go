//go:build e2e

package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"ticketing-system/backend/internal/store"
)

func TestBoardDisplaysPreseededTickets(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()
	st := scenario.Harness().Store()
	ctx := scenario.Harness().Context()

	// Pre-seed tickets directly into the database
	storyID := uuid.MustParse(seed.StoryID)
	projectID := uuid.MustParse(seed.ProjectID)
	backlogID := uuid.MustParse(seed.BacklogID)

	ticketTitles := []string{
		fmt.Sprintf("Preseeded Alpha %d", time.Now().UnixNano()),
		fmt.Sprintf("Preseeded Beta %d", time.Now().UnixNano()),
		fmt.Sprintf("Preseeded Gamma %d", time.Now().UnixNano()),
	}

	for _, title := range ticketTitles {
		_, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
			Title:   title,
			Type:    "feature",
			StoryID: storyID,
			StateID: &backlogID,
		})
		if err != nil {
			t.Fatalf("seed ticket %q: %v", title, err)
		}
	}

	// Navigate to the board and verify all pre-seeded tickets are displayed
	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/" + seed.ProjectID + "/board").
		WhenIClickRefresh().
		ThenISeeText(ticketTitles[0]).
		AndISeeText(ticketTitles[1]).
		AndISeeText(ticketTitles[2])
}

func TestBoardShowsTicketsInCorrectStates(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()
	st := scenario.Harness().Store()
	ctx := scenario.Harness().Context()

	storyID := uuid.MustParse(seed.StoryID)
	projectID := uuid.MustParse(seed.ProjectID)
	backlogID := uuid.MustParse(seed.BacklogID)
	inProgressID := uuid.MustParse(seed.InProgressID)
	doneID := uuid.MustParse(seed.DoneID)

	ts := time.Now().UnixNano()
	backlogTicket := fmt.Sprintf("Pending Task %d", ts)
	inProgressTicket := fmt.Sprintf("WIP Task %d", ts)
	doneTicket := fmt.Sprintf("Completed Task %d", ts)

	// Seed one ticket in each state
	_, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
		Title:   backlogTicket,
		Type:    "feature",
		StoryID: storyID,
		StateID: &backlogID,
	})
	if err != nil {
		t.Fatalf("seed backlog ticket: %v", err)
	}

	_, err = st.CreateTicket(ctx, projectID, store.TicketCreateInput{
		Title:   inProgressTicket,
		Type:    "bug",
		StoryID: storyID,
		StateID: &inProgressID,
	})
	if err != nil {
		t.Fatalf("seed in-progress ticket: %v", err)
	}

	_, err = st.CreateTicket(ctx, projectID, store.TicketCreateInput{
		Title:   doneTicket,
		Type:    "feature",
		StoryID: storyID,
		StateID: &doneID,
	})
	if err != nil {
		t.Fatalf("seed done ticket: %v", err)
	}

	// Navigate to board and verify all tickets and column headers
	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/" + seed.ProjectID + "/board").
		WhenIClickRefresh().
		// Verify workflow state columns are present
		ThenISeeText("Backlog").
		AndISeeText("In Progress").
		AndISeeText("Done").
		// Verify all seeded tickets are visible
		AndISeeText(backlogTicket).
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

	ts := time.Now().UnixNano()

	// Create a second story
	story2, err := st.CreateStory(ctx, projectID, store.StoryCreateInput{
		Title: fmt.Sprintf("Backend Work %d", ts),
	})
	if err != nil {
		t.Fatalf("create second story: %v", err)
	}

	// Create a third story
	story3, err := st.CreateStory(ctx, projectID, store.StoryCreateInput{
		Title: fmt.Sprintf("Frontend Work %d", ts),
	})
	if err != nil {
		t.Fatalf("create third story: %v", err)
	}

	// Seed tickets for default E2E Story
	defaultStoryID := uuid.MustParse(seed.StoryID)
	defaultTicket := fmt.Sprintf("Default Story Ticket %d", ts)
	if _, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
		Title:   defaultTicket,
		Type:    "feature",
		StoryID: defaultStoryID,
		StateID: &backlogID,
	}); err != nil {
		t.Fatalf("seed default story ticket: %v", err)
	}

	// Seed tickets for story 2
	story2Ticket1 := fmt.Sprintf("API Endpoint %d", ts)
	story2Ticket2 := fmt.Sprintf("Database Migration %d", ts)
	if _, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
		Title:    story2Ticket1,
		Type:     "feature",
		StoryID:  story2.ID,
		StateID:  &backlogID,
		Priority: "high",
	}); err != nil {
		t.Fatalf("seed story2 ticket 1: %v", err)
	}
	if _, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
		Title:    story2Ticket2,
		Type:     "feature",
		StoryID:  story2.ID,
		StateID:  &inProgressID,
		Priority: "urgent",
	}); err != nil {
		t.Fatalf("seed story2 ticket 2: %v", err)
	}

	// Seed tickets for story 3
	story3Ticket := fmt.Sprintf("UI Component %d", ts)
	if _, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
		Title:   story3Ticket,
		Type:    "bug",
		StoryID: story3.ID,
		StateID: &backlogID,
	}); err != nil {
		t.Fatalf("seed story3 ticket: %v", err)
	}

	// Navigate to board and verify everything
	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/" + seed.ProjectID + "/board").
		WhenIClickRefresh().
		// All three stories should be visible
		ThenISeeText("E2E Story").
		AndISeeText(story2.Title).
		AndISeeText(story3.Title).
		// All tickets should be visible
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
	st := scenario.Harness().Store()
	ctx := scenario.Harness().Context()

	projectID := uuid.MustParse(seed.ProjectID)
	storyID := uuid.MustParse(seed.StoryID)
	backlogID := uuid.MustParse(seed.BacklogID)

	ts := time.Now().UnixNano()

	// Create tickets with every combination of type and priority
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

	for _, tc := range tickets {
		_, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
			Title:    tc.title,
			Type:     tc.tickType,
			StoryID:  storyID,
			StateID:  &backlogID,
			Priority: tc.priority,
		})
		if err != nil {
			t.Fatalf("seed ticket %q: %v", tc.title, err)
		}
	}

	// Navigate to board and verify all tickets are visible
	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/" + seed.ProjectID + "/board").
		// Double refresh: first settles any stale DEMO project load, second ensures clean E2E data
		WhenIClickRefresh().
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

	// Create an assigned ticket
	_, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
		Title:      assignedTitle,
		Type:       "feature",
		StoryID:    storyID,
		StateID:    &backlogID,
		AssigneeID: &userID,
	})
	if err != nil {
		t.Fatalf("seed assigned ticket: %v", err)
	}

	// Create an unassigned ticket
	_, err = st.CreateTicket(ctx, projectID, store.TicketCreateInput{
		Title:   unassignedTitle,
		Type:    "bug",
		StoryID: storyID,
		StateID: &backlogID,
	})
	if err != nil {
		t.Fatalf("seed unassigned ticket: %v", err)
	}

	// Navigate to board and verify both tickets are visible
	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/" + seed.ProjectID + "/board").
		WhenIClickRefresh().
		ThenISeeText(assignedTitle).
		AndISeeText(unassignedTitle)
}
