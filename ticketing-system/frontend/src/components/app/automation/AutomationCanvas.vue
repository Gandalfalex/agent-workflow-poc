<script setup lang="ts">
import { ref, watch, markRaw } from 'vue'
import { VueFlow, useVueFlow, type Node, type Edge, type EdgeMouseEvent } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { Controls } from '@vue-flow/controls'
import { MiniMap } from '@vue-flow/minimap'
import TriggerNode from './TriggerNode.vue'
import ActionNode from './ActionNode.vue'
import BranchNode from './BranchNode.vue'
import '@vue-flow/core/dist/style.css'
import '@vue-flow/controls/dist/style.css'
import '@vue-flow/minimap/dist/style.css'

interface SimEvent {
  matched: boolean
  predictedActions?: Array<{ type: string; wouldSucceed: boolean; failureReason?: string | null }>
}

const props = defineProps<{
  modelValue: { nodes: Node[]; edges: Edge[] } | null
  workflowStates: { id: string; name: string }[]
  users: { id: string; name?: string | null; email: string }[]
  showValidation?: boolean
  simulatorEvent?: SimEvent | null
}>()

const emit = defineEmits<{
  'update:modelValue': [{ nodes: Node[]; edges: Edge[] }]
}>()

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const nodeTypes = markRaw({
  trigger: markRaw(TriggerNode),
  action: markRaw(ActionNode),
  branch: markRaw(BranchNode),
}) as any

const nodes = ref<Node[]>([])
const edges = ref<Edge[]>([])

function enrich(data: Record<string, any>): Record<string, any> {
  return {
    ...data,
    workflowStates: props.workflowStates,
    users: props.users,
    onRemove,
    onUpdate: emitUpdate,
  }
}

function stripRuntime(data: Record<string, any>): Record<string, any> {
  const { workflowStates: _a, users: _b, onRemove: _c, onUpdate: _d, ...rest } = data
  return rest
}

watch(
  () => props.modelValue,
  (val) => {
    if (val && val.nodes.length > 0) {
      nodes.value = val.nodes.map((n) => ({ ...n, data: enrich(n.data) }))
      edges.value = val.edges
    } else {
      nodes.value = [
        {
          id: 'trigger-1',
          type: 'trigger',
          position: { x: 240, y: 60 },
          data: enrich({ triggerEvent: 'ticket.created', condFromState: '', condToState: '' }),
        },
      ]
      edges.value = []
      emitUpdate()
    }
  },
  { immediate: true },
)

watch([() => props.workflowStates, () => props.users], () => {
  nodes.value = nodes.value.map((n) => ({ ...n, data: enrich(n.data as Record<string, any>) }))
})

function onRemove(nodeId: string) {
  edges.value = edges.value.filter((e) => e.source !== nodeId && e.target !== nodeId)
  nodes.value = nodes.value.filter((n) => n.id !== nodeId)
  emitUpdate()
}

function emitUpdate() {
  emit('update:modelValue', {
    nodes: nodes.value.map((n) => ({ ...n, data: stripRuntime(n.data as Record<string, any>) })),
    edges: edges.value,
  })
}

// Node validation
function isNodeValid(node: Node): boolean {
  const d = node.data as Record<string, any>
  if (node.type === 'trigger') return true
  if (node.type === 'branch') return !!(d.conditionField && d.conditionOperator)
  if (node.type === 'action') {
    if (d.actionType === 'set_state') return !!(d.params?.state_id)
    if (d.actionType === 'set_assignee') return !!(d.params?.assignee_id)
    if (d.actionType === 'add_comment') return !!(d.params?.body)
    return true
  }
  return true
}

function nodeClass(node: Node): string {
  const classes: string[] = []
  if (props.showValidation && !isNodeValid(node)) classes.push('vf-invalid')
  if (props.simulatorEvent) {
    if (node.type === 'trigger') {
      classes.push(props.simulatorEvent.matched ? 'vf-sim-hit' : 'vf-sim-miss')
    } else if (node.type === 'action' && props.simulatorEvent.matched) {
      const actionNodes = nodes.value.filter((n) => n.type === 'action')
      const idx = actionNodes.findIndex((n) => n.id === node.id)
      const pred = props.simulatorEvent.predictedActions?.[idx]
      if (pred !== undefined) classes.push(pred.wouldSucceed ? 'vf-sim-hit' : 'vf-sim-warn')
    }
  }
  return classes.join(' ')
}

const { onConnect, addEdges, removeEdges } = useVueFlow()

onConnect((params) => {
  const sourceBranch = nodes.value.find((n) => n.id === params.source && n.type === 'branch')
  const label = sourceBranch
    ? params.sourceHandle === 'true' ? '✓ true'
      : params.sourceHandle === 'false' ? '✗ false'
      : undefined
    : undefined

  addEdges([{
    ...params,
    id: `e-${Date.now()}`,
    ...(label ? {
      label,
      labelStyle: { fontSize: 10, fontWeight: 600, fill: label.startsWith('✓') ? '#16a34a' : '#ef4444' },
      labelBgStyle: { fill: 'transparent' },
    } : {}),
  }])
  emitUpdate()
})

function onEdgeDblClick({ edge }: EdgeMouseEvent) {
  removeEdges([edge as Edge])
  emitUpdate()
}

function addActionNode() {
  const id = `action-${Date.now()}`
  const lastY = nodes.value.reduce((max, n) => Math.max(max, n.position.y), 0)
  nodes.value = [
    ...nodes.value,
    {
      id,
      type: 'action',
      position: { x: 240, y: lastY + 180 },
      data: enrich({ actionType: 'add_comment', params: { body: '' } }),
    },
  ]
  emitUpdate()
}

function addBranchNode() {
  const id = `branch-${Date.now()}`
  const lastY = nodes.value.reduce((max, n) => Math.max(max, n.position.y), 0)
  nodes.value = [
    ...nodes.value,
    {
      id,
      type: 'branch',
      position: { x: 240, y: lastY + 180 },
      data: enrich({ conditionField: 'priority', conditionOperator: 'equals', conditionValue: 'urgent' }),
    },
  ]
  emitUpdate()
}
</script>

<template>
  <div data-testid="automation.canvas" class="relative rounded-xl border border-border" style="height: 580px;">
    <VueFlow
      v-model:nodes="nodes"
      v-model:edges="edges"
      :node-types="nodeTypes"
      :node-class="nodeClass"
      :default-edge-options="{ type: 'smoothstep', animated: true, style: { strokeWidth: 2 } }"
      :connection-line-style="{ strokeWidth: 2, stroke: 'hsl(var(--primary))' }"
      :snap-to-grid="true"
      :snap-grid="[20, 20]"
      :delete-key-code="'Delete'"
      :min-zoom="0.3"
      :max-zoom="2"
      fit-view-on-init
      class="bg-muted/10"
      @nodes-change="emitUpdate"
      @edges-change="emitUpdate"
      @edge-double-click="onEdgeDblClick"
    >
      <Background :gap="20" :size="1" pattern-color="hsl(var(--border))" />
      <Controls position="top-left" :show-interactive="false" />
      <MiniMap
        position="bottom-right"
        :height="90"
        :width="140"
        :node-stroke-width="3"
        pannable
        zoomable
      />
    </VueFlow>

    <!-- Add node toolbar -->
    <div class="absolute bottom-4 left-1/2 z-10 flex -translate-x-1/2 items-center gap-2 rounded-xl border border-border bg-background/95 px-3 py-1.5 shadow-md backdrop-blur">
      <span class="mr-1 text-[10px] text-muted-foreground">Add:</span>
      <button
        data-testid="automation.add_action_node"
        class="rounded-lg border border-border bg-muted/60 px-2.5 py-1 text-xs font-medium transition-colors hover:border-primary/40 hover:bg-primary/5"
        @click="addActionNode"
      >
        + Action
      </button>
      <button
        data-testid="automation.add_branch_node"
        class="rounded-lg border border-amber-500/40 bg-amber-500/10 px-2.5 py-1 text-xs font-medium text-amber-600 transition-colors hover:border-amber-500/70"
        @click="addBranchNode"
      >
        ⑂ Branch
      </button>
      <span class="ml-1 text-[10px] text-muted-foreground/50">· Drag handles to connect · Dbl-click edge to delete</span>
    </div>
  </div>
</template>

<style>
/* Larger, more visible handles */
.vue-flow__handle {
  width: 14px !important;
  height: 14px !important;
  border-radius: 50% !important;
  background: hsl(var(--primary)) !important;
  border: 2px solid hsl(var(--background)) !important;
  box-shadow: 0 0 0 2px hsl(var(--primary) / 0.5) !important;
  transition: transform 0.15s, box-shadow 0.15s !important;
}
.vue-flow__handle:hover {
  transform: scale(1.35) !important;
  box-shadow: 0 0 0 3px hsl(var(--primary) / 0.4) !important;
  cursor: crosshair !important;
}
.vue-flow__handle-left,
.vue-flow__handle-right {
  background: hsl(var(--amber-500, 245 158 11)) !important;
}
/* Branch handles: true=green, false=red */
.vue-flow__handle[data-handleid="true"] {
  background: #16a34a !important;
  box-shadow: 0 0 0 2px #16a34a66 !important;
}
.vue-flow__handle[data-handleid="false"] {
  background: #ef4444 !important;
  box-shadow: 0 0 0 2px #ef444466 !important;
}

/* Edges */
.vue-flow__edge-path {
  stroke-width: 2;
}
.vue-flow__edge.selected .vue-flow__edge-path {
  stroke: hsl(var(--primary)) !important;
  stroke-width: 2.5;
}
.vue-flow__edge:hover .vue-flow__edge-path {
  stroke-width: 3;
}

/* Connection line */
.vue-flow__connectionline path {
  stroke: hsl(var(--primary));
  stroke-width: 2;
  stroke-dasharray: 6 3;
}

/* Validation / simulator node outlines */
.vf-invalid > .vue-flow__node-default,
.vf-invalid > div {
  outline: 2px solid rgba(239, 68, 68, 0.65);
  outline-offset: 2px;
  border-radius: 12px;
}
.vf-sim-hit > div { outline: 2px solid rgba(34, 197, 94, 0.75); outline-offset: 2px; border-radius: 12px; }
.vf-sim-miss > div { outline: 2px solid rgba(239, 68, 68, 0.75); outline-offset: 2px; border-radius: 12px; }
.vf-sim-warn > div { outline: 2px solid rgba(245, 158, 11, 0.75); outline-offset: 2px; border-radius: 12px; }

/* Controls & MiniMap theming */
.vue-flow__controls {
  box-shadow: 0 2px 8px rgba(0,0,0,0.12);
  border-radius: 8px;
  overflow: hidden;
}
.vue-flow__minimap {
  border-radius: 8px;
  overflow: hidden;
  opacity: 0.85;
}
</style>
