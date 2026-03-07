package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrApprovalRequestNotFound = errors.New("approval request not found")
var ErrApprovalPolicyNotFound = errors.New("approval policy not found")

type ApprovalPolicy struct {
	ID            uuid.UUID
	ProjectID     uuid.UUID
	FromStateID   uuid.UUID
	ToStateID     uuid.UUID
	RequiredCount int
	Scope         string
	GroupID       *uuid.UUID
	Role          *string
	CreatedAt     time.Time
}

type ApprovalPolicyInput struct {
	FromStateID   uuid.UUID
	ToStateID     uuid.UUID
	RequiredCount int
	Scope         string
	GroupID       *uuid.UUID
	Role          *string
}

type ApprovalRequest struct {
	ID              uuid.UUID
	TicketID        uuid.UUID
	PolicyID        uuid.UUID
	FromStateID     uuid.UUID
	ToStateID       uuid.UUID
	RequestedBy     uuid.UUID
	RequestedByName string
	Status          string
	RequiredCount   int
	ApprovedCount   int
	Decisions       []ApprovalDecision
	CreatedAt       time.Time
	ResolvedAt      *time.Time
}

type ApprovalDecision struct {
	ID           uuid.UUID
	RequestID    uuid.UUID
	ApproverID   uuid.UUID
	ApproverName string
	Decision     string
	Reason       *string
	DecidedAt    time.Time
}

func (s *Store) GetApprovalPolicies(ctx context.Context, projectID uuid.UUID) ([]ApprovalPolicy, error) {
	return queryMany(ctx, s.db, mustSQL("approval_policies_by_project", nil), scanApprovalPolicy, projectID)
}

func (s *Store) ReplaceApprovalPolicies(ctx context.Context, projectID uuid.UUID, inputs []ApprovalPolicyInput) ([]ApprovalPolicy, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err := tx.Exec(ctx, mustSQL("approval_policy_delete_by_project", nil), projectID); err != nil {
		return nil, err
	}

	out := make([]ApprovalPolicy, 0, len(inputs))
	for _, inp := range inputs {
		rc := inp.RequiredCount
		if rc <= 0 {
			rc = 1
		}
		var p ApprovalPolicy
		row := tx.QueryRow(ctx, mustSQL("approval_policy_insert", nil),
			projectID, inp.FromStateID, inp.ToStateID, rc, inp.Scope, inp.GroupID, inp.Role)
		if err := row.Scan(&p.ID, &p.ProjectID, &p.FromStateID, &p.ToStateID,
			&p.RequiredCount, &p.Scope, &p.GroupID, &p.Role, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Store) LookupApprovalPolicy(ctx context.Context, projectID, fromStateID, toStateID uuid.UUID) (*ApprovalPolicy, error) {
	p, err := queryOne(ctx, s.db, mustSQL("approval_policy_lookup", nil), scanApprovalPolicy, projectID, fromStateID, toStateID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) GetPendingApprovalRequest(ctx context.Context, ticketID, toStateID uuid.UUID) (*ApprovalRequest, error) {
	req, err := queryOne(ctx, s.db, mustSQL("approval_request_pending_by_ticket", nil), scanApprovalRequest, ticketID, toStateID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	decisions, err := s.GetApprovalDecisions(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	req.Decisions = decisions
	return &req, nil
}

func (s *Store) GetApprovalRequestByID(ctx context.Context, requestID, ticketID uuid.UUID) (*ApprovalRequest, error) {
	req, err := queryOne(ctx, s.db, mustSQL("approval_request_by_id", nil), scanApprovalRequest, requestID, ticketID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrApprovalRequestNotFound
	}
	if err != nil {
		return nil, err
	}
	decisions, err := s.GetApprovalDecisions(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	req.Decisions = decisions
	return &req, nil
}

func (s *Store) ListApprovalRequests(ctx context.Context, ticketID uuid.UUID) ([]ApprovalRequest, error) {
	reqs, err := queryMany(ctx, s.db, mustSQL("approval_requests_by_ticket", nil), scanApprovalRequest, ticketID)
	if err != nil {
		return nil, err
	}
	for i := range reqs {
		decisions, err := s.GetApprovalDecisions(ctx, reqs[i].ID)
		if err != nil {
			return nil, err
		}
		reqs[i].Decisions = decisions
	}
	return reqs, nil
}

func (s *Store) CreateApprovalRequest(ctx context.Context, ticketID, policyID, fromStateID, toStateID, requestedBy uuid.UUID, requestedByName string, requiredCount int) (*ApprovalRequest, error) {
	req, err := queryOne(ctx, s.db, mustSQL("approval_request_insert", nil), scanApprovalRequest,
		ticketID, policyID, fromStateID, toStateID, requestedBy, requestedByName, requiredCount)
	if err != nil {
		return nil, err
	}
	req.Decisions = []ApprovalDecision{}
	return &req, nil
}

func (s *Store) GetApprovalDecisions(ctx context.Context, requestID uuid.UUID) ([]ApprovalDecision, error) {
	return queryMany(ctx, s.db, mustSQL("approval_decisions_by_request", nil), scanApprovalDecision, requestID)
}

func (s *Store) AddApprovalDecision(ctx context.Context, requestID, approverID uuid.UUID, approverName, decision string, reason *string) (*ApprovalDecision, error) {
	d, err := queryOne(ctx, s.db, mustSQL("approval_decision_insert", nil), scanApprovalDecision,
		requestID, approverID, approverName, decision, reason)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (s *Store) UpdateApprovalRequestStatus(ctx context.Context, requestID uuid.UUID, status string, approvedCount int) (*ApprovalRequest, error) {
	req, err := queryOne(ctx, s.db, mustSQL("approval_request_update_status", nil), scanApprovalRequest,
		requestID, status, approvedCount)
	if err != nil {
		return nil, err
	}
	decisions, err := s.GetApprovalDecisions(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	req.Decisions = decisions
	return &req, nil
}

func (s *Store) HasDuplicateApprovalDecision(ctx context.Context, requestID, approverID uuid.UUID) (bool, error) {
	var count int
	err := s.db.QueryRow(ctx, mustSQL("approval_check_duplicate_decision", nil), requestID, approverID).Scan(&count)
	return count > 0, err
}

func scanApprovalPolicy(row pgx.Row) (ApprovalPolicy, error) {
	var p ApprovalPolicy
	err := row.Scan(
		&p.ID, &p.ProjectID, &p.FromStateID, &p.ToStateID,
		&p.RequiredCount, &p.Scope, &p.GroupID, &p.Role, &p.CreatedAt,
	)
	return p, err
}

func scanApprovalRequest(row pgx.Row) (ApprovalRequest, error) {
	var r ApprovalRequest
	err := row.Scan(
		&r.ID, &r.TicketID, &r.PolicyID, &r.FromStateID, &r.ToStateID,
		&r.RequestedBy, &r.RequestedByName, &r.Status,
		&r.RequiredCount, &r.ApprovedCount, &r.CreatedAt, &r.ResolvedAt,
	)
	return r, err
}

func scanApprovalDecision(row pgx.Row) (ApprovalDecision, error) {
	var d ApprovalDecision
	err := row.Scan(
		&d.ID, &d.RequestID, &d.ApproverID, &d.ApproverName,
		&d.Decision, &d.Reason, &d.DecidedAt,
	)
	return d, err
}
