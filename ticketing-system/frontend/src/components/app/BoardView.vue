<script setup lang="ts">
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
    onDragStart: DragHandler;
    onDragEnd: () => void;
    onDropColumn: DropHandler;
    onDropCard: DropCardHandler;
    onQuickMoveNext: QuickMoveHandler;
    onQuickCyclePriority: QuickPriorityHandler;
    onQuickAssignToMe: QuickAssignHandler;
}>();

const { t } = useI18n();
</script>

<template>
    <section
        v-if="loading"
        class="rounded-3xl border border-border bg-card/80 p-6"
    >
        <p class="text-sm text-muted-foreground">{{ t("board.view.loading") }}</p>
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

        <div class="w-full pb-2">
            <BoardGridHeader :states="props.states" />

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
