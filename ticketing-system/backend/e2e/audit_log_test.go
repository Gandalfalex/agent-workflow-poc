//go:build e2e

package e2e

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAuditLogRecordsProjectCreation(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	h := scenario.Harness()

	// Create a project to generate an audit entry
	body := `{"key":"AUD1","name":"Audit Test Project","description":"for audit log test"}`
	resp, err := h.APIRequest("POST", "/projects", strings.NewReader(body))
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		t.Fatalf("expected 201 creating project, got %d", resp.StatusCode)
	}

	// Query audit log
	auditResp, err := h.APIRequest("GET", "/admin/audit-log?limit=50", nil)
	if err != nil {
		t.Fatalf("get audit log: %v", err)
	}
	defer auditResp.Body.Close()
	if auditResp.StatusCode != 200 {
		t.Fatalf("expected 200 from audit log, got %d", auditResp.StatusCode)
	}

	var result struct {
		Items []struct {
			Action       string `json:"action"`
			ResourceType string `json:"resourceType"`
			ResourceName string `json:"resourceName"`
		} `json:"items"`
		Total int `json:"total"`
	}
	if err := json.NewDecoder(auditResp.Body).Decode(&result); err != nil {
		t.Fatalf("decode audit log: %v", err)
	}

	found := false
	for _, entry := range result.Items {
		if entry.Action == "project.created" && entry.ResourceType == "project" &&
			strings.Contains(entry.ResourceName, "AUD1") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected project.created audit entry for AUD1, got %d entries", len(result.Items))
	}
}

func TestAuditLogViewerCannotAccess(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t, WithViewerUser())
	defer scenario.Close()

	h := scenario.Harness()

	resp, err := h.APIRequest("GET", "/admin/audit-log", nil)
	if err != nil {
		t.Fatalf("audit log request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 403 {
		t.Fatalf("expected 403 for viewer on audit log, got %d", resp.StatusCode)
	}
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
