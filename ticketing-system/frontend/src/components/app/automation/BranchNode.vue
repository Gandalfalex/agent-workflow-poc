<script setup lang="ts">
import { Handle, Position } from '@vue-flow/core'

const props = defineProps<{ data: any; id: string }>()

const fieldOptions = [
  { value: 'priority',    label: 'Priority' },
  { value: 'state_id',   label: 'State' },
  { value: 'assignee_id',label: 'Assignee' },
  { value: 'type',       label: 'Type' },
]

const opOptions = [
  { value: 'equals',       label: '= equals' },
  { value: 'not_equals',   label: '≠ not equals' },
  { value: 'is_empty',     label: 'is empty' },
  { value: 'is_not_empty', label: 'is not empty' },
]

const needsValue = ['equals', 'not_equals']

function onRemove() {
  props.data.onRemove?.(props.id)
}
</script>

<template>
  <div class="min-w-[230px] rounded-xl border-2 border-amber-500/50 bg-amber-500/5 shadow-sm">
    <Handle type="target" :position="Position.Top" />

    <!-- Drag handle / header -->
    <div class="flex cursor-grab items-center gap-2 rounded-t-xl border-b border-amber-500/30 bg-amber-500/10 px-3 py-1.5 active:cursor-grabbing">
      <span class="text-amber-500">⑂</span>
      <span class="text-[10px] font-bold uppercase tracking-widest text-amber-700/80">Branch</span>
      <span class="ml-auto text-[10px] text-amber-500/40">⠿</span>
      <button class="nodrag ml-1 rounded p-0.5 text-xs text-amber-700/40 transition-opacity hover:text-destructive" @click="onRemove">✕</button>
    </div>

    <div class="nodrag p-3 space-y-2">
      <div class="grid grid-cols-2 gap-2">
        <select v-model="props.data.conditionField" class="nodrag rounded-lg border border-amber-500/30 bg-background px-2 py-1.5 text-xs focus:outline-none" @change="props.data.onUpdate?.()">
          <option v-for="f in fieldOptions" :key="f.value" :value="f.value">{{ f.label }}</option>
        </select>
        <select v-model="props.data.conditionOperator" class="nodrag rounded-lg border border-amber-500/30 bg-background px-2 py-1.5 text-xs focus:outline-none" @change="props.data.onUpdate?.()">
          <option v-for="o in opOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
        </select>
      </div>

      <template v-if="needsValue.includes(props.data.conditionOperator)">
        <select v-if="props.data.conditionField === 'priority'" v-model="props.data.conditionValue" class="nodrag w-full rounded-lg border border-amber-500/30 bg-background px-2 py-1.5 text-xs focus:outline-none" @change="props.data.onUpdate?.()">
          <option value="low">Low</option>
          <option value="medium">Medium</option>
          <option value="high">High</option>
          <option value="urgent">Urgent</option>
        </select>
        <select v-else-if="props.data.conditionField === 'state_id'" v-model="props.data.conditionValue" class="nodrag w-full rounded-lg border border-amber-500/30 bg-background px-2 py-1.5 text-xs focus:outline-none" @change="props.data.onUpdate?.()">
          <option value="">Select state…</option>
          <option v-for="s in props.data.workflowStates" :key="s.id" :value="s.id">{{ s.name }}</option>
        </select>
        <select v-else-if="props.data.conditionField === 'type'" v-model="props.data.conditionValue" class="nodrag w-full rounded-lg border border-amber-500/30 bg-background px-2 py-1.5 text-xs focus:outline-none" @change="props.data.onUpdate?.()">
          <option value="feature">Feature</option>
          <option value="bug">Bug</option>
        </select>
        <input v-else v-model="props.data.conditionValue" class="nodrag w-full rounded-lg border border-amber-500/30 bg-background px-2 py-1.5 text-xs focus:outline-none" placeholder="Value…" @change="props.data.onUpdate?.()" />
      </template>

      <!-- Output labels -->
      <div class="flex justify-between pt-1 text-[10px] font-semibold">
        <span class="text-emerald-600">✓ true</span>
        <span class="text-red-500">✗ false</span>
      </div>
    </div>

    <!-- Two source handles with colored indicators -->
    <Handle id="true"  type="source" :position="Position.Bottom" style="left: 28%" />
    <Handle id="false" type="source" :position="Position.Bottom" style="left: 72%" />
  </div>
</template>
