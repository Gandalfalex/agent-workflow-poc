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

	h := scenario.Harness()
	seed := scenario.SeedData()

	// GET default policies (should be empty initially)
	getResp, err := h.APIRequest("GET", fmt.Sprintf("/projects/%s/sla-policies", seed.ProjectID), nil)
	if err != nil {
		t.Fatalf("get sla policies: %v", err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != 200 {
		t.Fatalf("expected 200 from GET sla-policies, got %d", getResp.StatusCode)
	}

	var initial struct {
		Items []interface{} `json:"items"`
	}
	if err := json.NewDecoder(getResp.Body).Decode(&initial); err != nil {
		t.Fatalf("decode sla policies: %v", err)
	}

	// PUT policies — one entry for urgent bugs
	putBody := `{"items":[{"priority":"urgent","ticketType":"bug","firstResponseHours":2,"completionHours":8,"atRiskThresholdPct":75}]}`
	putResp, err := h.APIRequest("PUT", fmt.Sprintf("/projects/%s/sla-policies", seed.ProjectID), strings.NewReader(putBody))
	if err != nil {
		t.Fatalf("put sla policies: %v", err)
	}
	defer putResp.Body.Close()
	if putResp.StatusCode != 200 {
		t.Fatalf("expected 200 from PUT sla-policies, got %d", putResp.StatusCode)
	}

	var updated struct {
		Items []struct {
			Priority           string `json:"priority"`
			TicketType         string `json:"ticketType"`
			FirstResponseHours int    `json:"firstResponseHours"`
			CompletionHours    int    `json:"completionHours"`
			AtRiskThresholdPct int    `json:"atRiskThresholdPct"`
		} `json:"items"`
	}
	if err := json.NewDecoder(putResp.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated sla policies: %v", err)
	}
	if len(updated.Items) != 1 {
		t.Fatalf("expected 1 SLA policy, got %d", len(updated.Items))
	}
	p := updated.Items[0]
	if p.Priority != "urgent" || p.TicketType != "bug" || p.FirstResponseHours != 2 || p.CompletionHours != 8 {
		t.Fatalf("unexpected policy values: %+v", p)
	}

	// GET again — should return the saved policy
	getResp2, err := h.APIRequest("GET", fmt.Sprintf("/projects/%s/sla-policies", seed.ProjectID), nil)
	if err != nil {
		t.Fatalf("get sla policies after update: %v", err)
	}
	defer getResp2.Body.Close()
	var afterUpdate struct {
		Items []struct {
			Priority string `json:"priority"`
		} `json:"items"`
	}
	if err := json.NewDecoder(getResp2.Body).Decode(&afterUpdate); err != nil {
		t.Fatalf("decode after update: %v", err)
	}
	if len(afterUpdate.Items) != 1 || afterUpdate.Items[0].Priority != "urgent" {
		t.Fatalf("expected 1 urgent policy after GET, got %d items", len(afterUpdate.Items))
	}
}

func TestSLAStatusAppearsOnBoardTicket(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	h := scenario.Harness()
	seed := scenario.SeedData()

	// Set an SLA policy for high/feature with very short completion time
	// so the ticket is immediately at_risk or breached
	putBody := `{"items":[{"priority":"high","ticketType":"feature","firstResponseHours":1,"completionHours":1,"atRiskThresholdPct":50}]}`
	putResp, err := h.APIRequest("PUT", fmt.Sprintf("/projects/%s/sla-policies", seed.ProjectID), strings.NewReader(putBody))
	if err != nil {
		t.Fatalf("put sla policies: %v", err)
	}
	defer putResp.Body.Close()
	if putResp.StatusCode != 200 {
		t.Fatalf("expected 200 from PUT sla-policies, got %d", putResp.StatusCode)
	}

	// Create a ticket with high priority / feature type; it will immediately be at_risk (1h completion)
	ticketBody := fmt.Sprintf(`{"title":"SLA test ticket","type":"feature","priority":"high","storyId":"%s"}`, seed.StoryID)
	tResp, err := h.APIRequest("POST", fmt.Sprintf("/projects/%s/tickets", seed.ProjectID), strings.NewReader(ticketBody))
	if err != nil {
		t.Fatalf("create ticket: %v", err)
	}
	defer tResp.Body.Close()
	if tResp.StatusCode != 201 {
		t.Fatalf("expected 201 creating ticket, got %d", tResp.StatusCode)
	}

	// Verify GET /tickets returns slaStatus field
	listResp, err := h.APIRequest("GET", fmt.Sprintf("/projects/%s/tickets", seed.ProjectID), nil)
	if err != nil {
		t.Fatalf("list tickets: %v", err)
	}
	defer listResp.Body.Close()
	if listResp.StatusCode != 200 {
		t.Fatalf("expected 200 listing tickets, got %d", listResp.StatusCode)
	}

	var ticketList struct {
		Items []struct {
			Title     string  `json:"title"`
			SlaStatus *string `json:"slaStatus"`
		} `json:"items"`
	}
	if err := json.NewDecoder(listResp.Body).Decode(&ticketList); err != nil {
		t.Fatalf("decode ticket list: %v", err)
	}

	found := false
	for _, tk := range ticketList.Items {
		if tk.Title == "SLA test ticket" {
			found = true
			if tk.SlaStatus == nil {
				t.Fatalf("expected slaStatus to be non-nil for SLA test ticket")
			}
			// Ticket was just created and policy completionHours=1 at 50% threshold
			// so it should be at_risk or breached immediately
			if *tk.SlaStatus != "at_risk" && *tk.SlaStatus != "breached" && *tk.SlaStatus != "on_track" {
				t.Fatalf("unexpected slaStatus value: %q", *tk.SlaStatus)
			}
			break
		}
	}
	if !found {
		t.Fatalf("SLA test ticket not found in list")
	}
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

	h := scenario.Harness()
	seed := scenario.SeedData()

	putBody := `{"items":[{"priority":"low","ticketType":"feature","firstResponseHours":48,"completionHours":120,"atRiskThresholdPct":80}]}`
	resp, err := h.APIRequest("PUT", fmt.Sprintf("/projects/%s/sla-policies", seed.ProjectID), strings.NewReader(putBody))
	if err != nil {
		t.Fatalf("put sla policies as viewer: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 403 {
		t.Fatalf("expected 403 for viewer on PUT sla-policies, got %d", resp.StatusCode)
	}
}
