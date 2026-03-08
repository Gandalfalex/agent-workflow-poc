<script setup lang="ts">
import { Handle, Position } from '@vue-flow/core'

const props = defineProps<{ data: any; id: string }>()

const actionColors: Record<string, string> = {
  set_state: 'blue',
  set_assignee: 'violet',
  set_priority: 'orange',
  add_comment: 'emerald',
  call_webhook: 'slate',
}

const actionIcons: Record<string, string> = {
  set_state: '⇒',
  set_assignee: '◎',
  set_priority: '▲',
  add_comment: '◆',
  call_webhook: '↗',
}

function onRemove() {
  props.data.onRemove?.(props.id)
}

function onTypeChange() {
  const defaults: Record<string, Record<string, string>> = {
    set_state: { state_id: '' },
    set_assignee: { assignee_id: '' },
    set_priority: { priority: 'medium' },
    add_comment: { body: '' },
    call_webhook: {},
  }
  props.data.params = defaults[props.data.actionType] ?? {}
}
</script>

<template>
  <div
    class="min-w-[220px] rounded-xl border bg-background shadow-sm"
    :class="`border-l-4 border-l-${actionColors[props.data.actionType] ?? 'slate'}-500/60`"
  >
    <Handle type="target" :position="Position.Top" />
    <div class="p-4">
      <div class="mb-3 flex items-center gap-2">
        <span>{{ actionIcons[props.data.actionType] }}</span>
        <span class="text-[10px] font-bold uppercase tracking-widest text-muted-foreground">Action</span>
        <div class="flex-1"></div>
        <button class="text-xs text-muted-foreground/50 hover:text-destructive" @click="onRemove">✕</button>
      </div>
      <select
        v-model="props.data.actionType"
        class="mb-3 w-full rounded-lg border border-border bg-muted px-2 py-1.5 text-xs"
        @change="onTypeChange"
      >
        <option value="set_state">Set State</option>
        <option value="set_assignee">Set Assignee</option>
        <option value="set_priority">Set Priority</option>
        <option value="add_comment">Add Comment</option>
        <option value="call_webhook">Call Webhook</option>
      </select>
      <div v-if="props.data.actionType === 'set_state'">
        <select v-model="props.data.params.state_id" class="w-full rounded-lg border border-border bg-muted px-2 py-1.5 text-xs">
          <option value="">Select state…</option>
          <option v-for="s in props.data.workflowStates" :key="s.id" :value="s.id">{{ s.name }}</option>
        </select>
      </div>
      <div v-else-if="props.data.actionType === 'set_assignee'">
        <select v-model="props.data.params.assignee_id" class="w-full rounded-lg border border-border bg-muted px-2 py-1.5 text-xs">
          <option value="">Select user…</option>
          <option v-for="u in props.data.users" :key="u.id" :value="u.id">{{ u.name || u.email }}</option>
        </select>
      </div>
      <div v-else-if="props.data.actionType === 'set_priority'">
        <select v-model="props.data.params.priority" class="w-full rounded-lg border border-border bg-muted px-2 py-1.5 text-xs">
          <option value="low">Low</option>
          <option value="medium">Medium</option>
          <option value="high">High</option>
          <option value="urgent">Urgent</option>
        </select>
      </div>
      <div v-else-if="props.data.actionType === 'add_comment'">
        <input
          v-model="props.data.params.body"
          class="w-full rounded-lg border border-border bg-muted px-2 py-1.5 text-xs"
          placeholder="Comment…"
        />
      </div>
      <div v-else-if="props.data.actionType === 'call_webhook'" class="text-[10px] text-muted-foreground">
        Fires all project webhooks.
      </div>
    </div>
    <Handle type="source" :position="Position.Bottom" />
  </div>
</template>
