<script setup lang="ts">
import { Handle, Position } from '@vue-flow/core'

const props = defineProps<{ data: any }>()

const triggerLabels: Record<string, string> = {
  'ticket.created': 'Ticket Created',
  'ticket.updated': 'Ticket Updated',
  'ticket.state_changed': 'State Changed',
  'ticket.priority_changed': 'Priority Changed',
  'ticket.assigned': 'Ticket Assigned',
  'ticket.deleted': 'Ticket Deleted',
}
</script>

<template>
  <div class="min-w-[200px] rounded-xl border-2 border-primary/50 bg-primary/5 p-4 shadow-sm">
    <div class="mb-2 text-[10px] font-bold uppercase tracking-widest text-primary/60">Trigger</div>
    <select v-model="props.data.triggerEvent" class="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm">
      <option v-for="(label, val) in triggerLabels" :key="val" :value="val">{{ label }}</option>
    </select>
    <template v-if="props.data.triggerEvent === 'ticket.state_changed'">
      <div class="mt-2 grid grid-cols-2 gap-2">
        <div>
          <label class="mb-1 block text-[10px] text-muted-foreground">From state</label>
          <select v-model="props.data.condFromState" class="w-full rounded border border-border bg-background px-2 py-1 text-xs">
            <option value="">Any</option>
            <option v-for="s in props.data.workflowStates" :key="s.id" :value="s.id">{{ s.name }}</option>
          </select>
        </div>
        <div>
          <label class="mb-1 block text-[10px] text-muted-foreground">To state</label>
          <select v-model="props.data.condToState" class="w-full rounded border border-border bg-background px-2 py-1 text-xs">
            <option value="">Any</option>
            <option v-for="s in props.data.workflowStates" :key="s.id" :value="s.id">{{ s.name }}</option>
          </select>
        </div>
      </div>
    </template>
    <Handle type="source" :position="Position.Bottom" />
  </div>
</template>
