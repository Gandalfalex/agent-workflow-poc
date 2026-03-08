//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"ticketing-system/backend/internal/store"
)

func TestSprintLifecycleStartCompleteAndRollover(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()
	st := scenario.Harness().Store()
	ctx := scenario.Harness().Context()

	projectID := uuid.MustParse(seed.ProjectID)
	storyID := uuid.MustParse(seed.StoryID)
	backlogID := uuid.MustParse(seed.BacklogID)

	now := time.Now().UTC()
	sprintStart := now.Format("2006-01-02")
	sprintEnd := now.AddDate(0, 0, 13).Format("2006-01-02")
	nextStart := now.AddDate(0, 0, 14).Format("2006-01-02")
	nextEnd := now.AddDate(0, 0, 27).Format("2006-01-02")

	var ticketID string
	var sprint1ID, sprint2ID string

	scenario.
		Given("a ticket exists in the backlog", func(s *Scenario) error {
			ticket, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
				Title:   fmt.Sprintf("Lifecycle Ticket %d", time.Now().UnixNano()),
				Type:    "feature",
				StoryID: storyID,
				StateID: &backlogID,
			})
			if err != nil {
				return fmt.Errorf("seed ticket: %w", err)
			}
			ticketID = ticket.ID.String()
			return nil
		}).
		And("a first sprint is created with that ticket", func(s *Scenario) error {
			body := fmt.Sprintf(`{"name":"Sprint One","startDate":"%s","endDate":"%s","ticketIds":["%s"]}`,
				sprintStart, sprintEnd, ticketID)
			resp, err := s.Harness().APIRequest(http.MethodPost, fmt.Sprintf("/projects/%s/sprints", seed.ProjectID), bytes.NewReader([]byte(body)))
			if err != nil {
				return fmt.Errorf("create sprint 1: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusCreated {
				return fmt.Errorf("create sprint 1 status: got %d", resp.StatusCode)
			}
			var sprint struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&sprint); err != nil {
				return fmt.Errorf("decode sprint 1: %w", err)
			}
			if sprint.Status != "planned" {
				return fmt.Errorf("expected status=planned, got %s", sprint.Status)
			}
			sprint1ID = sprint.ID
			return nil
		}).
		And("a second sprint is created as rollover target", func(s *Scenario) error {
			body := fmt.Sprintf(`{"name":"Sprint Two","startDate":"%s","endDate":"%s"}`, nextStart, nextEnd)
			resp, err := s.Harness().APIRequest(http.MethodPost, fmt.Sprintf("/projects/%s/sprints", seed.ProjectID), bytes.NewReader([]byte(body)))
			if err != nil {
				return fmt.Errorf("create sprint 2: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusCreated {
				return fmt.Errorf("create sprint 2 status: got %d", resp.StatusCode)
			}
			var sprint struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&sprint); err != nil {
				return fmt.Errorf("decode sprint 2: %w", err)
			}
			sprint2ID = sprint.ID
			return nil
		}).
		When("I start sprint 1", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest(http.MethodPost, fmt.Sprintf("/projects/%s/sprints/%s/start", seed.ProjectID, sprint1ID), nil)
			if err != nil {
				return fmt.Errorf("start sprint: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("start sprint status: got %d", resp.StatusCode)
			}
			var started struct {
				Status string `json:"status"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&started); err != nil {
				return fmt.Errorf("decode started sprint: %w", err)
			}
			if started.Status != "active" {
				return fmt.Errorf("expected status=active after start, got %s", started.Status)
			}
			return nil
		}).
		Then("starting sprint 2 while sprint 1 is active is rejected", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest(http.MethodPost, fmt.Sprintf("/projects/%s/sprints/%s/start", seed.ProjectID, sprint2ID), nil)
			if err != nil {
				return fmt.Errorf("start conflict request: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return fmt.Errorf("expected error starting second sprint while one is active, got 200")
			}
			return nil
		}).
		When("I complete sprint 1 and roll the ticket over to sprint 2", func(s *Scenario) error {
			payload := fmt.Sprintf(`{"moveTicketIds":["%s"]}`, ticketID)
			resp, err := s.Harness().APIRequest(http.MethodPost, fmt.Sprintf("/projects/%s/sprints/%s/complete", seed.ProjectID, sprint1ID), bytes.NewReader([]byte(payload)))
			if err != nil {
				return fmt.Errorf("complete sprint: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("complete sprint status: got %d", resp.StatusCode)
			}
			var completed struct {
				Status      string  `json:"status"`
				CompletedAt *string `json:"completedAt"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&completed); err != nil {
				return fmt.Errorf("decode completed sprint: %w", err)
			}
			if completed.Status != "completed" {
				return fmt.Errorf("expected status=completed, got %s", completed.Status)
			}
			if completed.CompletedAt == nil {
				return fmt.Errorf("expected completedAt to be set")
			}
			return nil
		}).
		Then("the ticket appears in sprint 2 after rollover", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest(http.MethodGet, fmt.Sprintf("/projects/%s/sprints", seed.ProjectID), nil)
			if err != nil {
				return fmt.Errorf("list sprints: %w", err)
			}
			defer resp.Body.Close()
			var sprintList struct {
				Items []struct {
					ID        string   `json:"id"`
					Status    string   `json:"status"`
					TicketIDs []string `json:"ticketIds"`
				} `json:"items"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&sprintList); err != nil {
				return fmt.Errorf("decode sprint list: %w", err)
			}
			for _, sp := range sprintList.Items {
				if sp.ID == sprint2ID {
					for _, tid := range sp.TicketIDs {
						if tid == ticketID {
							return nil
						}
					}
				}
			}
			return fmt.Errorf("expected ticket %s to be rolled over to sprint 2", ticketID)
		})

	// UI: sprint sidebar is visible on board and shows completed sprint section
	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		WhenIClickKey("board.sprint_sidebar_toggle").
		ThenISeeSelectorKey("board.sprint_sidebar").
		ThenISeeSelectorKey("sprint.completed_section")
}

func TestSprintPlannerForecastAndDashboardPanel(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()
	st := scenario.Harness().Store()
	ctx := scenario.Harness().Context()

	projectID := uuid.MustParse(seed.ProjectID)
	storyID := uuid.MustParse(seed.StoryID)
	backlogID := uuid.MustParse(seed.BacklogID)

	startDate := time.Now().UTC().Format("2006-01-02")
	endDate := time.Now().UTC().AddDate(0, 0, 13).Format("2006-01-02")

	scenario.
		Given("a ticket exists in the backlog", func(s *Scenario) error {
			_, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
				Title:   fmt.Sprintf("Sprint Planner Ticket %d", time.Now().UnixNano()),
				Type:    "feature",
				StoryID: storyID,
				StateID: &backlogID,
			})
			if err != nil {
				return fmt.Errorf("seed ticket: %w", err)
			}
			return nil
		}).
		And("a sprint is created with that ticket", func(s *Scenario) error {
			ticket, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
				Title:   fmt.Sprintf("Sprint Planner Ticket %d", time.Now().UnixNano()),
				Type:    "feature",
				StoryID: storyID,
				StateID: &backlogID,
			})
			if err != nil {
				return fmt.Errorf("seed sprint ticket: %w", err)
			}
			payload := fmt.Sprintf(`{"name":"Sprint Alpha","goal":"Ship planner","startDate":"%s","endDate":"%s","ticketIds":["%s"]}`,
				startDate, endDate, ticket.ID)
			resp, err := s.Harness().APIRequest(http.MethodPost, fmt.Sprintf("/projects/%s/sprints", seed.ProjectID), bytes.NewReader([]byte(payload)))
			if err != nil {
				return fmt.Errorf("create sprint request: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusCreated {
				return fmt.Errorf("create sprint status: got %d", resp.StatusCode)
			}
			return nil
		}).
		And("team capacity settings are configured", func(s *Scenario) error {
			payload := `{"items":[{"scope":"team","label":"Core Team","capacity":8}]}`
			resp, err := s.Harness().APIRequest(http.MethodPut, fmt.Sprintf("/projects/%s/capacity-settings", seed.ProjectID), bytes.NewReader([]byte(payload)))
			if err != nil {
				return fmt.Errorf("replace capacity request: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("replace capacity status: got %d", resp.StatusCode)
			}
			return nil
		}).
		When("I request a sprint forecast with 500 iterations", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest(http.MethodGet, fmt.Sprintf("/projects/%s/sprint-forecast?iterations=500", seed.ProjectID), nil)
			if err != nil {
				return fmt.Errorf("forecast request: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("forecast status: got %d", resp.StatusCode)
			}
			var forecast struct {
				CommittedTickets  int `json:"committedTickets"`
				Projected         int `json:"projectedCompletion"`
				Capacity          int `json:"capacity"`
				OverCapacityDelta int `json:"overCapacityDelta"`
				Iterations        int `json:"iterations"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&forecast); err != nil {
				return fmt.Errorf("decode forecast: %w", err)
			}
			if forecast.Iterations != 500 {
				return fmt.Errorf("iterations mismatch: got %d", forecast.Iterations)
			}
			if forecast.CommittedTickets <= 0 {
				return fmt.Errorf("expected committed tickets > 0, got %d", forecast.CommittedTickets)
			}
			return nil
		})

	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		WhenIClickKey("nav.dashboard_tab").
		ThenURLContains("/projects/" + seed.ProjectID + "/dashboard").
		ThenISeeSelectorKey("dashboard.sprint_forecast").
		AndISeeSelectorKey("dashboard.sprint_committed").
		AndISeeSelectorKey("dashboard.sprint_projected").
		AndISeeSelectorKey("dashboard.sprint_capacity").
		AndISeeSelectorKey("dashboard.sprint_delta")
}
