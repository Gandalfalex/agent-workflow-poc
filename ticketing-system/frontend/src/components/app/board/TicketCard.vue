<script setup lang="ts">
import type { TicketResponse } from "@/lib/api";
import { useI18n } from "@/lib/i18n";

const props = defineProps<{
    ticket: TicketResponse;
    stateId: string;
    rowId: string;
    canEditTickets: boolean;
    bulkSelectMode: boolean;
    selectedTicketIds: string[];
    onOpenTicket: (ticket: TicketResponse) => void;
    onToggleTicketSelection: (ticketId: string) => void;
    onDragStart: (ticketId: string) => void;
    onDragEnd: () => void;
    onDropCard: (ticketId: string, stateId: string, storyId: string) => void;
}>();

const { t } = useI18n();

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

const priorityStripeClass = (priority: string) => {
    switch (priority) {
        case "urgent":
            return "bg-red-500";
        case "high":
            return "bg-orange-500";
        case "medium":
            return "bg-amber-500";
        case "low":
        default:
            return "bg-slate-500";
    }
};

const assigneeInitials = (name?: string) => {
    if (!name) return "?";
    const parts = name.trim().split(/\s+/);
    if (parts.length >= 2) {
        return ((parts[0]?.[0] ?? "") + (parts[1]?.[0] ?? "")).toUpperCase();
    }
    return name.slice(0, 2).toUpperCase();
};

</script>

<template>
    <div
        class="group relative cursor-grab rounded-xl border-2 border-border bg-gradient-to-br from-background to-background/80 p-4 pl-7 shadow-sm transition-all hover:-translate-y-1 hover:shadow-lg hover:border-primary/40 hover:shadow-primary/5"
        :class="
            props.selectedTicketIds.includes(props.ticket.id)
                ? 'border-primary/70 ring-2 ring-primary/30 bg-primary/10'
                : ''
        "
        :draggable="props.canEditTickets && !props.bulkSelectMode"
        @dragstart="
            props.bulkSelectMode ? undefined : props.onDragStart(props.ticket.id)
        "
        @dragend="props.onDragEnd"
        @dragover.prevent
        @drop.prevent="
            props.onDropCard(props.ticket.id, props.stateId, props.rowId)
        "
        @click="
            props.bulkSelectMode
                ? props.onToggleTicketSelection(props.ticket.id)
                : props.onOpenTicket(props.ticket)
        "
    >
        <button
            v-if="props.bulkSelectMode"
            :data-testid="`board.ticket-select-${props.ticket.key}`"
            class="absolute left-2 top-2 z-10 flex h-5 w-5 items-center justify-center rounded border text-[10px] font-bold"
            :class="
                props.selectedTicketIds.includes(props.ticket.id)
                    ? 'border-primary bg-primary text-primary-foreground shadow-sm'
                    : 'border-border bg-background text-muted-foreground'
            "
            @click.stop="props.onToggleTicketSelection(props.ticket.id)"
        >
            {{ props.selectedTicketIds.includes(props.ticket.id) ? "✓" : "" }}
        </button>
        <div
            class="absolute inset-y-0 left-0 w-1 rounded-l-xl"
            :class="priorityStripeClass(props.ticket.priority)"
        />
        <div
            class="absolute left-2 top-8 text-muted-foreground/60"
            :title="props.canEditTickets ? t('board.view.dragCard') : ''"
        >
            ⋮⋮
        </div>

        <div class="mb-2 flex items-start justify-between gap-2">
            <div class="flex items-center gap-1.5">
                <span class="font-mono text-[10px] font-bold tracking-wider text-slate-300">
                    {{ props.ticket.key }}
                </span>
                <span
                    v-if="props.ticket.isBlocked"
                    data-testid="board.ticket-blocked-badge"
                    class="rounded-md border border-rose-400/40 bg-rose-500/10 px-1.5 py-0.5 text-[8px] font-bold uppercase tracking-wider text-rose-300"
                >
                    {{
                        t("board.view.blocked", {
                            count: props.ticket.blockedByCount,
                        })
                    }}
                </span>
            </div>
            <div class="flex items-center gap-1.5">
                <span
                    class="rounded-md border px-1.5 py-0.5 text-[8px] font-bold uppercase tracking-wider"
                    :class="priorityColor(props.ticket.priority)"
                >
                    {{ props.ticket.priority.substring(0, 3) }}
                </span>
                <span
                    class="rounded-md px-1.5 py-0.5 text-[8px] font-bold uppercase tracking-wider"
                    :class="typeColor(props.ticket.type)"
                    :title="props.ticket.type"
                >
                    {{ props.ticket.type === "bug" ? "🐛" : "✨" }}
                </span>
            </div>
        </div>

        <p
            class="mb-2 line-clamp-1 text-sm font-bold leading-snug text-foreground"
            :title="props.ticket.title"
        >
            {{ props.ticket.title }}
        </p>

        <p
            v-if="props.ticket.description"
            class="mb-3 line-clamp-2 text-xs leading-relaxed text-slate-400"
        >
            {{ props.ticket.description }}
        </p>

        <div class="flex items-center justify-between border-t border-border/50 pt-2">
            <span class="text-[9px] font-semibold uppercase tracking-wider text-muted-foreground">
                {{ props.ticket.type }}
            </span>
            <div class="flex items-center gap-1.5">
                <span
                    v-if="props.ticket.assignee?.name"
                    class="flex h-6 w-6 items-center justify-center rounded-full border border-primary/20 bg-gradient-to-br from-primary/20 to-primary/10 text-[9px] font-bold text-primary ring-2 ring-background"
                    :title="props.ticket.assignee.name"
                >
                    {{ assigneeInitials(props.ticket.assignee.name) }}
                </span>
                <span
                    v-else
                    class="flex h-6 w-6 items-center justify-center rounded-full border border-border bg-muted text-[9px] font-semibold text-muted-foreground"
                    :title="t('board.view.unassigned')"
                >
                    ?
                </span>
            </div>
        </div>
    </div>
</template>
