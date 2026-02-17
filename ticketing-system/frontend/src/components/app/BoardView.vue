<script setup lang="ts">
import type { StoryRow } from "@/lib/types";
import type { TicketResponse, WorkflowState } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { ref } from "vue";

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
    hasActiveFilter: boolean;
    searchQuery: string;
    onInitializeWorkflow: () => void;
    onOpenStoryModal: () => void;
    onClearFilter: () => void;
    onDeleteStory: DeleteStoryHandler;
    onOpenTicket: OpenTicketHandler;
    onOpenNewTicket: OpenNewTicketHandler;
    onDragStart: DragHandler;
    onDragEnd: () => void;
    onDropColumn: DropHandler;
    onDropCard: DropCardHandler;
}>();

const openStoryMenu = ref<string | null>(null);

const toggleStoryMenu = (storyId: string) => {
    openStoryMenu.value = openStoryMenu.value === storyId ? null : storyId;
};

const closeStoryMenu = () => {
    openStoryMenu.value = null;
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

const typeColor = (type: string) => {
    switch (type) {
        case "bug":
            return "bg-rose-500/15 text-rose-300";
        case "feature":
        default:
            return "bg-sky-500/15 text-sky-300";
    }
};

const assigneeInitials = (name?: string) => {
    if (!name) return "?";
    const parts = name.trim().split(/\s+/);
    if (parts.length >= 2)
        return ((parts[0]?.[0] ?? "") + (parts[1]?.[0] ?? "")).toUpperCase();
    return name.slice(0, 2).toUpperCase();
};
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
                <p class="mt-1 text-[11px]">
                    Stories group tickets horizontally across workflow states.
                </p>
            </div>
            <div class="flex items-center gap-3">
                <span
                    class="rounded-full bg-muted px-2 py-1 text-[10px] font-semibold"
                >
                    {{ storiesCount }} stories
                </span>
                <span
                    class="rounded-full bg-muted px-2 py-1 text-[10px] font-semibold"
                >
                    {{ ticketsCount }} tickets
                </span>
                <button
                    v-if="props.canEditTickets"
                    data-testid="board.create-story-button"
                    class="flex h-8 w-8 items-center justify-center rounded-full border border-border bg-background text-base font-semibold text-foreground transition hover:border-primary hover:text-primary"
                    @click="props.onOpenStoryModal"
                >
                    +
                </button>
            </div>
        </div>

        <div
            v-if="props.hasActiveFilter && props.storyRows.length === 0"
            class="rounded-3xl border border-border bg-card/70 p-6 text-sm text-muted-foreground"
        >
            <p class="font-semibold text-foreground">No matching tickets</p>
            <p class="mt-2">
                Nothing matched
                <span class="font-mono text-foreground">
                    "{{ props.searchQuery }}"
                </span>
                .
            </p>
            <button
                class="mt-4 rounded-xl border border-border bg-background px-3 py-2 text-xs font-semibold uppercase tracking-[0.2em] transition hover:border-primary hover:text-primary"
                @click="props.onClearFilter"
            >
                Clear filter
            </button>
        </div>

        <!-- Board grid -->
        <div>
            <div
                class="grid items-center gap-3 rounded-3xl border border-border bg-card/70 p-3 text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground"
                :style="{
                    'grid-template-columns':
                        '200px repeat(' +
                        states.length +
                        ', minmax(220px, 1fr))',
                }"
            >
                <div>Story</div>
                <div
                    v-for="state in states"
                    :key="state.id"
                    class="text-center"
                >
                    {{ state.name }}
                </div>
            </div>

            <div
                v-for="row in storyRows"
                :key="row.id"
                class="mt-4 grid gap-3 rounded-2xl border-2 p-3 border-border/60 bg-card/50 shadow-sm"
                :style="{
                    'grid-template-columns':
                        '200px repeat(' +
                        states.length +
                        ', minmax(220px, 1fr))',
                }"
            >
                <div
                    class="flex flex-col rounded-2xl border border-border p-4 bg-gradient-to-br from-card/90 to-card/70"
                >
                    <div class="flex items-center justify-between mb-3">
                        <p
                            class="text-[10px] uppercase tracking-[0.3em] font-bold text-primary/80"
                        >
                            Story
                        </p>
                        <div class="relative">
                            <button
                                v-if="props.canEditTickets"
                                class="rounded-full border border-border bg-background px-2 py-1 text-[12px] font-semibold uppercase tracking-[0.3em] text-muted-foreground transition hover:border-foreground hover:text-foreground"
                                aria-label="Story actions"
                                @click.stop="toggleStoryMenu(row.id)"
                            >
                                &#x22EE;
                            </button>
                            <div
                                v-if="
                                    props.canEditTickets &&
                                    openStoryMenu === row.id
                                "
                                class="dropdown-backdrop"
                                @click="closeStoryMenu"
                            ></div>
                            <div
                                v-if="
                                    props.canEditTickets &&
                                    openStoryMenu === row.id
                                "
                                class="absolute right-0 top-full z-50 mt-2 w-36 rounded-2xl border border-border bg-card/95 backdrop-blur p-2 text-xs shadow-lg"
                            >
                                <button
                                    class="w-full rounded-xl border border-destructive/40 bg-destructive/5 px-2 py-1 text-left text-[11px] font-semibold uppercase tracking-[0.2em] text-destructive transition hover:bg-destructive/10"
                                    @click.stop="
                                        closeStoryMenu();
                                        props.onDeleteStory(row.id);
                                    "
                                >
                                    Delete story
                                </button>
                            </div>
                        </div>
                    </div>
                    <p class="text-sm font-bold leading-tight text-foreground">
                        {{ row.title }}
                    </p>
                    <p
                        v-if="row.description"
                        class="mt-2 text-xs text-muted-foreground line-clamp-2 leading-relaxed"
                    >
                        {{ row.description }}
                    </p>
                    <div class="mt-auto pt-4 space-y-2">
                        <div class="flex items-center gap-2">
                            <div
                                class="h-1 flex-1 rounded-full bg-muted overflow-hidden"
                            >
                                <div
                                    class="h-full bg-primary/60 transition-all"
                                    :style="{
                                        width:
                                            states.reduce(
                                                (sum, state) =>
                                                    sum +
                                                    (row.ticketsByState[
                                                        state.id
                                                    ]?.length || 0),
                                                0,
                                            ) *
                                                10 +
                                            '%',
                                    }"
                                ></div>
                            </div>
                            <p
                                class="text-[10px] font-semibold text-muted-foreground whitespace-nowrap"
                            >
                                {{
                                    states.reduce(
                                        (sum, state) =>
                                            sum +
                                            (row.ticketsByState[state.id]
                                                ?.length || 0),
                                        0,
                                    )
                                }}
                                tickets
                            </p>
                        </div>
                        <button
                            v-if="props.canEditTickets"
                            data-testid="board.add-ticket-button"
                            class="w-full rounded-xl border-2 border-dashed border-border px-3 py-2.5 text-xs font-semibold text-muted-foreground transition-all hover:border-primary hover:bg-primary/5 hover:text-primary"
                            @click="
                                props.onOpenNewTicket(states[0]?.id, row.id)
                            "
                        >
                            + Add ticket
                        </button>
                    </div>
                </div>

                <div
                    v-for="state in states"
                    :key="state.id"
                    class="flex min-h-[160px] flex-col rounded-2xl border border-border bg-card/40 p-2.5"
                    @dragover.prevent
                    @drop.prevent="props.onDropColumn(state.id, row.id)"
                >
                    <div class="flex flex-1 flex-col gap-2.5">
                        <!-- Ticket card: enhanced design with better visual hierarchy -->
                        <div
                            v-for="ticket in row.ticketsByState[state.id]"
                            :key="ticket.id"
                            class="group relative cursor-grab rounded-xl border-2 border-border bg-gradient-to-br from-background to-background/80 p-3.5 shadow-sm transition-all hover:-translate-y-1 hover:shadow-lg hover:border-primary/40 hover:shadow-primary/5"
                            draggable="true"
                            @dragstart="props.onDragStart(ticket.id)"
                            @dragend="props.onDragEnd"
                            @dragover.prevent
                            @drop.prevent="
                                props.onDropCard(ticket.id, state.id, row.id)
                            "
                            @click="props.onOpenTicket(ticket)"
                        >
                            <!-- Priority indicator bar -->
                            <div
                                class="absolute top-0 left-0 right-0 h-1 rounded-t-xl"
                                :class="{
                                    'bg-red-500': ticket.priority === 'urgent',
                                    'bg-orange-500': ticket.priority === 'high',
                                    'bg-amber-500':
                                        ticket.priority === 'medium',
                                    'bg-slate-500': ticket.priority === 'low',
                                }"
                            ></div>

                            <div
                                class="flex items-start justify-between gap-2 mb-2"
                            >
                                <span
                                    class="text-[10px] font-bold text-muted-foreground tracking-wider"
                                >
                                    {{ ticket.key }}
                                </span>
                                <div class="flex items-center gap-1.5">
                                    <span
                                        class="rounded-md px-1.5 py-0.5 text-[8px] font-bold uppercase tracking-wider border"
                                        :class="priorityColor(ticket.priority)"
                                    >
                                        {{ ticket.priority.substring(0, 3) }}
                                    </span>
                                    <span
                                        class="rounded-md px-1.5 py-0.5 text-[8px] font-bold uppercase tracking-wider"
                                        :class="typeColor(ticket.type)"
                                    >
                                        {{
                                            ticket.type === "bug" ? "🐛" : "✨"
                                        }}
                                    </span>
                                </div>
                            </div>

                            <p
                                class="text-sm font-bold leading-snug text-foreground mb-2"
                            >
                                {{ ticket.title }}
                            </p>

                            <p
                                v-if="ticket.description"
                                class="text-xs text-muted-foreground line-clamp-2 leading-relaxed mb-3"
                            >
                                {{ ticket.description }}
                            </p>

                            <div
                                class="flex items-center justify-between pt-2 border-t border-border/50"
                            >
                                <span
                                    class="text-[9px] font-semibold uppercase tracking-wider text-muted-foreground"
                                >
                                    {{ ticket.type }}
                                </span>
                                <div class="flex items-center gap-1.5">
                                    <span
                                        v-if="ticket.assignee?.name"
                                        class="flex h-6 w-6 items-center justify-center rounded-full bg-gradient-to-br from-primary/20 to-primary/10 text-[9px] font-bold text-primary border border-primary/20 ring-2 ring-background"
                                        :title="ticket.assignee.name"
                                    >
                                        {{
                                            assigneeInitials(
                                                ticket.assignee.name,
                                            )
                                        }}
                                    </span>
                                    <span
                                        v-else
                                        class="flex h-6 w-6 items-center justify-center rounded-full bg-muted text-[9px] font-semibold text-muted-foreground border border-border"
                                        title="Unassigned"
                                    >
                                        ?
                                    </span>
                                </div>
                            </div>
                        </div>

                        <!-- Empty column state: drop zone only, no add button -->
                        <div
                            v-if="
                                (row.ticketsByState[state.id] || []).length ===
                                0
                            "
                            class="flex min-h-[100px] flex-col items-center justify-center rounded-xl border-2 border-dashed border-border/60 bg-background/20 px-4 py-6 text-center transition-colors hover:border-primary/40 hover:bg-primary/5"
                        >
                            <svg
                                class="w-8 h-8 text-muted-foreground/40 mb-2"
                                fill="none"
                                stroke="currentColor"
                                viewBox="0 0 24 24"
                            >
                                <path
                                    stroke-linecap="round"
                                    stroke-linejoin="round"
                                    stroke-width="2"
                                    d="M19 14l-7 7m0 0l-7-7m7 7V3"
                                ></path>
                            </svg>
                            <p
                                class="text-[10px] font-semibold text-muted-foreground uppercase tracking-wider"
                            >
                                Drop here
                            </p>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </section>
</template>
