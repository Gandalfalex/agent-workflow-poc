package automation

import (
	"context"
	"log"
	"strings"

	"ticketing-system/backend/internal/store"

	"github.com/google/uuid"
)

// Store is the minimal store interface the engine needs.
type Store interface {
	ListEnabledAutomationRules(ctx context.Context, projectID uuid.UUID) ([]store.AutomationRule, error)
	UpdateTicket(ctx context.Context, id uuid.UUID, input store.TicketUpdateInput) (store.Ticket, error)
	CreateComment(ctx context.Context, ticketID uuid.UUID, input store.CommentCreateInput) (store.Comment, error)
	BumpAutomationRuleStats(ctx context.Context, ruleID uuid.UUID) error
	CreateAutomationExecution(ctx context.Context, input store.AutomationExecutionCreateInput) error
}

// WebhookDispatcher dispatches webhook events.
type WebhookDispatcher interface {
	Dispatch(ctx context.Context, projectID uuid.UUID, event string, data any)
}

// EngineRunner is the interface exposed to handlers.
type EngineRunner interface {
	Fire(ctx context.Context, projectID uuid.UUID, event string, ticket store.Ticket, extra map[string]string)
}

// Engine evaluates automation rules fire-and-forget.
type Engine struct {
	store    Store
	webhooks WebhookDispatcher
}

// New creates a new Engine.
func New(s Store, w WebhookDispatcher) *Engine {
	return &Engine{store: s, webhooks: w}
}

// Fire evaluates all enabled rules for the project against the event and executes matching ones.
// It is intended to be called via `go engine.Fire(...)` from handlers.
func (e *Engine) Fire(ctx context.Context, projectID uuid.UUID, event string, ticket store.Ticket, extra map[string]string) {
	rules, err := e.store.ListEnabledAutomationRules(ctx, projectID)
	if err != nil {
		log.Printf("automation: list rules: %v", err)
		return
	}

	for _, rule := range rules {
		if rule.TriggerEvent != event {
			continue
		}
		if !matchConditions(rule.TriggerConditions, extra) {
			continue
		}
		e.executeRule(ctx, rule, ticket, event)
	}
}

func matchConditions(conditions map[string]string, extra map[string]string) bool {
	for k, v := range conditions {
		if v == "" {
			continue
		}
		if extra[k] != v {
			return false
		}
	}
	return true
}

func (e *Engine) executeRule(ctx context.Context, rule store.AutomationRule, ticket store.Ticket, event string) {
	results := make([]store.AutomationActionResult, 0, len(rule.Actions))
	for _, action := range rule.Actions {
		result := e.executeAction(ctx, rule.ProjectID, action, ticket)
		results = append(results, result)
	}

	if err := e.store.CreateAutomationExecution(ctx, store.AutomationExecutionCreateInput{
		RuleID:       rule.ID,
		TicketID:     ticket.ID,
		TriggerEvent: event,
		ActionsRun:   results,
	}); err != nil {
		log.Printf("automation: create execution for rule %s: %v", rule.ID, err)
	}

	if err := e.store.BumpAutomationRuleStats(ctx, rule.ID); err != nil {
		log.Printf("automation: bump stats for rule %s: %v", rule.ID, err)
	}
}

func (e *Engine) executeAction(ctx context.Context, projectID uuid.UUID, action store.AutomationAction, ticket store.Ticket) store.AutomationActionResult {
	result := store.AutomationActionResult{
		Type:   action.Type,
		Params: action.Params,
	}

	var execErr error
	switch action.Type {
	case "set_state":
		stateID, err := uuid.Parse(action.Params["state_id"])
		if err != nil {
			result.Error = "invalid state_id: " + err.Error()
			return result
		}
		_, execErr = e.store.UpdateTicket(ctx, ticket.ID, store.TicketUpdateInput{StateID: &stateID})

	case "set_assignee":
		assigneeID, err := uuid.Parse(action.Params["assignee_id"])
		if err != nil {
			result.Error = "invalid assignee_id: " + err.Error()
			return result
		}
		_, execErr = e.store.UpdateTicket(ctx, ticket.ID, store.TicketUpdateInput{AssigneeID: &assigneeID})

	case "set_priority":
		p := action.Params["priority"]
		_, execErr = e.store.UpdateTicket(ctx, ticket.ID, store.TicketUpdateInput{Priority: &p})

	case "add_comment":
		body := expandTemplate(action.Params["body"], ticket)
		_, execErr = e.store.CreateComment(ctx, ticket.ID, store.CommentCreateInput{
			AuthorName: "Automation",
			Message:    body,
		})

	case "call_webhook":
		payload := map[string]any{
			"ticketId":  ticket.ID,
			"ticketKey": ticket.Key,
		}
		e.webhooks.Dispatch(ctx, projectID, "automation.triggered", payload)

	default:
		result.Error = "unknown action type: " + action.Type
		return result
	}

	if execErr != nil {
		result.Error = execErr.Error()
	} else {
		result.Success = true
	}
	return result
}

func expandTemplate(tmpl string, ticket store.Ticket) string {
	r := strings.NewReplacer(
		"{{ticket.key}}", ticket.Key,
		"{{ticket.title}}", ticket.Title,
	)
	return r.Replace(tmpl)
}
