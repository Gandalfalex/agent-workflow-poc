/**
 * Client-side automation rule condition evaluator.
 * Mirrors the backend matchConditions logic in automation/engine.go.
 * Used to drive the interactive simulation step-through UI without round-trips.
 */

import type { SimulatedActionOutcome } from "./api";

export interface ConditionResult {
  key: string;
  expectedValue: string;
  actualValue: string;
  passed: boolean;
}

export interface EvalResult {
  matched: boolean;
  conditionResults: ConditionResult[];
  failureReason?: string;
  predictedActions: SimulatedActionOutcome[];
}

export interface RuleDefinition {
  triggerEvent: string;
  triggerConditions?: Record<string, string>;
  actions: Array<{ type: string; params: Record<string, string> }>;
}

export interface SimEvent {
  eventType: string;
  extra: Record<string, string>;
}

export function evaluateRule(rule: RuleDefinition, event: SimEvent): EvalResult {
  if (rule.triggerEvent !== event.eventType) {
    return {
      matched: false,
      conditionResults: [],
      failureReason: `trigger mismatch: rule expects "${rule.triggerEvent}", got "${event.eventType}"`,
      predictedActions: [],
    };
  }

  const conditionResults: ConditionResult[] = [];
  for (const [key, expected] of Object.entries(rule.triggerConditions ?? {})) {
    if (!expected) continue;
    const actual = event.extra[key] ?? "";
    conditionResults.push({ key, expectedValue: expected, actualValue: actual, passed: actual === expected });
  }

  const failing = conditionResults.find((c) => !c.passed);
  if (failing) {
    return {
      matched: false,
      conditionResults,
      failureReason: `condition "${failing.key}": expected "${failing.expectedValue}", got "${failing.actualValue}"`,
      predictedActions: [],
    };
  }

  return {
    matched: true,
    conditionResults,
    predictedActions: predictActions(rule.actions),
  };
}

function predictActions(
  actions: Array<{ type: string; params: Record<string, string> }>,
): SimulatedActionOutcome[] {
  return actions.map((a) => {
    const outcome: SimulatedActionOutcome = {
      type: a.type,
      params: a.params,
      wouldSucceed: true,
      failureReason: null,
    };
    switch (a.type) {
      case "set_state":
        if (!a.params["state_id"]) {
          outcome.wouldSucceed = false;
          outcome.failureReason = "missing state_id param";
        }
        break;
      case "set_assignee":
        if (!a.params["assignee_id"]) {
          outcome.wouldSucceed = false;
          outcome.failureReason = "missing assignee_id param";
        }
        break;
      case "set_priority":
        if (!a.params["priority"]) {
          outcome.wouldSucceed = false;
          outcome.failureReason = "missing priority param";
        }
        break;
      case "add_comment":
        if (!a.params["body"]) {
          outcome.wouldSucceed = false;
          outcome.failureReason = "missing body param";
        }
        break;
      case "call_webhook":
        break;
      default:
        outcome.wouldSucceed = false;
        outcome.failureReason = `unknown action type: ${a.type}`;
    }
    return outcome;
  });
}
