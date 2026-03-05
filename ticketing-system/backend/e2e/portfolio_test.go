//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"ticketing-system/backend/internal/store"
)

func TestPortfolioStatsAPIAndUI(t *testing.T) {
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

	// Seed an urgent ticket so the project shows as "at risk"
	_, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
		Title:    fmt.Sprintf("portfolio-urgent-%d", ts),
		Type:     "bug",
		StoryID:  storyID,
		StateID:  &backlogID,
		Priority: "urgent",
	})
	if err != nil {
		t.Fatalf("seed urgent ticket: %v", err)
	}

	// --- API: JSON response ---
	scenario.GivenAppIsRunning().WhenIGoToRoute("home").WhenILogInWithHarnessUser()

	resp, err := scenario.Harness().APIRequest(http.MethodGet, "/portfolio/stats", nil)
	if err != nil {
		t.Fatalf("portfolio stats API request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var stats struct {
		Totals struct {
			TotalProjects int `json:"totalProjects"`
			TotalOpen     int `json:"totalOpen"`
		} `json:"totals"`
		Projects []struct {
			ID                string  `json:"id"`
			ProjectKey        string  `json:"projectKey"`
			UrgentOpen        int     `json:"urgentOpen"`
			WeeklyThroughput  int     `json:"weeklyThroughput"`
			AvgCycleTimeHours float64 `json:"avgCycleTimeHours"`
		} `json:"projects"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		t.Fatalf("decode portfolio stats: %v", err)
	}
	if stats.Totals.TotalProjects == 0 {
		t.Fatal("expected at least 1 project in portfolio stats")
	}

	// Find our seeded project and assert new throughput fields are present
	found := false
	for _, p := range stats.Projects {
		if p.ID == seed.ProjectID {
			found = true
			if p.UrgentOpen == 0 {
				t.Errorf("expected urgentOpen > 0 for seeded project, got 0")
			}
			if p.WeeklyThroughput < 0 {
				t.Errorf("weeklyThroughput should be >= 0, got %d", p.WeeklyThroughput)
			}
			if p.AvgCycleTimeHours < 0 {
				t.Errorf("avgCycleTimeHours should be >= 0, got %f", p.AvgCycleTimeHours)
			}
		}
	}
	if !found {
		t.Fatalf("seeded project %s not found in portfolio stats", seed.ProjectID)
	}

	// --- API: CSV export ---
	csvResp, err := scenario.Harness().APIRequest(http.MethodGet, "/portfolio/stats?format=csv", nil)
	if err != nil {
		t.Fatalf("portfolio CSV request: %v", err)
	}
	defer csvResp.Body.Close()
	if csvResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for CSV, got %d", csvResp.StatusCode)
	}
	ct := csvResp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/csv") {
		t.Errorf("expected Content-Type text/csv, got %q", ct)
	}

	// --- UI: navigate to portfolio via nav tab ---
	scenario.
		WhenIClickKey("nav.portfolio_tab").
		ThenURLContains("/portfolio").
		ThenISeeSelectorKey("portfolio.page").
		ThenISeeSelectorKey("portfolio.stats_total")
}
