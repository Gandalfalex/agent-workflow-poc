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
    <div
        class="rounded-3xl border border-border bg-card/70 p-6 text-sm text-muted-foreground"
    >
        <p class="font-semibold text-foreground">{{ t("board.view.noMatchingTitle") }}</p>
        <p class="mt-2">
            {{ t("board.view.noMatchingBody") }}
            <span class="font-mono text-foreground">
                "{{ props.searchQuery }}"
            </span>
            .
        </p>
        <button
            class="mt-4 rounded-xl border border-border bg-background px-3 py-2 text-xs font-semibold uppercase tracking-[0.2em] transition hover:border-primary hover:text-primary"
            @click="emit('clear-filter')"
        >
            {{ t("board.view.clearFilter") }}
        </button>
    </div>
</template>
