//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"ticketing-system/backend/internal/store"
)

func TestIncidentTimelineAndPostmortemExport(t *testing.T) {
	t.Parallel()

	sc := NewScenario(t)
	defer sc.Close()

	seed := sc.SeedData()

	var ticketID, ticketKey string

	sc.
		Given("a bug ticket exists in backlog", func(s *Scenario) error {
			h := s.Harness()
			st := h.Store()
			ctx := h.Context()
			projectID := uuid.MustParse(seed.ProjectID)
			storyID := uuid.MustParse(seed.StoryID)
			backlogID := uuid.MustParse(seed.BacklogID)

			ts := time.Now().UnixNano()
			title := fmt.Sprintf("Incident Bridge Test %d", ts)
			created, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
				Title:   title,
				Type:    "bug",
				StoryID: storyID,
				StateID: &backlogID,
			})
			if err != nil {
				return fmt.Errorf("seed ticket: %w", err)
			}
			ticketID = created.ID.String()
			ticketKey = created.Key
			return nil
		}).
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/"+seed.ProjectID+"/board").
		WhenIClickRefresh().
		Then("the ticket is visible on the board", func(s *Scenario) error {
			return s.Harness().ExpectTextVisible(ticketKey)
		}).
		When("I open the ticket by its key text", func(s *Scenario) error {
			return s.Harness().page.GetByText(ticketKey).First().Click()
		}).
		ThenISeeSelectorKey("ticket.modal").
		WhenIClickKey("ticket.incident_enabled_checkbox").
		WhenISelectOptionByValueKey("ticket.incident_severity_select", "sev1").
		WhenIFillKey("ticket.incident_impact_input", "Checkout failures for all EU customers").
		WhenIFillKey("ticket.comment_input", "Initial response started").
		WhenIClickKey("ticket.post_comment_button").
		WhenIClickKey("ticket.save_button").
		WhenIClickRefresh().
		Then("the ticket is still visible after refresh", func(s *Scenario) error {
			return s.Harness().ExpectTextVisible(ticketKey)
		}).
		When("I reopen the ticket", func(s *Scenario) error {
			return s.Harness().page.GetByText(ticketKey).First().Click()
		}).
		ThenISeeSelectorKey("ticket.modal").
		ThenISeeText("changed incident severity from").
		Then("the incident timeline contains activity and webhook items", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest(http.MethodGet, "/tickets/"+ticketID+"/incident-timeline", nil)
			if err != nil {
				return fmt.Errorf("timeline request: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("timeline status=%d body=%s", resp.StatusCode, string(body))
			}
			var timeline struct {
				Items []struct {
					Type string `json:"type"`
				} `json:"items"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&timeline); err != nil {
				return fmt.Errorf("decode timeline: %w", err)
			}
			hasActivity := false
			hasWebhook := false
			for _, item := range timeline.Items {
				if item.Type == "activity" {
					hasActivity = true
				}
				if item.Type == "webhook" {
					hasWebhook = true
				}
			}
			if !hasActivity {
				return fmt.Errorf("expected incident timeline to include activity item")
			}
			if !hasWebhook {
				return fmt.Errorf("expected incident timeline to include webhook item")
			}
			return nil
		}).
		And("the postmortem export is valid markdown with all required sections", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest(http.MethodGet, "/tickets/"+ticketID+"/incident-postmortem", nil)
			if err != nil {
				return fmt.Errorf("postmortem request: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("postmortem status=%d body=%s", resp.StatusCode, string(body))
			}
			body, _ := io.ReadAll(resp.Body)
			if contentType := resp.Header.Get("Content-Type"); !strings.Contains(contentType, "text/markdown") {
				return fmt.Errorf("expected markdown content-type, got %q", contentType)
			}
			for _, needle := range []string{
				"# Postmortem Draft:",
				ticketKey,
				"## Incident Summary",
				"Severity: sev1",
				"## Timeline",
				"## Root Cause",
			} {
				if !bytes.Contains(body, []byte(needle)) {
					return fmt.Errorf("expected postmortem to contain %q; body=%s", needle, string(body))
				}
			}
			return nil
		})
}
