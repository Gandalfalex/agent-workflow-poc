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

	h := scenario.Harness()
	seed := scenario.SeedData()

	// GET policies — initially empty
	getResp, err := h.APIRequest("GET", fmt.Sprintf("/projects/%s/approval-policies", seed.ProjectID), nil)
	if err != nil {
		t.Fatalf("get policies: %v", err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != 200 {
		t.Fatalf("expected 200 from GET approval-policies, got %d", getResp.StatusCode)
	}

	var initial struct {
		Items []interface{} `json:"items"`
	}
	if err := json.NewDecoder(getResp.Body).Decode(&initial); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// PUT one policy: Backlog → In Progress requires 1 approval from any_member
	body := fmt.Sprintf(`{"items":[{"fromStateId":"%s","toStateId":"%s","requiredCount":1,"scope":"any_member"}]}`,
		seed.BacklogID, seed.InProgressID)
	putResp, err := h.APIRequest("PUT", fmt.Sprintf("/projects/%s/approval-policies", seed.ProjectID), strings.NewReader(body))
	if err != nil {
		t.Fatalf("put policies: %v", err)
	}
	defer putResp.Body.Close()
	if putResp.StatusCode != 200 {
		t.Fatalf("expected 200 from PUT approval-policies, got %d", putResp.StatusCode)
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
	if err := json.NewDecoder(putResp.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated: %v", err)
	}
	if len(updated.Items) != 1 {
		t.Fatalf("expected 1 policy, got %d", len(updated.Items))
	}
	if updated.Items[0].Scope != "any_member" {
		t.Errorf("expected scope any_member, got %s", updated.Items[0].Scope)
	}
	if updated.Items[0].RequiredCount != 1 {
		t.Errorf("expected requiredCount 1, got %d", updated.Items[0].RequiredCount)
	}

	// Clear policies
	clearResp, err := h.APIRequest("PUT", fmt.Sprintf("/projects/%s/approval-policies", seed.ProjectID), strings.NewReader(`{"items":[]}`))
	if err != nil {
		t.Fatalf("clear policies: %v", err)
	}
	defer clearResp.Body.Close()
	if clearResp.StatusCode != 200 {
		t.Fatalf("expected 200 clearing policies, got %d", clearResp.StatusCode)
	}
}

func TestApprovalGateBlocksTransition(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	h := scenario.Harness()
	seed := scenario.SeedData()

	// Create a policy: Backlog → In Progress requires 1 approval
	policyBody := fmt.Sprintf(`{"items":[{"fromStateId":"%s","toStateId":"%s","requiredCount":1,"scope":"any_member"}]}`,
		seed.BacklogID, seed.InProgressID)
	putResp, err := h.APIRequest("PUT", fmt.Sprintf("/projects/%s/approval-policies", seed.ProjectID), strings.NewReader(policyBody))
	if err != nil {
		t.Fatalf("put policy: %v", err)
	}
	defer putResp.Body.Close()
	if putResp.StatusCode != 200 {
		t.Fatalf("set policy: got %d", putResp.StatusCode)
	}

	// Create a ticket in Backlog state
	ticketBody := fmt.Sprintf(`{"title":"Gated ticket","type":"feature","priority":"medium","storyId":"%s","stateId":"%s"}`,
		seed.StoryID, seed.BacklogID)
	createResp, err := h.APIRequest("POST", fmt.Sprintf("/projects/%s/tickets", seed.ProjectID), strings.NewReader(ticketBody))
	if err != nil {
		t.Fatalf("create ticket: %v", err)
	}
	defer createResp.Body.Close()
	if createResp.StatusCode != 201 {
		t.Fatalf("expected 201 creating ticket, got %d", createResp.StatusCode)
	}
	var ticket struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createResp.Body).Decode(&ticket); err != nil {
		t.Fatalf("decode ticket: %v", err)
	}

	// Attempt to move to In Progress — should get 202 (approval required)
	moveBody := fmt.Sprintf(`{"stateId":"%s"}`, seed.InProgressID)
	moveResp, err := h.APIRequest("PATCH", fmt.Sprintf("/tickets/%s", ticket.ID), strings.NewReader(moveBody))
	if err != nil {
		t.Fatalf("move ticket: %v", err)
	}
	defer moveResp.Body.Close()
	if moveResp.StatusCode != 202 {
		t.Fatalf("expected 202 (approval required), got %d", moveResp.StatusCode)
	}

	var approvalResp struct {
		Status    string `json:"status"`
		RequestId string `json:"requestId"`
	}
	if err := json.NewDecoder(moveResp.Body).Decode(&approvalResp); err != nil {
		t.Fatalf("decode approval response: %v", err)
	}
	if approvalResp.Status != "approval_required" {
		t.Errorf("expected status approval_required, got %s", approvalResp.Status)
	}
	if approvalResp.RequestId == "" {
		t.Fatal("expected non-empty requestId")
	}

	// Second attempt — should get 409 (pending)
	moveResp2, err := h.APIRequest("PATCH", fmt.Sprintf("/tickets/%s", ticket.ID), strings.NewReader(moveBody))
	if err != nil {
		t.Fatalf("second move: %v", err)
	}
	defer moveResp2.Body.Close()
	if moveResp2.StatusCode != 409 {
		t.Fatalf("expected 409 (approval pending), got %d", moveResp2.StatusCode)
	}

	// List approval requests
	listResp, err := h.APIRequest("GET", fmt.Sprintf("/tickets/%s/approval-requests", ticket.ID), nil)
	if err != nil {
		t.Fatalf("list approval requests: %v", err)
	}
	defer listResp.Body.Close()
	if listResp.StatusCode != 200 {
		t.Fatalf("expected 200 listing requests, got %d", listResp.StatusCode)
	}
	var requests struct {
		Items []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"items"`
	}
	if err := json.NewDecoder(listResp.Body).Decode(&requests); err != nil {
		t.Fatalf("decode requests: %v", err)
	}
	if len(requests.Items) != 1 {
		t.Fatalf("expected 1 approval request, got %d", len(requests.Items))
	}
	if requests.Items[0].Status != "pending" {
		t.Errorf("expected pending status, got %s", requests.Items[0].Status)
	}
}

func TestApprovalApproveUnblocksTransition(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	h := scenario.Harness()
	seed := scenario.SeedData()

	// Set up policy
	policyBody := fmt.Sprintf(`{"items":[{"fromStateId":"%s","toStateId":"%s","requiredCount":1,"scope":"any_member"}]}`,
		seed.BacklogID, seed.InProgressID)
	putResp, err := h.APIRequest("PUT", fmt.Sprintf("/projects/%s/approval-policies", seed.ProjectID), strings.NewReader(policyBody))
	if err != nil {
		t.Fatalf("put policy: %v", err)
	}
	defer putResp.Body.Close()

	// Create ticket in Backlog
	ticketBody := fmt.Sprintf(`{"title":"Approve test ticket","type":"feature","priority":"medium","storyId":"%s","stateId":"%s"}`,
		seed.StoryID, seed.BacklogID)
	createResp, err := h.APIRequest("POST", fmt.Sprintf("/projects/%s/tickets", seed.ProjectID), strings.NewReader(ticketBody))
	if err != nil {
		t.Fatalf("create ticket: %v", err)
	}
	defer createResp.Body.Close()
	var ticket struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createResp.Body).Decode(&ticket); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// Trigger approval request
	moveBody := fmt.Sprintf(`{"stateId":"%s"}`, seed.InProgressID)
	moveResp, err := h.APIRequest("PATCH", fmt.Sprintf("/tickets/%s", ticket.ID), strings.NewReader(moveBody))
	if err != nil {
		t.Fatalf("move: %v", err)
	}
	defer moveResp.Body.Close()
	if moveResp.StatusCode != 202 {
		t.Fatalf("expected 202, got %d", moveResp.StatusCode)
	}
	var approvalResp struct {
		RequestId string `json:"requestId"`
	}
	if err := json.NewDecoder(moveResp.Body).Decode(&approvalResp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// Approve
	approveResp, err := h.APIRequest("POST",
		fmt.Sprintf("/tickets/%s/approval-requests/%s/approve", ticket.ID, approvalResp.RequestId),
		strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	defer approveResp.Body.Close()
	if approveResp.StatusCode != 200 {
		t.Fatalf("expected 200 from approve, got %d", approveResp.StatusCode)
	}

	var approvedReq struct {
		Status        string `json:"status"`
		ApprovedCount int    `json:"approvedCount"`
	}
	if err := json.NewDecoder(approveResp.Body).Decode(&approvedReq); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if approvedReq.Status != "approved" {
		t.Errorf("expected status approved, got %s", approvedReq.Status)
	}
	if approvedReq.ApprovedCount != 1 {
		t.Errorf("expected approvedCount 1, got %d", approvedReq.ApprovedCount)
	}

	// Now move should succeed (request resolved)
	moveResp2, err := h.APIRequest("PATCH", fmt.Sprintf("/tickets/%s", ticket.ID), strings.NewReader(moveBody))
	if err != nil {
		t.Fatalf("second move: %v", err)
	}
	defer moveResp2.Body.Close()
	if moveResp2.StatusCode != 200 {
		t.Fatalf("expected 200 after approval, got %d", moveResp2.StatusCode)
	}
}

func TestApprovalRejectKeepsBlocked(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	h := scenario.Harness()
	seed := scenario.SeedData()

	// Set up policy
	policyBody := fmt.Sprintf(`{"items":[{"fromStateId":"%s","toStateId":"%s","requiredCount":1,"scope":"any_member"}]}`,
		seed.BacklogID, seed.InProgressID)
	putResp, err := h.APIRequest("PUT", fmt.Sprintf("/projects/%s/approval-policies", seed.ProjectID), strings.NewReader(policyBody))
	if err != nil {
		t.Fatalf("put policy: %v", err)
	}
	defer putResp.Body.Close()

	// Create ticket in Backlog
	ticketBody := fmt.Sprintf(`{"title":"Reject test","type":"feature","priority":"medium","storyId":"%s","stateId":"%s"}`,
		seed.StoryID, seed.BacklogID)
	createResp, err := h.APIRequest("POST", fmt.Sprintf("/projects/%s/tickets", seed.ProjectID), strings.NewReader(ticketBody))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	defer createResp.Body.Close()
	var ticket struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createResp.Body).Decode(&ticket); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// Trigger approval request
	moveBody := fmt.Sprintf(`{"stateId":"%s"}`, seed.InProgressID)
	firstMove, err := h.APIRequest("PATCH", fmt.Sprintf("/tickets/%s", ticket.ID), strings.NewReader(moveBody))
	if err != nil {
		t.Fatalf("move: %v", err)
	}
	defer firstMove.Body.Close()
	if firstMove.StatusCode != 202 {
		t.Fatalf("expected 202, got %d", firstMove.StatusCode)
	}
	var approvalResp struct {
		RequestId string `json:"requestId"`
	}
	if err := json.NewDecoder(firstMove.Body).Decode(&approvalResp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// Reject
	rejectResp, err := h.APIRequest("POST",
		fmt.Sprintf("/tickets/%s/approval-requests/%s/reject", ticket.ID, approvalResp.RequestId),
		strings.NewReader(`{"reason":"not ready"}`))
	if err != nil {
		t.Fatalf("reject: %v", err)
	}
	defer rejectResp.Body.Close()
	if rejectResp.StatusCode != 200 {
		t.Fatalf("expected 200 from reject, got %d", rejectResp.StatusCode)
	}

	var rejectedReq struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(rejectResp.Body).Decode(&rejectedReq); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if rejectedReq.Status != "rejected" {
		t.Errorf("expected status rejected, got %s", rejectedReq.Status)
	}

	// Move should now trigger a new approval request (202), not 409
	secondMove, err := h.APIRequest("PATCH", fmt.Sprintf("/tickets/%s", ticket.ID), strings.NewReader(moveBody))
	if err != nil {
		t.Fatalf("second move: %v", err)
	}
	defer secondMove.Body.Close()
	if secondMove.StatusCode != 202 {
		t.Fatalf("expected new 202 after rejection, got %d", secondMove.StatusCode)
	}
}

func TestApprovalViewerCanReadPolicies(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	h := scenario.Harness()
	seed := scenario.SeedData()

	// Admin sets a policy
	policyBody := fmt.Sprintf(`{"items":[{"fromStateId":"%s","toStateId":"%s","requiredCount":1,"scope":"any_member"}]}`,
		seed.BacklogID, seed.InProgressID)
	putResp, err := h.APIRequest("PUT", fmt.Sprintf("/projects/%s/approval-policies", seed.ProjectID), strings.NewReader(policyBody))
	if err != nil {
		t.Fatalf("put policy: %v", err)
	}
	defer putResp.Body.Close()

	// Viewer can read policies
	getResp, err := h.APIRequest("GET", fmt.Sprintf("/projects/%s/approval-policies", seed.ProjectID), nil)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", getResp.StatusCode)
	}
}
