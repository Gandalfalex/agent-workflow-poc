<script setup lang="ts">
import { Handle, Position } from '@vue-flow/core'

const props = defineProps<{ data: any; id: string }>()

const actionConfig: Record<string, { icon: string; color: string; label: string }> = {
  set_state:    { icon: '⇒', color: 'blue',    label: 'Set State' },
  set_assignee: { icon: '◎', color: 'violet',  label: 'Set Assignee' },
  set_priority: { icon: '▲', color: 'orange',  label: 'Set Priority' },
  add_comment:  { icon: '◆', color: 'emerald', label: 'Add Comment' },
  call_webhook: { icon: '↗', color: 'slate',   label: 'Call Webhook' },
}

const borderColors: Record<string, string> = {
  set_state:    'border-l-blue-500/70',
  set_assignee: 'border-l-violet-500/70',
  set_priority: 'border-l-orange-500/70',
  add_comment:  'border-l-emerald-500/70',
  call_webhook: 'border-l-slate-500/70',
}

const headerColors: Record<string, string> = {
  set_state:    'bg-blue-500/8 border-blue-500/20 text-blue-600',
  set_assignee: 'bg-violet-500/8 border-violet-500/20 text-violet-600',
  set_priority: 'bg-orange-500/8 border-orange-500/20 text-orange-600',
  add_comment:  'bg-emerald-500/8 border-emerald-500/20 text-emerald-600',
  call_webhook: 'bg-slate-500/8 border-slate-500/20 text-slate-500',
}

function onRemove() {
  props.data.onRemove?.(props.id)
}

function onTypeChange() {
  const defaults: Record<string, Record<string, string>> = {
    set_state:    { state_id: '' },
    set_assignee: { assignee_id: '' },
    set_priority: { priority: 'medium' },
    add_comment:  { body: '' },
    call_webhook: {},
  }
  props.data.params = defaults[props.data.actionType] ?? {}
  props.data.onUpdate?.()
}
</script>

<template>
  <div class="min-w-[230px] rounded-xl border border-l-4 bg-background shadow-sm" :class="borderColors[props.data.actionType] ?? 'border-l-slate-500/70'">
    <Handle type="target" :position="Position.Top" />

    <!-- Drag handle / header -->
    <div
      class="flex cursor-grab items-center gap-2 rounded-tr-xl border-b px-3 py-1.5 active:cursor-grabbing"
      :class="headerColors[props.data.actionType] ?? 'bg-muted/40 border-border text-muted-foreground'"
    >
      <span class="text-sm">{{ actionConfig[props.data.actionType]?.icon ?? '◇' }}</span>
      <span class="text-[10px] font-bold uppercase tracking-widest">Action</span>
      <span class="ml-auto text-[10px] opacity-40">⠿</span>
      <button class="nodrag ml-1 rounded p-0.5 text-xs opacity-40 transition-opacity hover:opacity-100 hover:text-destructive" @click="onRemove">✕</button>
    </div>

    <div class="nodrag p-3 space-y-2">
      <select
        v-model="props.data.actionType"
        class="nodrag w-full rounded-lg border border-border bg-muted px-2 py-1.5 text-xs focus:outline-none"
        @change="onTypeChange"
      >
        <option value="set_state">Set State</option>
        <option value="set_assignee">Set Assignee</option>
        <option value="set_priority">Set Priority</option>
        <option value="add_comment">Add Comment</option>
        <option value="call_webhook">Call Webhook</option>
      </select>

      <div v-if="props.data.actionType === 'set_state'">
        <select v-model="props.data.params.state_id" class="nodrag w-full rounded-lg border border-border bg-muted px-2 py-1.5 text-xs focus:outline-none" @change="props.data.onUpdate?.()">
          <option value="">Select state…</option>
          <option v-for="s in props.data.workflowStates" :key="s.id" :value="s.id">{{ s.name }}</option>
        </select>
      </div>
      <div v-else-if="props.data.actionType === 'set_assignee'">
        <select v-model="props.data.params.assignee_id" class="nodrag w-full rounded-lg border border-border bg-muted px-2 py-1.5 text-xs focus:outline-none" @change="props.data.onUpdate?.()">
          <option value="">Select user…</option>
          <option v-for="u in props.data.users" :key="u.id" :value="u.id">{{ u.name || u.email }}</option>
        </select>
      </div>
      <div v-else-if="props.data.actionType === 'set_priority'">
        <select v-model="props.data.params.priority" class="nodrag w-full rounded-lg border border-border bg-muted px-2 py-1.5 text-xs focus:outline-none" @change="props.data.onUpdate?.()">
          <option value="low">Low</option>
          <option value="medium">Medium</option>
          <option value="high">High</option>
          <option value="urgent">Urgent</option>
        </select>
      </div>
      <div v-else-if="props.data.actionType === 'add_comment'">
        <input
          v-model="props.data.params.body"
          class="nodrag w-full rounded-lg border border-border bg-muted px-2 py-1.5 text-xs focus:outline-none"
          placeholder="Comment text…"
          @change="props.data.onUpdate?.()"
        />
      </div>
      <div v-else-if="props.data.actionType === 'call_webhook'" class="text-[10px] text-muted-foreground">
        Fires all configured project webhooks.
      </div>
    </div>

    <Handle type="source" :position="Position.Bottom" />
  </div>
</template>
