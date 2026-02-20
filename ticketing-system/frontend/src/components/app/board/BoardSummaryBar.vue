<script setup lang="ts">
import { useI18n } from "@/lib/i18n";

const props = defineProps<{
    storiesCount: number;
    ticketsCount: number;
    canEditTickets: boolean;
    bulkSelectMode: boolean;
    selectedCount: number;
}>();

const emit = defineEmits<{
    (e: "toggle-bulk-select-mode"): void;
    (e: "clear-ticket-selection"): void;
    (e: "open-story-modal"): void;
}>();

const { t } = useI18n();
</script>

<template>
    <div
        class="flex items-center justify-between rounded-3xl border border-border bg-card/70 px-5 py-4 text-xs text-muted-foreground"
    >
        <div>
            <p class="text-xs font-semibold uppercase tracking-[0.3em]">
                {{ t("board.view.storySection") }}
            </p>
            <p class="mt-1 text-[11px]">
                {{ t("board.view.storyHelp") }}
            </p>
        </div>
        <div class="flex items-center gap-3">
            <span class="rounded-full bg-muted px-2 py-1 text-[10px] font-semibold">
                {{ t("board.view.storiesCount", { count: props.storiesCount }) }}
            </span>
            <span class="rounded-full bg-muted px-2 py-1 text-[10px] font-semibold">
                {{ t("board.view.ticketsCount", { count: props.ticketsCount }) }}
            </span>
            <button
                v-if="props.canEditTickets"
                data-testid="board.bulk-toggle-button"
                class="rounded-lg border border-border bg-background px-2.5 py-1.5 text-[10px] font-semibold uppercase tracking-[0.15em] text-muted-foreground transition hover:border-foreground hover:text-foreground"
                @click="emit('toggle-bulk-select-mode')"
            >
                {{
                    props.bulkSelectMode
                        ? t("board.view.exitSelect")
                        : t("board.view.selectTickets")
                }}
            </button>
            <span
                v-if="props.bulkSelectMode"
                data-testid="board.bulk-selected-count"
                class="rounded-full bg-primary/15 px-2 py-1 text-[10px] font-semibold text-primary"
            >
                {{ t("board.view.selectedCount", { count: props.selectedCount }) }}
            </span>
            <button
                v-if="props.bulkSelectMode"
                data-testid="board.bulk-clear-selection-button"
                class="rounded-lg border border-border bg-background px-2.5 py-1.5 text-[10px] font-semibold uppercase tracking-[0.15em] text-muted-foreground transition hover:border-foreground hover:text-foreground"
                @click="emit('clear-ticket-selection')"
            >
                {{ t("common.clear") }}
            </button>
            <button
                v-if="props.canEditTickets"
                data-testid="board.create-story-button"
                class="flex h-8 w-8 items-center justify-center rounded-full border border-border bg-background text-base font-semibold text-foreground transition hover:border-primary hover:text-primary"
                :title="t('board.view.addStory')"
                @click="emit('open-story-modal')"
            >
                +
            </button>
        </div>
    </div>
</template>
