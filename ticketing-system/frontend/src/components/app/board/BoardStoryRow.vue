<script setup lang="ts">
import type { StoryRow } from "@/lib/types";
import type { TicketResponse, WorkflowState } from "@/lib/api";
import StoryInfoCell from "@/components/app/board/StoryInfoCell.vue";
import TicketCard from "@/components/app/board/TicketCard.vue";
import EmptyDropZone from "@/components/app/board/EmptyDropZone.vue";
import { ref } from "vue";

type DropHandler = (stateId: string, storyId: string) => void;
type DropCardHandler = (
    ticketId: string,
    stateId: string,
    storyId: string,
) => void;

const props = defineProps<{
    row: StoryRow;
    states: WorkflowState[];
    dragging: boolean;
    canEditTickets: boolean;
    canQuickAssignToMe: boolean;
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
    onQuickMoveNext: (ticket: TicketResponse, nextStateId: string) => void;
    onQuickCyclePriority: (ticket: TicketResponse) => void;
    onQuickAssignToMe: (ticket: TicketResponse) => void;
}>();

const dragOverStateId = ref<string | null>(null);
</script>

<template>
    <div
        class="mt-5 grid w-full gap-4 rounded-2xl border-2 border-border/60 bg-card/50 p-4 shadow-sm"
        :style="{
            'grid-template-columns':
                'minmax(160px, 15vw) repeat(' +
                props.states.length +
                ', minmax(0, 1fr))',
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
            v-for="(state, idx) in props.states"
            :key="state.id"
            class="flex min-h-[180px] flex-col rounded-2xl border p-3.5 transition-all duration-150"
            :class="
                dragOverStateId === state.id
                    ? 'border-primary/60 bg-primary/8 shadow-[inset_0_0_0_2px_rgba(var(--primary)/0.2)]'
                    : 'border-border bg-card/35'
            "
            @dragenter.prevent="dragOverStateId = state.id"
            @dragleave="dragOverStateId = null"
            @dragover.prevent
            @drop.prevent="dragOverStateId = null; props.onDropColumn(state.id, props.row.id)"
        >
            <TransitionGroup
                tag="div"
                class="flex flex-1 flex-col gap-2.5"
                :enter-active-class="props.dragging ? '' : 'transition-all duration-200 ease-out'"
                :enter-from-class="props.dragging ? '' : 'opacity-0 translate-y-2'"
                :enter-to-class="props.dragging ? '' : 'opacity-100 translate-y-0'"
                move-class="transition-all duration-200"
                appear
            >
                <TicketCard
                    v-for="(ticket, i) in props.row.ticketsByState[state.id]"
                    :key="ticket.id"
                    :style="{ transitionDelay: `${i * 40}ms` }"
                    :ticket="ticket"
                    :state-id="state.id"
                    :row-id="props.row.id"
                    :next-state-id="props.states[idx + 1]?.id || ''"
                    :can-edit-tickets="props.canEditTickets"
                    :can-quick-assign-to-me="props.canQuickAssignToMe"
                    :bulk-select-mode="props.bulkSelectMode"
                    :selected-ticket-ids="props.selectedTicketIds"
                    :on-open-ticket="props.onOpenTicket"
                    :on-toggle-ticket-selection="props.onToggleTicketSelection"
                    :on-drag-start="props.onDragStart"
                    :on-drag-end="props.onDragEnd"
                    :on-drop-card="props.onDropCard"
                    :on-quick-move-next="props.onQuickMoveNext"
                    :on-quick-cycle-priority="props.onQuickCyclePriority"
                    :on-quick-assign-to-me="props.onQuickAssignToMe"
                />

                <EmptyDropZone
                    v-if="(props.row.ticketsByState[state.id] || []).length === 0"
                    key="empty-drop"
                    :active="dragOverStateId === state.id"
                />
            </TransitionGroup>
        </div>
    </div>
</template>
