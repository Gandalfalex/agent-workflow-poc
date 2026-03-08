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

	var stats struct {
		Totals struct {
			TotalProjects int `json:"totalProjects"`
			TotalOpen     int `json:"totalOpen"`
		} `json:"totals"`
		Projects []struct {
			ID               string  `json:"id"`
			ProjectKey       string  `json:"projectKey"`
			UrgentOpen       int     `json:"urgentOpen"`
			WeeklyThroughput int     `json:"weeklyThroughput"`
			AvgCycleTimeHours float64 `json:"avgCycleTimeHours"`
		} `json:"projects"`
	}

	scenario.
		Given("an urgent ticket is seeded so the project appears at risk", func(s *Scenario) error {
			st := s.Harness().Store()
			ctx := s.Harness().Context()
			projectID := uuid.MustParse(seed.ProjectID)
			storyID := uuid.MustParse(seed.StoryID)
			backlogID := uuid.MustParse(seed.BacklogID)
			ts := time.Now().UnixNano()
			_, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
				Title:    fmt.Sprintf("portfolio-urgent-%d", ts),
				Type:     "bug",
				StoryID:  storyID,
				StateID:  &backlogID,
				Priority: "urgent",
			})
			if err != nil {
				return fmt.Errorf("seed urgent ticket: %w", err)
			}
			return nil
		}).
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInWithHarnessUser().
		When("I request the portfolio stats JSON API", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest(http.MethodGet, "/portfolio/stats", nil)
			if err != nil {
				return fmt.Errorf("portfolio stats API request: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("expected 200, got %d", resp.StatusCode)
			}
			if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
				return fmt.Errorf("decode portfolio stats: %w", err)
			}
			return nil
		}).
		Then("the response contains at least one project", func(s *Scenario) error {
			if stats.Totals.TotalProjects == 0 {
				return fmt.Errorf("expected at least 1 project in portfolio stats")
			}
			return nil
		}).
		And("the seeded project has urgentOpen > 0 and non-negative throughput metrics", func(s *Scenario) error {
			for _, p := range stats.Projects {
				if p.ID != seed.ProjectID {
					continue
				}
				if p.UrgentOpen == 0 {
					return fmt.Errorf("expected urgentOpen > 0 for seeded project, got 0")
				}
				if p.WeeklyThroughput < 0 {
					return fmt.Errorf("weeklyThroughput should be >= 0, got %d", p.WeeklyThroughput)
				}
				if p.AvgCycleTimeHours < 0 {
					return fmt.Errorf("avgCycleTimeHours should be >= 0, got %f", p.AvgCycleTimeHours)
				}
				return nil
			}
			return fmt.Errorf("seeded project %s not found in portfolio stats", seed.ProjectID)
		}).
		When("I request the portfolio stats CSV export", func(s *Scenario) error {
			csvResp, err := s.Harness().APIRequest(http.MethodGet, "/portfolio/stats?format=csv", nil)
			if err != nil {
				return fmt.Errorf("portfolio CSV request: %w", err)
			}
			defer csvResp.Body.Close()
			if csvResp.StatusCode != http.StatusOK {
				return fmt.Errorf("expected 200 for CSV, got %d", csvResp.StatusCode)
			}
			ct := csvResp.Header.Get("Content-Type")
			if !strings.HasPrefix(ct, "text/csv") {
				return fmt.Errorf("expected Content-Type text/csv, got %q", ct)
			}
			return nil
		}).
		WhenIClickKey("nav.portfolio_tab").
		ThenURLContains("/portfolio").
		ThenISeeSelectorKey("portfolio.page").
		ThenISeeSelectorKey("portfolio.stats_total")
}
