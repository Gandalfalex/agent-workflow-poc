<script setup lang="ts">
import { ref, watch, markRaw } from 'vue'
import { VueFlow, useVueFlow, type Node, type Edge } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import TriggerNode from './TriggerNode.vue'
import ActionNode from './ActionNode.vue'
import BranchNode from './BranchNode.vue'
import '@vue-flow/core/dist/style.css'

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
  }
}

function stripRuntime(data: Record<string, any>): Record<string, any> {
  const { workflowStates: _a, users: _b, onRemove: _c, ...rest } = data
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
          position: { x: 200, y: 50 },
          data: enrich({ triggerEvent: 'ticket.created', condFromState: '', condToState: '' }),
        },
      ]
      edges.value = []
      emitUpdate()
    }
  },
  { immediate: true },
)

// Re-enrich when workflowStates/users arrive async
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

// Validation
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

// Node class for validation + simulator highlights
function nodeClass(node: Node): string {
  const classes: string[] = []

  if (props.showValidation && !isNodeValid(node)) {
    classes.push('vf-invalid')
  }

  if (props.simulatorEvent) {
    if (node.type === 'trigger') {
      classes.push(props.simulatorEvent.matched ? 'vf-sim-hit' : 'vf-sim-miss')
    } else if (node.type === 'action' && props.simulatorEvent.matched) {
      // Map action nodes in order to predictedActions
      const actionNodes = nodes.value.filter((n) => n.type === 'action')
      const idx = actionNodes.findIndex((n) => n.id === node.id)
      const pred = props.simulatorEvent.predictedActions?.[idx]
      if (pred !== undefined) {
        classes.push(pred.wouldSucceed ? 'vf-sim-hit' : 'vf-sim-warn')
      }
    }
  }

  return classes.join(' ')
}

const { onConnect, addEdges } = useVueFlow()

onConnect((params) => {
  // Add true/false labels for edges coming from branch nodes
  const sourceBranch = nodes.value.find((n) => n.id === params.source && n.type === 'branch')
  const label = sourceBranch
    ? params.sourceHandle === 'true' ? '✓ true' : params.sourceHandle === 'false' ? '✗ false' : undefined
    : undefined

  addEdges([{
    ...params,
    id: `e-${Date.now()}`,
    ...(label ? { label, labelStyle: { fontSize: 10, fill: label.startsWith('✓') ? '#16a34a' : '#ef4444' }, labelBgStyle: { fill: 'transparent' } } : {}),
  }])
  emitUpdate()
})

function addActionNode() {
  const id = `action-${Date.now()}`
  nodes.value = [
    ...nodes.value,
    {
      id,
      type: 'action',
      position: { x: 200, y: nodes.value.length * 160 + 50 },
      data: enrich({ actionType: 'set_state', params: { state_id: '' } }),
    },
  ]
  emitUpdate()
}

function addBranchNode() {
  const id = `branch-${Date.now()}`
  nodes.value = [
    ...nodes.value,
    {
      id,
      type: 'branch',
      position: { x: 200, y: nodes.value.length * 160 + 50 },
      data: enrich({ conditionField: 'priority', conditionOperator: 'equals', conditionValue: 'urgent' }),
    },
  ]
  emitUpdate()
}
</script>

<template>
  <div data-testid="automation.canvas" class="relative overflow-hidden rounded-xl border border-border" style="height: 520px;">
    <VueFlow
      v-model:nodes="nodes"
      v-model:edges="edges"
      :node-types="nodeTypes"
      :node-class="nodeClass"
      :default-edge-options="{ type: 'smoothstep', animated: true }"
      fit-view-on-init
      class="bg-muted/10"
      @nodes-change="emitUpdate"
      @edges-change="emitUpdate"
    >
      <Background :gap="20" :size="1" pattern-color="hsl(var(--border))" />
    </VueFlow>

    <!-- Toolbar -->
    <div class="absolute bottom-4 left-1/2 z-10 flex -translate-x-1/2 gap-2">
      <button
        data-testid="automation.add_action_node"
        class="rounded-lg border border-border bg-background/95 px-3 py-1.5 text-xs font-medium shadow-sm backdrop-blur transition-colors hover:border-primary/40"
        @click="addActionNode"
      >
        + Action
      </button>
      <button
        data-testid="automation.add_branch_node"
        class="rounded-lg border border-amber-500/40 bg-amber-500/10 px-3 py-1.5 text-xs font-medium text-amber-600 shadow-sm backdrop-blur transition-colors hover:border-amber-500/70"
        @click="addBranchNode"
      >
        ⑂ Branch
      </button>
    </div>
  </div>
</template>

<style>
/* Validation ring for invalid nodes */
.vf-invalid .vue-flow__node-default,
.vf-invalid > div {
  outline: 2px solid rgba(239, 68, 68, 0.6);
  outline-offset: 2px;
  border-radius: 12px;
}
/* Simulator hit (green) */
.vf-sim-hit > div {
  outline: 2px solid rgba(34, 197, 94, 0.7);
  outline-offset: 2px;
  border-radius: 12px;
}
/* Simulator miss (red) */
.vf-sim-miss > div {
  outline: 2px solid rgba(239, 68, 68, 0.7);
  outline-offset: 2px;
  border-radius: 12px;
}
/* Simulator warn (amber) */
.vf-sim-warn > div {
  outline: 2px solid rgba(245, 158, 11, 0.7);
  outline-offset: 2px;
  border-radius: 12px;
}
</style>
