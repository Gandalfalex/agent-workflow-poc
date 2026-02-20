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

func TestIncidentTimelineAndPostmortemExport(t *testing.T) {
	t.Parallel()

	sc := NewScenario(t)
	defer sc.Close()

	seed := sc.SeedData()
	h := sc.Harness()
	ctx := h.Context()
	st := h.Store()

	projectID := uuid.MustParse(seed.ProjectID)
	storyID := uuid.MustParse(seed.StoryID)
	backlogID := uuid.MustParse(seed.BacklogID)

	ts := time.Now().UnixNano()
	title := fmt.Sprintf("Incident Bridge Test %d", ts)
	created, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
		Title:   title,
		Type:    "bug",
		StoryID: storyID,
		StateID: &backlogID,
	})
	if err != nil {
		t.Fatalf("seed ticket: %v", err)
	}

	sc.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/"+seed.ProjectID+"/board").
		WhenIClickRefresh().
		ThenISeeText(title).
		WhenIClickTicketByText(title).
		ThenISeeSelectorKey("ticket.modal").
		WhenIClickKey("ticket.incident_enabled_checkbox").
		WhenISelectOptionByValueKey("ticket.incident_severity_select", "sev1").
		WhenIFillKey("ticket.incident_impact_input", "Checkout failures for all EU customers").
		WhenIFillKey("ticket.comment_input", "Initial response started").
		WhenIClickKey("ticket.post_comment_button").
		WhenIClickKey("ticket.save_button").
		WhenIClickRefresh().
		ThenISeeText(title).
		WhenIClickTicketByText(title).
		ThenISeeSelectorKey("ticket.modal").
		ThenISeeText("changed incident severity from")

	respTimeline, err := h.APIRequest(http.MethodGet, "/tickets/"+created.ID.String()+"/incident-timeline", nil)
	if err != nil {
		t.Fatalf("timeline request: %v", err)
	}
	defer respTimeline.Body.Close()
	if respTimeline.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(respTimeline.Body)
		t.Fatalf("timeline status=%d body=%s", respTimeline.StatusCode, string(body))
	}
	var timeline struct {
		Items []struct {
			Type string `json:"type"`
		} `json:"items"`
	}
	if err := json.NewDecoder(respTimeline.Body).Decode(&timeline); err != nil {
		t.Fatalf("decode timeline: %v", err)
	}
	hasActivity := false
	hasWebhook := false
	for _, item := range timeline.Items {
		if item.Type == "activity" {
			hasActivity = true
		}
		if item.Type == "webhook" {
			hasWebhook = true
		}
	}
	if !hasActivity {
		t.Fatalf("expected incident timeline to include activity item")
	}
	if !hasWebhook {
		t.Fatalf("expected incident timeline to include webhook item")
	}

	respPostmortem, err := h.APIRequest(http.MethodGet, "/tickets/"+created.ID.String()+"/incident-postmortem", nil)
	if err != nil {
		t.Fatalf("postmortem request: %v", err)
	}
	defer respPostmortem.Body.Close()
	if respPostmortem.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(respPostmortem.Body)
		t.Fatalf("postmortem status=%d body=%s", respPostmortem.StatusCode, string(body))
	}
	body, _ := io.ReadAll(respPostmortem.Body)
	markdown := string(body)
	if contentType := respPostmortem.Header.Get("Content-Type"); !strings.Contains(contentType, "text/markdown") {
		t.Fatalf("expected markdown content-type, got %q", contentType)
	}
	for _, needle := range []string{
		"# Postmortem Draft:",
		created.Key,
		"## Incident Summary",
		"Severity: sev1",
		"## Timeline",
		"## Root Cause",
	} {
		if !bytes.Contains(body, []byte(needle)) {
			t.Fatalf("expected postmortem to contain %q; body=%s", needle, markdown)
		}
	}
}
