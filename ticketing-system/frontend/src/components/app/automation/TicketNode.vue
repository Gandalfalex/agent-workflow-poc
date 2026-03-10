<script setup lang="ts">
// Read-only node for the dependency graph overlay.
// No handles — connections are not supported here.
const props = defineProps<{
  id: string
  data: {
    key: string
    title: string
    isCurrent: boolean
    onClick: (id: string) => void
  }
}>()

function shortTitle(t: string) {
  const s = t.trim()
  return s.length <= 22 ? s : s.slice(0, 21) + '…'
}
</script>

<template>
  <div
    :data-testid="`ticket.dependency-graph-node-${props.id}`"
    class="flex w-[160px] cursor-pointer select-none flex-col justify-center rounded-lg border px-3 py-1.5 shadow-sm transition-colors"
    :class="data.isCurrent
      ? 'border-blue-400/60 bg-blue-900/40 text-blue-100 hover:bg-blue-900/60'
      : 'border-border bg-card hover:border-primary/40 hover:bg-muted/40'"
    @click="data.onClick(props.id)"
  >
    <span class="text-[11px] font-bold leading-none" :class="data.isCurrent ? 'text-blue-200' : 'text-foreground/80'">
      {{ data.key }}
    </span>
    <span class="mt-0.5 text-[10px] leading-snug" :class="data.isCurrent ? 'text-blue-300/80' : 'text-muted-foreground'">
      {{ shortTitle(data.title) }}
    </span>
  </div>
</template>
