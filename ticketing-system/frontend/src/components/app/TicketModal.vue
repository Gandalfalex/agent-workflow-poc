<script setup lang="ts">
import { Button } from "@/components/ui/button";
import { MarkdownEditor } from "@/components/ui/markdown-editor";
import { marked } from "marked";
import { computed, ref, watch } from "vue";
import type {
    Attachment,
    DependencyRelationType,
    IncidentTimelineItem,
    TicketActivity,
    TicketComment,
    TicketDependency,
    TicketIncidentSeverity,
    TicketDependencyGraphResponse,
    TicketPriority,
    TicketType,
    WorkflowState,
} from "@/lib/api";
import { downloadTicketAttachmentUrl } from "@/lib/api";

type TicketEditor = {
    title: string;
    description: string;
    priority: TicketPriority;
    stateId: string;
    type: TicketType;
    incidentEnabled: boolean;
    incidentSeverity: TicketIncidentSeverity | undefined;
    incidentImpact: string;
    incidentCommanderId: string;
};

const props = defineProps<{
    show: boolean;
    ticketKey: string;
    editor: TicketEditor;
    states: WorkflowState[];
    priorities: TicketPriority[];
    ticketTypes: TicketType[];
    ticketSaving: boolean;
    ticketError: string;
    activities: TicketActivity[];
    comments: TicketComment[];
    commentDraft: string;
    commentSaving: boolean;
    commentError: string;
    currentUserId?: string;
    attachments: Attachment[];
    attachmentUploading: boolean;
    attachmentError: string;
    dependencies: TicketDependency[];
    dependenciesLoading: boolean;
    dependencyGraph: TicketDependencyGraphResponse;
    dependencyGraphLoading: boolean;
    dependencyOptions: Array<{ id: string; key: string; title: string }>;
    dependencyRelationDraft: DependencyRelationType;
    dependencyTicketIdDraft: string;
    dependencySaving: boolean;
    dependencyError: string;
    incidentTimeline: IncidentTimelineItem[];
    incidentTimelineLoading: boolean;
    assigneeOptions: Array<{ id: string; name: string }>;
    projectId: string;
    ticketId: string;
    readOnly?: boolean;
}>();

const emit = defineEmits<{
    (e: "update:editor", value: TicketEditor): void;
    (e: "update:commentDraft", value: string): void;
    (e: "close"): void;
    (e: "save"): void;
    (e: "delete"): void;
    (e: "add-comment"): void;
    (e: "upload-attachment", file: File): void;
    (e: "delete-attachment", attachmentId: string): void;
    (e: "update:dependencyRelationDraft", value: DependencyRelationType): void;
    (e: "update:dependencyTicketIdDraft", value: string): void;
    (e: "add-dependency"): void;
    (e: "delete-dependency", dependencyId: string): void;
    (e: "export-postmortem"): void;
    (e: "open-dependency-ticket", ticketId: string): void;
}>();

const fileInput = ref<HTMLInputElement | null>(null);
const incidentExpanded = ref(false);
const showDependencyGraphOverlay = ref(false);

const triggerFileUpload = () => {
    fileInput.value?.click();
};

const onFileSelected = (event: Event) => {
    const input = event.target as HTMLInputElement;
    const file = input.files?.[0];
    if (file) {
        emit("upload-attachment", file);
        input.value = "";
    }
};

const formatFileSize = (bytes: number): string => {
    if (bytes < 1024) return bytes + " B";
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
    return (bytes / (1024 * 1024)).toFixed(1) + " MB";
};

const downloadUrl = (att: Attachment): string => {
    return downloadTicketAttachmentUrl(props.projectId, props.ticketId, att.id);
};

const menuOpen = ref(false);
const incidentTimelineVisibleItems = computed(() => {
    if (incidentExpanded.value) return props.incidentTimeline;
    return props.incidentTimeline.slice(-6);
});
const canExpandIncidentTimeline = computed(
    () => props.incidentTimeline.length > 6,
);

watch(
    () => props.show,
    (show) => {
        if (show) {
            incidentExpanded.value = false;
        } else {
            showDependencyGraphOverlay.value = false;
        }
    },
);

const updateEditor = (patch: Partial<TicketEditor>) => {
    emit("update:editor", { ...props.editor, ...patch });
};

const isCurrentUser = (userId: string) => {
    return props.currentUserId === userId;
};

const priorityColor = (priority: string) => {
    switch (priority) {
        case "urgent":
            return "bg-red-500/20 text-red-300 border-red-500/30";
        case "high":
            return "bg-orange-500/20 text-orange-300 border-orange-500/30";
        case "medium":
            return "bg-amber-500/20 text-amber-300 border-amber-500/30";
        case "low":
        default:
            return "bg-slate-500/15 text-slate-400 border-slate-500/20";
    }
};

type GraphNodeLayout = {
    id: string;
    key: string;
    title: string;
    depth: number;
    x: number;
    y: number;
    width: number;
    height: number;
    isCurrent: boolean;
};

type GraphEdgeLayout = {
    id: string;
    path: string;
    relationType: string;
    color: string;
};

const dependencyGraphLayout = computed(() => {
    const nodes = props.dependencyGraph.nodes || [];
    const edges = props.dependencyGraph.edges || [];
    if (!nodes.length) {
        return {
            width: 640,
            height: 180,
            nodes: [] as GraphNodeLayout[],
            edges: [] as GraphEdgeLayout[],
        };
    }

    const grouped = new Map<number, typeof nodes>();
    for (const node of nodes) {
        const list = grouped.get(node.depth) || [];
        list.push(node);
        grouped.set(node.depth, list);
    }
    const depths = Array.from(grouped.keys()).sort((a, b) => a - b);
    for (const depth of depths) {
        const list = grouped.get(depth) || [];
        list.sort((a, b) => a.ticket.key.localeCompare(b.ticket.key));
        grouped.set(depth, list);
    }

    const cardWidth = 150;
    const cardHeight = 38;
    const colGap = 26;
    const rowGap = 14;
    const marginX = 18;
    const marginY = 16;
    const maxRows = Math.max(...depths.map((d) => (grouped.get(d) || []).length));
    const width = Math.max(520, marginX * 2 + depths.length * cardWidth + (depths.length - 1) * colGap);
    const height = Math.max(150, marginY * 2 + maxRows * cardHeight + (maxRows - 1) * rowGap);

    const nodeLayouts: GraphNodeLayout[] = [];
    const byID = new Map<string, GraphNodeLayout>();
    depths.forEach((depth, colIndex) => {
        const list = grouped.get(depth) || [];
        list.forEach((node, rowIndex) => {
            const layout: GraphNodeLayout = {
                id: node.ticket.id,
                key: node.ticket.key,
                title: node.ticket.title,
                depth: node.depth,
                x: marginX + colIndex * (cardWidth + colGap),
                y: marginY + rowIndex * (cardHeight + rowGap),
                width: cardWidth,
                height: cardHeight,
                isCurrent: node.ticket.id === props.ticketId,
            };
            nodeLayouts.push(layout);
            byID.set(layout.id, layout);
        });
    });

    const colorForRelation = (relationType: string) => {
        switch (relationType) {
            case "blocks":
                return "#f97316";
            case "blocked_by":
                return "#38bdf8";
            default:
                return "#94a3b8";
        }
    };

    const edgeLayouts: GraphEdgeLayout[] = [];
    for (const edge of edges) {
        const source = byID.get(edge.sourceTicketId);
        const target = byID.get(edge.targetTicketId);
        if (!source || !target) continue;
        const sx = source.x + source.width;
        const sy = source.y + source.height / 2;
        const tx = target.x;
        const ty = target.y + target.height / 2;
        const cx = sx + (tx - sx) / 2;
        edgeLayouts.push({
            id: edge.id,
            relationType: edge.relationType,
            color: colorForRelation(edge.relationType),
            path: `M ${sx} ${sy} C ${cx} ${sy}, ${cx} ${ty}, ${tx} ${ty}`,
        });
    }

    return {
        width,
        height,
        nodes: nodeLayouts,
        edges: edgeLayouts,
    };
});

const shortTitle = (title: string) => {
    const trimmed = title.trim();
    if (trimmed.length <= 28) return trimmed;
    return trimmed.slice(0, 27) + "...";
};

const dependencyNodeTestId = (nodeID: string) =>
    `ticket.dependency-graph-node-${nodeID}`;

const openDependencyTicket = (ticketId: string) => {
    showDependencyGraphOverlay.value = false;
    emit("open-dependency-ticket", ticketId);
};
</script>

<template>
    <div
        v-if="props.show"
        data-testid="ticket.modal"
        class="fixed inset-0 z-30 flex items-center justify-center bg-black/50 px-6"
        @click.self="emit('close')"
    >
        <div
            class="flex w-full max-h-[92vh] max-w-5xl flex-col rounded-3xl border border-border bg-card shadow-xl"
        >
            <!-- Header: fixed at top -->
            <div class="flex items-center justify-between px-6 pt-5 pb-4 border-b border-border flex-shrink-0">
                <div class="flex items-center gap-4">
                    <div>
                        <p
                            class="text-[10px] uppercase tracking-[0.3em] text-muted-foreground"
                        >
                            Ticket
                        </p>
                        <h2 class="text-xl font-semibold">
                            {{ props.ticketKey }}
                        </h2>
                    </div>
                    <span
                        class="rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase border"
                        :class="priorityColor(props.editor.priority)"
                    >
                        {{ props.editor.priority }}
                    </span>
                </div>
                <div class="flex items-center gap-2">
                    <Button variant="ghost" size="sm" @click="emit('close')">
                        Close
                    </Button>
                    <div v-if="!props.readOnly" class="relative">
                        <button
                            class="rounded-full border border-border bg-background px-2 py-1 text-lg font-semibold text-muted-foreground transition hover:border-foreground hover:text-foreground cursor-pointer"
                            aria-label="Ticket actions"
                            @click.stop="menuOpen = !menuOpen"
                        >
                            &#x22EE;
                        </button>
                        <div
                            v-if="menuOpen"
                            class="dropdown-backdrop"
                            @click="menuOpen = false"
                        ></div>
                        <div
                            v-if="menuOpen"
                            class="absolute right-0 top-full mt-2 w-40 rounded-2xl border border-border bg-card/95 backdrop-blur p-2 text-xs z-50 shadow-lg"
                        >
                            <Button
                                data-testid="ticket.delete-button"
                                variant="outline"
                                size="sm"
                                class="w-full border-destructive/30 text-destructive hover:bg-destructive/5"
                                :disabled="props.ticketSaving"
                                @click.stop="menuOpen = false; emit('delete')"
                            >
                                Delete ticket
                            </Button>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Body: scrollable two-column layout -->
            <div
                class="flex-1 min-h-0 grid gap-0 lg:grid-cols-[1fr_1fr]"
            >
                <!-- Left: Form fields (independently scrollable) -->
                <div class="overflow-y-auto p-6 border-r border-border">
                    <div class="grid gap-4">
                        <div>
                            <label
                                class="text-xs font-semibold text-muted-foreground"
                                >Title</label
                            >
                            <input
                                data-testid="ticket.title-input"
                                :value="props.editor.title"
                                type="text"
                                class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                                :disabled="props.readOnly"
                                @input="
                                    updateEditor({
                                        title: ($event.target as HTMLInputElement)
                                            .value,
                                    })
                                "
                            />
                        </div>
                        <div>
                            <label
                                class="text-xs font-semibold text-muted-foreground"
                                >Description</label
                            >
                            <MarkdownEditor
                                :model-value="props.editor.description"
                                @update:model-value="updateEditor({ description: $event })"
                                :rows="7"
                                placeholder="Describe the ticket..."
                                data-testid="ticket.description-input"
                                show-preview
                            />
                        </div>
                        <div class="grid gap-4 sm:grid-cols-3">
                            <div>
                                <label
                                    class="text-xs font-semibold text-muted-foreground"
                                    >Type</label
                                >
                                <select
                                    data-testid="ticket.type-select"
                                    :value="props.editor.type"
                                    class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                                    :disabled="props.readOnly"
                                    @change="
                                        updateEditor({
                                            type: (
                                                $event.target as HTMLSelectElement
                                            ).value as TicketType,
                                        })
                                    "
                                >
                                    <option
                                        v-for="type in props.ticketTypes"
                                        :key="type"
                                        :value="type"
                                    >
                                        {{ type }}
                                    </option>
                                </select>
                            </div>
                            <div>
                                <label
                                    class="text-xs font-semibold text-muted-foreground"
                                    >Priority</label
                                >
                                <select
                                    data-testid="ticket.priority-select"
                                    :value="props.editor.priority"
                                    class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                                    :disabled="props.readOnly"
                                    @change="
                                        updateEditor({
                                            priority: (
                                                $event.target as HTMLSelectElement
                                            ).value as TicketPriority,
                                        })
                                    "
                                >
                                    <option
                                        v-for="priority in props.priorities"
                                        :key="priority"
                                        :value="priority"
                                    >
                                        {{ priority }}
                                    </option>
                                </select>
                            </div>
                            <div>
                                <label
                                    class="text-xs font-semibold text-muted-foreground"
                                    >State</label
                                >
                                <select
                                    data-testid="ticket.state-select"
                                    :value="props.editor.stateId"
                                    class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                                    :disabled="props.readOnly"
                                    @change="
                                        updateEditor({
                                            stateId: (
                                                $event.target as HTMLSelectElement
                                            ).value,
                                        })
                                    "
                                >
                                    <option
                                        v-for="state in props.states"
                                        :key="state.id"
                                        :value="state.id"
                                    >
                                        {{ state.name }}
                                    </option>
                                </select>
                            </div>
                        </div>
                        <div class="rounded-xl border border-border bg-background/40 p-3" data-testid="ticket.incident-section">
                            <div class="flex items-center justify-between">
                                <label class="text-xs font-semibold text-muted-foreground">Incident mode</label>
                                <input
                                    data-testid="ticket.incident-enabled-checkbox"
                                    type="checkbox"
                                    :checked="props.editor.incidentEnabled"
                                    :disabled="props.readOnly"
                                    @change="
                                        updateEditor({
                                            incidentEnabled: ($event.target as HTMLInputElement).checked,
                                        })
                                    "
                                />
                            </div>
                            <div v-if="props.editor.incidentEnabled" class="mt-3 grid gap-3 sm:grid-cols-3">
                                <div>
                                    <label class="text-xs font-semibold text-muted-foreground">Severity</label>
                                    <select
                                        data-testid="ticket.incident-severity-select"
                                        :value="props.editor.incidentSeverity || ''"
                                        class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                                        :disabled="props.readOnly"
                                        @change="
                                            updateEditor({
                                                incidentSeverity: (($event.target as HTMLSelectElement).value || undefined) as TicketIncidentSeverity | undefined,
                                            })
                                        "
                                    >
                                        <option value="">Select</option>
                                        <option value="sev1">sev1</option>
                                        <option value="sev2">sev2</option>
                                        <option value="sev3">sev3</option>
                                        <option value="sev4">sev4</option>
                                    </select>
                                </div>
                                <div>
                                    <label class="text-xs font-semibold text-muted-foreground">Commander</label>
                                    <select
                                        data-testid="ticket.incident-commander-select"
                                        :value="props.editor.incidentCommanderId"
                                        class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                                        :disabled="props.readOnly"
                                        @change="
                                            updateEditor({
                                                incidentCommanderId: ($event.target as HTMLSelectElement).value,
                                            })
                                        "
                                    >
                                        <option value="">Unassigned</option>
                                        <option
                                            v-for="assignee in props.assigneeOptions"
                                            :key="assignee.id"
                                            :value="assignee.id"
                                        >
                                            {{ assignee.name }}
                                        </option>
                                    </select>
                                </div>
                                <div class="sm:col-span-3">
                                    <label class="text-xs font-semibold text-muted-foreground">Impact</label>
                                    <textarea
                                        data-testid="ticket.incident-impact-input"
                                        :value="props.editor.incidentImpact"
                                        rows="2"
                                        class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                                        :disabled="props.readOnly"
                                        @input="
                                            updateEditor({
                                                incidentImpact: ($event.target as HTMLTextAreaElement).value,
                                            })
                                        "
                                    />
                                </div>
                            </div>
                        </div>
                        <div
                            v-if="props.ticketError"
                            class="rounded-2xl border border-border bg-secondary/60 px-3 py-2 text-xs"
                        >
                            {{ props.ticketError }}
                        </div>

                        <!-- Attachments -->
                        <div data-testid="ticket.attachments-section">
                            <div class="flex items-center justify-between">
                                <label class="text-xs font-semibold text-muted-foreground">Attachments</label>
                                <button
                                    v-if="!props.readOnly"
                                    data-testid="ticket.upload-attachment-button"
                                    type="button"
                                    class="text-xs text-primary hover:text-primary/80 transition font-semibold"
                                    :disabled="props.attachmentUploading"
                                    @click="triggerFileUpload"
                                >
                                    {{ props.attachmentUploading ? "Uploading..." : "+ Upload" }}
                                </button>
                            </div>
                            <input
                                ref="fileInput"
                                data-testid="ticket.file-input"
                                type="file"
                                class="hidden"
                                @change="onFileSelected"
                            />
                            <div v-if="props.attachments.length" class="mt-2 space-y-1.5">
                                <div
                                    v-for="att in props.attachments"
                                    :key="att.id"
                                    data-testid="ticket.attachment-item"
                                    class="flex items-center justify-between rounded-xl border border-border bg-background px-3 py-2"
                                >
                                    <div class="flex items-center gap-2 min-w-0">
                                        <a
                                            data-testid="ticket.attachment-download-link"
                                            :href="downloadUrl(att)"
                                            target="_blank"
                                            class="text-xs text-primary hover:underline truncate"
                                        >{{ att.filename }}</a>
                                        <span class="text-[10px] text-muted-foreground whitespace-nowrap">{{ formatFileSize(att.size) }}</span>
                                    </div>
                                    <button
                                        v-if="!props.readOnly"
                                        data-testid="ticket.attachment-delete-button"
                                        type="button"
                                        class="text-[10px] text-destructive hover:text-destructive/80 transition ml-2 whitespace-nowrap"
                                        @click="emit('delete-attachment', att.id)"
                                    >
                                        Delete
                                    </button>
                                </div>
                            </div>
                            <p v-else class="mt-2 text-[10px] text-muted-foreground">No files attached.</p>
                            <p v-if="props.attachmentError" class="mt-1 text-xs text-destructive">{{ props.attachmentError }}</p>
                        </div>

                        <div data-testid="ticket.dependencies-section" class="mt-3">
                            <div class="flex items-center justify-between">
                                <label class="text-xs font-semibold text-muted-foreground">Dependencies</label>
                                <span
                                    v-if="props.dependenciesLoading"
                                    class="text-[10px] text-muted-foreground"
                                >
                                    Loading...
                                </span>
                            </div>
                            <div
                                v-if="!props.readOnly"
                                class="mt-2 grid grid-cols-1 gap-2 sm:grid-cols-[140px_minmax(0,1fr)_auto]"
                            >
                                <select
                                    data-testid="ticket.dependency-relation-select"
                                    :value="props.dependencyRelationDraft"
                                    class="rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                                    @change="
                                        emit(
                                            'update:dependencyRelationDraft',
                                            ($event.target as HTMLSelectElement).value as DependencyRelationType,
                                        )
                                    "
                                >
                                    <option value="blocks">blocks</option>
                                    <option value="blocked_by">blocked_by</option>
                                    <option value="related">related</option>
                                </select>
                                <select
                                    data-testid="ticket.dependency-ticket-select"
                                    :value="props.dependencyTicketIdDraft"
                                    class="rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                                    @change="
                                        emit(
                                            'update:dependencyTicketIdDraft',
                                            ($event.target as HTMLSelectElement).value,
                                        )
                                    "
                                >
                                    <option value="">Select ticket</option>
                                    <option
                                        v-for="opt in props.dependencyOptions"
                                        :key="opt.id"
                                        :value="opt.id"
                                    >
                                        {{ opt.key }} - {{ opt.title }}
                                    </option>
                                </select>
                                <Button
                                    data-testid="ticket.dependency-add-button"
                                    variant="outline"
                                    size="sm"
                                    :disabled="props.dependencySaving || !props.dependencyTicketIdDraft"
                                    @click="emit('add-dependency')"
                                >
                                    Add
                                </Button>
                            </div>
                            <p
                                v-if="props.dependencyError"
                                class="mt-1 text-xs text-destructive"
                            >
                                {{ props.dependencyError }}
                            </p>
                            <div v-if="props.dependencies.length" class="mt-2 space-y-1.5">
                                <div
                                    v-for="dep in props.dependencies"
                                    :key="dep.id"
                                    data-testid="ticket.dependency-item"
                                    class="flex items-center justify-between rounded-xl border border-border bg-background px-3 py-2 text-xs"
                                >
                                    <div class="min-w-0">
                                        <span class="font-semibold text-foreground">{{ dep.relationType }}</span>
                                        <span class="ml-2 text-muted-foreground">
                                            {{
                                                dep.relatedTicket?.key
                                                    ? dep.relatedTicket.key + " - " + dep.relatedTicket.title
                                                    : dep.relatedTicketId
                                            }}
                                        </span>
                                    </div>
                                    <button
                                        v-if="!props.readOnly"
                                        data-testid="ticket.dependency-delete-button"
                                        type="button"
                                        class="text-[10px] text-destructive hover:text-destructive/80 transition ml-2 whitespace-nowrap"
                                        @click="emit('delete-dependency', dep.id)"
                                    >
                                        Remove
                                    </button>
                                </div>
                            </div>
                            <p
                                v-else
                                class="mt-2 text-[10px] text-muted-foreground"
                            >
                                No dependencies.
                            </p>
                            <div data-testid="ticket.dependency-graph" class="mt-3 rounded-xl border border-border bg-background px-3 py-2">
                                <p class="text-[10px] font-semibold uppercase tracking-[0.15em] text-muted-foreground">
                                    Graph (2-hop)
                                </p>
                                <p
                                    v-if="props.dependencyGraphLoading"
                                    class="mt-1 text-[10px] text-muted-foreground"
                                >
                                    Loading graph...
                                </p>
                                <template v-else>
                                    <p class="mt-1 text-[10px] text-muted-foreground">
                                        {{ props.dependencyGraph.nodes.length }} nodes · {{ props.dependencyGraph.edges.length }} edges
                                    </p>
                                    <Button
                                        data-testid="ticket.dependency-graph-open-button"
                                        variant="outline"
                                        size="sm"
                                        class="mt-2"
                                        :disabled="!dependencyGraphLayout.nodes.length"
                                        @click="showDependencyGraphOverlay = true"
                                    >
                                        Open graph
                                    </Button>
                                </template>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Right: Activity + Comments (independently scrollable) -->
                <div class="flex flex-col min-h-0 overflow-hidden">
                    <!-- Activity Timeline -->
                    <div
                        v-if="props.activities.length"
                        class="flex-shrink-0 border-b border-border"
                    >
                        <div class="px-6 py-3 flex items-center gap-2">
                            <span class="text-sm font-semibold text-foreground">Activity</span>
                        </div>
                        <div
                            data-testid="ticket.activity-timeline"
                            class="px-6 pb-3 space-y-1 max-h-40 overflow-y-auto"
                        >
                            <div
                                v-for="activity in props.activities"
                                :key="activity.id"
                                data-testid="ticket.activity-item"
                                class="flex items-start gap-2 text-xs text-muted-foreground"
                            >
                                <span class="mt-0.5 h-1.5 w-1.5 rounded-full bg-muted-foreground flex-shrink-0"></span>
                                <span>
                                    <span class="font-medium text-foreground">{{ activity.actorName }}</span>
                                    <span v-if="activity.action === 'state_changed'">
                                        changed state from
                                        <span class="font-medium text-foreground">{{ activity.oldValue }}</span>
                                        to
                                        <span class="font-medium text-foreground">{{ activity.newValue }}</span>
                                    </span>
                                    <span v-else-if="activity.action === 'priority_changed'">
                                        changed priority from
                                        <span class="font-medium text-foreground">{{ activity.oldValue }}</span>
                                        to
                                        <span class="font-medium text-foreground">{{ activity.newValue }}</span>
                                    </span>
                                    <span v-else-if="activity.action === 'assignee_changed'">
                                        <span v-if="activity.newValue">assigned to <span class="font-medium text-foreground">{{ activity.newValue }}</span></span>
                                        <span v-else>removed assignee</span>
                                    </span>
                                    <span v-else-if="activity.action === 'type_changed'">
                                        changed type from
                                        <span class="font-medium text-foreground">{{ activity.oldValue }}</span>
                                        to
                                        <span class="font-medium text-foreground">{{ activity.newValue }}</span>
                                    </span>
                                    <span v-else-if="activity.action === 'title_changed'">
                                        renamed ticket
                                    </span>
                                    <span v-else-if="activity.action === 'incident_severity_changed'">
                                        changed incident severity from
                                        <span class="font-medium text-foreground">{{ activity.oldValue || "-" }}</span>
                                        to
                                        <span class="font-medium text-foreground">{{ activity.newValue || "-" }}</span>
                                    </span>
                                    <span v-else>{{ activity.action }}</span>
                                    <span class="ml-1 text-muted-foreground/60">· {{ new Date(activity.createdAt).toLocaleString() }}</span>
                                </span>
                            </div>
                        </div>
                    </div>
                    <div
                        v-if="props.editor.incidentEnabled"
                        class="flex-shrink-0 border-b border-border"
                    >
                        <div class="px-6 py-3 flex items-center justify-between gap-2">
                            <div class="flex items-center gap-2">
                                <span class="text-sm font-semibold text-foreground">Incident timeline</span>
                                <span class="text-[10px] text-muted-foreground">
                                    {{ props.incidentTimeline.length }} events
                                </span>
                            </div>
                            <div class="flex items-center gap-2">
                                <Button
                                    v-if="canExpandIncidentTimeline"
                                    data-testid="ticket.incident-toggle-button"
                                    variant="ghost"
                                    size="sm"
                                    @click="incidentExpanded = !incidentExpanded"
                                >
                                    {{ incidentExpanded ? "Collapse" : "Expand" }}
                                </Button>
                                <Button
                                    data-testid="ticket.export-postmortem-button"
                                    variant="outline"
                                    size="sm"
                                    @click="emit('export-postmortem')"
                                >
                                    Export postmortem
                                </Button>
                            </div>
                        </div>
                        <div
                            data-testid="ticket.incident-timeline"
                            class="px-6 pb-3 space-y-1 max-h-28 overflow-y-auto"
                        >
                            <p
                                v-if="props.incidentTimelineLoading"
                                class="text-xs text-muted-foreground"
                            >
                                Loading timeline...
                            </p>
                            <div
                                v-for="item in incidentTimelineVisibleItems"
                                :key="item.id"
                                data-testid="ticket.incident-item"
                                class="text-xs text-muted-foreground break-words"
                            >
                                <span class="font-medium text-foreground">{{ item.title }}</span>
                                <span v-if="item.body"> · {{ item.body }}</span>
                                <span class="ml-1 text-muted-foreground/60">· {{ new Date(item.createdAt).toLocaleString() }}</span>
                            </div>
                            <p
                                v-if="!props.incidentTimelineLoading && props.incidentTimeline.length === 0"
                                class="text-xs text-muted-foreground"
                            >
                                No incident timeline events yet.
                            </p>
                        </div>
                    </div>

                    <div
                        class="flex items-center justify-between px-6 py-3 flex-shrink-0 border-b border-border"
                    >
                        <span class="text-sm font-semibold text-foreground"
                            >Comments</span
                        >
                        <span
                            v-if="props.commentSaving"
                            class="text-xs text-muted-foreground"
                            >Saving...</span
                        >
                    </div>

                    <div
                        v-if="props.comments.length"
                        class="flex-1 space-y-2 overflow-y-auto px-6 py-3 min-h-0"
                    >
                        <div
                            v-for="comment in props.comments"
                            :key="comment.id"
                            :class="[
                                'rounded-xl px-4 py-3 max-w-[85%]',
                                isCurrentUser(comment.authorId)
                                    ? 'ml-auto bg-primary/10 border border-primary/30'
                                    : 'bg-card border border-border',
                            ]"
                        >
                            <div class="flex items-center justify-between mb-2">
                                <span class="text-xs font-semibold">{{
                                    comment.authorName
                                }}</span>
                                <span
                                    class="text-[10px] text-muted-foreground ml-2"
                                >
                                    {{
                                        new Date(
                                            comment.createdAt,
                                        ).toLocaleString()
                                    }}
                                </span>
                            </div>
                            <div
                                class="text-xs text-foreground prose prose-sm dark:prose-invert max-w-none"
                                v-html="marked(comment.message)"
                            ></div>
                        </div>
                    </div>
                    <div
                        v-else
                        class="flex-1 flex items-center justify-center text-xs text-muted-foreground min-h-0"
                    >
                        No comments yet.
                    </div>

                    <div v-if="!props.readOnly" class="border-t border-border px-6 py-3 flex-shrink-0">
                        <label
                            class="text-[10px] font-semibold text-muted-foreground block mb-2"
                            >Add comment (Markdown)</label
                        >
                        <MarkdownEditor
                            :model-value="props.commentDraft"
                            @update:model-value="emit('update:commentDraft', $event)"
                            :rows="2"
                            placeholder="Progress, blockers, notes..."
                            data-testid="ticket.comment-input"
                            compact
                            :show-preview="false"
                        />
                        <div class="mt-2 flex items-center gap-3">
                            <Button
                                data-testid="ticket.post-comment-button"
                                size="sm"
                                :disabled="
                                    props.commentSaving ||
                                    !props.commentDraft.trim().length
                                "
                                @click="emit('add-comment')"
                            >
                                {{
                                    props.commentSaving
                                        ? "Posting..."
                                        : "Post"
                                }}
                            </Button>
                            <span
                                v-if="props.commentError"
                                class="text-xs text-destructive"
                                >{{ props.commentError }}</span
                            >
                        </div>
                    </div>
                </div>
            </div>

            <!-- Footer: fixed at bottom -->
            <div class="flex items-center justify-end gap-2 px-6 py-4 border-t border-border flex-shrink-0">
                <Button variant="ghost" size="sm" @click="emit('close')">
                    Cancel
                </Button>
                <Button
                    v-if="!props.readOnly"
                    data-testid="ticket.save-button"
                    size="sm"
                    :disabled="props.ticketSaving"
                    @click="emit('save')"
                >
                    {{ props.ticketSaving ? "Saving..." : "Save changes" }}
                </Button>
            </div>
        </div>
    </div>
    <div
        v-if="showDependencyGraphOverlay"
        data-testid="ticket.dependency-graph-overlay"
        class="fixed inset-0 z-40 flex items-center justify-center bg-black/70 p-6"
        @click.self="showDependencyGraphOverlay = false"
    >
        <div class="flex h-[74vh] w-full max-w-5xl flex-col rounded-2xl border border-border bg-card shadow-2xl">
            <div class="flex items-center justify-between border-b border-border px-5 py-3">
                <div>
                    <p class="text-xs uppercase tracking-[0.2em] text-muted-foreground">Dependency graph</p>
                    <p class="text-sm text-foreground">{{ props.ticketKey }} · 2-hop view</p>
                </div>
                <Button
                    data-testid="ticket.dependency-graph-close-button"
                    variant="ghost"
                    size="sm"
                    @click="showDependencyGraphOverlay = false"
                >
                    Close
                </Button>
            </div>
            <div class="flex-1 overflow-auto p-4">
                <div v-if="dependencyGraphLayout.nodes.length" class="min-w-fit">
                    <svg
                        class="h-[56vh]"
                        :viewBox="`0 0 ${dependencyGraphLayout.width} ${dependencyGraphLayout.height}`"
                        role="img"
                        aria-label="Ticket dependency graph"
                    >
                        <path
                            v-for="edge in dependencyGraphLayout.edges"
                            :key="edge.id"
                            :d="edge.path"
                            :stroke="edge.color"
                            stroke-width="2"
                            fill="none"
                            stroke-linecap="round"
                            opacity="0.9"
                        />
                        <g
                            v-for="node in dependencyGraphLayout.nodes"
                            :key="node.id"
                            :data-testid="dependencyNodeTestId(node.id)"
                            class="cursor-pointer"
                            @click="openDependencyTicket(node.id)"
                        >
                            <rect
                                :x="node.x"
                                :y="node.y"
                                :width="node.width"
                                :height="node.height"
                                rx="8"
                                :fill="node.isCurrent ? '#1d4ed8' : '#0f172a'"
                                :stroke="node.isCurrent ? '#93c5fd' : '#334155'"
                                stroke-width="1.5"
                            />
                            <text
                                :x="node.x + 10"
                                :y="node.y + 17"
                                font-size="11"
                                font-weight="700"
                                :fill="node.isCurrent ? '#eff6ff' : '#e2e8f0'"
                            >
                                {{ node.key }}
                            </text>
                            <text
                                :x="node.x + 10"
                                :y="node.y + 33"
                                font-size="10"
                                :fill="node.isCurrent ? '#dbeafe' : '#94a3b8'"
                            >
                                {{ shortTitle(node.title) }}
                            </text>
                        </g>
                    </svg>
                </div>
                <p v-else class="text-sm text-muted-foreground">No dependency graph data yet.</p>
            </div>
            <div class="border-t border-border px-5 py-3 text-[11px] text-muted-foreground">
                <span class="inline-flex items-center gap-1 mr-3">
                    <span class="h-2 w-2 rounded-full bg-orange-400"></span>
                    blocks
                </span>
                <span class="inline-flex items-center gap-1 mr-3">
                    <span class="h-2 w-2 rounded-full bg-sky-400"></span>
                    blocked_by
                </span>
                <span class="inline-flex items-center gap-1">
                    <span class="h-2 w-2 rounded-full bg-slate-400"></span>
                    related
                </span>
            </div>
        </div>
    </div>
</template>
