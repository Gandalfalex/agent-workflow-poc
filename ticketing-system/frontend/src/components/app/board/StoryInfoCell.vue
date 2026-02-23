<script setup lang="ts">
import { computed, ref } from "vue";
import type { StoryRow } from "@/lib/types";
import type { WorkflowState } from "@/lib/api";
import { useI18n } from "@/lib/i18n";

const props = defineProps<{
    row: StoryRow;
    states: WorkflowState[];
    canEditTickets: boolean;
    onDeleteStory: (storyId: string) => void;
    onOpenNewTicket: (stateId?: string, storyId?: string) => void;
}>();

const { t } = useI18n();
const menuOpen = ref(false);

const totalTickets = computed(() =>
    props.states.reduce(
        (sum, state) => sum + (props.row.ticketsByState[state.id]?.length || 0),
        0,
    ),
);

const usedStoryPoints = computed(() => {
    let sum = 0;
    for (const state of props.states) {
        for (const ticket of props.row.ticketsByState[state.id] || []) {
            if (ticket.storyPoints != null) sum += ticket.storyPoints;
        }
    }
    return sum;
});

const segmentWidth = (stateId: string) => {
    const count = props.row.ticketsByState[stateId]?.length || 0;
    return (count / Math.max(1, totalTickets.value)) * 100;
};

const closeMenu = () => {
    menuOpen.value = false;
};
</script>

<template>
    <div
        class="flex flex-col rounded-2xl border border-border bg-gradient-to-br from-card/90 to-card/70 p-4"
    >
        <div class="mb-3 flex items-center justify-between">
            <p class="text-[10px] font-bold uppercase tracking-[0.3em] text-primary/80">
                {{ t("board.view.storyLabel") }}
            </p>
            <div class="relative">
                <button
                    v-if="props.canEditTickets"
                    class="rounded-full border border-border bg-background px-2 py-1 text-[12px] font-semibold uppercase tracking-[0.3em] text-muted-foreground transition hover:border-foreground hover:text-foreground"
                    :aria-label="t('board.view.storyActions')"
                    @click.stop="menuOpen = !menuOpen"
                >
                    &#x22EE;
                </button>
                <div
                    v-if="props.canEditTickets && menuOpen"
                    class="dropdown-backdrop"
                    @click="closeMenu"
                ></div>
                <div
                    v-if="props.canEditTickets && menuOpen"
                    class="absolute right-0 top-full z-50 mt-2 w-36 rounded-2xl border border-border bg-card/95 p-2 text-xs shadow-lg backdrop-blur"
                >
                    <button
                        class="w-full rounded-xl border border-destructive/40 bg-destructive/5 px-2 py-1 text-left text-[11px] font-semibold uppercase tracking-[0.2em] text-destructive transition hover:bg-destructive/10"
                        @click.stop="
                            closeMenu();
                            props.onDeleteStory(props.row.id);
                        "
                    >
                        {{ t("board.view.deleteStory") }}
                    </button>
                </div>
            </div>
        </div>

        <p class="text-sm font-bold leading-tight text-foreground">
            {{ props.row.title }}
        </p>
        <p
            v-if="props.row.description"
            class="mt-2 line-clamp-2 text-xs leading-relaxed text-muted-foreground"
        >
            {{ props.row.description }}
        </p>

        <div
            v-if="props.row.storyPoints != null"
            class="mt-2 flex items-center gap-2"
        >
            <div class="flex h-1.5 flex-1 overflow-hidden rounded-full bg-muted">
                <div
                    class="h-full rounded-full transition-all"
                    :class="usedStoryPoints > (props.row.storyPoints || 0) ? 'bg-amber-500' : 'bg-violet-500'"
                    :style="{ width: Math.min(100, (usedStoryPoints / Math.max(1, props.row.storyPoints || 1)) * 100) + '%' }"
                />
            </div>
            <span class="whitespace-nowrap text-[10px] font-semibold" :class="usedStoryPoints > (props.row.storyPoints || 0) ? 'text-amber-400' : 'text-violet-400'">
                {{ usedStoryPoints }}/{{ props.row.storyPoints }} SP
            </span>
            <span
                v-if="usedStoryPoints > (props.row.storyPoints || 0)"
                class="rounded-md border border-amber-400/30 bg-amber-500/10 px-1 py-0.5 text-[8px] font-bold uppercase text-amber-300"
            >
                {{ t("board.view.overBudget") }}
            </span>
        </div>

        <div class="mt-auto space-y-2 pt-4">
            <div class="flex items-center gap-2">
                <div class="flex h-1.5 flex-1 overflow-hidden rounded-full bg-muted">
                    <div
                        v-for="(state, idx) in props.states"
                        :key="`progress-${props.row.id}-${state.id}`"
                        class="h-full transition-all"
                        :class="idx % 2 === 0 ? 'bg-primary/70' : 'bg-primary/40'"
                        :style="{ width: segmentWidth(state.id) + '%' }"
                    />
                </div>
                <p class="whitespace-nowrap text-[10px] font-semibold text-muted-foreground">
                    {{ t("board.view.ticketCount", { count: totalTickets }) }}
                </p>
            </div>
            <button
                v-if="props.canEditTickets"
                data-testid="board.add-ticket-button"
                class="w-full rounded-xl border-2 border-dashed border-border px-3 py-2.5 text-xs font-semibold text-muted-foreground transition-all hover:border-primary hover:bg-primary/5 hover:text-primary"
                @click="props.onOpenNewTicket(props.states[0]?.id, props.row.id)"
            >
                {{ t("board.view.addTicket") }}
            </button>
        </div>
    </div>
</template>
