//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

// postAiTriageSuggestion creates an AI triage suggestion via the API and returns its ID.
func postAiTriageSuggestion(s *Scenario, projectID, payload string) (string, error) {
	resp, err := s.Harness().APIRequest(
		http.MethodPost,
		fmt.Sprintf("/projects/%s/ai-triage/suggestions", projectID),
		bytes.NewReader([]byte(payload)),
	)
	if err != nil {
		return "", fmt.Errorf("suggestion request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("suggestion status: got %d", resp.StatusCode)
	}
	var suggestion struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&suggestion); err != nil {
		return "", fmt.Errorf("decode suggestion: %w", err)
	}
	if suggestion.ID == "" {
		return "", fmt.Errorf("missing suggestion id")
	}
	return suggestion.ID, nil
}

// postAiTriageDecision submits a decision for an AI triage suggestion.
func postAiTriageDecision(s *Scenario, projectID, suggestionID, payload string) error {
	resp, err := s.Harness().APIRequest(
		http.MethodPost,
		fmt.Sprintf("/projects/%s/ai-triage/suggestions/%s/decision", projectID, suggestionID),
		bytes.NewReader([]byte(payload)),
	)
	if err != nil {
		return fmt.Errorf("decision request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("decision status: got %d", resp.StatusCode)
	}
	return nil
}

func TestAiTriageToggleSuggestionAndDecision(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()

	var suggestionID string

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
		ThenISeeSelectorKey("new_ticket.ai_suggestion_panel").
		When("I submit an AI triage suggestion via the API", func(s *Scenario) error {
			id, err := postAiTriageSuggestion(s, seed.ProjectID,
				`{"title":"critical prod error","description":"sev1 outage in payment path","type":"bug"}`)
			if err != nil {
				return err
			}
			suggestionID = id
			return nil
		}).
		Then("the suggestion is created with a valid ID", func(s *Scenario) error {
			if suggestionID == "" {
				return fmt.Errorf("expected non-empty suggestion ID")
			}
			return nil
		}).
		When("I submit a decision for the suggestion", func(s *Scenario) error {
			return postAiTriageDecision(s, seed.ProjectID, suggestionID,
				`{"acceptedFields":["summary","priority"],"rejectedFields":["assignee","state"]}`)
		}).
		Then("the decision is accepted with status 201", func(s *Scenario) error {
			// Decision was validated inside postAiTriageDecision; reaching here means success.
			return nil
		})
}
