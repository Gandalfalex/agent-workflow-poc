<script setup lang="ts">
import { ref, watch, markRaw } from 'vue'
import { type Node, type Edge, type Connection } from '@vue-flow/core'
import GraphCanvas from './GraphCanvas.vue'
import TriggerNode from './TriggerNode.vue'
import ActionNode from './ActionNode.vue'
import BranchNode from './BranchNode.vue'

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
  action:  markRaw(ActionNode),
  branch:  markRaw(BranchNode),
}) as any

// Internal enriched graph — nodes carry runtime callbacks + context data.
const internalGraph = ref<{ nodes: Node[]; edges: Edge[] }>({ nodes: [], edges: [] })
const canvasRef = ref<InstanceType<typeof GraphCanvas> | null>(null)

// Prevent the watch from re-enriching after our own emitStripped() round-trip.
let skipNextModelValueUpdate = false

// ── Enrichment / stripping ───────────────────────────────────────────────────

function enrich(data: Record<string, any>): Record<string, any> {
  return {
    ...data,
    workflowStates: props.workflowStates,
    users: props.users,
    onRemove,
    onUpdate: emitStripped,
  }
}

function stripRuntime(data: Record<string, any>): Record<string, any> {
  const { workflowStates: _a, users: _b, onRemove: _c, onUpdate: _d, ...rest } = data
  return rest
}

function emitStripped() {
  emit('update:modelValue', {
    nodes: internalGraph.value.nodes.map(n => ({ ...n, data: stripRuntime(n.data as Record<string, any>) })),
    edges: internalGraph.value.edges,
  })
}

// ── Sync from parent ─────────────────────────────────────────────────────────

watch(
  () => props.modelValue,
  (val) => {
    if (skipNextModelValueUpdate) {
      skipNextModelValueUpdate = false
      return
    }
    if (val && val.nodes.length > 0) {
      internalGraph.value = {
        nodes: val.nodes.map(n => ({ ...n, data: enrich(n.data as Record<string, any>) })),
        edges: val.edges,
      }
    } else {
      internalGraph.value = {
        nodes: [{
          id: 'trigger-1',
          type: 'trigger',
          position: { x: 240, y: 60 },
          data: enrich({ triggerEvent: 'ticket.created', condFromState: '', condToState: '' }),
        }],
        edges: [],
      }
      emitStripped()
    }
  },
  { immediate: true },
)

// Re-enrich when context props change (new states/users added)
watch([() => props.workflowStates, () => props.users], () => {
  internalGraph.value = {
    nodes: internalGraph.value.nodes.map(n => ({ ...n, data: enrich(n.data as Record<string, any>) })),
    edges: internalGraph.value.edges,
  }
})

// ── Callbacks injected into node data ────────────────────────────────────────

function onRemove(nodeId: string) {
  internalGraph.value = {
    nodes: internalGraph.value.nodes.filter(n => n.id !== nodeId),
    edges: internalGraph.value.edges.filter(e => e.source !== nodeId && e.target !== nodeId),
  }
  emitStripped()
}

// ── Handle updates emitted by GraphCanvas ────────────────────────────────────

function onGraphUpdate(val: { nodes: Node[]; edges: Edge[] }) {
  skipNextModelValueUpdate = true
  internalGraph.value = val
  emitStripped()
}

// ── Edge creation (branch-aware labels) ──────────────────────────────────────

function createEdge(params: Connection, currentNodes: Node[]): Edge {
  const sourceBranch = currentNodes.find(n => n.id === params.source && n.type === 'branch')
  const label = sourceBranch
    ? params.sourceHandle === 'true'  ? '✓ true'
    : params.sourceHandle === 'false' ? '✗ false'
    : undefined
    : undefined
  return {
    ...params,
    id: `e-${Date.now()}`,
    ...(label ? {
      label,
      labelStyle: { fontSize: 10, fontWeight: 600, fill: label.startsWith('✓') ? '#16a34a' : '#ef4444' },
      labelBgStyle: { fill: 'transparent' },
    } : {}),
  } as Edge
}

// ── Node validation / styling ─────────────────────────────────────────────────

function isNodeValid(node: Node): boolean {
  const d = node.data as Record<string, any>
  if (node.type === 'trigger') return true
  if (node.type === 'branch') return !!(d.conditionField && d.conditionOperator)
  if (node.type === 'action') {
    if (d.actionType === 'set_state')    return !!(d.params?.state_id)
    if (d.actionType === 'set_assignee') return !!(d.params?.assignee_id)
    if (d.actionType === 'add_comment')  return !!(d.params?.body)
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
      const actionNodes = internalGraph.value.nodes.filter(n => n.type === 'action')
      const idx = actionNodes.findIndex(n => n.id === node.id)
      const pred = props.simulatorEvent.predictedActions?.[idx]
      if (pred !== undefined) classes.push(pred.wouldSucceed ? 'vf-sim-hit' : 'vf-sim-warn')
    }
  }
  return classes.join(' ')
}

// ── Add nodes ─────────────────────────────────────────────────────────────────

function addActionNode() {
  const id = `action-${Date.now()}`
  const lastY = internalGraph.value.nodes.reduce((max, n) => Math.max(max, n.position.y), 0)
  internalGraph.value = {
    nodes: [...internalGraph.value.nodes, {
      id,
      type: 'action',
      position: { x: 240, y: lastY + 180 },
      data: enrich({ actionType: 'add_comment', params: { body: '' } }),
    }],
    edges: internalGraph.value.edges,
  }
  emitStripped()
  setTimeout(() => canvasRef.value?.fitView({ padding: 0.2, duration: 300 }), 50)
}

function addBranchNode() {
  const id = `branch-${Date.now()}`
  const lastY = internalGraph.value.nodes.reduce((max, n) => Math.max(max, n.position.y), 0)
  internalGraph.value = {
    nodes: [...internalGraph.value.nodes, {
      id,
      type: 'branch',
      position: { x: 240, y: lastY + 180 },
      data: enrich({ conditionField: 'priority', conditionOperator: 'equals', conditionValue: 'urgent' }),
    }],
    edges: internalGraph.value.edges,
  }
  emitStripped()
  setTimeout(() => canvasRef.value?.fitView({ padding: 0.2, duration: 300 }), 50)
}
</script>

<template>
  <div data-testid="automation.canvas" class="relative rounded-xl border border-border" style="height: 580px;">
    <GraphCanvas
      ref="canvasRef"
      :model-value="internalGraph"
      :node-types="nodeTypes"
      :node-class="nodeClass"
      :show-minimap="true"
      :default-edge-options="{ type: 'smoothstep', animated: true, style: { strokeWidth: 2 } }"
      :create-edge="createEdge"
      @update:model-value="onGraphUpdate"
    />

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
