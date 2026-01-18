<script setup lang="ts">
import type { StoryRow } from "@/lib/types";
import type { TicketResponse, WorkflowState } from "@/lib/api";
import { Button } from "@/components/ui/button";

type DragHandler = (ticketId: string) => void;
type DropHandler = (stateId: string) => void;
type DropCardHandler = (ticketId: string, stateId: string) => void;
type OpenTicketHandler = (ticket: TicketResponse) => void;
type OpenNewTicketHandler = (stateId?: string, storyId?: string) => void;
type DeleteStoryHandler = (storyId: string) => void;

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
    onInitializeWorkflow: () => void;
    onOpenStoryModal: () => void;
    onDeleteStory: DeleteStoryHandler;
    onOpenTicket: OpenTicketHandler;
    onOpenNewTicket: OpenNewTicketHandler;
    onDragStart: DragHandler;
    onDragEnd: () => void;
    onDropColumn: DropHandler;
    onDropCard: DropCardHandler;
}>();
</script>

<template>
    <section
        v-if="loading"
        class="rounded-3xl border border-border bg-card/80 p-6"
    >
        <p class="text-sm text-muted-foreground">Loading board...</p>
    </section>

    <section
        v-if="!loading && states.length === 0"
        class="rounded-3xl border border-border bg-card/80 p-6 shadow-sm"
    >
        <p class="text-xs uppercase tracking-[0.3em] text-muted-foreground">
            Setup
        </p>
        <h2 class="mt-2 text-2xl font-semibold">Create your first workflow</h2>
        <p class="mt-2 text-sm text-muted-foreground">
            The board is empty because no workflow states exist yet. Initialize
            a default board to start adding tickets.
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
                {{ workflowSetupBusy ? "Creating..." : "Initialize board" }}
            </Button>
        </div>
    </section>

    <section
        v-if="!loading && states.length > 0"
        class="grid gap-4"
        :style="{ '--cols': states.length }"
    >
        <div
            class="flex items-center justify-between rounded-3xl border border-border bg-card/70 px-4 py-3 text-xs text-muted-foreground"
        >
            <div>
                <p class="text-xs font-semibold uppercase tracking-[0.3em]">
                    Board stories
                </p>
                <p class="mt-2 text-[11px]">
                    All stories belong to this board and group tickets
                    horizontally.
                </p>
            </div>
            <div class="flex items-center gap-3">
                <span
                    class="rounded-full bg-muted px-2 py-1 text-[10px] font-semibold"
                >
                    {{ storiesCount }} stories
                </span>
                <button
                    class="flex h-8 w-8 items-center justify-center rounded-full border border-border bg-background text-base font-semibold text-foreground transition hover:border-primary hover:text-primary"
                    @click="props.onOpenStoryModal"
                >
                    +
                </button>
            </div>
        </div>
        <div
            class="grid items-center gap-3 rounded-3xl border border-border bg-card/70 p-3 text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground [grid-template-columns:220px_repeat(var(--cols),minmax(0,1fr))]"
        >
            <div>Story</div>
            <div v-for="state in states" :key="state.id" class="text-center">
                {{ state.name }}
            </div>
        </div>

        <div
            v-for="row in storyRows"
            :key="row.id"
            class="grid gap-3 [grid-template-columns:220px_repeat(var(--cols),minmax(0,1fr))]"
        >
            <div class="rounded-3xl border border-border bg-card/70 p-3">
                <div class="flex items-center justify-between">
                    <p
                        class="text-xs uppercase tracking-[0.3em] text-muted-foreground"
                    >
                        {{ row.isUngrouped ? "Ungrouped" : "Story" }}
                    </p>
                    <details v-if="!row.isUngrouped" class="relative">
                        <summary
                            class="list-none rounded-full border border-border bg-background px-2 py-1 text-[12px] font-semibold uppercase tracking-[0.3em] text-muted-foreground transition hover:border-foreground hover:text-foreground"
                            aria-label="Story actions"
                        >
                            ⋮
                        </summary>
                        <div
                            class="absolute right-0 top-full mt-2 w-36 rounded-2xl border border-border bg-card/80 p-2 text-xs"
                        >
                            <button
                                class="w-full rounded-xl border border-destructive/40 bg-destructive/5 px-2 py-1 text-left text-[11px] font-semibold uppercase tracking-[0.2em] text-destructive transition hover:bg-destructive/10"
                                @click.stop="props.onDeleteStory(row.id)"
                            >
                                Delete story
                            </button>
                        </div>
                    </details>
                </div>
                <p class="mt-2 text-sm font-semibold">{{ row.title }}</p>
                <p
                    v-if="row.description"
                    class="mt-2 text-xs text-muted-foreground"
                >
                    {{ row.description }}
                </p>
                <p class="mt-3 text-xs text-muted-foreground">
                    {{
                        states.reduce(
                            (sum, state) =>
                                sum +
                                (row.ticketsByState[state.id]?.length || 0),
                            0,
                        )
                    }}
                    tickets
                </p>
            </div>

            <div
                v-for="state in states"
                :key="state.id"
                class="flex min-h-[220px] flex-col rounded-3xl border border-border bg-card/70 p-3"
                @dragover.prevent
                @drop.prevent="props.onDropColumn(state.id)"
            >
                <div class="flex flex-1 flex-col gap-3">
                    <div
                        v-for="ticket in row.ticketsByState[state.id]"
                        :key="ticket.id"
                        class="group cursor-grab rounded-2xl border border-border bg-background p-4 shadow-sm transition hover:-translate-y-0.5 hover:shadow-md"
                        draggable="true"
                        @dragstart="props.onDragStart(ticket.id)"
                        @dragend="props.onDragEnd"
                        @dragover.prevent
                        @drop.prevent="props.onDropCard(ticket.id, state.id)"
                    >
                        <div class="flex items-center justify-between">
                            <div class="flex items-center gap-2">
                                <span
                                    class="text-xs font-semibold text-muted-foreground"
                                >
                                    {{ ticket.key }}
                                </span>
                                <span
                                    class="rounded-full bg-muted px-2 py-1 text-[10px] font-semibold uppercase tracking-wide"
                                >
                                    {{ ticket.type }}
                                </span>
                            </div>
                            <span
                                class="rounded-full bg-muted px-2 py-1 text-[10px] font-semibold capitalize"
                            >
                                {{ ticket.priority }}
                            </span>
                        </div>
                        <p class="mt-2 text-sm font-semibold">
                            {{ ticket.title }}
                        </p>
                        <p class="mt-2 text-xs text-muted-foreground">
                            {{ ticket.description }}
                        </p>
                        <div class="mt-3 flex justify-end">
                            <Button
                                variant="ghost"
                                size="sm"
                                @click.stop="props.onOpenTicket(ticket)"
                            >
                                Open
                            </Button>
                        </div>
                        <div
                            class="mt-4 flex items-center justify-between text-xs text-muted-foreground"
                        >
                            <span>Assignee</span>
                            <span
                                class="rounded-full bg-muted px-2 py-1 text-[10px] font-semibold"
                            >
                                {{ ticket.assignee?.name || "Unassigned" }}
                            </span>
                        </div>
                    </div>
                    <button
                        class="mt-auto rounded-2xl border border-dashed border-border px-3 py-3 text-left text-xs text-muted-foreground transition hover:border-primary hover:text-primary"
                        @click="
                            props.onOpenNewTicket(
                                state.id,
                                row.isUngrouped ? undefined : row.id,
                            )
                        "
                    >
                        + Add ticket
                    </button>
                </div>
            </div>
        </div>
    </section>
</template>
