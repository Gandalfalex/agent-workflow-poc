//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"ticketing-system/backend/internal/store"
)

type bulkOpAPIResponse struct {
	Action       string `json:"action"`
	Total        int    `json:"total"`
	SuccessCount int    `json:"successCount"`
	ErrorCount   int    `json:"errorCount"`
	Results      []struct {
		TicketID  string  `json:"ticketId"`
		Success   bool    `json:"success"`
		ErrorCode *string `json:"errorCode"`
		Message   *string `json:"message"`
	} `json:"results"`
}

func seedBulkTicket(ctx context.Context, st *store.Store, projectID, storyID, stateID uuid.UUID, title, ticketType string) (store.Ticket, error) {
	ticket, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
		Title:   title,
		Type:    ticketType,
		StoryID: storyID,
		StateID: &stateID,
	})
	if err != nil {
		return store.Ticket{}, fmt.Errorf("seed ticket %q: %w", title, err)
	}
	return ticket, nil
}

func TestBulkTicketOperationsAdminUI(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()
	st := scenario.Harness().Store()
	ctx := scenario.Harness().Context()

	projectID := uuid.MustParse(seed.ProjectID)
	storyID := uuid.MustParse(seed.StoryID)
	backlogID := uuid.MustParse(seed.BacklogID)
	inProgressID := uuid.MustParse(seed.InProgressID)
	viewerUserID := uuid.MustParse(seed.ViewerUserID)

	ts := time.Now().UnixNano()

	var ticketA, ticketB store.Ticket

	scenario.
		Given("three tickets are seeded into the project", func(s *Scenario) error {
			lookup, err := st.CreateTicket(ctx, projectID, store.TicketCreateInput{
				Title:      fmt.Sprintf("Bulk Lookup %d", ts),
				Type:       "feature",
				StoryID:    storyID,
				StateID:    &backlogID,
				AssigneeID: &viewerUserID,
			})
			if err != nil {
				return fmt.Errorf("seed lookup ticket: %w", err)
			}
			_ = lookup

			a, err := seedBulkTicket(ctx, st, projectID, storyID, backlogID, fmt.Sprintf("Bulk A %d", ts), "feature")
			if err != nil {
				return err
			}
			ticketA = a

			b, err := seedBulkTicket(ctx, st, projectID, storyID, backlogID, fmt.Sprintf("Bulk B %d", ts), "bug")
			if err != nil {
				return err
			}
			ticketB = b
			return nil
		}).
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/"+seed.ProjectID+"/board").
		WhenIClickRefresh().
		WhenIClickKey("board.bulk_toggle_button").
		ThenISeeSelectorKey("board.bulk_action_select").
		When("I select two tickets for bulk operation", func(s *Scenario) error {
			if err := clickBulkTicketKey(s.Harness(), ticketA.Key); err != nil {
				return err
			}
			return clickBulkTicketKey(s.Harness(), ticketB.Key)
		}).
		WhenISelectOptionByValueKey("board.bulk_action_select", "move_state").
		WhenISelectOptionByValueKey("board.bulk_state_select", inProgressID.String()).
		WhenIClickKey("board.bulk_apply_button").
		Then("tickets are moved to in-progress", func(s *Scenario) error {
			return waitForTicketsState(s.Harness().Context(), st, ticketA.ID, ticketB.ID, inProgressID)
		}).
		When("I reselect the same tickets", func(s *Scenario) error {
			if err := clickBulkTicketKey(s.Harness(), ticketA.Key); err != nil {
				return err
			}
			return clickBulkTicketKey(s.Harness(), ticketB.Key)
		}).
		WhenISelectOptionByValueKey("board.bulk_action_select", "assign").
		WhenISelectOptionByValueKey("board.bulk_assignee_select", viewerUserID.String()).
		WhenIClickKey("board.bulk_apply_button").
		Then("tickets are assigned to viewer user", func(s *Scenario) error {
			return waitForTicketsAssignee(s.Harness().Context(), st, viewerUserID, ticketA.ID, ticketB.ID)
		}).
		When("I reselect the same tickets again", func(s *Scenario) error {
			if err := clickBulkTicketKey(s.Harness(), ticketA.Key); err != nil {
				return err
			}
			return clickBulkTicketKey(s.Harness(), ticketB.Key)
		}).
		WhenISelectOptionByValueKey("board.bulk_action_select", "set_priority").
		WhenISelectOptionByValueKey("board.bulk_priority_select", "urgent").
		WhenIClickKey("board.bulk_apply_button").
		Then("tickets have urgent priority", func(s *Scenario) error {
			return waitForTicketsPriority(s.Harness().Context(), st, "urgent", ticketA.ID, ticketB.ID)
		}).
		When("I reselect the same tickets one last time", func(s *Scenario) error {
			if err := clickBulkTicketKey(s.Harness(), ticketA.Key); err != nil {
				return err
			}
			return clickBulkTicketKey(s.Harness(), ticketB.Key)
		}).
		WhenISelectOptionByValueKey("board.bulk_action_select", "delete").
		WhenIClickKey("board.bulk_apply_button").
		Then("tickets are deleted", func(s *Scenario) error {
			return waitForTicketsDeleted(s.Harness().Context(), st, ticketA.ID, ticketB.ID)
		})
}

func TestBulkTicketOperationViewerGetsPerTicketFailures(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t, WithViewerUser())
	defer scenario.Close()

	seed := scenario.SeedData()
	st := scenario.Harness().Store()
	ctx := scenario.Harness().Context()

	projectID := uuid.MustParse(seed.ProjectID)
	storyID := uuid.MustParse(seed.StoryID)
	backlogID := uuid.MustParse(seed.BacklogID)

	var ticketID string
	var bulkResponse bulkOpAPIResponse

	scenario.
		Given("a ticket is seeded into the project", func(s *Scenario) error {
			ticket, err := seedBulkTicket(ctx, st, projectID, storyID, backlogID,
				fmt.Sprintf("Viewer Bulk %d", time.Now().UnixNano()), "feature")
			if err != nil {
				return err
			}
			ticketID = ticket.ID.String()
			return nil
		}).
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInWithHarnessUser().
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/" + seed.ProjectID + "/board").
		When("a viewer submits a bulk delete request via API", func(s *Scenario) error {
			reqBody := map[string]any{
				"action":    "delete",
				"ticketIds": []string{ticketID},
			}
			raw, err := json.Marshal(reqBody)
			if err != nil {
				return fmt.Errorf("marshal request: %w", err)
			}
			resp, err := s.Harness().APIRequest(
				http.MethodPost,
				fmt.Sprintf("/projects/%s/tickets/bulk", seed.ProjectID),
				bytes.NewReader(raw),
			)
			if err != nil {
				return fmt.Errorf("bulk API request: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("expected status 200, got %d", resp.StatusCode)
			}
			if err := json.NewDecoder(resp.Body).Decode(&bulkResponse); err != nil {
				return fmt.Errorf("decode bulk response: %w", err)
			}
			return nil
		}).
		Then("the response summary shows one total and one error", func(s *Scenario) error {
			if bulkResponse.Total != 1 || bulkResponse.SuccessCount != 0 || bulkResponse.ErrorCount != 1 {
				return fmt.Errorf("unexpected summary: %+v", bulkResponse)
			}
			return nil
		}).
		And("the response contains a per-ticket insufficient_role failure", func(s *Scenario) error {
			if len(bulkResponse.Results) != 1 || bulkResponse.Results[0].Success {
				return fmt.Errorf("expected one failed result, got %+v", bulkResponse.Results)
			}
			if bulkResponse.Results[0].ErrorCode == nil || *bulkResponse.Results[0].ErrorCode != "insufficient_role" {
				return fmt.Errorf("expected insufficient_role error, got %+v", bulkResponse.Results[0].ErrorCode)
			}
			return nil
		})
}

func clickBulkTicketKey(h *Harness, ticketKey string) error {
	return h.Click(fmt.Sprintf("[data-testid=\"board.ticket-select-%s\"]", ticketKey))
}

func waitForTicketsState(ctx context.Context, st *store.Store, a, b, stateID uuid.UUID) error {
	return waitForTicketCondition(5*time.Second, func() (bool, error) {
		ta, err := st.GetTicket(ctx, a)
		if err != nil {
			return false, err
		}
		tb, err := st.GetTicket(ctx, b)
		if err != nil {
			return false, err
		}
		return ta.StateID == stateID && tb.StateID == stateID, nil
	}, "tickets did not reach target state")
}

func waitForTicketsAssignee(ctx context.Context, st *store.Store, assigneeID uuid.UUID, ids ...uuid.UUID) error {
	return waitForTicketCondition(5*time.Second, func() (bool, error) {
		for _, id := range ids {
			ticket, err := st.GetTicket(ctx, id)
			if err != nil {
				return false, err
			}
			if ticket.AssigneeID == nil || *ticket.AssigneeID != assigneeID {
				return false, nil
			}
		}
		return true, nil
	}, "tickets did not get assigned")
}

func waitForTicketsPriority(ctx context.Context, st *store.Store, priority string, ids ...uuid.UUID) error {
	return waitForTicketCondition(5*time.Second, func() (bool, error) {
		for _, id := range ids {
			ticket, err := st.GetTicket(ctx, id)
			if err != nil {
				return false, err
			}
			if ticket.Priority != priority {
				return false, nil
			}
		}
		return true, nil
	}, "tickets did not get priority")
}

func waitForTicketsDeleted(ctx context.Context, st *store.Store, ids ...uuid.UUID) error {
	return waitForTicketCondition(5*time.Second, func() (bool, error) {
		for _, id := range ids {
			_, err := st.GetTicket(ctx, id)
			if err == nil {
				return false, nil
			}
			if err != nil && err != pgx.ErrNoRows {
				return false, err
			}
		}
		return true, nil
	}, "tickets were not deleted")
}

func waitForTicketCondition(timeout time.Duration, check func() (bool, error), msg string) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		ok, err := check()
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("%s", msg)
}
