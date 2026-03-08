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
	seed := scenario.SeedData()

	var created struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		Label     string `json:"label"`
		FieldType string `json:"fieldType"`
		Required  bool   `json:"required"`
	}

	scenario.
		When("I create a text custom field named customer_tier", func(s *Scenario) error {
			body := `{"name":"customer_tier","label":"Customer Tier","fieldType":"text","required":false}`
			resp, err := s.Harness().APIRequest("POST", fmt.Sprintf("/projects/%s/custom-fields", seed.ProjectID), strings.NewReader(body))
			if err != nil {
				return fmt.Errorf("create custom field: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 201 {
				return fmt.Errorf("expected 201 creating custom field, got %d", resp.StatusCode)
			}
			return json.NewDecoder(resp.Body).Decode(&created)
		}).
		Then("the created field has name customer_tier and fieldType text", func(s *Scenario) error {
			if created.Name != "customer_tier" {
				return fmt.Errorf("expected name customer_tier, got %s", created.Name)
			}
			if created.FieldType != "text" {
				return fmt.Errorf("expected fieldType text, got %s", created.FieldType)
			}
			return nil
		}).
		And("the field appears in the custom fields list", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("GET", fmt.Sprintf("/projects/%s/custom-fields", seed.ProjectID), nil)
			if err != nil {
				return fmt.Errorf("list custom fields: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200 listing custom fields, got %d", resp.StatusCode)
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
			for _, f := range list.Items {
				if f.Name == "customer_tier" {
					return nil
				}
			}
			return fmt.Errorf("expected customer_tier in custom fields list, got %d items", len(list.Items))
		})
}

func TestCustomFieldsTicketValues(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()
	seed := scenario.SeedData()

	var ticketID string
	var cfID string

	scenario.
		Given("a ticket exists in the project", func(s *Scenario) error {
			projectID := uuid.MustParse(seed.ProjectID)
			storyID := uuid.MustParse(seed.StoryID)
			backlogID := uuid.MustParse(seed.BacklogID)
			st := s.Harness().Store()
			ctx := s.Harness().Context()
			ticket, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
				Title:   "Custom Field Value Test Ticket",
				Type:    "feature",
				StoryID: storyID,
				StateID: &backlogID,
			})
			if err != nil {
				return fmt.Errorf("create ticket: %w", err)
			}
			ticketID = ticket.ID.String()
			return nil
		}).
		And("an enum custom field impact_level exists", func(s *Scenario) error {
			body := `{"name":"impact_level","label":"Impact Level","fieldType":"enum","options":[{"value":"low","label":"Low"},{"value":"high","label":"High"}]}`
			resp, err := s.Harness().APIRequest("POST", fmt.Sprintf("/projects/%s/custom-fields", seed.ProjectID), strings.NewReader(body))
			if err != nil {
				return fmt.Errorf("create custom field: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 201 {
				return fmt.Errorf("expected 201, got %d", resp.StatusCode)
			}
			var cf struct {
				ID string `json:"id"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&cf); err != nil {
				return fmt.Errorf("decode field: %w", err)
			}
			cfID = cf.ID
			return nil
		}).
		When("I set impact_level=high on the ticket", func(s *Scenario) error {
			body := fmt.Sprintf(`{"values":[{"fieldId":"%s","valueText":"high"}]}`, cfID)
			resp, err := s.Harness().APIRequest("PUT", fmt.Sprintf("/tickets/%s/custom-field-values", ticketID), strings.NewReader(body))
			if err != nil {
				return fmt.Errorf("set custom field values: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200 setting values, got %d", resp.StatusCode)
			}
			return nil
		}).
		Then("GET custom-field-values returns impact_level=high for the ticket", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("GET", fmt.Sprintf("/tickets/%s/custom-field-values", ticketID), nil)
			if err != nil {
				return fmt.Errorf("get custom field values: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200 getting values, got %d", resp.StatusCode)
			}
			var values struct {
				Items []struct {
					FieldID   string  `json:"fieldId"`
					FieldName string  `json:"fieldName"`
					ValueText *string `json:"valueText"`
				} `json:"items"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&values); err != nil {
				return fmt.Errorf("decode values: %w", err)
			}
			for _, v := range values.Items {
				if v.FieldID == cfID && v.ValueText != nil && *v.ValueText == "high" {
					return nil
				}
			}
			return fmt.Errorf("expected impact_level=high in ticket values, got %d items", len(values.Items))
		})
}

func TestCustomFieldsViewerCanAccess(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t, WithViewerUser())
	defer scenario.Close()
	seed := scenario.SeedData()

	scenario.
		When("a viewer lists custom fields", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("GET", fmt.Sprintf("/projects/%s/custom-fields", seed.ProjectID), nil)
			if err != nil {
				return fmt.Errorf("list custom fields as viewer: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200 for viewer, got %d", resp.StatusCode)
			}
			return nil
		}).
		Then("the request succeeds with 200", func(s *Scenario) error {
			return nil
		})
}

func TestCustomFieldsViewerCannotCreate(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t, WithViewerUser())
	defer scenario.Close()
	seed := scenario.SeedData()

	scenario.
		When("a viewer attempts to create a custom field", func(s *Scenario) error {
			body := `{"name":"restricted_field","label":"Restricted","fieldType":"text"}`
			resp, err := s.Harness().APIRequest("POST", fmt.Sprintf("/projects/%s/custom-fields", seed.ProjectID), strings.NewReader(body))
			if err != nil {
				return fmt.Errorf("create custom field as viewer: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 403 {
				return fmt.Errorf("expected 403 for viewer create, got %d", resp.StatusCode)
			}
			return nil
		}).
		Then("the request is rejected with 403", func(s *Scenario) error {
			return nil
		})
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
