//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestReleaseCRUD(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	h := scenario.Harness()
	seed := scenario.SeedData()

	// Create a release
	createBody := `{"name":"v1.0.0","version":"1.0.0","status":"planned","notes":"Initial release"}`
	createResp, err := h.APIRequest("POST", fmt.Sprintf("/projects/%s/releases", seed.ProjectID), strings.NewReader(createBody))
	if err != nil {
		t.Fatalf("create release: %v", err)
	}
	defer createResp.Body.Close()
	if createResp.StatusCode != 201 {
		t.Fatalf("expected 201 from POST releases, got %d", createResp.StatusCode)
	}

	var created struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Version string `json:"version"`
		Status  string `json:"status"`
	}
	if err := json.NewDecoder(createResp.Body).Decode(&created); err != nil {
		t.Fatalf("decode create release: %v", err)
	}
	if created.Name != "v1.0.0" {
		t.Errorf("expected name v1.0.0, got %s", created.Name)
	}
	if created.Status != "planned" {
		t.Errorf("expected status planned, got %s", created.Status)
	}
	if created.ID == "" {
		t.Fatal("expected non-empty release ID")
	}

	// List releases
	listResp, err := h.APIRequest("GET", fmt.Sprintf("/projects/%s/releases", seed.ProjectID), nil)
	if err != nil {
		t.Fatalf("list releases: %v", err)
	}
	defer listResp.Body.Close()
	if listResp.StatusCode != 200 {
		t.Fatalf("expected 200 from GET releases, got %d", listResp.StatusCode)
	}

	var list struct {
		Items []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"items"`
	}
	if err := json.NewDecoder(listResp.Body).Decode(&list); err != nil {
		t.Fatalf("decode list releases: %v", err)
	}
	found := false
	for _, r := range list.Items {
		if r.ID == created.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("created release not found in list")
	}

	// Update release
	updateBody := `{"status":"in_progress","notes":"Updated notes"}`
	updateResp, err := h.APIRequest("PATCH", fmt.Sprintf("/projects/%s/releases/%s", seed.ProjectID, created.ID), strings.NewReader(updateBody))
	if err != nil {
		t.Fatalf("update release: %v", err)
	}
	defer updateResp.Body.Close()
	if updateResp.StatusCode != 200 {
		t.Fatalf("expected 200 from PATCH release, got %d", updateResp.StatusCode)
	}

	var updated struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(updateResp.Body).Decode(&updated); err != nil {
		t.Fatalf("decode update release: %v", err)
	}
	if updated.Status != "in_progress" {
		t.Errorf("expected status in_progress, got %s", updated.Status)
	}

	// Delete release
	deleteResp, err := h.APIRequest("DELETE", fmt.Sprintf("/projects/%s/releases/%s", seed.ProjectID, created.ID), nil)
	if err != nil {
		t.Fatalf("delete release: %v", err)
	}
	defer deleteResp.Body.Close()
	if deleteResp.StatusCode != 204 {
		t.Fatalf("expected 204 from DELETE release, got %d", deleteResp.StatusCode)
	}
}

func TestReleaseExport(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	h := scenario.Harness()
	seed := scenario.SeedData()

	// Create a release
	createBody := `{"name":"Export Release","version":"2.0.0","status":"planned"}`
	createResp, err := h.APIRequest("POST", fmt.Sprintf("/projects/%s/releases", seed.ProjectID), strings.NewReader(createBody))
	if err != nil {
		t.Fatalf("create release: %v", err)
	}
	defer createResp.Body.Close()
	if createResp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d", createResp.StatusCode)
	}

	var created struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createResp.Body).Decode(&created); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// Export as markdown
	exportResp, err := h.APIRequest("GET", fmt.Sprintf("/projects/%s/releases/%s/export?format=markdown", seed.ProjectID, created.ID), nil)
	if err != nil {
		t.Fatalf("export release: %v", err)
	}
	defer exportResp.Body.Close()
	if exportResp.StatusCode != 200 {
		t.Fatalf("expected 200 from export, got %d", exportResp.StatusCode)
	}

	var export struct {
		Content string `json:"content"`
		Name    string `json:"name"`
	}
	if err := json.NewDecoder(exportResp.Body).Decode(&export); err != nil {
		t.Fatalf("decode export: %v", err)
	}
	if !strings.Contains(export.Content, "Export Release") {
		t.Errorf("expected markdown to contain release name, got: %s", export.Content)
	}

	// Export as JSON
	exportJSONResp, err := h.APIRequest("GET", fmt.Sprintf("/projects/%s/releases/%s/export?format=json", seed.ProjectID, created.ID), nil)
	if err != nil {
		t.Fatalf("export release json: %v", err)
	}
	defer exportJSONResp.Body.Close()
	if exportJSONResp.StatusCode != 200 {
		t.Fatalf("expected 200 from json export, got %d", exportJSONResp.StatusCode)
	}
}

func TestReleaseLinkedToTicket(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	h := scenario.Harness()
	seed := scenario.SeedData()

	// Create a release
	createBody := `{"name":"Linked Release","status":"planned"}`
	createResp, err := h.APIRequest("POST", fmt.Sprintf("/projects/%s/releases", seed.ProjectID), strings.NewReader(createBody))
	if err != nil {
		t.Fatalf("create release: %v", err)
	}
	defer createResp.Body.Close()
	if createResp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d", createResp.StatusCode)
	}
	var release struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createResp.Body).Decode(&release); err != nil {
		t.Fatalf("decode release: %v", err)
	}

	// Create a ticket to link
	ticketBody := fmt.Sprintf(`{"title":"Release test ticket","type":"feature","priority":"medium","storyId":"%s"}`, seed.StoryID)
	ticketCreateResp, err := h.APIRequest("POST", fmt.Sprintf("/projects/%s/tickets", seed.ProjectID), strings.NewReader(ticketBody))
	if err != nil {
		t.Fatalf("create ticket: %v", err)
	}
	defer ticketCreateResp.Body.Close()
	if ticketCreateResp.StatusCode != 201 {
		t.Fatalf("expected 201 creating ticket, got %d", ticketCreateResp.StatusCode)
	}
	var ticket struct {
		ID  string `json:"id"`
		Key string `json:"key"`
	}
	if err := json.NewDecoder(ticketCreateResp.Body).Decode(&ticket); err != nil {
		t.Fatalf("decode ticket: %v", err)
	}

	// Link ticket to release
	updateTicketBody := fmt.Sprintf(`{"releaseId":"%s"}`, release.ID)
	updateTicketResp, err := h.APIRequest("PATCH", fmt.Sprintf("/tickets/%s", ticket.ID), strings.NewReader(updateTicketBody))
	if err != nil {
		t.Fatalf("update ticket: %v", err)
	}
	defer updateTicketResp.Body.Close()
	if updateTicketResp.StatusCode != 200 {
		t.Fatalf("expected 200 from PATCH ticket, got %d", updateTicketResp.StatusCode)
	}

	var updatedTicket struct {
		ReleaseID *string `json:"releaseId"`
	}
	if err := json.NewDecoder(updateTicketResp.Body).Decode(&updatedTicket); err != nil {
		t.Fatalf("decode ticket: %v", err)
	}
	if updatedTicket.ReleaseID == nil || *updatedTicket.ReleaseID != release.ID {
		t.Errorf("expected ticket releaseId=%s, got %v", release.ID, updatedTicket.ReleaseID)
	}

	// Check that ticket appears in release export
	exportResp, err := h.APIRequest("GET", fmt.Sprintf("/projects/%s/releases/%s/export", seed.ProjectID, release.ID), nil)
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	defer exportResp.Body.Close()
	if exportResp.StatusCode != 200 {
		t.Fatalf("expected 200 from export, got %d", exportResp.StatusCode)
	}

	var export struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(exportResp.Body).Decode(&export); err != nil {
		t.Fatalf("decode export: %v", err)
	}
	if !strings.Contains(export.Content, ticket.Key) {
		t.Errorf("expected export to contain ticket key %s, content: %s", ticket.Key, export.Content)
	}
}

func TestReleaseViewerCanAccess(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	h := scenario.Harness()
	seed := scenario.SeedData()

	// Create release as admin
	createBody := `{"name":"Viewer Test Release","status":"planned"}`
	createResp, err := h.APIRequest("POST", fmt.Sprintf("/projects/%s/releases", seed.ProjectID), strings.NewReader(createBody))
	if err != nil {
		t.Fatalf("create release: %v", err)
	}
	defer createResp.Body.Close()
	if createResp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d", createResp.StatusCode)
	}

	// Viewer can list releases
	listResp, err := h.APIRequest("GET", fmt.Sprintf("/projects/%s/releases", seed.ProjectID), nil)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	defer listResp.Body.Close()
	if listResp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", listResp.StatusCode)
	}
}
