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
	h := scenario.Harness()

	projectID := uuid.MustParse(seed.ProjectID)
	storyID := uuid.MustParse(seed.StoryID)
	backlogID := uuid.MustParse(seed.BacklogID)

	ticket, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
		Title:   fmt.Sprintf("Lifecycle Ticket %d", time.Now().UnixNano()),
		Type:    "feature",
		StoryID: storyID,
		StateID: &backlogID,
	})
	if err != nil {
		t.Fatalf("seed ticket: %v", err)
	}

	now := time.Now().UTC()
	sprintStart := now.Format("2006-01-02")
	sprintEnd := now.AddDate(0, 0, 13).Format("2006-01-02")
	nextStart := now.AddDate(0, 0, 14).Format("2006-01-02")
	nextEnd := now.AddDate(0, 0, 27).Format("2006-01-02")

	// Create first sprint with the ticket
	createResp, err := h.APIRequest(
		http.MethodPost,
		fmt.Sprintf("/projects/%s/sprints", seed.ProjectID),
		bytes.NewReader([]byte(fmt.Sprintf(`{"name":"Sprint One","startDate":"%s","endDate":"%s","ticketIds":["%s"]}`,
			sprintStart, sprintEnd, ticket.ID))),
	)
	if err != nil {
		t.Fatalf("create sprint 1: %v", err)
	}
	defer createResp.Body.Close()
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("create sprint 1 status: got %d", createResp.StatusCode)
	}
	var sprint1 struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(createResp.Body).Decode(&sprint1); err != nil {
		t.Fatalf("decode sprint 1: %v", err)
	}
	if sprint1.Status != "planned" {
		t.Fatalf("expected status=planned, got %s", sprint1.Status)
	}

	// Create second sprint (rollover target)
	create2Resp, err := h.APIRequest(
		http.MethodPost,
		fmt.Sprintf("/projects/%s/sprints", seed.ProjectID),
		bytes.NewReader([]byte(fmt.Sprintf(`{"name":"Sprint Two","startDate":"%s","endDate":"%s"}`,
			nextStart, nextEnd))),
	)
	if err != nil {
		t.Fatalf("create sprint 2: %v", err)
	}
	defer create2Resp.Body.Close()
	if create2Resp.StatusCode != http.StatusCreated {
		t.Fatalf("create sprint 2 status: got %d", create2Resp.StatusCode)
	}
	var sprint2 struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(create2Resp.Body).Decode(&sprint2); err != nil {
		t.Fatalf("decode sprint 2: %v", err)
	}

	// Start sprint 1 → should transition to active
	startResp, err := h.APIRequest(
		http.MethodPost,
		fmt.Sprintf("/projects/%s/sprints/%s/start", seed.ProjectID, sprint1.ID),
		nil,
	)
	if err != nil {
		t.Fatalf("start sprint: %v", err)
	}
	defer startResp.Body.Close()
	if startResp.StatusCode != http.StatusOK {
		t.Fatalf("start sprint status: got %d", startResp.StatusCode)
	}
	var started struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(startResp.Body).Decode(&started); err != nil {
		t.Fatalf("decode started sprint: %v", err)
	}
	if started.Status != "active" {
		t.Fatalf("expected status=active after start, got %s", started.Status)
	}

	// Starting sprint 2 while sprint 1 is active should fail
	startConflictResp, err := h.APIRequest(
		http.MethodPost,
		fmt.Sprintf("/projects/%s/sprints/%s/start", seed.ProjectID, sprint2.ID),
		nil,
	)
	if err != nil {
		t.Fatalf("start conflict request: %v", err)
	}
	defer startConflictResp.Body.Close()
	if startConflictResp.StatusCode == http.StatusOK {
		t.Fatalf("expected error starting second sprint while one is active, got 200")
	}

	// Complete sprint 1, rolling the ticket over to sprint 2
	completePayload := fmt.Sprintf(`{"moveTicketIds":["%s"]}`, ticket.ID)
	completeResp, err := h.APIRequest(
		http.MethodPost,
		fmt.Sprintf("/projects/%s/sprints/%s/complete", seed.ProjectID, sprint1.ID),
		bytes.NewReader([]byte(completePayload)),
	)
	if err != nil {
		t.Fatalf("complete sprint: %v", err)
	}
	defer completeResp.Body.Close()
	if completeResp.StatusCode != http.StatusOK {
		t.Fatalf("complete sprint status: got %d", completeResp.StatusCode)
	}
	var completed struct {
		Status      string  `json:"status"`
		CompletedAt *string `json:"completedAt"`
	}
	if err := json.NewDecoder(completeResp.Body).Decode(&completed); err != nil {
		t.Fatalf("decode completed sprint: %v", err)
	}
	if completed.Status != "completed" {
		t.Fatalf("expected status=completed, got %s", completed.Status)
	}
	if completed.CompletedAt == nil {
		t.Fatalf("expected completedAt to be set")
	}

	// Verify ticket was rolled over to sprint 2
	getSprint2Resp, err := h.APIRequest(
		http.MethodGet,
		fmt.Sprintf("/projects/%s/sprints", seed.ProjectID),
		nil,
	)
	if err != nil {
		t.Fatalf("list sprints: %v", err)
	}
	defer getSprint2Resp.Body.Close()
	var sprintList struct {
		Items []struct {
			ID        string   `json:"id"`
			Status    string   `json:"status"`
			TicketIDs []string `json:"ticketIds"`
		} `json:"items"`
	}
	if err := json.NewDecoder(getSprint2Resp.Body).Decode(&sprintList); err != nil {
		t.Fatalf("decode sprint list: %v", err)
	}
	var found bool
	for _, s := range sprintList.Items {
		if s.ID == sprint2.ID {
			for _, tid := range s.TicketIDs {
				if tid == ticket.ID.String() {
					found = true
				}
			}
		}
	}
	if !found {
		t.Fatalf("expected ticket %s to be rolled over to sprint 2", ticket.ID)
	}

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

	ticket, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
		Title:   fmt.Sprintf("Sprint Planner Ticket %d", time.Now().UnixNano()),
		Type:    "feature",
		StoryID: storyID,
		StateID: &backlogID,
	})
	if err != nil {
		t.Fatalf("seed ticket: %v", err)
	}

	startDate := time.Now().UTC().Format("2006-01-02")
	endDate := time.Now().UTC().AddDate(0, 0, 13).Format("2006-01-02")
	createSprintPayload := fmt.Sprintf(`{"name":"Sprint Alpha","goal":"Ship planner","startDate":"%s","endDate":"%s","ticketIds":["%s"]}`,
		startDate, endDate, ticket.ID,
	)
	createSprintResp, err := scenario.Harness().APIRequest(
		http.MethodPost,
		fmt.Sprintf("/projects/%s/sprints", seed.ProjectID),
		bytes.NewReader([]byte(createSprintPayload)),
	)
	if err != nil {
		t.Fatalf("create sprint request: %v", err)
	}
	defer createSprintResp.Body.Close()
	if createSprintResp.StatusCode != http.StatusCreated {
		t.Fatalf("create sprint status: got %d", createSprintResp.StatusCode)
	}

	capacityPayload := `{"items":[{"scope":"team","label":"Core Team","capacity":8}]}`
	capacityResp, err := scenario.Harness().APIRequest(
		http.MethodPut,
		fmt.Sprintf("/projects/%s/capacity-settings", seed.ProjectID),
		bytes.NewReader([]byte(capacityPayload)),
	)
	if err != nil {
		t.Fatalf("replace capacity request: %v", err)
	}
	defer capacityResp.Body.Close()
	if capacityResp.StatusCode != http.StatusOK {
		t.Fatalf("replace capacity status: got %d", capacityResp.StatusCode)
	}

	forecastResp, err := scenario.Harness().APIRequest(
		http.MethodGet,
		fmt.Sprintf("/projects/%s/sprint-forecast?iterations=500", seed.ProjectID),
		nil,
	)
	if err != nil {
		t.Fatalf("forecast request: %v", err)
	}
	defer forecastResp.Body.Close()
	if forecastResp.StatusCode != http.StatusOK {
		t.Fatalf("forecast status: got %d", forecastResp.StatusCode)
	}
	var forecast struct {
		CommittedTickets  int `json:"committedTickets"`
		Projected         int `json:"projectedCompletion"`
		Capacity          int `json:"capacity"`
		OverCapacityDelta int `json:"overCapacityDelta"`
		Iterations        int `json:"iterations"`
	}
	if err := json.NewDecoder(forecastResp.Body).Decode(&forecast); err != nil {
		t.Fatalf("decode forecast: %v", err)
	}
	if forecast.Iterations != 500 {
		t.Fatalf("iterations mismatch: got %d", forecast.Iterations)
	}
	if forecast.CommittedTickets <= 0 {
		t.Fatalf("expected committed tickets > 0, got %d", forecast.CommittedTickets)
	}

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
