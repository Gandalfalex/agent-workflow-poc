<script setup lang="ts">
import { useI18n } from "@/lib/i18n";
import type { WorkflowState } from "@/lib/api";

const props = defineProps<{
    states: WorkflowState[];
}>();

const { t } = useI18n();
</script>

<template>
    <div
        class="grid min-w-max items-center gap-4 rounded-3xl border border-border bg-card/70 p-4 text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground"
        :style="{
            'grid-template-columns':
                'minmax(170px, 15vw) repeat(' +
                props.states.length +
                ', minmax(260px, 1fr))',
        }"
    >
        <div>{{ t("board.view.storyGroup") }}</div>
        <div
            v-for="state in props.states"
            :key="state.id"
            class="text-center"
        >
            {{ state.name }}
        </div>
    </div>
</template>
