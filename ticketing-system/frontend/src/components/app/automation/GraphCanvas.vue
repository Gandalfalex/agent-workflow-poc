<script setup lang="ts">
import { ref, watch } from 'vue'
import {
  VueFlow, useVueFlow,
  type Node, type Edge, type EdgeMouseEvent, type NodeChange, type Connection,
} from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { Controls } from '@vue-flow/controls'
import { MiniMap } from '@vue-flow/minimap'
import '@vue-flow/core/dist/style.css'
import '@vue-flow/controls/dist/style.css'
import '@vue-flow/minimap/dist/style.css'

const props = defineProps<{
  modelValue: { nodes: Node[]; edges: Edge[] } | null
  nodeTypes?: Record<string, any>
  readonly?: boolean
  showMinimap?: boolean
  nodeClass?: (node: Node) => string
  defaultEdgeOptions?: Record<string, any>
  /** Called when a new connection is made; return the Edge to add or null to cancel. */
  createEdge?: (params: Connection, nodes: Node[]) => Edge | null
}>()

const emit = defineEmits<{
  'update:modelValue': [{ nodes: Node[]; edges: Edge[] }]
}>()

const nodes = ref<Node[]>([])
const edges = ref<Edge[]>([])
const isNodeDragging = ref(false)

watch(
  () => props.modelValue,
  (val) => {
    if (isNodeDragging.value) return
    if (val && val.nodes.length > 0) {
      nodes.value = val.nodes
      edges.value = val.edges
    } else if (!val) {
      nodes.value = []
      edges.value = []
    }
  },
  { immediate: true },
)

function emitUpdate() {
  if (props.readonly) return
  emit('update:modelValue', { nodes: nodes.value, edges: edges.value })
}

const { onConnect, addEdges, removeEdges, fitView, onNodeDragStart, onNodeDragStop } = useVueFlow()

onNodeDragStart(() => { isNodeDragging.value = true })
onNodeDragStop(() => {
  isNodeDragging.value = false
  emitUpdate()
})

function onNodesChange(changes: NodeChange[]) {
  // Skip position-change events — VueFlow manages those internally during drag.
  // Only emit for structural changes (add, remove, select, dimensions).
  if (!isNodeDragging.value && changes.some(c => c.type !== 'position')) {
    emitUpdate()
  }
}

onConnect((params) => {
  if (props.readonly) return
  const edge = props.createEdge
    ? props.createEdge(params as Connection, nodes.value)
    : ({ ...params, id: `e-${Date.now()}` } as Edge)
  if (edge) addEdges([edge])
  // emitUpdate handled via @edges-change
})

function onEdgeDblClick({ edge }: EdgeMouseEvent) {
  if (props.readonly) return
  removeEdges([edge as Edge])
  // emitUpdate handled via @edges-change
}

defineExpose({ fitView })
</script>

<template>
  <VueFlow
    v-model:nodes="nodes"
    v-model:edges="edges"
    :node-types="nodeTypes"
    :node-class="nodeClass"
    :default-edge-options="defaultEdgeOptions ?? {
      type: 'smoothstep',
      animated: !readonly,
      style: { strokeWidth: 2 },
    }"
    :connection-line-style="{ strokeWidth: 2, stroke: 'hsl(var(--primary))' }"
    :snap-to-grid="!readonly"
    :snap-grid="[20, 20]"
    :delete-key-code="readonly ? null : 'Delete'"
    :nodes-draggable="!readonly"
    :nodes-connectable="!readonly"
    :elements-selectable="!readonly"
    :min-zoom="0.2"
    :max-zoom="4"
    fit-view-on-init
    class="bg-muted/10"
    @nodes-change="onNodesChange"
    @edges-change="emitUpdate"
    @edge-double-click="onEdgeDblClick"
  >
    <Background :gap="20" :size="1" pattern-color="hsl(var(--border))" />
    <Controls position="top-left" :show-interactive="false" />
    <MiniMap
      v-if="showMinimap"
      position="bottom-right"
      :height="90"
      :width="140"
      :node-stroke-width="3"
      pannable
      zoomable
    />
    <slot />
  </VueFlow>
</template>

<style>
/* ── Handles ── */
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
.vue-flow__handle[data-handleid="true"]  { background: #16a34a !important; box-shadow: 0 0 0 2px #16a34a66 !important; }
.vue-flow__handle[data-handleid="false"] { background: #ef4444 !important; box-shadow: 0 0 0 2px #ef444466 !important; }

/* ── Edges ── */
.vue-flow__edge-path { stroke-width: 2; }
.vue-flow__edge.selected .vue-flow__edge-path { stroke: hsl(var(--primary)) !important; stroke-width: 2.5; }
.vue-flow__edge:hover .vue-flow__edge-path { stroke-width: 3; }
.vue-flow__connectionline path { stroke: hsl(var(--primary)); stroke-width: 2; stroke-dasharray: 6 3; }

/* ── Validation / simulator outlines ── */
.vf-invalid > .vue-flow__node-default,
.vf-invalid > div { outline: 2px solid rgba(239,68,68,0.65); outline-offset: 2px; border-radius: 12px; }
.vf-sim-hit  > div { outline: 2px solid rgba(34,197,94,0.75);  outline-offset: 2px; border-radius: 12px; }
.vf-sim-miss > div { outline: 2px solid rgba(239,68,68,0.75);  outline-offset: 2px; border-radius: 12px; }
.vf-sim-warn > div { outline: 2px solid rgba(245,158,11,0.75); outline-offset: 2px; border-radius: 12px; }

/* ── Controls / MiniMap ── */
.vue-flow__controls  { box-shadow: 0 2px 8px rgba(0,0,0,0.12); border-radius: 8px; overflow: hidden; }
.vue-flow__minimap   { border-radius: 8px; overflow: hidden; opacity: 0.85; }
</style>
