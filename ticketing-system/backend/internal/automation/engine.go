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
	var results []store.AutomationActionResult
	if rule.Graph != nil {
		var err error
		results, err = e.executeRuleGraph(ctx, rule.ProjectID, rule.Graph, ticket, map[string]string{})
		if err != nil {
			log.Printf("automation: graph execution for rule %s: %v", rule.ID, err)
		}
	} else {
		results = make([]store.AutomationActionResult, 0, len(rule.Actions))
		for _, action := range rule.Actions {
			result := e.executeAction(ctx, rule.ProjectID, action, ticket)
			results = append(results, result)
		}
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

func (e *Engine) executeRuleGraph(ctx context.Context, projectID uuid.UUID, graph *store.AutomationGraph, ticket store.Ticket, extra map[string]string) ([]store.AutomationActionResult, error) {
	results := []store.AutomationActionResult{}

	// Build adjacency map: "nodeId" or "nodeId:handle" -> []targetNodeID
	edgeMap := map[string][]string{}
	for _, edge := range graph.Edges {
		key := edge.Source
		if edge.SourceHandle != nil && *edge.SourceHandle != "" {
			key = edge.Source + ":" + *edge.SourceHandle
		}
		edgeMap[key] = append(edgeMap[key], edge.Target)
	}

	// Build node map
	nodeMap := map[string]*store.AutomationGraphNode{}
	for i := range graph.Nodes {
		nodeMap[graph.Nodes[i].ID] = &graph.Nodes[i]
	}

	// Find trigger node
	var startID string
	for _, node := range graph.Nodes {
		if node.Type == "trigger" {
			startID = node.ID
			break
		}
	}
	if startID == "" {
		return results, nil
	}

	// BFS traversal
	queue := []string{startID}
	visited := map[string]bool{}
	for len(queue) > 0 {
		nodeID := queue[0]
		queue = queue[1:]
		if visited[nodeID] {
			continue
		}
		visited[nodeID] = true

		node := nodeMap[nodeID]
		if node == nil {
			continue
		}

		switch node.Type {
		case "trigger":
			for _, next := range edgeMap[nodeID] {
				queue = append(queue, next)
			}
		case "action":
			actionType, _ := node.Data["actionType"].(string)
			params := map[string]string{}
			if p, ok := node.Data["params"].(map[string]any); ok {
				for k, v := range p {
					if s, ok := v.(string); ok {
						params[k] = s
					}
				}
			}
			action := store.AutomationAction{Type: actionType, Params: params}
			result := e.executeAction(ctx, projectID, action, ticket)
			results = append(results, result)
			for _, next := range edgeMap[nodeID] {
				queue = append(queue, next)
			}
		case "branch":
			condField, _ := node.Data["conditionField"].(string)
			condOp, _ := node.Data["conditionOperator"].(string)
			condVal, _ := node.Data["conditionValue"].(string)
			matched := evaluateBranchCondition(condField, condOp, condVal, ticket, extra)
			handle := "false"
			if matched {
				handle = "true"
			}
			for _, next := range edgeMap[node.ID+":"+handle] {
				queue = append(queue, next)
			}
		}
	}
	return results, nil
}

func evaluateBranchCondition(field, op, value string, ticket store.Ticket, extra map[string]string) bool {
	var actual string
	switch field {
	case "priority":
		actual = string(ticket.Priority)
	case "state_id":
		actual = ticket.StateID.String()
	case "assignee_id":
		if ticket.AssigneeID != nil {
			actual = ticket.AssigneeID.String()
		}
	case "type":
		actual = string(ticket.Type)
	default:
		actual = extra[field]
	}
	switch op {
	case "equals":
		return actual == value
	case "not_equals":
		return actual != value
	case "is_empty":
		return actual == "" || actual == "00000000-0000-0000-0000-000000000000"
	case "is_not_empty":
		return actual != "" && actual != "00000000-0000-0000-0000-000000000000"
	}
	return false
}
