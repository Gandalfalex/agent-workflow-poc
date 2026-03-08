//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// createRelease is a helper that POSTs a new release and returns its ID.
func createRelease(s *Scenario, projectID, body string) (string, error) {
	resp, err := s.Harness().APIRequest("POST", fmt.Sprintf("/projects/%s/releases", projectID), strings.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("POST release: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		return "", fmt.Errorf("expected 201, got %d", resp.StatusCode)
	}
	var created struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}
	return created.ID, nil
}

func TestReleaseCRUD(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()

	var releaseID string
	var createdName, createdStatus string

	scenario.
		When("I create a release v1.0.0 with status planned", func(s *Scenario) error {
			body := `{"name":"v1.0.0","version":"1.0.0","status":"planned","notes":"Initial release"}`
			resp, err := s.Harness().APIRequest("POST", fmt.Sprintf("/projects/%s/releases", seed.ProjectID), strings.NewReader(body))
			if err != nil {
				return fmt.Errorf("create release: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 201 {
				return fmt.Errorf("expected 201, got %d", resp.StatusCode)
			}
			var created struct {
				ID      string `json:"id"`
				Name    string `json:"name"`
				Version string `json:"version"`
				Status  string `json:"status"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
				return fmt.Errorf("decode: %w", err)
			}
			releaseID = created.ID
			createdName = created.Name
			createdStatus = created.Status
			return nil
		}).
		Then("the release is returned with the correct name and status", func(s *Scenario) error {
			if createdName != "v1.0.0" {
				return fmt.Errorf("expected name v1.0.0, got %s", createdName)
			}
			if createdStatus != "planned" {
				return fmt.Errorf("expected status planned, got %s", createdStatus)
			}
			if releaseID == "" {
				return fmt.Errorf("expected non-empty release ID")
			}
			return nil
		}).
		And("the release appears in the project release list", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("GET", fmt.Sprintf("/projects/%s/releases", seed.ProjectID), nil)
			if err != nil {
				return fmt.Errorf("list releases: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200, got %d", resp.StatusCode)
			}
			var list struct {
				Items []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"items"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
				return fmt.Errorf("decode list: %w", err)
			}
			for _, r := range list.Items {
				if r.ID == releaseID {
					return nil
				}
			}
			return fmt.Errorf("created release not found in list")
		}).
		When("I update the release status to in_progress", func(s *Scenario) error {
			body := `{"status":"in_progress","notes":"Updated notes"}`
			resp, err := s.Harness().APIRequest("PATCH", fmt.Sprintf("/projects/%s/releases/%s", seed.ProjectID, releaseID), strings.NewReader(body))
			if err != nil {
				return fmt.Errorf("update release: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200, got %d", resp.StatusCode)
			}
			var updated struct {
				Status string `json:"status"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
				return fmt.Errorf("decode update: %w", err)
			}
			if updated.Status != "in_progress" {
				return fmt.Errorf("expected status in_progress, got %s", updated.Status)
			}
			return nil
		}).
		Then("the release is deleted successfully", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("DELETE", fmt.Sprintf("/projects/%s/releases/%s", seed.ProjectID, releaseID), nil)
			if err != nil {
				return fmt.Errorf("delete release: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 204 {
				return fmt.Errorf("expected 204, got %d", resp.StatusCode)
			}
			return nil
		})
}

func TestReleaseExport(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()

	var releaseID string

	scenario.
		Given("a release named Export Release exists", func(s *Scenario) error {
			var err error
			releaseID, err = createRelease(s, seed.ProjectID, `{"name":"Export Release","version":"2.0.0","status":"planned"}`)
			return err
		}).
		When("I export the release as markdown", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("GET", fmt.Sprintf("/projects/%s/releases/%s/export?format=markdown", seed.ProjectID, releaseID), nil)
			if err != nil {
				return fmt.Errorf("export release: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200, got %d", resp.StatusCode)
			}
			var export struct {
				Content string `json:"content"`
				Name    string `json:"name"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&export); err != nil {
				return fmt.Errorf("decode export: %w", err)
			}
			if !strings.Contains(export.Content, "Export Release") {
				return fmt.Errorf("expected markdown to contain release name, got: %s", export.Content)
			}
			return nil
		}).
		And("I export the release as JSON", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("GET", fmt.Sprintf("/projects/%s/releases/%s/export?format=json", seed.ProjectID, releaseID), nil)
			if err != nil {
				return fmt.Errorf("export release json: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200 from json export, got %d", resp.StatusCode)
			}
			return nil
		})
}

func TestReleaseLinkedToTicket(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()

	var releaseID string
	var ticketID, ticketKey string

	scenario.
		Given("a release named Linked Release exists", func(s *Scenario) error {
			var err error
			releaseID, err = createRelease(s, seed.ProjectID, `{"name":"Linked Release","status":"planned"}`)
			return err
		}).
		And("a ticket exists in the project", func(s *Scenario) error {
			body := fmt.Sprintf(`{"title":"Release test ticket","type":"feature","priority":"medium","storyId":"%s"}`, seed.StoryID)
			resp, err := s.Harness().APIRequest("POST", fmt.Sprintf("/projects/%s/tickets", seed.ProjectID), strings.NewReader(body))
			if err != nil {
				return fmt.Errorf("create ticket: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 201 {
				return fmt.Errorf("expected 201 creating ticket, got %d", resp.StatusCode)
			}
			var ticket struct {
				ID  string `json:"id"`
				Key string `json:"key"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&ticket); err != nil {
				return fmt.Errorf("decode ticket: %w", err)
			}
			ticketID = ticket.ID
			ticketKey = ticket.Key
			return nil
		}).
		When("I assign the ticket to the release", func(s *Scenario) error {
			body := fmt.Sprintf(`{"releaseId":"%s"}`, releaseID)
			resp, err := s.Harness().APIRequest("PATCH", fmt.Sprintf("/tickets/%s", ticketID), strings.NewReader(body))
			if err != nil {
				return fmt.Errorf("update ticket: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200 from PATCH ticket, got %d", resp.StatusCode)
			}
			var updatedTicket struct {
				ReleaseID *string `json:"releaseId"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&updatedTicket); err != nil {
				return fmt.Errorf("decode ticket: %w", err)
			}
			if updatedTicket.ReleaseID == nil || *updatedTicket.ReleaseID != releaseID {
				return fmt.Errorf("expected ticket releaseId=%s, got %v", releaseID, updatedTicket.ReleaseID)
			}
			return nil
		}).
		Then("the ticket key appears in the release export", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("GET", fmt.Sprintf("/projects/%s/releases/%s/export", seed.ProjectID, releaseID), nil)
			if err != nil {
				return fmt.Errorf("export: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200 from export, got %d", resp.StatusCode)
			}
			var export struct {
				Content string `json:"content"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&export); err != nil {
				return fmt.Errorf("decode export: %w", err)
			}
			if !strings.Contains(export.Content, ticketKey) {
				return fmt.Errorf("expected export to contain ticket key %s, content: %s", ticketKey, export.Content)
			}
			return nil
		})
}

func TestReleaseViewerCanAccess(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()

	scenario.
		Given("a release exists in the project", func(s *Scenario) error {
			_, err := createRelease(s, seed.ProjectID, `{"name":"Viewer Test Release","status":"planned"}`)
			return err
		}).
		When("I list releases", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("GET", fmt.Sprintf("/projects/%s/releases", seed.ProjectID), nil)
			if err != nil {
				return fmt.Errorf("list: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200, got %d", resp.StatusCode)
			}
			return nil
		}).
		Then("the request succeeds", func(s *Scenario) error {
			return nil
		})
}
