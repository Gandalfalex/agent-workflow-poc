package automation

import (
	"fmt"

	"ticketing-system/backend/internal/store"
)

// ConditionResult holds the evaluation result for a single condition.
type ConditionResult struct {
	Key           string
	ExpectedValue string
	ActualValue   string
	Passed        bool
}

// EvalResult is the result of evaluating a rule against an event.
type EvalResult struct {
	Matched          bool
	ConditionResults []ConditionResult
	FailureReason    *string
	PredictedActions []store.SimulatedActionOutcome
}

// EvaluateRule evaluates a rule against an event without any side effects.
// It mirrors the logic in matchConditions and executeAction but never touches the DB.
func EvaluateRule(rule store.AutomationRule, event string, extra map[string]string) EvalResult {
	if rule.TriggerEvent != event {
		reason := fmt.Sprintf("trigger event mismatch: rule expects %q, got %q", rule.TriggerEvent, event)
		return EvalResult{Matched: false, FailureReason: &reason}
	}

	condResults := make([]ConditionResult, 0, len(rule.TriggerConditions))
	for k, v := range rule.TriggerConditions {
		if v == "" {
			continue
		}
		actual := extra[k]
		condResults = append(condResults, ConditionResult{
			Key:           k,
			ExpectedValue: v,
			ActualValue:   actual,
			Passed:        actual == v,
		})
	}

	for _, c := range condResults {
		if !c.Passed {
			reason := fmt.Sprintf("condition %q: expected %q, got %q", c.Key, c.ExpectedValue, c.ActualValue)
			return EvalResult{Matched: false, ConditionResults: condResults, FailureReason: &reason}
		}
	}

	outcomes := predictActions(rule.Actions)
	return EvalResult{Matched: true, ConditionResults: condResults, PredictedActions: outcomes}
}

func predictActions(actions []store.AutomationAction) []store.SimulatedActionOutcome {
	out := make([]store.SimulatedActionOutcome, 0, len(actions))
	for _, a := range actions {
		outcome := store.SimulatedActionOutcome{
			Type:         a.Type,
			Params:       a.Params,
			WouldSucceed: true,
		}
		switch a.Type {
		case "set_state":
			if a.Params["state_id"] == "" {
				reason := "missing state_id param"
				outcome.WouldSucceed = false
				outcome.FailureReason = &reason
			}
		case "set_assignee":
			if a.Params["assignee_id"] == "" {
				reason := "missing assignee_id param"
				outcome.WouldSucceed = false
				outcome.FailureReason = &reason
			}
		case "set_priority":
			if a.Params["priority"] == "" {
				reason := "missing priority param"
				outcome.WouldSucceed = false
				outcome.FailureReason = &reason
			}
		case "add_comment":
			if a.Params["body"] == "" {
				reason := "missing body param"
				outcome.WouldSucceed = false
				outcome.FailureReason = &reason
			}
		case "call_webhook":
			// webhooks always predicted as would-succeed in dry-run
		default:
			reason := fmt.Sprintf("unknown action type: %s", a.Type)
			outcome.WouldSucceed = false
			outcome.FailureReason = &reason
		}
		out = append(out, outcome)
	}
	return out
}
