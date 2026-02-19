//go:build e2e

package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"ticketing-system/backend/internal/store"
)

func TestDashboardShowsRecentActivity(t *testing.T) {
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

	ts := time.Now().UnixNano()
	title := fmt.Sprintf("Dashboard Activity Ticket %d", ts)

	ticket, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
		Title:   title,
		Type:    "feature",
		StoryID: storyID,
		StateID: &backlogID,
	})
	if err != nil {
		t.Fatalf("seed ticket: %v", err)
	}

	// Record an activity by updating the ticket state
	_, err = st.UpdateTicket(ctx, ticket.ID, store.TicketUpdateInput{
		StateID: &inProgressID,
	})
	if err != nil {
		t.Fatalf("update ticket: %v", err)
	}

	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/"+seed.ProjectID+"/board").
		WhenIClickKey("nav.dashboard_tab").
		ThenURLContains("/projects/"+seed.ProjectID+"/dashboard").
		ThenISeeSelectorKey("dashboard.recent_activity")
}

func TestDashboardNavigationFromBoard(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()
	st := scenario.Harness().Store()
	ctx := scenario.Harness().Context()

	projectID := uuid.MustParse(seed.ProjectID)
	storyID := uuid.MustParse(seed.StoryID)
	backlogID := uuid.MustParse(seed.BacklogID)
	_, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
		Title:   "Nav Test Ticket",
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
		ThenURLContains("/projects/"+seed.ProjectID+"/board").
		WhenIClickKey("nav.dashboard_tab").
		ThenURLContains("/projects/"+seed.ProjectID+"/dashboard").
		ThenISeeSelectorKey("dashboard.recent_activity").
		WhenIClickKey("nav.board_tab").
		ThenURLContains("/projects/"+seed.ProjectID+"/board")
}
