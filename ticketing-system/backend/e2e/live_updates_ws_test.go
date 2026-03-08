//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/websocket"
)

type wsProjectEvent struct {
	Type      string         `json:"type"`
	ProjectID string         `json:"projectId"`
	Payload   map[string]any `json:"payload"`
}

func TestProjectEventsWebSocketDeliversLiveUpdates(t *testing.T) {
	t.Parallel()

	sc := NewScenario(t)
	defer sc.Close()

	seed := sc.SeedData()

	var conn *websocket.Conn

	sc.
		Given("a WebSocket connection is established to the project events stream", func(s *Scenario) error {
			c, err := dialProjectEventsWS(s.Harness(), seed.ProjectID, e2eUserToken)
			if err != nil {
				return fmt.Errorf("dial project events ws: %w", err)
			}
			conn = c
			return nil
		}).
		When("all notifications are marked read and a new ticket is created", func(s *Scenario) error {
			if err := apiMarkAllRead(s.Harness(), seed.ProjectID); err != nil {
				return fmt.Errorf("mark all read: %w", err)
			}
			if err := apiCreateTicketForWS(s.Harness(), seed.ProjectID, seed.StoryID); err != nil {
				return fmt.Errorf("create ticket for ws: %w", err)
			}
			return nil
		}).
		Then("all expected live event types are received over the WebSocket within 5 seconds", func(s *Scenario) error {
			defer conn.Close()
			expected := map[string]bool{
				"notifications.changed":      false,
				"notifications.unread_count": false,
				"board.refresh":              false,
				"activity.changed":           false,
			}
			if err := waitForProjectEventTypes(conn, 5*time.Second, expected); err != nil {
				return fmt.Errorf("wait for ws events: %w", err)
			}
			return nil
		})
}

func TestProjectEventsWebSocketFallbackToPollingEndpoints(t *testing.T) {
	t.Parallel()

	sc := NewScenario(t)
	defer sc.Close()

	seed := sc.SeedData()

	sc.
		When("I request the WebSocket endpoint without an Upgrade header", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest(http.MethodGet, fmt.Sprintf("/projects/%s/events/ws", seed.ProjectID), nil)
			if err != nil {
				return fmt.Errorf("request ws endpoint without upgrade: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusUpgradeRequired {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("expected 426 from ws endpoint without upgrade, got %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
			}
			return nil
		}).
		Then("the polling unread-count endpoint is still reachable as a fallback", func(s *Scenario) error {
			countResp, err := s.Harness().APIRequest(http.MethodGet, fmt.Sprintf("/projects/%s/notifications/unread-count", seed.ProjectID), nil)
			if err != nil {
				return fmt.Errorf("poll unread count after ws upgrade failure: %w", err)
			}
			defer countResp.Body.Close()
			if countResp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(countResp.Body)
				return fmt.Errorf("expected polling unread-count endpoint to succeed, got %d: %s", countResp.StatusCode, strings.TrimSpace(string(body)))
			}
			return nil
		})
}

func dialProjectEventsWS(h *Harness, projectID, token string) (*websocket.Conn, error) {
	base, err := url.Parse(h.BaseURL())
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}
	base.Path = fmt.Sprintf("/rest/v1/projects/%s/events/ws", projectID)
	base.RawQuery = ""
	if base.Scheme == "https" {
		base.Scheme = "wss"
	} else {
		base.Scheme = "ws"
	}

	cfg, err := websocket.NewConfig(base.String(), "http://localhost/")
	if err != nil {
		return nil, fmt.Errorf("new websocket config: %w", err)
	}
	cfg.Header = http.Header{}
	cfg.Header.Set("Cookie", fmt.Sprintf("ticketing_session=%s", token))
	return websocket.DialConfig(cfg)
}

func apiCreateTicketForWS(h *Harness, projectID, storyID string) error {
	payload := map[string]any{
		"title":   fmt.Sprintf("WS Ticket %d", time.Now().UnixNano()),
		"storyId": storyID,
		"type":    "feature",
	}
	raw, _ := json.Marshal(payload)
	resp, err := h.APIRequest(http.MethodPost, fmt.Sprintf("/projects/%s/tickets", projectID), bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("create ticket request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create ticket status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

func waitForProjectEventTypes(conn *websocket.Conn, timeout time.Duration, expected map[string]bool) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		_ = conn.SetReadDeadline(time.Now().Add(750 * time.Millisecond))
		var evt wsProjectEvent
		if err := websocket.JSON.Receive(conn, &evt); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "timeout") {
				continue
			}
			return fmt.Errorf("receive ws event: %w", err)
		}
		if _, ok := expected[evt.Type]; ok {
			expected[evt.Type] = true
		}
		all := true
		for _, seen := range expected {
			if !seen {
				all = false
				break
			}
		}
		if all {
			return nil
		}
	}
	missing := make([]string, 0, len(expected))
	for typ, seen := range expected {
		if !seen {
			missing = append(missing, typ)
		}
	}
	return fmt.Errorf("missing expected ws events: %v", missing)
}
