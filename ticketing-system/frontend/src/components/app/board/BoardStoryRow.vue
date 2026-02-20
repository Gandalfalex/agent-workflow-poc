<script setup lang="ts">
import type { StoryRow } from "@/lib/types";
import type { TicketResponse, WorkflowState } from "@/lib/api";
import StoryInfoCell from "@/components/app/board/StoryInfoCell.vue";
import TicketCard from "@/components/app/board/TicketCard.vue";
import EmptyDropZone from "@/components/app/board/EmptyDropZone.vue";

type DropHandler = (stateId: string, storyId: string) => void;
type DropCardHandler = (
    ticketId: string,
    stateId: string,
    storyId: string,
) => void;

const props = defineProps<{
    row: StoryRow;
    states: WorkflowState[];
    canEditTickets: boolean;
    bulkSelectMode: boolean;
    selectedTicketIds: string[];
    onDeleteStory: (storyId: string) => void;
    onOpenNewTicket: (stateId?: string, storyId?: string) => void;
    onOpenTicket: (ticket: TicketResponse) => void;
    onToggleTicketSelection: (ticketId: string) => void;
    onDragStart: (ticketId: string) => void;
    onDragEnd: () => void;
    onDropColumn: DropHandler;
    onDropCard: DropCardHandler;
}>();
</script>

<template>
    <div
        class="mt-5 grid min-w-max gap-4 rounded-2xl border-2 border-border/60 bg-card/50 p-4 shadow-sm"
        :style="{
            'grid-template-columns':
                'minmax(170px, 15vw) repeat(' +
                props.states.length +
                ', minmax(260px, 1fr))',
        }"
    >
        <StoryInfoCell
            :row="props.row"
            :states="props.states"
            :can-edit-tickets="props.canEditTickets"
            :on-delete-story="props.onDeleteStory"
            :on-open-new-ticket="props.onOpenNewTicket"
        />

        <div
            v-for="state in props.states"
            :key="state.id"
            class="flex min-h-[180px] flex-col rounded-2xl border border-border bg-card/35 p-3.5"
            @dragover.prevent
            @drop.prevent="props.onDropColumn(state.id, props.row.id)"
        >
            <div class="flex flex-1 flex-col gap-2.5">
                <TicketCard
                    v-for="ticket in props.row.ticketsByState[state.id]"
                    :key="ticket.id"
                    :ticket="ticket"
                    :state-id="state.id"
                    :row-id="props.row.id"
                    :can-edit-tickets="props.canEditTickets"
                    :bulk-select-mode="props.bulkSelectMode"
                    :selected-ticket-ids="props.selectedTicketIds"
                    :on-open-ticket="props.onOpenTicket"
                    :on-toggle-ticket-selection="props.onToggleTicketSelection"
                    :on-drag-start="props.onDragStart"
                    :on-drag-end="props.onDragEnd"
                    :on-drop-card="props.onDropCard"
                />

                <EmptyDropZone
                    v-if="(props.row.ticketsByState[state.id] || []).length === 0"
                />
            </div>
        </div>
    </div>
</template>
