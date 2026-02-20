<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import BoardView from "@/components/app/BoardView.vue";
import BoardToolbar from "@/components/app/board/BoardToolbar.vue";
import BoardBulkActionBar from "@/components/app/board/BoardBulkActionBar.vue";
import NewTicketModal from "@/components/app/NewTicketModal.vue";
import StoryModal from "@/components/app/StoryModal.vue";
import TicketModal from "@/components/app/TicketModal.vue";
import { useBoardStore } from "@/stores/board";
import { useSessionStore } from "@/stores/session";
import { useAdminStore } from "@/stores/admin";
import {
    createAiTriageSuggestion,
    getProjectAiTriageSettings,
    getTicketIncidentPostmortem,
    listTicketIncidentTimeline,
    recordAiTriageSuggestionDecision,
} from "@/lib/api";
import type { StoryRow } from "@/lib/types";
import type {
    AiTriageSuggestion,
    BulkTicketAction,
    BulkTicketOperationRequest,
    BoardFilter,
    BoardFilterPreset,
    DependencyRelationType,
    IncidentTimelineItem,
    TicketPriority,
    TicketIncidentSeverity,
    TicketResponse,
    TicketType,
} from "@/lib/api";

const props = defineProps<{ projectId: string }>();

const boardStore = useBoardStore();
const sessionStore = useSessionStore();
const adminStore = useAdminStore();
const route = useRoute();
const router = useRouter();

const showNewTicket = ref(false);
const showStoryModal = ref(false);
const boardSearch = ref("");
const filterStateId = ref("");
const filterAssigneeId = ref("");
const filterPriority = ref("");
const filterType = ref("");
const filterBlocked = ref(false);
const presetName = ref("");
const activePresetId = ref("");
const presetBusy = ref(false);
const presetMessage = ref("");
const bulkSelectMode = ref(false);
const selectedTicketIds = ref<string[]>([]);
const bulkAction = ref<BulkTicketAction>("move_state");
const bulkStateId = ref("");
const bulkAssigneeId = ref("");
const bulkPriority = ref<TicketPriority>("medium");
const bulkBusy = ref(false);
const bulkMessage = ref("");
const boardToolbarRef = ref<{ focusSearch: () => void } | null>(null);
const showFilterPanel = ref(true);
const showShortcutHelp = ref(false);
const showPresetEditor = ref(false);
const selectedTicket = ref<TicketResponse | null>(null);
const ticketEditor = ref({
    title: "",
    description: "",
    priority: "medium" as TicketPriority,
    stateId: "",
    type: "feature" as TicketType,
    incidentEnabled: false,
    incidentSeverity: undefined as TicketIncidentSeverity | undefined,
    incidentImpact: "",
    incidentCommanderId: "",
});
const incidentTimeline = ref<IncidentTimelineItem[]>([]);
const incidentTimelineLoading = ref(false);
const ticketSaving = ref(false);
const ticketError = ref("");
const dependencyRelationDraft = ref<DependencyRelationType>("blocks");
const dependencyTicketIdDraft = ref("");
const dependencySaving = ref(false);
const dependencyError = ref("");
const commentDraft = ref("");
const newTicket = ref({
    title: "",
    description: "",
    priority: "medium" as TicketPriority,
    type: "feature" as TicketType,
    storyId: "",
    assignee: "",
    stateId: "",
});
const aiTriageEnabled = ref(false);
const aiTriageLoading = ref(false);
const aiTriageBusy = ref(false);
const aiTriageError = ref("");
const aiTriageSuggestion = ref<AiTriageSuggestion | null>(null);
const aiFieldSelection = ref({
    summary: true,
    priority: true,
    state: true,
    assignee: false,
});
const newStory = ref({
    title: "",
    description: "",
});
const storySaving = ref(false);

const states = computed(() => boardStore.states);
const tickets = computed(() => boardStore.tickets);
const stories = computed(() => boardStore.stories);
const apiMode = computed(() => boardStore.apiMode);
const loading = computed(() => boardStore.loading);
const errorMessage = computed(() => boardStore.errorMessage);
const webhooks = computed(() => boardStore.webhooks);
const workflowSetupBusy = computed(() => boardStore.workflowSetupBusy);
const workflowSetupError = computed(() => boardStore.workflowSetupError);
const ticketComments = computed(() => boardStore.ticketComments);
const ticketActivities = computed(() => boardStore.ticketActivities);
const commentSaving = computed(() => boardStore.commentSaving);
const commentError = computed(() => boardStore.commentError);
const ticketAttachments = computed(() => boardStore.ticketAttachments);
const ticketDependencies = computed(() => boardStore.ticketDependencies);
const ticketDependenciesLoading = computed(
    () => boardStore.ticketDependenciesLoading,
);
const ticketDependencyGraph = computed(() => boardStore.ticketDependencyGraph);
const ticketDependencyGraphLoading = computed(
    () => boardStore.ticketDependencyGraphLoading,
);
const attachmentUploading = computed(() => boardStore.attachmentUploading);
const attachmentError = computed(() => boardStore.attachmentError);
const storiesCount = computed(() => boardStore.stories.length);
const groupMembers = computed(() => adminStore.groupMembers);
const canEditTickets = computed(() => boardStore.canEditTickets);
const boardFilterPresets = computed(() => boardStore.boardFilterPresets);
const boardFilterPresetsLoading = computed(
    () => boardStore.boardFilterPresetsLoading,
);
const boardFilterPresetsError = computed(() => boardStore.boardFilterPresetsError);

const persistedPresetKey = computed(
    () => `board.active-preset.${props.projectId || "unknown"}`,
);

const priorities: TicketPriority[] = ["low", "medium", "high", "urgent"];
const ticketTypes: TicketType[] = ["feature", "bug"];

const defaultWorkflow = [
    { name: "Backlog", order: 1, isDefault: true, isClosed: false },
    { name: "In Progress", order: 2, isDefault: false, isClosed: false },
    { name: "Review", order: 3, isDefault: false, isClosed: false },
    { name: "Done", order: 4, isDefault: false, isClosed: true },
];

const handleAuthError = (err: unknown) => {
    const error = err as Error & { status?: number };
    if (error.status === 401 || error.status === 403) {
        sessionStore.reset();
        return true;
    }
    return false;
};

const refreshBoard = async () => {
    if (!props.projectId) return;
    try {
        await boardStore.loadCurrentUserRole(props.projectId);
        await boardStore.loadBoard(props.projectId);
        await boardStore.loadStories(props.projectId);
        await boardStore.loadWebhooks(props.projectId);
        await boardStore.loadBoardFilterPresets(props.projectId);
    } catch (err) {
        handleAuthError(err);
    }
};

const assigneeOptions = computed(() => {
    const seen = new Map<string, string>();
    tickets.value.forEach((ticket) => {
        if (!ticket.assignee?.id || !ticket.assignee.name) return;
        if (!seen.has(ticket.assignee.id)) {
            seen.set(ticket.assignee.id, ticket.assignee.name);
        }
    });
    return Array.from(seen.entries()).map(([id, name]) => ({ id, name }));
});

const currentFilters = computed<BoardFilter>(() => ({
    q: boardSearch.value.trim() || undefined,
    stateId: filterStateId.value || undefined,
    assigneeId: filterAssigneeId.value || undefined,
    priority: (filterPriority.value as TicketPriority) || undefined,
    type: (filterType.value as TicketType) || undefined,
    blocked: filterBlocked.value || undefined,
}));

const hasActiveFilter = computed(
    () =>
        !!currentFilters.value.q ||
        !!currentFilters.value.stateId ||
        !!currentFilters.value.assigneeId ||
        !!currentFilters.value.priority ||
        !!currentFilters.value.type ||
        !!currentFilters.value.blocked,
);

const initializeWorkflow = async () => {
    if (!props.projectId) return;
    try {
        await boardStore.initializeWorkflow(props.projectId, defaultWorkflow);
    } catch (err) {
        handleAuthError(err);
    }
};

const ticketsByState = computed(() => {
    const map: Record<string, TicketResponse[]> = {};
    states.value.forEach((state) => {
        map[state.id] = [];
    });
    tickets.value.forEach((ticket) => {
        const bucket = map[ticket.stateId];
        if (bucket) {
            bucket.push(ticket);
        } else {
            map[ticket.stateId] = [ticket];
        }
    });
    Object.values(map).forEach((list) =>
        list.sort((a, b) => a.position - b.position),
    );
    return map;
});

const storyRows = computed<StoryRow[]>(() => {
    if (states.value.length === 0) {
        return [];
    }

    const createBuckets = () => {
        const buckets: Record<string, TicketResponse[]> = {};
        states.value.forEach((state) => {
            buckets[state.id] = [];
        });
        return buckets;
    };

    const rows = new Map<string, StoryRow>();
    stories.value.forEach((story) => {
        rows.set(story.id, {
            id: story.id,
            title: story.title,
            description: story.description || undefined,
            ticketsByState: createBuckets(),
        });
    });

    tickets.value.forEach((ticket) => {
        const row = rows.get(ticket.storyId);
        if (!row) return;
        const bucket = row.ticketsByState[ticket.stateId];
        if (bucket) {
            bucket.push(ticket);
        } else {
            row.ticketsByState[ticket.stateId] = [ticket];
        }
    });

    rows.forEach((row) => {
        Object.values(row.ticketsByState).forEach((bucket) =>
            bucket.sort((a, b) => a.position - b.position),
        );
    });

    return Array.from(rows.values());
});

const filteredStoryRows = computed<StoryRow[]>(() => {
    const query = boardSearch.value.trim().toLowerCase();
    const hasAssigneeFilter = !!filterAssigneeId.value;
    const hasStateFilter = !!filterStateId.value;
    const hasPriorityFilter = !!filterPriority.value;
    const hasTypeFilter = !!filterType.value;
    const hasBlockedFilter = !!filterBlocked.value;
    const hasStructuredFilter =
        hasAssigneeFilter ||
        hasStateFilter ||
        hasPriorityFilter ||
        hasTypeFilter ||
        hasBlockedFilter;
    if (
        !query &&
        !hasAssigneeFilter &&
        !hasStateFilter &&
        !hasPriorityFilter &&
        !hasTypeFilter &&
        !hasBlockedFilter
    ) {
        return storyRows.value;
    }

    const contains = (value?: string | null) =>
        (value || "").toLowerCase().includes(query);

    return storyRows.value
        .map((row) => {
            const storyMatches =
                contains(row.title) || contains(row.description);
            const nextBuckets: Record<string, TicketResponse[]> = {};
            let ticketMatches = 0;

            states.value.forEach((state) => {
                const filtered = (row.ticketsByState[state.id] || []).filter(
                    (ticket) => {
                        if (
                            hasStateFilter &&
                            ticket.stateId !== filterStateId.value
                        ) {
                            return false;
                        }
                        if (
                            hasAssigneeFilter &&
                            ticket.assignee?.id !== filterAssigneeId.value
                        ) {
                            return false;
                        }
                        if (
                            hasPriorityFilter &&
                            ticket.priority !== filterPriority.value
                        ) {
                            return false;
                        }
                        if (hasTypeFilter && ticket.type !== filterType.value) {
                            return false;
                        }
                        if (hasBlockedFilter && !ticket.isBlocked) {
                            return false;
                        }
                        if (!query) {
                            return true;
                        }
                        return (
                            contains(ticket.key) ||
                            contains(ticket.title) ||
                            contains(ticket.description) ||
                            contains(ticket.assignee?.name)
                        );
                    },
                );
                nextBuckets[state.id] = filtered;
                ticketMatches += filtered.length;
            });

            if (query && storyMatches && ticketMatches === 0 && !hasStructuredFilter) {
                states.value.forEach((state) => {
                    nextBuckets[state.id] = row.ticketsByState[state.id] || [];
                });
                ticketMatches = Object.values(nextBuckets).reduce(
                    (sum, bucket) => sum + bucket.length,
                    0,
                );
            }

            if (!storyMatches && ticketMatches === 0) {
                return null;
            }

            return {
                ...row,
                ticketsByState: nextBuckets,
            };
        })
        .filter((row): row is StoryRow => row !== null);
});
const visibleTicketCount = computed(() =>
    filteredStoryRows.value.reduce(
        (sum, row) =>
            sum +
            states.value.reduce(
                (stateSum, state) =>
                    stateSum + (row.ticketsByState[state.id]?.length || 0),
                0,
            ),
        0,
    ),
);

const canSubmit = computed(
    () =>
        newTicket.value.title.trim().length > 0 &&
        newTicket.value.storyId.length > 0,
);
const canCreateStory = computed(() => newStory.value.title.trim().length > 0);
const dependencyOptions = computed(() =>
    tickets.value
        .filter((ticket) => ticket.id !== selectedTicket.value?.id)
        .map((ticket) => ({
            id: ticket.id,
            key: ticket.key,
            title: ticket.title,
        })),
);

const openTicket = async (ticket: TicketResponse) => {
    selectedTicket.value = ticket;
    ticketEditor.value = {
        title: ticket.title,
        description: ticket.description || "",
        priority: ticket.priority,
        stateId: ticket.stateId,
        type: ticket.type,
        incidentEnabled: !!ticket.incidentEnabled,
        incidentSeverity: ticket.incidentSeverity || undefined,
        incidentImpact: ticket.incidentImpact || "",
        incidentCommanderId: ticket.incidentCommanderId || "",
    };
    ticketError.value = "";
    dependencyError.value = "";
    dependencyRelationDraft.value = "blocks";
    dependencyTicketIdDraft.value = "";
    commentDraft.value = "";
    boardStore.clearComments();
    incidentTimeline.value = [];
    incidentTimelineLoading.value = true;
    if (apiMode.value !== "demo") {
        try {
            if (stories.value.length === 0 && props.projectId) {
                await boardStore.loadStories(props.projectId);
            }
            await boardStore.loadTicketComments(ticket.id);
            await boardStore.loadTicketActivities(ticket.id);
            if (ticket.incidentEnabled) {
                const timeline = await listTicketIncidentTimeline(ticket.id);
                incidentTimeline.value = timeline.items;
            } else {
                incidentTimeline.value = [];
            }
            await boardStore.loadTicketAttachments(props.projectId, ticket.id);
            await boardStore.loadTicketDependencies(ticket.id);
            if (props.projectId) {
                await boardStore.loadDependencyGraph(props.projectId, {
                    rootTicketId: ticket.id,
                    depth: 2,
                });
            }
        } catch (err) {
            handleAuthError(err);
        }
    }
    incidentTimelineLoading.value = false;
};

const closeTicket = () => {
    selectedTicket.value = null;
    ticketError.value = "";
    dependencyError.value = "";
    dependencyRelationDraft.value = "blocks";
    dependencyTicketIdDraft.value = "";
    commentDraft.value = "";
    boardStore.clearComments();
    incidentTimeline.value = [];
    incidentTimelineLoading.value = false;
};

const addDependencySubmit = async () => {
    if (!selectedTicket.value || !dependencyTicketIdDraft.value) return;
    dependencySaving.value = true;
    dependencyError.value = "";
    try {
        await boardStore.createTicketDependency(selectedTicket.value.id, {
            relatedTicketId: dependencyTicketIdDraft.value,
            relationType: dependencyRelationDraft.value,
        });
        dependencyTicketIdDraft.value = "";
        await boardStore.loadTicketDependencies(selectedTicket.value.id);
        if (props.projectId) {
            await Promise.all([
                boardStore.loadDependencyGraph(props.projectId, {
                    rootTicketId: selectedTicket.value.id,
                    depth: 2,
                }),
                boardStore.loadBoard(props.projectId),
                boardStore.loadDashboardStats(props.projectId),
            ]);
        }
    } catch (err) {
        if (!handleAuthError(err)) {
            dependencyError.value = "Unable to add dependency.";
        }
    } finally {
        dependencySaving.value = false;
    }
};

const deleteDependencySubmit = async (dependencyId: string) => {
    if (!selectedTicket.value) return;
    dependencySaving.value = true;
    dependencyError.value = "";
    try {
        await boardStore.removeTicketDependency(selectedTicket.value.id, dependencyId);
        await boardStore.loadTicketDependencies(selectedTicket.value.id);
        if (props.projectId) {
            await Promise.all([
                boardStore.loadDependencyGraph(props.projectId, {
                    rootTicketId: selectedTicket.value.id,
                    depth: 2,
                }),
                boardStore.loadBoard(props.projectId),
                boardStore.loadDashboardStats(props.projectId),
            ]);
        }
    } catch (err) {
        if (!handleAuthError(err)) {
            dependencyError.value = "Unable to remove dependency.";
        }
    } finally {
        dependencySaving.value = false;
    }
};

const saveTicket = async () => {
    if (!selectedTicket.value || ticketSaving.value) return;
    ticketSaving.value = true;
    ticketError.value = "";

    const payload = {
        title: ticketEditor.value.title.trim(),
        description: ticketEditor.value.description.trim() || undefined,
        priority: ticketEditor.value.priority,
        stateId: ticketEditor.value.stateId,
        type: ticketEditor.value.type,
        incidentEnabled: ticketEditor.value.incidentEnabled,
        incidentSeverity: ticketEditor.value.incidentEnabled
            ? ticketEditor.value.incidentSeverity
            : undefined,
        incidentImpact: ticketEditor.value.incidentEnabled
            ? ticketEditor.value.incidentImpact.trim() || undefined
            : undefined,
        incidentCommanderId:
            ticketEditor.value.incidentEnabled &&
            ticketEditor.value.incidentCommanderId
                ? ticketEditor.value.incidentCommanderId
                : undefined,
    };

    try {
        await boardStore.updateTicket(
            selectedTicket.value.id,
            payload,
        );
        closeTicket();
    } catch (err) {
        if (!handleAuthError(err)) {
            ticketError.value = "Unable to update ticket.";
        }
    } finally {
        ticketSaving.value = false;
    }
};

const createStorySubmit = async () => {
    if (!canCreateStory.value || storySaving.value || !props.projectId) return;
    storySaving.value = true;

    const payload = {
        title: newStory.value.title.trim(),
        description: newStory.value.description.trim() || undefined,
    };

    try {
        const created = await boardStore.createStory(props.projectId, payload);
        if (showNewTicket.value) {
            newTicket.value.storyId = created.id;
        }
        newStory.value = { title: "", description: "" };
        showStoryModal.value = false;
    } catch (err) {
        handleAuthError(err);
    } finally {
        storySaving.value = false;
    }
};

const addCommentSubmit = async () => {
    if (
        !selectedTicket.value ||
        !commentDraft.value.trim() ||
        commentSaving.value
    ) {
        return;
    }
    const message = commentDraft.value.trim();
    try {
        const authorName = sessionStore.user?.name || "Demo User";
        await boardStore.addTicketComment(
            selectedTicket.value.id,
            message,
            authorName,
        );
        commentDraft.value = "";
    } catch (err) {
        handleAuthError(err);
    }
};

const uploadAttachmentHandler = async (file: File) => {
    if (!selectedTicket.value || !props.projectId) return;
    try {
        await boardStore.uploadAttachment(
            props.projectId,
            selectedTicket.value.id,
            file,
        );
    } catch (err) {
        handleAuthError(err);
    }
};

const deleteAttachmentHandler = async (attachmentId: string) => {
    if (!selectedTicket.value || !props.projectId) return;
    try {
        await boardStore.removeAttachment(
            props.projectId,
            selectedTicket.value.id,
            attachmentId,
        );
    } catch (err) {
        handleAuthError(err);
    }
};

const openNewTicket = async (stateId?: string, storyId?: string) => {
    const fallbackState = states.value[0]?.id || "";
    newTicket.value = {
        title: "",
        description: "",
        priority: "medium" as TicketPriority,
        type: "feature" as TicketType,
        storyId: storyId || "",
        assignee: "",
        stateId: stateId || fallbackState,
    };
    aiTriageSuggestion.value = null;
    aiTriageError.value = "";
    aiFieldSelection.value = {
        summary: true,
        priority: true,
        state: true,
        assignee: false,
    };
    showNewTicket.value = true;
    if (apiMode.value !== "demo" && props.projectId) {
        await boardStore.loadStories(props.projectId);
        // Load all groups and their members for assignee selection
        try {
            await adminStore.loadGroups();
            // Load members for each group and collect them
            if (adminStore.groups.length > 0) {
                adminStore.clearGroupMembers();
                const allMembers: Record<string, any> = {};
                for (const group of adminStore.groups) {
                    await adminStore.loadGroupMembers(group.id);
                    adminStore.groupMembers.forEach((member) => {
                        allMembers[member.userId] = member;
                    });
                }
                // Update groupMembers with all unique members by reloading all
                if (Object.keys(allMembers).length > 0) {
                    // Force a reactive update by creating a new array
                    adminStore.$patch({
                        groupMembers: Object.values(allMembers),
                    });
                }
            }
        } catch (err) {
            // Silently fail group loading, it's not critical
        }
        aiTriageLoading.value = true;
        try {
            const settings = await getProjectAiTriageSettings(props.projectId);
            aiTriageEnabled.value = settings.enabled;
        } catch {
            aiTriageEnabled.value = false;
        } finally {
            aiTriageLoading.value = false;
        }
    }
};

const applyAiSuggestionToDraft = () => {
    const suggestion = aiTriageSuggestion.value;
    if (!suggestion) return;
    const patch: Partial<typeof newTicket.value> = {};
    if (aiFieldSelection.value.summary && suggestion.summary?.trim()) {
        patch.description = suggestion.summary;
    }
    if (aiFieldSelection.value.priority) {
        patch.priority = suggestion.priority;
    }
    if (aiFieldSelection.value.state) {
        patch.stateId = suggestion.stateId;
    }
    if (aiFieldSelection.value.assignee && suggestion.assigneeId) {
        patch.assignee = suggestion.assigneeId;
    }
    newTicket.value = { ...newTicket.value, ...patch };
};

const requestAiTriageSuggestion = async () => {
    if (!props.projectId || aiTriageBusy.value || !aiTriageEnabled.value) return;
    aiTriageBusy.value = true;
    aiTriageError.value = "";
    try {
        const suggestion = await createAiTriageSuggestion(props.projectId, {
            title: newTicket.value.title.trim(),
            description: newTicket.value.description.trim() || undefined,
            type: newTicket.value.type,
        });
        aiTriageSuggestion.value = suggestion;
        applyAiSuggestionToDraft();
    } catch (err) {
        if (!handleAuthError(err)) {
            aiTriageError.value = "Unable to generate AI triage suggestion.";
        }
    } finally {
        aiTriageBusy.value = false;
    }
};

const toggleAiField = (
    field: "summary" | "priority" | "state" | "assignee",
    value: boolean,
) => {
    aiFieldSelection.value = { ...aiFieldSelection.value, [field]: value };
    applyAiSuggestionToDraft();
};

const clearBoardFilters = () => {
    boardSearch.value = "";
    filterStateId.value = "";
    filterAssigneeId.value = "";
    filterPriority.value = "";
    filterType.value = "";
    filterBlocked.value = false;
    activePresetId.value = "";
    presetName.value = "";
};

const toggleFilterPanel = () => {
    showFilterPanel.value = !showFilterPanel.value;
};

const openPresetEditor = () => {
    showPresetEditor.value = true;
    if (!presetName.value.trim() && activePresetId.value) {
        const active = boardFilterPresets.value.find(
            (item) => item.id === activePresetId.value,
        );
        if (active) {
            presetName.value = active.name;
        }
    }
};

const cancelPresetEditor = () => {
    showPresetEditor.value = false;
    presetName.value = "";
};

const savePresetFromEditor = async () => {
    await savePreset();
    if (activePresetId.value) {
        showPresetEditor.value = false;
    }
};

const toggleBulkSelectMode = () => {
    bulkSelectMode.value = !bulkSelectMode.value;
    if (!bulkSelectMode.value) {
        selectedTicketIds.value = [];
        bulkMessage.value = "";
    }
};

const toggleTicketSelection = (ticketId: string) => {
    if (selectedTicketIds.value.includes(ticketId)) {
        selectedTicketIds.value = selectedTicketIds.value.filter((id) => id !== ticketId);
        return;
    }
    selectedTicketIds.value = [...selectedTicketIds.value, ticketId];
};

const clearTicketSelection = () => {
    selectedTicketIds.value = [];
};

const applyBulkOptimistic = (payload: BulkTicketOperationRequest) => {
    if (payload.action === "delete") {
        boardStore.tickets = boardStore.tickets.filter(
            (ticket) => !payload.ticketIds.includes(ticket.id),
        );
        return;
    }
    if (payload.action === "move_state" && payload.stateId) {
        boardStore.tickets = boardStore.tickets.map((ticket) =>
            payload.ticketIds.includes(ticket.id)
                ? { ...ticket, stateId: payload.stateId as string }
                : ticket,
        );
        return;
    }
    if (payload.action === "assign" && payload.assigneeId) {
        const assigneeName = assigneeOptions.value.find(
            (item) => item.id === payload.assigneeId,
        )?.name;
        boardStore.tickets = boardStore.tickets.map((ticket) =>
            payload.ticketIds.includes(ticket.id)
                ? {
                      ...ticket,
                      assigneeId: payload.assigneeId,
                      assignee: assigneeName
                          ? { id: payload.assigneeId as string, name: assigneeName }
                          : ticket.assignee,
                  }
                : ticket,
        );
        return;
    }
    if (payload.action === "set_priority" && payload.priority) {
        boardStore.tickets = boardStore.tickets.map((ticket) =>
            payload.ticketIds.includes(ticket.id)
                ? { ...ticket, priority: payload.priority as TicketPriority }
                : ticket,
        );
    }
};

const applyBulkAction = async () => {
    if (!props.projectId || bulkBusy.value) return;
    if (!bulkSelectMode.value || selectedTicketIds.value.length === 0) {
        bulkMessage.value = "Select at least one ticket.";
        return;
    }

    const payload: BulkTicketOperationRequest = {
        action: bulkAction.value,
        ticketIds: [...selectedTicketIds.value],
    };
    if (bulkAction.value === "move_state") {
        if (!bulkStateId.value) {
            bulkMessage.value = "Choose a target state.";
            return;
        }
        payload.stateId = bulkStateId.value;
    } else if (bulkAction.value === "assign") {
        if (!bulkAssigneeId.value) {
            bulkMessage.value = "Choose an assignee.";
            return;
        }
        payload.assigneeId = bulkAssigneeId.value;
    } else if (bulkAction.value === "set_priority") {
        payload.priority = bulkPriority.value;
    }

    const snapshot = boardStore.tickets.map((ticket) => ({ ...ticket })) as TicketResponse[];
    bulkBusy.value = true;
    bulkMessage.value = "";
    applyBulkOptimistic(payload);

    try {
        const result = await boardStore.bulkTicketOperation(props.projectId, payload);
        if (result.errorCount === 0) {
            bulkMessage.value = `Updated ${result.successCount} ticket(s).`;
            selectedTicketIds.value = [];
            return;
        }

        const reconciled = [...snapshot];
        for (const item of result.results) {
            if (!item.success) continue;
            if (result.action === "delete") {
                const idx = reconciled.findIndex((ticket) => ticket.id === item.ticketId);
                if (idx >= 0) reconciled.splice(idx, 1);
                continue;
            }
            if (!item.ticket) continue;
            const idx = reconciled.findIndex((ticket) => ticket.id === item.ticketId);
            if (idx >= 0) {
                reconciled[idx] = item.ticket;
            }
        }
        boardStore.tickets = reconciled;
        const failed = result.results
            .filter((item) => !item.success)
            .map((item) => item.ticketId);
        selectedTicketIds.value = failed;
        bulkMessage.value = `Updated ${result.successCount}/${result.total}. ${result.errorCount} failed.`;
    } catch (err) {
        boardStore.tickets = snapshot;
        if (!handleAuthError(err)) {
            bulkMessage.value = "Bulk operation failed.";
        }
    } finally {
        bulkBusy.value = false;
    }
};

const applyPreset = (preset: BoardFilterPreset) => {
    presetName.value = preset.name;
    activePresetId.value = preset.id;
    boardSearch.value = preset.filters.q || "";
    filterStateId.value = preset.filters.stateId || "";
    filterAssigneeId.value = preset.filters.assigneeId || "";
    filterPriority.value = preset.filters.priority || "";
    filterType.value = preset.filters.type || "";
    filterBlocked.value = !!preset.filters.blocked;
    localStorage.setItem(persistedPresetKey.value, preset.id);
};

const onPresetChange = () => {
    const preset = boardFilterPresets.value.find(
        (item) => item.id === activePresetId.value,
    );
    if (!preset) return;
    applyPreset(preset);
};

const savePreset = async () => {
    if (!props.projectId || presetBusy.value) return;
    const name = presetName.value.trim();
    if (!name) {
        presetMessage.value = "Enter a preset name.";
        return;
    }
    presetBusy.value = true;
    presetMessage.value = "";
    try {
        const created = await boardStore.createBoardFilterPreset(props.projectId, {
            name,
            filters: currentFilters.value,
        });
        applyPreset(created);
        presetMessage.value = "Preset saved.";
    } catch (err) {
        if (!handleAuthError(err)) {
            presetMessage.value = "Unable to save preset.";
        }
    } finally {
        presetBusy.value = false;
    }
};

const renamePreset = async () => {
    if (!props.projectId || !activePresetId.value || presetBusy.value) return;
    const name = presetName.value.trim();
    if (!name) {
        presetMessage.value = "Enter a preset name.";
        return;
    }
    presetBusy.value = true;
    presetMessage.value = "";
    try {
        const updated = await boardStore.updateBoardFilterPreset(
            props.projectId,
            activePresetId.value,
            { name, filters: currentFilters.value },
        );
        applyPreset(updated);
        presetMessage.value = "Preset renamed.";
    } catch (err) {
        if (!handleAuthError(err)) {
            presetMessage.value = "Unable to rename preset.";
        }
    } finally {
        presetBusy.value = false;
    }
};

const deletePreset = async () => {
    if (!props.projectId || !activePresetId.value || presetBusy.value) return;
    presetBusy.value = true;
    presetMessage.value = "";
    try {
        await boardStore.deleteBoardFilterPreset(props.projectId, activePresetId.value);
        clearBoardFilters();
        localStorage.removeItem(persistedPresetKey.value);
        presetMessage.value = "Preset deleted.";
    } catch (err) {
        if (!handleAuthError(err)) {
            presetMessage.value = "Unable to delete preset.";
        }
    } finally {
        presetBusy.value = false;
    }
};

const sharePreset = async () => {
    if (!props.projectId || !activePresetId.value || presetBusy.value) return;
    presetBusy.value = true;
    presetMessage.value = "";
    try {
        const updated = await boardStore.updateBoardFilterPreset(
            props.projectId,
            activePresetId.value,
            { generateShareToken: true },
        );
        applyPreset(updated);
        if (updated.shareToken) {
            const origin = window.location.origin;
            const url = `${origin}/projects/${props.projectId}/board?share=${encodeURIComponent(updated.shareToken)}`;
            let copied = false;
            try {
                await navigator.clipboard.writeText(url);
                copied = true;
            } catch {
                copied = false;
            }
            await router.replace({
                path: `/projects/${props.projectId}/board`,
                query: { share: updated.shareToken },
            });
            presetMessage.value = copied
                ? "Share link copied."
                : "Share link generated.";
        }
    } catch (err) {
        if (!handleAuthError(err)) {
            presetMessage.value = "Unable to generate share link.";
        }
    } finally {
        presetBusy.value = false;
    }
};

const shouldHandleGlobalShortcut = (event: KeyboardEvent) => {
    const target = event.target as HTMLElement | null;
    if (!target) return true;
    const tag = target.tagName.toLowerCase();
    if (tag === "input" || tag === "textarea" || tag === "select") {
        return false;
    }
    if (target.isContentEditable) {
        return false;
    }
    return true;
};

const onGlobalKeydown = async (event: KeyboardEvent) => {
    if (!shouldHandleGlobalShortcut(event)) {
        return;
    }
    if (event.key === "/") {
        event.preventDefault();
        await nextTick();
        boardToolbarRef.value?.focusSearch();
        return;
    }
    if (event.key.toLowerCase() === "n") {
        if (
            !canEditTickets.value ||
            showNewTicket.value ||
            showStoryModal.value ||
            selectedTicket.value
        ) {
            return;
        }
        event.preventDefault();
        await openNewTicket();
    }
};

const closeNewTicket = () => {
    showNewTicket.value = false;
};

const openStoryModal = async () => {
    showStoryModal.value = true;
    if (apiMode.value !== "demo" && props.projectId) {
        await boardStore.loadStories(props.projectId);
    }
};

const deleteStorySubmit = async (storyId: string) => {
    if (!storyId || storyId === "ungrouped") return;
    const story = stories.value.find((item) => item.id === storyId);
    const ticketCount = tickets.value.filter(
        (ticket) => ticket.storyId === storyId,
    ).length;
    const label = story?.title ? `"${story.title}"` : "this story";
    if (!window.confirm(`Delete ${label} and ${ticketCount} ticket(s)?`)) {
        return;
    }
    try {
        await boardStore.removeStory(storyId);
    } catch (err) {
        handleAuthError(err);
    }
};

const createTicketSubmit = async () => {
    if (!canSubmit.value || !props.projectId) return;
    try {
        if (aiTriageSuggestion.value) {
            const accepted: Array<"summary" | "priority" | "state" | "assignee"> = [];
            const rejected: Array<"summary" | "priority" | "state" | "assignee"> = [];
            (["summary", "priority", "state", "assignee"] as const).forEach((field) => {
                if (aiFieldSelection.value[field]) {
                    accepted.push(field);
                } else {
                    rejected.push(field);
                }
            });
            await recordAiTriageSuggestionDecision(
                props.projectId,
                aiTriageSuggestion.value.id,
                {
                    acceptedFields: accepted,
                    rejectedFields: rejected,
                },
            );
        }
        await boardStore.createTicket(props.projectId, {
            title: newTicket.value.title.trim(),
            description: newTicket.value.description.trim(),
            type: newTicket.value.type,
            storyId: newTicket.value.storyId,
            stateId: newTicket.value.stateId,
            priority: newTicket.value.priority,
        });
        closeNewTicket();
    } catch (err) {
        handleAuthError(err);
    }
};

const deleteTicketSubmit = async () => {
    if (!selectedTicket.value) return;
    const label = selectedTicket.value.key || "this ticket";
    if (!window.confirm(`Delete ${label}?`)) {
        return;
    }
    try {
        await boardStore.removeTicket(selectedTicket.value.id);
        closeTicket();
    } catch (err) {
        if (!handleAuthError(err)) {
            ticketError.value = "Unable to delete ticket.";
        }
    }
};

const draggingId = ref<string | null>(null);

const onDragStart = (ticketId: string) => {
    draggingId.value = ticketId;
};

const onDragEnd = () => {
    draggingId.value = null;
};

const moveToState = async (
    ticketId: string,
    stateId: string,
    storyId: string,
    position: number,
) => {
    try {
        await boardStore.updateTicket(ticketId, {
            stateId,
            storyId,
            position,
        });
    } catch (err) {
        handleAuthError(err);
    }
};

const onDropColumn = async (stateId: string, storyId: string) => {
    if (!draggingId.value) return;
    const column = ticketsByState.value[stateId] || [];
    const position = column.length + 1;
    await moveToState(draggingId.value, stateId, storyId, position);
    draggingId.value = null;
};

const onDropCard = async (
    targetId: string,
    stateId: string,
    storyId: string,
) => {
    if (!draggingId.value || draggingId.value === targetId) return;
    const moving = tickets.value.find(
        (ticket) => ticket.id === draggingId.value,
    );
    if (!moving) return;
    const targetStateTickets = ticketsByState.value[stateId] || [];
    const nextPosition = targetStateTickets.length + 1;
    await moveToState(moving.id, stateId, storyId, nextPosition);
    draggingId.value = null;
};

const updateNewTicket = (value: typeof newTicket.value) => {
    newTicket.value = value;
};

const updateTicketEditor = (value: typeof ticketEditor.value) => {
    ticketEditor.value = value;
};

const updateCommentDraft = (value: string) => {
    commentDraft.value = value;
};

const exportIncidentPostmortem = async () => {
    if (!selectedTicket.value) return;
    const markdown = await getTicketIncidentPostmortem(selectedTicket.value.id);
    const blob = new Blob([markdown], { type: "text/markdown;charset=utf-8" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = `${selectedTicket.value.key}-postmortem.md`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
};

const openDependencyTicketFromGraph = async (ticketId: string) => {
    if (!ticketId) return;
    const graphTicket = ticketDependencyGraph.value.nodes.find(
        (node) => node.ticket.id === ticketId,
    )?.ticket;
    if (graphTicket) {
        await openTicket(graphTicket);
        return;
    }
    const ticket = tickets.value.find((item) => item.id === ticketId);
    if (ticket) {
        await openTicket(ticket);
    }
};

const updateNewStory = (value: typeof newStory.value) => {
    newStory.value = value;
};

onMounted(() => {
    window.addEventListener("keydown", onGlobalKeydown);
    refreshBoard().then(async () => {
        if (!props.projectId) return;
        const shareToken =
            typeof route.query.share === "string" ? route.query.share : "";
        if (shareToken) {
            try {
                const shared = await boardStore.resolveSharedBoardFilterPreset(
                    props.projectId,
                    shareToken,
                );
                applyPreset(shared);
                presetMessage.value = `Applied shared preset "${shared.name}".`;
                return;
            } catch (err) {
                if (!handleAuthError(err)) {
                    presetMessage.value = "Shared preset not found.";
                }
            }
        }
        const savedId = localStorage.getItem(persistedPresetKey.value);
        if (!savedId) return;
        const preset = boardFilterPresets.value.find((item) => item.id === savedId);
        if (preset) {
            applyPreset(preset);
        } else {
            localStorage.removeItem(persistedPresetKey.value);
        }
    });
});

onUnmounted(() => {
    window.removeEventListener("keydown", onGlobalKeydown);
});

watch(
    () => props.projectId,
    async (value) => {
        if (!value) return;
        await refreshBoard();
    },
);
</script>

<template>
    <BoardToolbar
        ref="boardToolbarRef"
        :board-search="boardSearch"
        :active-preset-id="activePresetId"
        :has-active-filter="hasActiveFilter"
        :show-filter-panel="showFilterPanel"
        :show-shortcut-help="showShortcutHelp"
        :show-preset-editor="showPresetEditor"
        :filter-state-id="filterStateId"
        :filter-assignee-id="filterAssigneeId"
        :filter-priority="filterPriority"
        :filter-type="filterType"
        :filter-blocked="filterBlocked"
        :preset-name="presetName"
        :preset-busy="presetBusy"
        :preset-message="presetMessage"
        :board-filter-presets-loading="boardFilterPresetsLoading"
        :board-filter-presets-error="boardFilterPresetsError"
        :states="states"
        :assignee-options="assigneeOptions"
        :priorities="priorities"
        :ticket-types="ticketTypes"
        :board-filter-presets="boardFilterPresets"
        @update:board-search="boardSearch = $event"
        @update:active-preset-id="activePresetId = $event"
        @toggle-filter-panel="toggleFilterPanel"
        @toggle-shortcut-help="showShortcutHelp = !showShortcutHelp"
        @clear-filters="clearBoardFilters"
        @preset-change="onPresetChange"
        @update:filter-state-id="filterStateId = $event"
        @update:filter-assignee-id="filterAssigneeId = $event"
        @update:filter-priority="filterPriority = $event"
        @update:filter-type="filterType = $event"
        @update:filter-blocked="filterBlocked = $event"
        @open-preset-editor="openPresetEditor"
        @update:preset-name="presetName = $event"
        @save-preset-from-editor="savePresetFromEditor"
        @cancel-preset-editor="cancelPresetEditor"
        @rename-preset="renamePreset"
        @delete-preset="deletePreset"
        @share-preset="sharePreset"
    />

    <section
        v-if="errorMessage"
        class="rounded-2xl border border-border bg-secondary/60 px-4 py-3 text-sm"
    >
        {{ errorMessage }}
    </section>

    <BoardView
        :loading="loading"
        :states="states"
        :story-rows="filteredStoryRows"
        :stories-count="storiesCount"
        :tickets-count="visibleTicketCount"
        :webhooks-count="webhooks.length"
        :api-mode="apiMode"
        :workflow-setup-busy="workflowSetupBusy"
        :workflow-setup-error="workflowSetupError"
        :can-edit-tickets="canEditTickets"
        :bulk-select-mode="bulkSelectMode"
        :selected-ticket-ids="selectedTicketIds"
        :has-active-filter="hasActiveFilter"
        :search-query="boardSearch"
        :on-initialize-workflow="initializeWorkflow"
        :on-open-story-modal="openStoryModal"
        :on-toggle-bulk-select-mode="toggleBulkSelectMode"
        :on-toggle-ticket-selection="toggleTicketSelection"
        :on-clear-ticket-selection="clearTicketSelection"
        :on-delete-story="deleteStorySubmit"
        :on-open-ticket="openTicket"
        :on-open-new-ticket="openNewTicket"
        :on-clear-filter="clearBoardFilters"
        :on-drag-start="onDragStart"
        :on-drag-end="onDragEnd"
        :on-drop-column="onDropColumn"
        :on-drop-card="onDropCard"
    />
    <BoardBulkActionBar
        v-if="canEditTickets && bulkSelectMode"
        :selected-count="selectedTicketIds.length"
        :bulk-action="bulkAction"
        :bulk-state-id="bulkStateId"
        :bulk-assignee-id="bulkAssigneeId"
        :bulk-priority="bulkPriority"
        :bulk-busy="bulkBusy"
        :bulk-message="bulkMessage"
        :states="states"
        :priorities="priorities"
        :assignee-options="assigneeOptions"
        @update:bulk-action="bulkAction = $event"
        @update:bulk-state-id="bulkStateId = $event"
        @update:bulk-assignee-id="bulkAssigneeId = $event"
        @update:bulk-priority="bulkPriority = $event"
        @apply="applyBulkAction"
    />

    <NewTicketModal
        v-if="canEditTickets"
        :show="showNewTicket"
        :ticket="newTicket"
        :states="states"
        :stories="stories"
        :priorities="priorities"
        :ticket-types="ticketTypes"
        :can-submit="canSubmit"
        :group-members="groupMembers"
        :ai-triage-enabled="aiTriageEnabled"
        :ai-triage-loading="aiTriageLoading"
        :ai-triage-busy="aiTriageBusy"
        :ai-triage-error="aiTriageError"
        :ai-triage-suggestion="aiTriageSuggestion"
        :ai-field-selection="aiFieldSelection"
        @update:ticket="updateNewTicket"
        @close="closeNewTicket"
        @create="createTicketSubmit"
        @request-ai-triage="requestAiTriageSuggestion"
        @toggle-ai-field="toggleAiField"
    />

    <StoryModal
        v-if="canEditTickets"
        :show="showStoryModal"
        :story="newStory"
        :can-create="canCreateStory"
        :story-saving="storySaving"
        :story-error="boardStore.storyError"
        @update:story="updateNewStory"
        @close="showStoryModal = false"
        @create="createStorySubmit"
    />

    <TicketModal
        :show="!!selectedTicket"
        :ticket-key="selectedTicket?.key || ''"
        :editor="ticketEditor"
        :states="states"
        :priorities="priorities"
        :ticket-types="ticketTypes"
        :ticket-saving="ticketSaving"
        :ticket-error="ticketError"
        :activities="ticketActivities"
        :comments="ticketComments"
        :comment-draft="commentDraft"
        :comment-saving="commentSaving"
        :comment-error="commentError"
        :current-user-id="sessionStore.user?.id"
        :attachments="ticketAttachments"
        :attachment-uploading="attachmentUploading"
        :attachment-error="attachmentError"
        :dependencies="ticketDependencies"
        :dependencies-loading="ticketDependenciesLoading"
        :dependency-graph="ticketDependencyGraph"
        :dependency-graph-loading="ticketDependencyGraphLoading"
        :dependency-options="dependencyOptions"
        :dependency-relation-draft="dependencyRelationDraft"
        :dependency-ticket-id-draft="dependencyTicketIdDraft"
        :dependency-saving="dependencySaving"
        :dependency-error="dependencyError"
        :incident-timeline="incidentTimeline"
        :incident-timeline-loading="incidentTimelineLoading"
        :assignee-options="assigneeOptions"
        :project-id="props.projectId"
        :ticket-id="selectedTicket?.id || ''"
        :read-only="!canEditTickets"
        @update:editor="updateTicketEditor"
        @update:commentDraft="updateCommentDraft"
        @update:dependencyRelationDraft="(value) => (dependencyRelationDraft = value)"
        @update:dependencyTicketIdDraft="(value) => (dependencyTicketIdDraft = value)"
        @close="closeTicket"
        @save="saveTicket"
        @delete="deleteTicketSubmit"
        @add-comment="addCommentSubmit"
        @add-dependency="addDependencySubmit"
        @delete-dependency="deleteDependencySubmit"
        @export-postmortem="exportIncidentPostmortem"
        @open-dependency-ticket="openDependencyTicketFromGraph"
        @upload-attachment="uploadAttachmentHandler"
        @delete-attachment="deleteAttachmentHandler"
    />
</template>
