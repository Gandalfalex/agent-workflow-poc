//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestAuditLogRecordsProjectCreation(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	var auditItems []struct {
		Action       string `json:"action"`
		ResourceType string `json:"resourceType"`
		ResourceName string `json:"resourceName"`
	}

	scenario.
		When("I create a project to generate an audit entry", func(s *Scenario) error {
			body := `{"key":"AUD1","name":"Audit Test Project","description":"for audit log test"}`
			resp, err := s.Harness().APIRequest("POST", "/projects", strings.NewReader(body))
			if err != nil {
				return fmt.Errorf("create project: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 201 {
				return fmt.Errorf("expected 201 creating project, got %d", resp.StatusCode)
			}
			return nil
		}).
		When("I query the audit log", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("GET", "/admin/audit-log?limit=50", nil)
			if err != nil {
				return fmt.Errorf("get audit log: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200 from audit log, got %d", resp.StatusCode)
			}
			var result struct {
				Items []struct {
					Action       string `json:"action"`
					ResourceType string `json:"resourceType"`
					ResourceName string `json:"resourceName"`
				} `json:"items"`
				Total int `json:"total"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return fmt.Errorf("decode audit log: %w", err)
			}
			auditItems = result.Items
			return nil
		}).
		Then("a project.created audit entry for AUD1 is present", func(s *Scenario) error {
			for _, entry := range auditItems {
				if entry.Action == "project.created" && entry.ResourceType == "project" &&
					strings.Contains(entry.ResourceName, "AUD1") {
					return nil
				}
			}
			return fmt.Errorf("expected project.created audit entry for AUD1, got %d entries", len(auditItems))
		})
}

func TestAuditLogViewerCannotAccess(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t, WithViewerUser())
	defer scenario.Close()

	scenario.
		When("a viewer requests the audit log", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("GET", "/admin/audit-log", nil)
			if err != nil {
				return fmt.Errorf("audit log request: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 403 {
				return fmt.Errorf("expected 403 for viewer on audit log, got %d", resp.StatusCode)
			}
			return nil
		}).
		Then("access is denied", func(s *Scenario) error {
			return nil
		})
}

func TestAuditLogUIShowsAuditTab(t *testing.T) {
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
		ThenISeeSelectorKey("settings.audit_tab").
		WhenIClickKey("settings.audit_tab").
		ThenISeeSelectorKey("settings.audit_view")
}
