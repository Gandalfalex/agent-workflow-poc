//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

type boardFilterPresetAPI struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	ShareToken *string `json:"shareToken"`
	Filters    struct {
		AssigneeID *string `json:"assigneeId"`
		StateID    *string `json:"stateId"`
		Priority   *string `json:"priority"`
		Type       *string `json:"type"`
		Q          *string `json:"q"`
	} `json:"filters"`
}

type boardFilterPresetListAPI struct {
	Items []boardFilterPresetAPI `json:"items"`
}

func TestBoardFilterPresetFlow(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()
	searchQuery := fmt.Sprintf("Preset Search %d", time.Now().UnixNano())
	presetName := fmt.Sprintf("Preset %d", time.Now().UnixNano())
	renamedPreset := fmt.Sprintf("Renamed %d", time.Now().UnixNano())

	presetID := ""
	shareToken := ""

	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/"+seed.ProjectID+"/board").
		WhenIClickKey("board.filter_toggle_button").
		WhenISelectOptionByValueKey("board.filter_priority_select", "high").
		WhenIFillKey("board.filter_search_input", searchQuery).
		WhenIClickKey("board.preset_open_editor_button").
		WhenIFillKey("board.preset_name_input", presetName).
		WhenIClickKey("board.preset_save_button").
		Then("preset saved via API", func(s *Scenario) error {
			list, err := waitForPresetByName(s.Harness(), seed.ProjectID, presetName, 5*time.Second)
			if err != nil {
				return err
			}
			var preset *boardFilterPresetAPI
			for i := range list.Items {
				if list.Items[i].Name == presetName {
					preset = &list.Items[i]
					break
				}
			}
			if preset == nil {
				return fmt.Errorf("preset %q not found after save", presetName)
			}
			presetID = preset.ID
			if preset.Name != presetName {
				return fmt.Errorf("expected preset name %q, got %q", presetName, preset.Name)
			}
			if preset.Filters.Priority == nil || *preset.Filters.Priority != "high" {
				return fmt.Errorf("expected priority high in saved preset")
			}
			if preset.Filters.Q == nil || *preset.Filters.Q != searchQuery {
				return fmt.Errorf("expected query %q in saved preset", searchQuery)
			}
			return nil
		}).
		WhenISelectOptionByValueKey("board.filter_priority_select", "low").
		WhenIFillKey("board.filter_search_input", "").
		When("I apply the saved preset", func(s *Scenario) error {
			return s.Harness().SelectOptionByValueKey("board.preset_select", presetID)
		}).
		Then("saved preset reapplies board filters", func(s *Scenario) error {
			priority, err := inputValueByKey(s.Harness(), "board.filter_priority_select")
			if err != nil {
				return err
			}
			if priority != "high" {
				return fmt.Errorf("expected priority high after preset apply, got %q", priority)
			}
			query, err := inputValueByKey(s.Harness(), "board.filter_search_input")
			if err != nil {
				return err
			}
			if query != searchQuery {
				return fmt.Errorf("expected query %q after preset apply, got %q", searchQuery, query)
			}
			return nil
		}).
		WhenIClickKey("board.preset_open_editor_button").
		WhenIFillKey("board.preset_name_input", renamedPreset).
		WhenIClickKey("board.preset_rename_button").
		Then("preset rename persists via API", func(s *Scenario) error {
			name, err := waitForPresetNameByID(s.Harness(), seed.ProjectID, presetID, renamedPreset, 5*time.Second)
			if err != nil {
				return err
			}
			if name != renamedPreset {
				return fmt.Errorf("expected renamed preset %q, got %q", renamedPreset, name)
			}
			return nil
		}).
		WhenIClickKey("board.preset_share_button").
		ThenURLContains("/projects/"+seed.ProjectID+"/board?share=").
		Then("share token is present in URL", func(s *Scenario) error {
			parsed, err := url.Parse(s.Harness().page.URL())
			if err != nil {
				return err
			}
			shareToken = strings.TrimSpace(parsed.Query().Get("share"))
			if shareToken == "" {
				return fmt.Errorf("expected share token in URL")
			}
			return nil
		}).
		Then("shared token resolves preset via API", func(s *Scenario) error {
			preset, err := getSharedBoardFilterPreset(s.Harness(), seed.ProjectID, shareToken)
			if err != nil {
				return err
			}
			if preset.ID != presetID {
				return fmt.Errorf("expected shared preset id %q, got %q", presetID, preset.ID)
			}
			if preset.Filters.Priority == nil || *preset.Filters.Priority != "high" {
				return fmt.Errorf("expected shared preset priority high")
			}
			if preset.Filters.Q == nil || *preset.Filters.Q != searchQuery {
				return fmt.Errorf("expected shared preset query %q", searchQuery)
			}
			return nil
		}).
		WhenISelectOptionByValueKey("board.filter_priority_select", "low").
		WhenIFillKey("board.filter_search_input", "manual override").
		WhenISelectOptionByValueKey("board.preset_select", presetID).
		Then("shared preset values are applied in-app", func(s *Scenario) error {
			priority, err := inputValueByKey(s.Harness(), "board.filter_priority_select")
			if err != nil {
				return err
			}
			if priority != "high" {
				return fmt.Errorf("expected priority high from shared link, got %q", priority)
			}
			query, err := inputValueByKey(s.Harness(), "board.filter_search_input")
			if err != nil {
				return err
			}
			if query != searchQuery {
				return fmt.Errorf("expected query %q from shared link, got %q", searchQuery, query)
			}
			return nil
		}).
		WhenIClickRefresh().
		Then("active preset remains selected after reload", func(s *Scenario) error {
			value, err := inputValueByKey(s.Harness(), "board.preset_select")
			if err != nil {
				return err
			}
			if value != presetID {
				return fmt.Errorf("expected active preset %q after reload, got %q", presetID, value)
			}
			return nil
		}).
		WhenIClickKey("board.preset_delete_button").
		Then("preset delete persists via API", func(s *Scenario) error {
			list, err := waitForPresetDeletion(s.Harness(), seed.ProjectID, presetID, 5*time.Second)
			if err != nil {
				return err
			}
			for _, item := range list.Items {
				if item.ID == presetID {
					return fmt.Errorf("preset %q still present after delete", presetID)
				}
			}
			return nil
		})
}

func inputValueByKey(h *Harness, key string) (string, error) {
	selector, err := h.Selector(key)
	if err != nil {
		return "", err
	}
	if err := h.WaitVisible(selector); err != nil {
		return "", err
	}
	return h.page.Locator(selector).InputValue()
}

func listBoardFilterPresets(h *Harness, projectID string) (boardFilterPresetListAPI, error) {
	resp, err := h.APIRequest(http.MethodGet, fmt.Sprintf("/projects/%s/board-filters", projectID), nil)
	if err != nil {
		return boardFilterPresetListAPI{}, fmt.Errorf("list presets request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return boardFilterPresetListAPI{}, fmt.Errorf("list presets status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var payload boardFilterPresetListAPI
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return boardFilterPresetListAPI{}, fmt.Errorf("decode presets response: %w", err)
	}
	return payload, nil
}

func waitForPresetByName(h *Harness, projectID, name string, timeout time.Duration) (boardFilterPresetListAPI, error) {
	deadline := time.Now().Add(timeout)
	var last boardFilterPresetListAPI
	for time.Now().Before(deadline) {
		list, err := listBoardFilterPresets(h, projectID)
		if err != nil {
			return boardFilterPresetListAPI{}, err
		}
		last = list
		for _, item := range list.Items {
			if item.Name == name {
				return list, nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	return last, fmt.Errorf("preset %q not found before timeout", name)
}

func waitForPresetNameByID(h *Harness, projectID, presetID, expectedName string, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	lastName := ""
	for time.Now().Before(deadline) {
		list, err := listBoardFilterPresets(h, projectID)
		if err != nil {
			return "", err
		}
		for _, item := range list.Items {
			if item.ID == presetID {
				lastName = item.Name
				if item.Name == expectedName {
					return item.Name, nil
				}
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	if lastName == "" {
		return "", fmt.Errorf("preset %q not found before timeout", presetID)
	}
	return lastName, fmt.Errorf("preset %q name did not become %q before timeout (last=%q)", presetID, expectedName, lastName)
}

func getSharedBoardFilterPreset(h *Harness, projectID, token string) (boardFilterPresetAPI, error) {
	resp, err := h.APIRequest(http.MethodGet, fmt.Sprintf("/projects/%s/board-filters/shared/%s", projectID, token), nil)
	if err != nil {
		return boardFilterPresetAPI{}, fmt.Errorf("shared preset request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return boardFilterPresetAPI{}, fmt.Errorf("shared preset status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var payload boardFilterPresetAPI
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return boardFilterPresetAPI{}, fmt.Errorf("decode shared preset response: %w", err)
	}
	return payload, nil
}

func waitForPresetDeletion(h *Harness, projectID, presetID string, timeout time.Duration) (boardFilterPresetListAPI, error) {
	deadline := time.Now().Add(timeout)
	var last boardFilterPresetListAPI
	for time.Now().Before(deadline) {
		list, err := listBoardFilterPresets(h, projectID)
		if err != nil {
			return boardFilterPresetListAPI{}, err
		}
		last = list
		found := false
		for _, item := range list.Items {
			if item.ID == presetID {
				found = true
				break
			}
		}
		if !found {
			return list, nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return last, fmt.Errorf("preset %q still present before timeout", presetID)
}
