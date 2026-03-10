<script setup lang="ts">
import { computed } from "vue";
import type { StoryRow } from "@/lib/types";
import type { TicketResponse, WorkflowState } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { useI18n } from "@/lib/i18n";
import BoardStoryRow from "@/components/app/board/BoardStoryRow.vue";
import BoardSummaryBar from "@/components/app/board/BoardSummaryBar.vue";
import BoardNoResults from "@/components/app/board/BoardNoResults.vue";
import BoardGridHeader from "@/components/app/board/BoardGridHeader.vue";

type DragHandler = (ticketId: string) => void;
type DropHandler = (stateId: string, storyId: string) => void;
type DropCardHandler = (
    ticketId: string,
    stateId: string,
    storyId: string,
) => void;
type OpenTicketHandler = (ticket: TicketResponse) => void;
type OpenNewTicketHandler = (stateId?: string, storyId?: string) => void;
type DeleteStoryHandler = (storyId: string) => void;
type QuickMoveHandler = (ticket: TicketResponse, nextStateId: string) => void;
type QuickPriorityHandler = (ticket: TicketResponse) => void;
type QuickAssignHandler = (ticket: TicketResponse) => void;

const props = defineProps<{
    loading: boolean;
    states: WorkflowState[];
    storyRows: StoryRow[];
    storiesCount: number;
    ticketsCount: number;
    webhooksCount: number;
    apiMode: "live" | "demo";
    workflowSetupBusy: boolean;
    workflowSetupError: string;
    canEditTickets: boolean;
    canQuickAssignToMe: boolean;
    bulkSelectMode: boolean;
    selectedTicketIds: string[];
    hasActiveFilter: boolean;
    searchQuery: string;
    onInitializeWorkflow: () => void;
    onOpenStoryModal: () => void;
    onToggleBulkSelectMode: () => void;
    onToggleTicketSelection: (ticketId: string) => void;
    onClearTicketSelection: () => void;
    onClearFilter: () => void;
    onDeleteStory: DeleteStoryHandler;
    onOpenTicket: OpenTicketHandler;
    onOpenNewTicket: OpenNewTicketHandler;
    dragging: boolean;
    onDragStart: DragHandler;
    onDragEnd: () => void;
    onDropColumn: DropHandler;
    onDropCard: DropCardHandler;
    onQuickMoveNext: QuickMoveHandler;
    onQuickCyclePriority: QuickPriorityHandler;
    onQuickAssignToMe: QuickAssignHandler;
}>();

const { t } = useI18n();

const ticketCountByState = computed(() => {
    const counts: Record<string, number> = {};
    for (const row of props.storyRows) {
        for (const [stateId, tickets] of Object.entries(row.ticketsByState)) {
            counts[stateId] = (counts[stateId] ?? 0) + tickets.length;
        }
    }
    return counts;
});
</script>

<template>
    <section
        v-if="loading"
        class="rounded-3xl border border-border bg-card/80 p-6 animate-pulse"
    >
        <div class="flex items-center justify-between mb-6">
            <div class="h-8 bg-muted rounded-lg w-48"></div>
            <div class="h-8 bg-muted rounded-lg w-32"></div>
        </div>
        <div class="flex gap-4 mb-4">
            <div v-for="i in 4" :key="i" class="flex-1 h-10 bg-muted/70 rounded-xl"></div>
        </div>
        <div class="space-y-4">
            <div v-for="row in 3" :key="row" class="border border-border/50 rounded-xl p-4">
                <div class="h-6 bg-muted rounded w-1/3 mb-4"></div>
                <div class="grid grid-cols-4 gap-4">
                    <div v-for="col in 4" :key="col" class="space-y-3">
                        <div class="h-24 bg-muted/60 rounded-xl"></div>
                        <div class="h-24 bg-muted/60 rounded-xl"></div>
                    </div>
                </div>
            </div>
        </div>
    </section>

    <section
        v-if="!loading && states.length === 0"
        class="rounded-3xl border border-border bg-card/80 p-6 shadow-sm"
    >
        <p class="text-xs uppercase tracking-[0.3em] text-muted-foreground">
            {{ t("board.view.setup") }}
        </p>
        <h2 class="mt-2 text-2xl font-semibold">{{ t("board.view.createWorkflow") }}</h2>
        <p class="mt-2 text-sm text-muted-foreground">
            {{ t("board.view.emptyWorkflow") }}
        </p>
        <div
            v-if="workflowSetupError"
            class="mt-4 rounded-2xl border border-border bg-secondary/60 px-3 py-2 text-xs"
        >
            {{ workflowSetupError }}
        </div>
        <div class="mt-6">
            <Button
                :disabled="workflowSetupBusy"
                @click="props.onInitializeWorkflow"
            >
                {{ workflowSetupBusy ? t("board.view.creating") : t("board.view.initialize") }}
            </Button>
        </div>
    </section>

    <section
        v-if="!loading && states.length > 0"
        class="grid gap-6"
        :style="{ '--cols': states.length }"
    >
        <BoardSummaryBar
            :stories-count="props.storiesCount"
            :tickets-count="props.ticketsCount"
            :can-edit-tickets="props.canEditTickets"
            :bulk-select-mode="props.bulkSelectMode"
            :selected-count="props.selectedTicketIds.length"
            @toggle-bulk-select-mode="props.onToggleBulkSelectMode"
            @clear-ticket-selection="props.onClearTicketSelection"
            @open-story-modal="props.onOpenStoryModal"
        />

        <BoardNoResults
            v-if="props.hasActiveFilter && props.storyRows.length === 0"
            :search-query="props.searchQuery"
            @clear-filter="props.onClearFilter"
        />

        <div class="w-full overflow-x-auto pb-2">
            <BoardGridHeader :states="props.states" :ticket-count-by-state="ticketCountByState" />

            <BoardStoryRow
                v-for="row in props.storyRows"
                :key="row.id"
                :row="row"
                :states="props.states"
                :can-edit-tickets="props.canEditTickets"
                :can-quick-assign-to-me="props.canQuickAssignToMe"
                :bulk-select-mode="props.bulkSelectMode"
                :selected-ticket-ids="props.selectedTicketIds"
                :on-delete-story="props.onDeleteStory"
                :on-open-new-ticket="props.onOpenNewTicket"
                :on-open-ticket="props.onOpenTicket"
                :on-toggle-ticket-selection="props.onToggleTicketSelection"
                :dragging="props.dragging"
                :on-drag-start="props.onDragStart"
                :on-drag-end="props.onDragEnd"
                :on-drop-column="props.onDropColumn"
                :on-drop-card="props.onDropCard"
                :on-quick-move-next="props.onQuickMoveNext"
                :on-quick-cycle-priority="props.onQuickCyclePriority"
                :on-quick-assign-to-me="props.onQuickAssignToMe"
            />
        </div>
    </section>
</template>
