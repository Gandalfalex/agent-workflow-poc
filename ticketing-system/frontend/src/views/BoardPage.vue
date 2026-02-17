<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from "vue";
import BoardView from "@/components/app/BoardView.vue";
import NewTicketModal from "@/components/app/NewTicketModal.vue";
import StoryModal from "@/components/app/StoryModal.vue";
import TicketModal from "@/components/app/TicketModal.vue";
import { useBoardStore } from "@/stores/board";
import { useSessionStore } from "@/stores/session";
import { useAdminStore } from "@/stores/admin";
import type { StoryRow } from "@/lib/types";
import type { TicketResponse, TicketPriority, TicketType } from "@/lib/api";

const props = defineProps<{ projectId: string }>();

const boardStore = useBoardStore();
const sessionStore = useSessionStore();
const adminStore = useAdminStore();

const showNewTicket = ref(false);
const showStoryModal = ref(false);
const boardSearch = ref("");
const boardSearchInput = ref<HTMLInputElement | null>(null);
const selectedTicket = ref<TicketResponse | null>(null);
const ticketEditor = ref({
    title: "",
    description: "",
    priority: "medium" as TicketPriority,
    stateId: "",
    type: "feature" as TicketType,
});
const ticketSaving = ref(false);
const ticketError = ref("");
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
const commentSaving = computed(() => boardStore.commentSaving);
const commentError = computed(() => boardStore.commentError);
const ticketAttachments = computed(() => boardStore.ticketAttachments);
const attachmentUploading = computed(() => boardStore.attachmentUploading);
const attachmentError = computed(() => boardStore.attachmentError);
const storiesCount = computed(() => boardStore.stories.length);
const groupMembers = computed(() => adminStore.groupMembers);
const canEditTickets = computed(() => boardStore.canEditTickets);

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
        await boardStore.loadBoard(props.projectId);
        await boardStore.loadStories(props.projectId);
        await boardStore.loadWebhooks(props.projectId);
    } catch (err) {
        handleAuthError(err);
    }
};

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

const hasActiveSearch = computed(() => boardSearch.value.trim().length > 0);
const filteredStoryRows = computed<StoryRow[]>(() => {
    const query = boardSearch.value.trim().toLowerCase();
    if (!query) {
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
                    (ticket) =>
                        contains(ticket.key) ||
                        contains(ticket.title) ||
                        contains(ticket.description) ||
                        contains(ticket.assignee?.name),
                );
                nextBuckets[state.id] = filtered;
                ticketMatches += filtered.length;
            });

            if (storyMatches && ticketMatches === 0) {
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

const openTicket = async (ticket: TicketResponse) => {
    selectedTicket.value = ticket;
    ticketEditor.value = {
        title: ticket.title,
        description: ticket.description || "",
        priority: ticket.priority,
        stateId: ticket.stateId,
        type: ticket.type,
    };
    ticketError.value = "";
    commentDraft.value = "";
    boardStore.clearComments();
    if (apiMode.value !== "demo") {
        if (stories.value.length === 0 && props.projectId) {
            await boardStore.loadStories(props.projectId);
        }
        await boardStore.loadTicketComments(ticket.id);
        await boardStore.loadTicketAttachments(props.projectId, ticket.id);
    }
};

const closeTicket = () => {
    selectedTicket.value = null;
    ticketError.value = "";
    commentDraft.value = "";
    boardStore.clearComments();
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
    };

    try {
        const updated = await boardStore.updateTicket(
            selectedTicket.value.id,
            payload,
        );
        selectedTicket.value = updated;
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
    }
};

const clearBoardSearch = () => {
    boardSearch.value = "";
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
        boardSearchInput.value?.focus();
        boardSearchInput.value?.select();
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

const updateNewStory = (value: typeof newStory.value) => {
    newStory.value = value;
};

onMounted(() => {
    window.addEventListener("keydown", onGlobalKeydown);
    refreshBoard();
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
    <section class="rounded-2xl border border-border bg-card/70 px-4 py-2.5">
        <div class="flex items-center gap-3">
            <div class="relative flex-1">
                <input
                    ref="boardSearchInput"
                    v-model="boardSearch"
                    type="text"
                    placeholder="Filter tickets..."
                    class="w-full rounded-xl border border-input bg-background pl-3 pr-16 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                />
                <div class="absolute right-3 top-1/2 -translate-y-1/2 flex items-center gap-1.5 pointer-events-none text-muted-foreground">
                    <kbd class="rounded bg-muted px-1.5 py-0.5 text-[9px] font-semibold">/</kbd>
                    <kbd class="rounded bg-muted px-1.5 py-0.5 text-[9px] font-semibold">N</kbd>
                </div>
            </div>
            <button
                v-if="hasActiveSearch"
                class="rounded-lg border border-border bg-background px-2.5 py-1.5 text-[10px] font-semibold uppercase tracking-[0.15em] text-muted-foreground transition hover:border-foreground hover:text-foreground whitespace-nowrap"
                @click="clearBoardSearch"
            >
                Clear
            </button>
        </div>
    </section>

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
        :has-active-filter="hasActiveSearch"
        :search-query="boardSearch"
        :on-initialize-workflow="initializeWorkflow"
        :on-open-story-modal="openStoryModal"
        :on-delete-story="deleteStorySubmit"
        :on-open-ticket="openTicket"
        :on-open-new-ticket="openNewTicket"
        :on-clear-filter="clearBoardSearch"
        :on-drag-start="onDragStart"
        :on-drag-end="onDragEnd"
        :on-drop-column="onDropColumn"
        :on-drop-card="onDropCard"
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
        @update:ticket="updateNewTicket"
        @close="closeNewTicket"
        @create="createTicketSubmit"
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
        :comments="ticketComments"
        :comment-draft="commentDraft"
        :comment-saving="commentSaving"
        :comment-error="commentError"
        :current-user-id="sessionStore.user?.id"
        :attachments="ticketAttachments"
        :attachment-uploading="attachmentUploading"
        :attachment-error="attachmentError"
        :project-id="props.projectId"
        :ticket-id="selectedTicket?.id || ''"
        :read-only="!canEditTickets"
        @update:editor="updateTicketEditor"
        @update:commentDraft="updateCommentDraft"
        @close="closeTicket"
        @save="saveTicket"
        @delete="deleteTicketSubmit"
        @add-comment="addCommentSubmit"
        @upload-attachment="uploadAttachmentHandler"
        @delete-attachment="deleteAttachmentHandler"
    />
</template>
