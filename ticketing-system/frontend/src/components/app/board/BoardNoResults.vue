<script setup lang="ts">
import { useI18n } from "@/lib/i18n";

const props = defineProps<{
    searchQuery: string;
}>();

const emit = defineEmits<{
    (e: "clear-filter"): void;
}>();

const { t } = useI18n();
</script>

<template>
    <div class="flex flex-col items-center gap-4 rounded-3xl border border-border bg-card/70 px-6 py-14 text-center">
        <div class="flex h-14 w-14 items-center justify-center rounded-2xl bg-muted text-2xl text-muted-foreground">
            🔍
        </div>
        <div>
            <p class="font-semibold text-foreground">{{ t("board.view.noMatchingTitle") }}</p>
            <p class="mt-1.5 text-sm text-muted-foreground">
                {{ t("board.view.noMatchingBody") }}
                <span class="font-mono font-medium text-foreground">"{{ props.searchQuery }}"</span>.
            </p>
        </div>
        <button
            class="rounded-xl border border-border bg-background px-4 py-2 text-xs font-semibold uppercase tracking-[0.2em] transition hover:border-primary hover:text-primary active:scale-[0.98]"
            @click="emit('clear-filter')"
        >
            {{ t("board.view.clearFilter") }}
        </button>
    </div>
</template>
