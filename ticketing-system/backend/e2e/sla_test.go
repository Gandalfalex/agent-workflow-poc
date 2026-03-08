//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestSLAPoliciesAPIGetAndUpdate(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()
	seed := scenario.SeedData()

	var updated struct {
		Items []struct {
			Priority           string `json:"priority"`
			TicketType         string `json:"ticketType"`
			FirstResponseHours int    `json:"firstResponseHours"`
			CompletionHours    int    `json:"completionHours"`
			AtRiskThresholdPct int    `json:"atRiskThresholdPct"`
		} `json:"items"`
	}

	scenario.
		Given("the project has no SLA policies configured", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("GET", fmt.Sprintf("/projects/%s/sla-policies", seed.ProjectID), nil)
			if err != nil {
				return fmt.Errorf("get sla policies: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200 from GET sla-policies, got %d", resp.StatusCode)
			}
			var initial struct {
				Items []interface{} `json:"items"`
			}
			return json.NewDecoder(resp.Body).Decode(&initial)
		}).
		When("I PUT an SLA policy for urgent bugs with 2h first response and 8h completion", func(s *Scenario) error {
			body := `{"items":[{"priority":"urgent","ticketType":"bug","firstResponseHours":2,"completionHours":8,"atRiskThresholdPct":75}]}`
			resp, err := s.Harness().APIRequest("PUT", fmt.Sprintf("/projects/%s/sla-policies", seed.ProjectID), strings.NewReader(body))
			if err != nil {
				return fmt.Errorf("put sla policies: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200 from PUT sla-policies, got %d", resp.StatusCode)
			}
			return json.NewDecoder(resp.Body).Decode(&updated)
		}).
		Then("the response contains exactly one policy with the correct values", func(s *Scenario) error {
			if len(updated.Items) != 1 {
				return fmt.Errorf("expected 1 SLA policy, got %d", len(updated.Items))
			}
			p := updated.Items[0]
			if p.Priority != "urgent" || p.TicketType != "bug" || p.FirstResponseHours != 2 || p.CompletionHours != 8 {
				return fmt.Errorf("unexpected policy values: %+v", p)
			}
			return nil
		}).
		And("a subsequent GET returns the saved policy", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("GET", fmt.Sprintf("/projects/%s/sla-policies", seed.ProjectID), nil)
			if err != nil {
				return fmt.Errorf("get sla policies after update: %w", err)
			}
			defer resp.Body.Close()
			var afterUpdate struct {
				Items []struct {
					Priority string `json:"priority"`
				} `json:"items"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&afterUpdate); err != nil {
				return fmt.Errorf("decode after update: %w", err)
			}
			if len(afterUpdate.Items) != 1 || afterUpdate.Items[0].Priority != "urgent" {
				return fmt.Errorf("expected 1 urgent policy after GET, got %d items", len(afterUpdate.Items))
			}
			return nil
		})
}

func TestSLAStatusAppearsOnBoardTicket(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()
	seed := scenario.SeedData()

	scenario.
		Given("an SLA policy exists for high/feature with 1h completion at 50% at-risk threshold", func(s *Scenario) error {
			body := `{"items":[{"priority":"high","ticketType":"feature","firstResponseHours":1,"completionHours":1,"atRiskThresholdPct":50}]}`
			resp, err := s.Harness().APIRequest("PUT", fmt.Sprintf("/projects/%s/sla-policies", seed.ProjectID), strings.NewReader(body))
			if err != nil {
				return fmt.Errorf("put sla policies: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200 from PUT sla-policies, got %d", resp.StatusCode)
			}
			return nil
		}).
		When("I create a high-priority feature ticket", func(s *Scenario) error {
			body := fmt.Sprintf(`{"title":"SLA test ticket","type":"feature","priority":"high","storyId":"%s"}`, seed.StoryID)
			resp, err := s.Harness().APIRequest("POST", fmt.Sprintf("/projects/%s/tickets", seed.ProjectID), strings.NewReader(body))
			if err != nil {
				return fmt.Errorf("create ticket: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 201 {
				return fmt.Errorf("expected 201 creating ticket, got %d", resp.StatusCode)
			}
			return nil
		}).
		Then("the ticket list includes a non-nil slaStatus for the new ticket", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("GET", fmt.Sprintf("/projects/%s/tickets", seed.ProjectID), nil)
			if err != nil {
				return fmt.Errorf("list tickets: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200 listing tickets, got %d", resp.StatusCode)
			}
			var ticketList struct {
				Items []struct {
					Title     string  `json:"title"`
					SlaStatus *string `json:"slaStatus"`
				} `json:"items"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&ticketList); err != nil {
				return fmt.Errorf("decode ticket list: %w", err)
			}
			for _, tk := range ticketList.Items {
				if tk.Title == "SLA test ticket" {
					if tk.SlaStatus == nil {
						return fmt.Errorf("expected slaStatus to be non-nil for SLA test ticket")
					}
					if *tk.SlaStatus != "at_risk" && *tk.SlaStatus != "breached" && *tk.SlaStatus != "on_track" {
						return fmt.Errorf("unexpected slaStatus value: %q", *tk.SlaStatus)
					}
					return nil
				}
			}
			return fmt.Errorf("SLA test ticket not found in list")
		})
}

func TestSLAUIShowsSLATab(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()

	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInWithHarnessUser().
		ThenISeeSelectorKey("nav.board_tab").
		WhenIGoToRoute("settings", map[string]string{"projectId": seed.ProjectID}).
		ThenISeeSelectorKey("settings.sla_tab").
		WhenIClickKey("settings.sla_tab").
		ThenISeeSelectorKey("settings.sla_view")
}

func TestSLAViewerCannotUpdatePolicies(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t, WithViewerUser())
	defer scenario.Close()
	seed := scenario.SeedData()

	scenario.
		When("a viewer attempts to PUT SLA policies", func(s *Scenario) error {
			body := `{"items":[{"priority":"low","ticketType":"feature","firstResponseHours":48,"completionHours":120,"atRiskThresholdPct":80}]}`
			resp, err := s.Harness().APIRequest("PUT", fmt.Sprintf("/projects/%s/sla-policies", seed.ProjectID), strings.NewReader(body))
			if err != nil {
				return fmt.Errorf("put sla policies as viewer: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 403 {
				return fmt.Errorf("expected 403 for viewer on PUT sla-policies, got %d", resp.StatusCode)
			}
			return nil
		}).
		Then("the request is rejected with 403", func(s *Scenario) error {
			return nil
		})
}
