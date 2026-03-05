package store

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type AutomationAction struct {
	Type   string            `json:"type"`
	Params map[string]string `json:"params"`
}

type AutomationRule struct {
	ID                 uuid.UUID
	ProjectID          uuid.UUID
	Name               string
	Enabled            bool
	TriggerEvent       string
	TriggerConditions  map[string]string
	Actions            []AutomationAction
	ExecutionCount     int
	LastExecutedAt     *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type AutomationRuleCreateInput struct {
	Name              string
	Enabled           bool
	TriggerEvent      string
	TriggerConditions map[string]string
	Actions           []AutomationAction
}

type AutomationRuleUpdateInput struct {
	Name              *string
	Enabled           *bool
	TriggerEvent      *string
	TriggerConditions map[string]string
	Actions           []AutomationAction
	HasActions        bool
}

type AutomationActionResult struct {
	Type    string            `json:"type"`
	Params  map[string]string `json:"params"`
	Success bool              `json:"success"`
	Error   string            `json:"error,omitempty"`
}

type AutomationExecution struct {
	ID           uuid.UUID
	RuleID       uuid.UUID
	TicketID     uuid.UUID
	TriggerEvent string
	ActionsRun   []AutomationActionResult
	CreatedAt    time.Time
}

type AutomationExecutionCreateInput struct {
	RuleID       uuid.UUID
	TicketID     uuid.UUID
	TriggerEvent string
	ActionsRun   []AutomationActionResult
}

func (s *Store) ListAutomationRules(ctx context.Context, projectID uuid.UUID) ([]AutomationRule, error) {
	return queryMany(ctx, s.db, mustSQL("automation_rules_list", nil), scanAutomationRule, projectID)
}

func (s *Store) ListEnabledAutomationRules(ctx context.Context, projectID uuid.UUID) ([]AutomationRule, error) {
	return queryMany(ctx, s.db, mustSQL("automation_rules_list_enabled", nil), scanAutomationRule, projectID)
}

func (s *Store) GetAutomationRule(ctx context.Context, id, projectID uuid.UUID) (AutomationRule, error) {
	return queryOne(ctx, s.db, mustSQL("automation_rules_get", nil), scanAutomationRule, id, projectID)
}

func (s *Store) CreateAutomationRule(ctx context.Context, projectID uuid.UUID, input AutomationRuleCreateInput) (AutomationRule, error) {
	condJSON, _ := json.Marshal(input.TriggerConditions)
	actJSON, _ := json.Marshal(input.Actions)

	var id uuid.UUID
	if err := s.db.QueryRow(ctx, mustSQL("automation_rules_insert", nil),
		projectID, input.Name, input.Enabled, input.TriggerEvent, condJSON, actJSON,
	).Scan(&id); err != nil {
		return AutomationRule{}, err
	}
	return s.GetAutomationRule(ctx, id, projectID)
}

func (s *Store) UpdateAutomationRule(ctx context.Context, id, projectID uuid.UUID, input AutomationRuleUpdateInput) (AutomationRule, error) {
	updates := []string{"updated_at = now()"}
	args := []any{}
	arg := func(value any) string {
		args = append(args, value)
		return "$" + strconv.Itoa(len(args))
	}

	if input.Name != nil {
		updates = append(updates, "name = "+arg(strings.TrimSpace(*input.Name)))
	}
	if input.Enabled != nil {
		updates = append(updates, "enabled = "+arg(*input.Enabled))
	}
	if input.TriggerEvent != nil {
		updates = append(updates, "trigger_event = "+arg(*input.TriggerEvent))
	}
	if input.TriggerConditions != nil {
		condJSON, _ := json.Marshal(input.TriggerConditions)
		updates = append(updates, "trigger_conditions = "+arg(condJSON))
	}
	if input.HasActions {
		actJSON, _ := json.Marshal(input.Actions)
		updates = append(updates, "actions = "+arg(actJSON))
	}

	args = append(args, id, projectID)
	query := mustSQL("automation_rules_update", map[string]any{
		"Updates":    strings.Join(updates, ", "),
		"IDArg":      len(args) - 1,
		"ProjectArg": len(args),
	})
	if err := execOne(ctx, s.db, query, pgx.ErrNoRows, args...); err != nil {
		return AutomationRule{}, err
	}
	return s.GetAutomationRule(ctx, id, projectID)
}

func (s *Store) DeleteAutomationRule(ctx context.Context, id, projectID uuid.UUID) error {
	return execOne(ctx, s.db, mustSQL("automation_rules_delete", nil), pgx.ErrNoRows, id, projectID)
}

func (s *Store) BumpAutomationRuleStats(ctx context.Context, ruleID uuid.UUID) error {
	_, err := s.db.Exec(ctx, mustSQL("automation_rules_bump_stats", nil), ruleID)
	return err
}

func (s *Store) CreateAutomationExecution(ctx context.Context, input AutomationExecutionCreateInput) error {
	actJSON, _ := json.Marshal(input.ActionsRun)
	_, err := s.db.Exec(ctx, mustSQL("automation_executions_insert", nil),
		input.RuleID, input.TicketID, input.TriggerEvent, actJSON,
	)
	return err
}

func (s *Store) ListAutomationExecutions(ctx context.Context, ruleID uuid.UUID) ([]AutomationExecution, error) {
	return queryMany(ctx, s.db, mustSQL("automation_executions_list", nil), scanAutomationExecution, ruleID)
}

func scanAutomationRule(row pgx.Row) (AutomationRule, error) {
	var r AutomationRule
	var condRaw, actRaw []byte
	if err := row.Scan(
		&r.ID,
		&r.ProjectID,
		&r.Name,
		&r.Enabled,
		&r.TriggerEvent,
		&condRaw,
		&actRaw,
		&r.ExecutionCount,
		&r.LastExecutedAt,
		&r.CreatedAt,
		&r.UpdatedAt,
	); err != nil {
		return r, err
	}
	if len(condRaw) > 0 {
		_ = json.Unmarshal(condRaw, &r.TriggerConditions)
	}
	if r.TriggerConditions == nil {
		r.TriggerConditions = map[string]string{}
	}
	if len(actRaw) > 0 {
		_ = json.Unmarshal(actRaw, &r.Actions)
	}
	if r.Actions == nil {
		r.Actions = []AutomationAction{}
	}
	return r, nil
}

func scanAutomationExecution(row pgx.Row) (AutomationExecution, error) {
	var e AutomationExecution
	var actRaw []byte
	if err := row.Scan(
		&e.ID,
		&e.RuleID,
		&e.TicketID,
		&e.TriggerEvent,
		&actRaw,
		&e.CreatedAt,
	); err != nil {
		return e, err
	}
	if len(actRaw) > 0 {
		_ = json.Unmarshal(actRaw, &e.ActionsRun)
	}
	if e.ActionsRun == nil {
		e.ActionsRun = []AutomationActionResult{}
	}
	return e, nil
}
