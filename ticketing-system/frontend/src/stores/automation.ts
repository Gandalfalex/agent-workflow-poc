import { defineStore } from "pinia";
import { ref } from "vue";
import type {
  AutomationRule,
  AutomationRuleCreateRequest,
  AutomationRuleUpdateRequest,
  AutomationExecution,
} from "../lib/api";
import {
  listAutomationRules,
  createAutomationRule as apiCreateAutomationRule,
  updateAutomationRule as apiUpdateAutomationRule,
  deleteAutomationRule as apiDeleteAutomationRule,
  listAutomationExecutions,
} from "../lib/api";

export const useAutomationStore = defineStore("automation", () => {
  const rules = ref<AutomationRule[]>([]);
  const executionsByRule = ref<Record<string, AutomationExecution[]>>({});
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function loadRules(projectId: string) {
    loading.value = true;
    error.value = null;
    try {
      const res = await listAutomationRules(projectId);
      rules.value = res.items;
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : "Failed to load rules";
    } finally {
      loading.value = false;
    }
  }

  async function createRule(projectId: string, payload: AutomationRuleCreateRequest): Promise<AutomationRule | null> {
    try {
      const rule = await apiCreateAutomationRule(projectId, payload);
      rules.value.push(rule);
      return rule;
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : "Failed to create rule";
      return null;
    }
  }

  async function updateRule(projectId: string, ruleId: string, payload: AutomationRuleUpdateRequest): Promise<AutomationRule | null> {
    try {
      const rule = await apiUpdateAutomationRule(projectId, ruleId, payload);
      const idx = rules.value.findIndex((r) => r.id === ruleId);
      if (idx !== -1) rules.value[idx] = rule;
      return rule;
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : "Failed to update rule";
      return null;
    }
  }

  async function deleteRule(projectId: string, ruleId: string): Promise<boolean> {
    try {
      await apiDeleteAutomationRule(projectId, ruleId);
      rules.value = rules.value.filter((r) => r.id !== ruleId);
      return true;
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : "Failed to delete rule";
      return false;
    }
  }

  async function loadExecutions(projectId: string, ruleId: string) {
    try {
      const res = await listAutomationExecutions(projectId, ruleId);
      executionsByRule.value[ruleId] = res.items;
    } catch {
      executionsByRule.value[ruleId] = [];
    }
  }

  return {
    rules,
    executionsByRule,
    loading,
    error,
    loadRules,
    createRule,
    updateRule,
    deleteRule,
    loadExecutions,
  };
});
