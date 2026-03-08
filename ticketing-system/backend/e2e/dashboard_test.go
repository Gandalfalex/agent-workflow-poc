//go:build e2e

package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"ticketing-system/backend/internal/store"
)

// seedDashboardTicket creates a ticket and moves it to in-progress, recording activity.
func seedDashboardTicket(s *Scenario, seed SeedData, title string) error {
	st := s.Harness().Store()
	ctx := s.Harness().Context()

	projectID := uuid.MustParse(seed.ProjectID)
	storyID := uuid.MustParse(seed.StoryID)
	backlogID := uuid.MustParse(seed.BacklogID)
	inProgressID := uuid.MustParse(seed.InProgressID)

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
}

func TestDashboardShowsRecentActivity(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()

	scenario.
		Given("a ticket exists and has been moved to in-progress to generate activity", func(s *Scenario) error {
			ts := time.Now().UnixNano()
			return seedDashboardTicket(s, seed, fmt.Sprintf("Dashboard Activity Ticket %d", ts))
		}).
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

	scenario.
		Given("a ticket exists in the project", func(s *Scenario) error {
			st := s.Harness().Store()
			ctx := s.Harness().Context()
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
				return fmt.Errorf("seed ticket: %w", err)
			}
			return nil
		}).
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
