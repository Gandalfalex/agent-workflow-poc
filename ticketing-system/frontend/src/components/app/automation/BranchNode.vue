<script setup lang="ts">
import { Handle, Position } from '@vue-flow/core'

const props = defineProps<{ data: any; id: string }>()

const fieldOptions = [
  { value: 'priority', label: 'Priority' },
  { value: 'state_id', label: 'State' },
  { value: 'assignee_id', label: 'Assignee' },
  { value: 'type', label: 'Type' },
]

const opOptions = [
  { value: 'equals', label: '=' },
  { value: 'not_equals', label: '≠' },
  { value: 'is_empty', label: 'is empty' },
  { value: 'is_not_empty', label: 'is not empty' },
]

const needsValue = ['equals', 'not_equals']

function onRemove() {
  props.data.onRemove?.(props.id)
}
</script>

<template>
  <div class="min-w-[220px] rounded-xl border-2 border-amber-500/40 bg-amber-500/5 shadow-sm">
    <Handle type="target" :position="Position.Top" />
    <div class="p-4">
      <div class="mb-3 flex items-center gap-2">
        <span class="text-amber-500">⑂</span>
        <span class="text-[10px] font-bold uppercase tracking-widest text-amber-600/70">Branch</span>
        <div class="flex-1"></div>
        <button class="text-xs text-muted-foreground/50 hover:text-destructive" @click="onRemove">✕</button>
      </div>
      <div class="space-y-2">
        <select v-model="props.data.conditionField" class="w-full rounded-lg border border-amber-500/30 bg-background px-2 py-1.5 text-xs">
          <option v-for="f in fieldOptions" :key="f.value" :value="f.value">{{ f.label }}</option>
        </select>
        <select v-model="props.data.conditionOperator" class="w-full rounded-lg border border-amber-500/30 bg-background px-2 py-1.5 text-xs">
          <option v-for="o in opOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
        </select>
        <template v-if="needsValue.includes(props.data.conditionOperator)">
          <template v-if="props.data.conditionField === 'priority'">
            <select v-model="props.data.conditionValue" class="w-full rounded-lg border border-amber-500/30 bg-background px-2 py-1.5 text-xs">
              <option value="low">Low</option>
              <option value="medium">Medium</option>
              <option value="high">High</option>
              <option value="urgent">Urgent</option>
            </select>
          </template>
          <template v-else-if="props.data.conditionField === 'state_id'">
            <select v-model="props.data.conditionValue" class="w-full rounded-lg border border-amber-500/30 bg-background px-2 py-1.5 text-xs">
              <option value="">Select state…</option>
              <option v-for="s in props.data.workflowStates" :key="s.id" :value="s.id">{{ s.name }}</option>
            </select>
          </template>
          <template v-else-if="props.data.conditionField === 'type'">
            <select v-model="props.data.conditionValue" class="w-full rounded-lg border border-amber-500/30 bg-background px-2 py-1.5 text-xs">
              <option value="feature">Feature</option>
              <option value="bug">Bug</option>
            </select>
          </template>
          <template v-else>
            <input
              v-model="props.data.conditionValue"
              class="w-full rounded-lg border border-amber-500/30 bg-background px-2 py-1.5 text-xs"
              placeholder="Value…"
            />
          </template>
        </template>
      </div>
      <!-- True/False handle labels -->
      <div class="mt-3 flex justify-between text-[10px] text-muted-foreground">
        <span class="text-emerald-600">✓ true</span>
        <span class="text-red-500">✗ false</span>
      </div>
    </div>
    <!-- Two output handles: true (left) and false (right) -->
    <Handle id="true" type="source" :position="Position.Bottom" style="left: 30%" />
    <Handle id="false" type="source" :position="Position.Bottom" style="left: 70%" />
  </div>
</template>
