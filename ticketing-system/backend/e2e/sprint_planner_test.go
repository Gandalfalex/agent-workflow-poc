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
