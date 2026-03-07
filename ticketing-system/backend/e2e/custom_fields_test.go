//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"ticketing-system/backend/internal/store"
)

func TestCustomFieldsAPICreateAndList(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	h := scenario.Harness()
	seed := scenario.SeedData()

	// Create a text custom field
	body := fmt.Sprintf(`{"name":"customer_tier","label":"Customer Tier","fieldType":"text","required":false}`)
	resp, err := h.APIRequest("POST", fmt.Sprintf("/projects/%s/custom-fields", seed.ProjectID), strings.NewReader(body))
	if err != nil {
		t.Fatalf("create custom field: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		t.Fatalf("expected 201 creating custom field, got %d", resp.StatusCode)
	}

	var created struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		Label     string `json:"label"`
		FieldType string `json:"fieldType"`
		Required  bool   `json:"required"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("decode created field: %v", err)
	}
	if created.Name != "customer_tier" {
		t.Fatalf("expected name customer_tier, got %s", created.Name)
	}
	if created.FieldType != "text" {
		t.Fatalf("expected fieldType text, got %s", created.FieldType)
	}

	// List custom fields
	listResp, err := h.APIRequest("GET", fmt.Sprintf("/projects/%s/custom-fields", seed.ProjectID), nil)
	if err != nil {
		t.Fatalf("list custom fields: %v", err)
	}
	defer listResp.Body.Close()
	if listResp.StatusCode != 200 {
		t.Fatalf("expected 200 listing custom fields, got %d", listResp.StatusCode)
	}

	var list struct {
		Items []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"items"`
	}
	if err := json.NewDecoder(listResp.Body).Decode(&list); err != nil {
		t.Fatalf("decode list: %v", err)
	}

	found := false
	for _, f := range list.Items {
		if f.Name == "customer_tier" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected customer_tier in custom fields list, got %d items", len(list.Items))
	}
}

func TestCustomFieldsTicketValues(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	h := scenario.Harness()
	seed := scenario.SeedData()
	st := h.Store()
	ctx := h.Context()

	// Create a ticket to attach values to
	projectID := uuid.MustParse(seed.ProjectID)
	storyID := uuid.MustParse(seed.StoryID)
	backlogID := uuid.MustParse(seed.BacklogID)
	ticket, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
		Title:   "Custom Field Value Test Ticket",
		Type:    "feature",
		StoryID: storyID,
		StateID: &backlogID,
	})
	if err != nil {
		t.Fatalf("create ticket: %v", err)
	}
	ticketID := ticket.ID.String()

	// Create a custom field
	cfBody := `{"name":"impact_level","label":"Impact Level","fieldType":"enum","options":[{"value":"low","label":"Low"},{"value":"high","label":"High"}]}`
	cfResp, err := h.APIRequest("POST", fmt.Sprintf("/projects/%s/custom-fields", seed.ProjectID), strings.NewReader(cfBody))
	if err != nil {
		t.Fatalf("create custom field: %v", err)
	}
	defer cfResp.Body.Close()
	if cfResp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d", cfResp.StatusCode)
	}
	var cf struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(cfResp.Body).Decode(&cf); err != nil {
		t.Fatalf("decode field: %v", err)
	}

	// Set custom field value on ticket
	setBody := fmt.Sprintf(`{"values":[{"fieldId":"%s","valueText":"high"}]}`, cf.ID)
	setResp, err := h.APIRequest("PUT", fmt.Sprintf("/tickets/%s/custom-field-values", ticketID), strings.NewReader(setBody))
	if err != nil {
		t.Fatalf("set custom field values: %v", err)
	}
	defer setResp.Body.Close()
	if setResp.StatusCode != 200 {
		t.Fatalf("expected 200 setting values, got %d", setResp.StatusCode)
	}

	// Get custom field values
	getResp, err := h.APIRequest("GET", fmt.Sprintf("/tickets/%s/custom-field-values", ticketID), nil)
	if err != nil {
		t.Fatalf("get custom field values: %v", err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != 200 {
		t.Fatalf("expected 200 getting values, got %d", getResp.StatusCode)
	}

	var values struct {
		Items []struct {
			FieldID   string  `json:"fieldId"`
			FieldName string  `json:"fieldName"`
			ValueText *string `json:"valueText"`
		} `json:"items"`
	}
	if err := json.NewDecoder(getResp.Body).Decode(&values); err != nil {
		t.Fatalf("decode values: %v", err)
	}

	found := false
	for _, v := range values.Items {
		if v.FieldID == cf.ID && v.ValueText != nil && *v.ValueText == "high" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected impact_level=high in ticket values, got %d items", len(values.Items))
	}
}

func TestCustomFieldsViewerCanAccess(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t, WithViewerUser())
	defer scenario.Close()

	h := scenario.Harness()
	seed := scenario.SeedData()

	// Viewer can list custom fields
	resp, err := h.APIRequest("GET", fmt.Sprintf("/projects/%s/custom-fields", seed.ProjectID), nil)
	if err != nil {
		t.Fatalf("list custom fields as viewer: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 for viewer, got %d", resp.StatusCode)
	}
}

func TestCustomFieldsViewerCannotCreate(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t, WithViewerUser())
	defer scenario.Close()

	h := scenario.Harness()
	seed := scenario.SeedData()

	// Viewer cannot create custom fields
	body := `{"name":"restricted_field","label":"Restricted","fieldType":"text"}`
	resp, err := h.APIRequest("POST", fmt.Sprintf("/projects/%s/custom-fields", seed.ProjectID), strings.NewReader(body))
	if err != nil {
		t.Fatalf("create custom field as viewer: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 403 {
		t.Fatalf("expected 403 for viewer create, got %d", resp.StatusCode)
	}
}

func TestCustomFieldsUISettingsTab(t *testing.T) {
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
		ThenISeeSelectorKey("settings.custom_fields_tab").
		WhenIClickKey("settings.custom_fields_tab").
		ThenISeeSelectorKey("settings.custom_fields_view")
}
