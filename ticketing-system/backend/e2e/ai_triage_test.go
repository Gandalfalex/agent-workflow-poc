//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestAiTriageToggleSuggestionAndDecision(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()
	h := scenario.Harness()

	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		WhenINavigateToSettings().
		ThenISeeSelectorKey("settings.ai_triage_toggle").
		WhenIClickKey("settings.ai_triage_toggle").
		WhenINavigateToBoard().
		WhenIClickKey("board.add_ticket_button").
		ThenISeeSelectorKey("new_ticket.modal").
		WhenIFillKey("new_ticket.title_input", "AI triage ticket").
		WhenIClickKey("new_ticket.ai_suggest_button").
		ThenISeeSelectorKey("new_ticket.ai_suggestion_panel")

	suggestionPayload := `{"title":"critical prod error","description":"sev1 outage in payment path","type":"bug"}`
	suggestionResp, err := h.APIRequest(
		http.MethodPost,
		fmt.Sprintf("/projects/%s/ai-triage/suggestions", seed.ProjectID),
		bytes.NewReader([]byte(suggestionPayload)),
	)
	if err != nil {
		t.Fatalf("suggestion request: %v", err)
	}
	defer suggestionResp.Body.Close()
	if suggestionResp.StatusCode != http.StatusCreated {
		t.Fatalf("suggestion status: got %d", suggestionResp.StatusCode)
	}
	var suggestion struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(suggestionResp.Body).Decode(&suggestion); err != nil {
		t.Fatalf("decode suggestion: %v", err)
	}
	if suggestion.ID == "" {
		t.Fatalf("missing suggestion id")
	}

	decisionPayload := `{"acceptedFields":["summary","priority"],"rejectedFields":["assignee","state"]}`
	decisionResp, err := h.APIRequest(
		http.MethodPost,
		fmt.Sprintf("/projects/%s/ai-triage/suggestions/%s/decision", seed.ProjectID, suggestion.ID),
		bytes.NewReader([]byte(decisionPayload)),
	)
	if err != nil {
		t.Fatalf("decision request: %v", err)
	}
	defer decisionResp.Body.Close()
	if decisionResp.StatusCode != http.StatusCreated {
		t.Fatalf("decision status: got %d", decisionResp.StatusCode)
	}
}
