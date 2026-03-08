//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestApprovalPolicyCRUD(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()
	seed := scenario.SeedData()

	var initial struct {
		Items []interface{} `json:"items"`
	}
	var updated struct {
		Items []struct {
			ID            string `json:"id"`
			FromStateId   string `json:"fromStateId"`
			ToStateId     string `json:"toStateId"`
			RequiredCount int    `json:"requiredCount"`
			Scope         string `json:"scope"`
		} `json:"items"`
	}

	scenario.
		Given("no approval policies exist for the project", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("GET", fmt.Sprintf("/projects/%s/approval-policies", seed.ProjectID), nil)
			if err != nil {
				return fmt.Errorf("get policies: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200 from GET approval-policies, got %d", resp.StatusCode)
			}
			return json.NewDecoder(resp.Body).Decode(&initial)
		}).
		When("I PUT a policy requiring 1 approval for Backlog → In Progress", func(s *Scenario) error {
			body := fmt.Sprintf(`{"items":[{"fromStateId":"%s","toStateId":"%s","requiredCount":1,"scope":"any_member"}]}`,
				seed.BacklogID, seed.InProgressID)
			resp, err := s.Harness().APIRequest("PUT", fmt.Sprintf("/projects/%s/approval-policies", seed.ProjectID), strings.NewReader(body))
			if err != nil {
				return fmt.Errorf("put policies: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200 from PUT approval-policies, got %d", resp.StatusCode)
			}
			return json.NewDecoder(resp.Body).Decode(&updated)
		}).
		Then("the response contains exactly one policy", func(s *Scenario) error {
			if len(updated.Items) != 1 {
				return fmt.Errorf("expected 1 policy, got %d", len(updated.Items))
			}
			return nil
		}).
		And("the policy has scope any_member and requiredCount 1", func(s *Scenario) error {
			if updated.Items[0].Scope != "any_member" {
				return fmt.Errorf("expected scope any_member, got %s", updated.Items[0].Scope)
			}
			if updated.Items[0].RequiredCount != 1 {
				return fmt.Errorf("expected requiredCount 1, got %d", updated.Items[0].RequiredCount)
			}
			return nil
		}).
		And("clearing the policies returns 200", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("PUT", fmt.Sprintf("/projects/%s/approval-policies", seed.ProjectID), strings.NewReader(`{"items":[]}`))
			if err != nil {
				return fmt.Errorf("clear policies: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200 clearing policies, got %d", resp.StatusCode)
			}
			return nil
		})
}

func TestApprovalGateBlocksTransition(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()
	seed := scenario.SeedData()

	var ticketID string
	var requestID string

	scenario.
		Given("an approval policy requires 1 approval for Backlog → In Progress", func(s *Scenario) error {
			body := fmt.Sprintf(`{"items":[{"fromStateId":"%s","toStateId":"%s","requiredCount":1,"scope":"any_member"}]}`,
				seed.BacklogID, seed.InProgressID)
			resp, err := s.Harness().APIRequest("PUT", fmt.Sprintf("/projects/%s/approval-policies", seed.ProjectID), strings.NewReader(body))
			if err != nil {
				return fmt.Errorf("put policy: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("set policy: got %d", resp.StatusCode)
			}
			return nil
		}).
		And("a ticket exists in Backlog", func(s *Scenario) error {
			body := fmt.Sprintf(`{"title":"Gated ticket","type":"feature","priority":"medium","storyId":"%s","stateId":"%s"}`,
				seed.StoryID, seed.BacklogID)
			resp, err := s.Harness().APIRequest("POST", fmt.Sprintf("/projects/%s/tickets", seed.ProjectID), strings.NewReader(body))
			if err != nil {
				return fmt.Errorf("create ticket: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 201 {
				return fmt.Errorf("expected 201 creating ticket, got %d", resp.StatusCode)
			}
			var ticket struct {
				ID string `json:"id"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&ticket); err != nil {
				return fmt.Errorf("decode ticket: %w", err)
			}
			ticketID = ticket.ID
			return nil
		}).
		When("I attempt to move the ticket to In Progress", func(s *Scenario) error {
			body := fmt.Sprintf(`{"stateId":"%s"}`, seed.InProgressID)
			resp, err := s.Harness().APIRequest("PATCH", fmt.Sprintf("/tickets/%s", ticketID), strings.NewReader(body))
			if err != nil {
				return fmt.Errorf("move ticket: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 202 {
				return fmt.Errorf("expected 202 (approval required), got %d", resp.StatusCode)
			}
			var approvalResp struct {
				Status    string `json:"status"`
				RequestId string `json:"requestId"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&approvalResp); err != nil {
				return fmt.Errorf("decode approval response: %w", err)
			}
			if approvalResp.Status != "approval_required" {
				return fmt.Errorf("expected status approval_required, got %s", approvalResp.Status)
			}
			if approvalResp.RequestId == "" {
				return fmt.Errorf("expected non-empty requestId")
			}
			requestID = approvalResp.RequestId
			return nil
		}).
		Then("the transition is blocked and an approval request is created", func(s *Scenario) error {
			if requestID == "" {
				return fmt.Errorf("expected a requestId to be captured")
			}
			return nil
		}).
		And("a second move attempt returns 409 (approval already pending)", func(s *Scenario) error {
			body := fmt.Sprintf(`{"stateId":"%s"}`, seed.InProgressID)
			resp, err := s.Harness().APIRequest("PATCH", fmt.Sprintf("/tickets/%s", ticketID), strings.NewReader(body))
			if err != nil {
				return fmt.Errorf("second move: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 409 {
				return fmt.Errorf("expected 409 (approval pending), got %d", resp.StatusCode)
			}
			return nil
		}).
		And("the ticket's approval requests list contains one pending request", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("GET", fmt.Sprintf("/tickets/%s/approval-requests", ticketID), nil)
			if err != nil {
				return fmt.Errorf("list approval requests: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200 listing requests, got %d", resp.StatusCode)
			}
			var requests struct {
				Items []struct {
					ID     string `json:"id"`
					Status string `json:"status"`
				} `json:"items"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&requests); err != nil {
				return fmt.Errorf("decode requests: %w", err)
			}
			if len(requests.Items) != 1 {
				return fmt.Errorf("expected 1 approval request, got %d", len(requests.Items))
			}
			if requests.Items[0].Status != "pending" {
				return fmt.Errorf("expected pending status, got %s", requests.Items[0].Status)
			}
			return nil
		})
}

func TestApprovalApproveUnblocksTransition(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()
	seed := scenario.SeedData()

	var ticketID string
	var requestID string

	scenario.
		Given("an approval policy requires 1 approval for Backlog → In Progress", func(s *Scenario) error {
			body := fmt.Sprintf(`{"items":[{"fromStateId":"%s","toStateId":"%s","requiredCount":1,"scope":"any_member"}]}`,
				seed.BacklogID, seed.InProgressID)
			resp, err := s.Harness().APIRequest("PUT", fmt.Sprintf("/projects/%s/approval-policies", seed.ProjectID), strings.NewReader(body))
			if err != nil {
				return fmt.Errorf("put policy: %w", err)
			}
			defer resp.Body.Close()
			return nil
		}).
		And("a ticket in Backlog has a pending approval request", func(s *Scenario) error {
			ticketBody := fmt.Sprintf(`{"title":"Approve test ticket","type":"feature","priority":"medium","storyId":"%s","stateId":"%s"}`,
				seed.StoryID, seed.BacklogID)
			createResp, err := s.Harness().APIRequest("POST", fmt.Sprintf("/projects/%s/tickets", seed.ProjectID), strings.NewReader(ticketBody))
			if err != nil {
				return fmt.Errorf("create ticket: %w", err)
			}
			defer createResp.Body.Close()
			var ticket struct {
				ID string `json:"id"`
			}
			if err := json.NewDecoder(createResp.Body).Decode(&ticket); err != nil {
				return fmt.Errorf("decode: %w", err)
			}
			ticketID = ticket.ID

			moveBody := fmt.Sprintf(`{"stateId":"%s"}`, seed.InProgressID)
			moveResp, err := s.Harness().APIRequest("PATCH", fmt.Sprintf("/tickets/%s", ticketID), strings.NewReader(moveBody))
			if err != nil {
				return fmt.Errorf("move: %w", err)
			}
			defer moveResp.Body.Close()
			if moveResp.StatusCode != 202 {
				return fmt.Errorf("expected 202, got %d", moveResp.StatusCode)
			}
			var approvalResp struct {
				RequestId string `json:"requestId"`
			}
			if err := json.NewDecoder(moveResp.Body).Decode(&approvalResp); err != nil {
				return fmt.Errorf("decode: %w", err)
			}
			requestID = approvalResp.RequestId
			return nil
		}).
		When("I approve the pending request", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("POST",
				fmt.Sprintf("/tickets/%s/approval-requests/%s/approve", ticketID, requestID),
				strings.NewReader(`{}`))
			if err != nil {
				return fmt.Errorf("approve: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200 from approve, got %d", resp.StatusCode)
			}
			var approvedReq struct {
				Status        string `json:"status"`
				ApprovedCount int    `json:"approvedCount"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&approvedReq); err != nil {
				return fmt.Errorf("decode: %w", err)
			}
			if approvedReq.Status != "approved" {
				return fmt.Errorf("expected status approved, got %s", approvedReq.Status)
			}
			if approvedReq.ApprovedCount != 1 {
				return fmt.Errorf("expected approvedCount 1, got %d", approvedReq.ApprovedCount)
			}
			return nil
		}).
		Then("moving the ticket to In Progress succeeds with 200", func(s *Scenario) error {
			moveBody := fmt.Sprintf(`{"stateId":"%s"}`, seed.InProgressID)
			resp, err := s.Harness().APIRequest("PATCH", fmt.Sprintf("/tickets/%s", ticketID), strings.NewReader(moveBody))
			if err != nil {
				return fmt.Errorf("second move: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200 after approval, got %d", resp.StatusCode)
			}
			return nil
		})
}

func TestApprovalRejectKeepsBlocked(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()
	seed := scenario.SeedData()

	var ticketID string
	var requestID string

	scenario.
		Given("an approval policy requires 1 approval for Backlog → In Progress", func(s *Scenario) error {
			body := fmt.Sprintf(`{"items":[{"fromStateId":"%s","toStateId":"%s","requiredCount":1,"scope":"any_member"}]}`,
				seed.BacklogID, seed.InProgressID)
			resp, err := s.Harness().APIRequest("PUT", fmt.Sprintf("/projects/%s/approval-policies", seed.ProjectID), strings.NewReader(body))
			if err != nil {
				return fmt.Errorf("put policy: %w", err)
			}
			defer resp.Body.Close()
			return nil
		}).
		And("a ticket in Backlog has a pending approval request", func(s *Scenario) error {
			ticketBody := fmt.Sprintf(`{"title":"Reject test","type":"feature","priority":"medium","storyId":"%s","stateId":"%s"}`,
				seed.StoryID, seed.BacklogID)
			createResp, err := s.Harness().APIRequest("POST", fmt.Sprintf("/projects/%s/tickets", seed.ProjectID), strings.NewReader(ticketBody))
			if err != nil {
				return fmt.Errorf("create: %w", err)
			}
			defer createResp.Body.Close()
			var ticket struct {
				ID string `json:"id"`
			}
			if err := json.NewDecoder(createResp.Body).Decode(&ticket); err != nil {
				return fmt.Errorf("decode: %w", err)
			}
			ticketID = ticket.ID

			moveBody := fmt.Sprintf(`{"stateId":"%s"}`, seed.InProgressID)
			firstMove, err := s.Harness().APIRequest("PATCH", fmt.Sprintf("/tickets/%s", ticketID), strings.NewReader(moveBody))
			if err != nil {
				return fmt.Errorf("move: %w", err)
			}
			defer firstMove.Body.Close()
			if firstMove.StatusCode != 202 {
				return fmt.Errorf("expected 202, got %d", firstMove.StatusCode)
			}
			var approvalResp struct {
				RequestId string `json:"requestId"`
			}
			if err := json.NewDecoder(firstMove.Body).Decode(&approvalResp); err != nil {
				return fmt.Errorf("decode: %w", err)
			}
			requestID = approvalResp.RequestId
			return nil
		}).
		When("I reject the pending request", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("POST",
				fmt.Sprintf("/tickets/%s/approval-requests/%s/reject", ticketID, requestID),
				strings.NewReader(`{"reason":"not ready"}`))
			if err != nil {
				return fmt.Errorf("reject: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200 from reject, got %d", resp.StatusCode)
			}
			var rejectedReq struct {
				Status string `json:"status"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&rejectedReq); err != nil {
				return fmt.Errorf("decode: %w", err)
			}
			if rejectedReq.Status != "rejected" {
				return fmt.Errorf("expected status rejected, got %s", rejectedReq.Status)
			}
			return nil
		}).
		Then("a new move attempt creates a fresh approval request (202)", func(s *Scenario) error {
			moveBody := fmt.Sprintf(`{"stateId":"%s"}`, seed.InProgressID)
			resp, err := s.Harness().APIRequest("PATCH", fmt.Sprintf("/tickets/%s", ticketID), strings.NewReader(moveBody))
			if err != nil {
				return fmt.Errorf("second move: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 202 {
				return fmt.Errorf("expected new 202 after rejection, got %d", resp.StatusCode)
			}
			return nil
		})
}

func TestApprovalViewerCanReadPolicies(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()
	seed := scenario.SeedData()

	scenario.
		Given("an admin has set an approval policy", func(s *Scenario) error {
			body := fmt.Sprintf(`{"items":[{"fromStateId":"%s","toStateId":"%s","requiredCount":1,"scope":"any_member"}]}`,
				seed.BacklogID, seed.InProgressID)
			resp, err := s.Harness().APIRequest("PUT", fmt.Sprintf("/projects/%s/approval-policies", seed.ProjectID), strings.NewReader(body))
			if err != nil {
				return fmt.Errorf("put policy: %w", err)
			}
			defer resp.Body.Close()
			return nil
		}).
		When("I read the approval policies", func(s *Scenario) error {
			resp, err := s.Harness().APIRequest("GET", fmt.Sprintf("/projects/%s/approval-policies", seed.ProjectID), nil)
			if err != nil {
				return fmt.Errorf("get: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("expected 200, got %d", resp.StatusCode)
			}
			return nil
		}).
		Then("the response is 200", func(s *Scenario) error {
			return nil
		})
}
