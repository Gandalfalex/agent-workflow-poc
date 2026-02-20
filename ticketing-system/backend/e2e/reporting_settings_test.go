//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestReportingTabAndExportEndpoints(t *testing.T) {
	t.Parallel()

	sc := NewScenario(t)
	defer sc.Close()

	seed := sc.SeedData()
	id, pw := sc.Harness().LoginCredentials()

	sc.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs(id, pw).
		ThenISeeSelectorKey("nav.board_tab").
		WhenISelectProjectByID(seed.ProjectID).
		WhenIClickRefresh().
		ThenISeeSelectorKey("nav.settings_tab").
		WhenINavigateToSettings().
		WhenIClickKey("settings.reporting_tab").
		ThenISeeSelectorKey("reporting.view").
		AndISeeSelectorKey("reporting.from_input").
		AndISeeSelectorKey("reporting.to_input").
		AndISeeSelectorKey("reporting.reload_button").
		AndISeeSelectorKey("reporting.export_json_button").
		AndISeeSelectorKey("reporting.export_csv_button").
		AndISeeSelectorKey("reporting.throughput_list").
		Then("the reporting summary and export endpoints return data", func(s *Scenario) error {
			to := time.Now().UTC().Format("2006-01-02")
			from := time.Now().UTC().AddDate(0, 0, -13).Format("2006-01-02")

			summaryPath := fmt.Sprintf("/projects/%s/reporting/summary?from=%s&to=%s", seed.ProjectID, from, to)
			summaryResp, err := s.Harness().APIRequest(http.MethodGet, summaryPath, nil)
			if err != nil {
				return fmt.Errorf("GET reporting summary: %w", err)
			}
			defer summaryResp.Body.Close()
			if summaryResp.StatusCode != http.StatusOK {
				return fmt.Errorf("expected summary status 200, got %d", summaryResp.StatusCode)
			}

			var summary struct {
				From            string `json:"from"`
				To              string `json:"to"`
				ThroughputByDay []struct {
					Date  string `json:"date"`
					Value int    `json:"value"`
				} `json:"throughputByDay"`
			}
			if err := json.NewDecoder(summaryResp.Body).Decode(&summary); err != nil {
				return fmt.Errorf("decode reporting summary: %w", err)
			}
			if summary.From == "" || summary.To == "" {
				return fmt.Errorf("summary range is empty")
			}
			if len(summary.ThroughputByDay) == 0 {
				return fmt.Errorf("expected throughput rows in summary")
			}

			jsonPath := fmt.Sprintf("/projects/%s/reporting/export?from=%s&to=%s&format=json", seed.ProjectID, from, to)
			jsonResp, err := s.Harness().APIRequest(http.MethodGet, jsonPath, nil)
			if err != nil {
				return fmt.Errorf("GET reporting export json: %w", err)
			}
			defer jsonResp.Body.Close()
			if jsonResp.StatusCode != http.StatusOK {
				return fmt.Errorf("expected export json status 200, got %d", jsonResp.StatusCode)
			}
			if !strings.Contains(strings.ToLower(jsonResp.Header.Get("Content-Type")), "application/json") {
				return fmt.Errorf("expected export json content-type application/json, got %q", jsonResp.Header.Get("Content-Type"))
			}

			var exported struct {
				GeneratedAt string `json:"generatedAt"`
				Summary     struct {
					From string `json:"from"`
					To   string `json:"to"`
				} `json:"summary"`
			}
			if err := json.NewDecoder(jsonResp.Body).Decode(&exported); err != nil {
				return fmt.Errorf("decode export json: %w", err)
			}
			if exported.GeneratedAt == "" {
				return fmt.Errorf("export json missing generatedAt")
			}
			if exported.Summary.From == "" || exported.Summary.To == "" {
				return fmt.Errorf("export json missing summary range")
			}

			csvPath := fmt.Sprintf("/projects/%s/reporting/export?from=%s&to=%s&format=csv", seed.ProjectID, from, to)
			csvResp, err := s.Harness().APIRequest(http.MethodGet, csvPath, nil)
			if err != nil {
				return fmt.Errorf("GET reporting export csv: %w", err)
			}
			defer csvResp.Body.Close()
			if csvResp.StatusCode != http.StatusOK {
				return fmt.Errorf("expected export csv status 200, got %d", csvResp.StatusCode)
			}
			if !strings.Contains(strings.ToLower(csvResp.Header.Get("Content-Type")), "text/csv") {
				return fmt.Errorf("expected export csv content-type text/csv, got %q", csvResp.Header.Get("Content-Type"))
			}
			body, err := io.ReadAll(csvResp.Body)
			if err != nil {
				return fmt.Errorf("read export csv body: %w", err)
			}
			text := string(body)
			if !strings.Contains(text, "section,date,label,value") {
				return fmt.Errorf("csv header missing section/date/label/value columns")
			}
			if !strings.Contains(text, "throughput_by_day") {
				return fmt.Errorf("csv body missing throughput rows")
			}

			return nil
		})
}
