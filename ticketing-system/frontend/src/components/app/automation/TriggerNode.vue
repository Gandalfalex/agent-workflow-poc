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
  <div class="min-w-[220px] rounded-xl border-2 border-primary/50 bg-primary/5 shadow-sm">
    <!-- Drag handle -->
    <div class="flex cursor-grab items-center gap-2 rounded-t-xl border-b border-primary/20 bg-primary/10 px-3 py-1.5 active:cursor-grabbing">
      <span class="text-[10px] font-bold uppercase tracking-widest text-primary/70">⚡ Trigger</span>
      <span class="ml-auto text-[10px] text-primary/40">⠿</span>
    </div>
    <div class="nodrag p-3">
      <select v-model="props.data.triggerEvent" class="nodrag w-full rounded-lg border border-primary/20 bg-background px-2 py-1.5 text-sm focus:outline-none focus:ring-1 focus:ring-primary/40" @change="props.data.onUpdate?.()">
        <option v-for="(label, val) in triggerLabels" :key="val" :value="val">{{ label }}</option>
      </select>
      <template v-if="props.data.triggerEvent === 'ticket.state_changed'">
        <div class="mt-2 grid grid-cols-2 gap-2">
          <div>
            <label class="mb-1 block text-[10px] text-muted-foreground">From</label>
            <select v-model="props.data.condFromState" class="nodrag w-full rounded border border-primary/20 bg-background px-2 py-1 text-xs focus:outline-none" @change="props.data.onUpdate?.()">
              <option value="">Any</option>
              <option v-for="s in props.data.workflowStates" :key="s.id" :value="s.id">{{ s.name }}</option>
            </select>
          </div>
          <div>
            <label class="mb-1 block text-[10px] text-muted-foreground">To</label>
            <select v-model="props.data.condToState" class="nodrag w-full rounded border border-primary/20 bg-background px-2 py-1 text-xs focus:outline-none" @change="props.data.onUpdate?.()">
              <option value="">Any</option>
              <option v-for="s in props.data.workflowStates" :key="s.id" :value="s.id">{{ s.name }}</option>
            </select>
          </div>
        </div>
      </template>
    </div>
    <Handle type="source" :position="Position.Bottom" />
  </div>
</template>
