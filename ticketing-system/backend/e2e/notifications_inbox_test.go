//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"ticketing-system/backend/internal/store"
)

func TestInboxNotificationsMentionsAssignmentsAndPreferences(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()
	st := scenario.Harness().Store()
	ctx := scenario.Harness().Context()

	projectID := uuid.MustParse(seed.ProjectID)
	storyID := uuid.MustParse(seed.StoryID)
	backlogID := uuid.MustParse(seed.BacklogID)
	viewerID := uuid.MustParse(seed.ViewerUserID)

	ticket, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
		Title:   fmt.Sprintf("Notif Ticket %d", time.Now().UnixNano()),
		Type:    "feature",
		StoryID: storyID,
		StateID: &backlogID,
	})
	if err != nil {
		t.Fatalf("seed ticket: %v", err)
	}

	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/"+seed.ProjectID+"/board").
		Then("admin assigns and mentions viewer", func(s *Scenario) error {
			if err := apiUpdateTicketAssignee(s.Harness(), ticket.ID.String(), viewerID.String()); err != nil {
				return err
			}
			return apiAddComment(s.Harness(), ticket.ID.String(), "@NormalUser please review this")
		}).
		Then("viewer has unread notifications after assignment and mention", func(s *Scenario) error {
			count, err := st.CountUnreadNotifications(ctx, projectID, viewerID)
			if err != nil {
				return fmt.Errorf("count viewer unread notifications: %w", err)
			}
			if count < 2 {
				return fmt.Errorf("expected at least 2 unread notifications, got %d", count)
			}
			return nil
		}).
		WhenIClickLogout().
		WhenILogInAs("NormalUser", "viewer123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/"+seed.ProjectID+"/board").
		WhenIClickKey("nav.inbox_button").
		ThenISeeSelectorKey("nav.inbox_panel").
		ThenISeeText("assigned you").
		AndISeeText("mentioned you").
		Then("viewer marks all read", func(s *Scenario) error {
			return apiMarkAllRead(s.Harness(), seed.ProjectID)
		}).
		Then("viewer unread count is zero after mark all", func(s *Scenario) error {
			return waitUnreadCount(s.Harness(), seed.ProjectID, 0, 5*time.Second)
		}).
		Then("viewer disables mention notifications", func(s *Scenario) error {
			return apiUpdateNotificationPreferences(s.Harness(), seed.ProjectID, map[string]any{
				"mentionEnabled": false,
			})
		}).
		WhenIClickLogout().
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		Then("admin sends a new mention", func(s *Scenario) error {
			return apiAddComment(s.Harness(), ticket.ID.String(), "@NormalUser follow-up mention")
		}).
		WhenIClickLogout().
		WhenILogInAs("NormalUser", "viewer123").
		WhenISelectProjectByID(seed.ProjectID).
		Then("viewer does not get mention notifications when mention preference disabled", func(s *Scenario) error {
			return waitUnreadCount(s.Harness(), seed.ProjectID, 0, 5*time.Second)
		})
}

func apiUpdateTicketAssignee(h *Harness, ticketID, assigneeID string) error {
	body := map[string]any{"assigneeId": assigneeID}
	raw, _ := json.Marshal(body)
	resp, err := h.APIRequest(http.MethodPatch, "/tickets/"+ticketID, bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("update ticket assignee request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		payload, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update ticket assignee status %d: %s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}
	return nil
}

func apiAddComment(h *Harness, ticketID, message string) error {
	body := map[string]any{"message": message}
	raw, _ := json.Marshal(body)
	resp, err := h.APIRequest(http.MethodPost, "/tickets/"+ticketID+"/comments", bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("add comment request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		payload, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("add comment status %d: %s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}
	return nil
}

func getUnreadCount(h *Harness, projectID string) (int, error) {
	resp, err := h.APIRequest(http.MethodGet, "/projects/"+projectID+"/notifications/unread-count", nil)
	if err != nil {
		return 0, fmt.Errorf("unread count request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		payload, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("unread count status %d: %s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}
	var result struct {
		Count int `json:"count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decode unread count: %w", err)
	}
	return result.Count, nil
}

func apiMarkAllRead(h *Harness, projectID string) error {
	resp, err := h.APIRequest(http.MethodPost, "/projects/"+projectID+"/notifications/read-all", nil)
	if err != nil {
		return fmt.Errorf("mark all read request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		payload, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("mark all read status %d: %s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}
	return nil
}

func apiUpdateNotificationPreferences(h *Harness, projectID string, payload map[string]any) error {
	raw, _ := json.Marshal(payload)
	resp, err := h.APIRequest(http.MethodPatch, "/projects/"+projectID+"/notification-preferences", bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("update notification preferences request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update notification preferences status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

func waitUnreadCount(h *Harness, projectID string, want int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	last := -1
	for time.Now().Before(deadline) {
		count, err := getUnreadCount(h, projectID)
		if err != nil {
			return err
		}
		last = count
		if count == want {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("expected unread count %d, got %d", want, last)
}
