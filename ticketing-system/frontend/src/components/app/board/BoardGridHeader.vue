<script setup lang="ts">
import { useI18n } from "@/lib/i18n";
import type { WorkflowState } from "@/lib/api";

const props = defineProps<{
    states: WorkflowState[];
    ticketCountByState?: Record<string, number>;
}>();

const { t } = useI18n();
</script>

<template>
    <div
        class="grid w-full items-center gap-4 rounded-3xl border border-border bg-card/70 p-4 text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground"
        :style="{
            'grid-template-columns':
                'minmax(160px, 15vw) repeat(' +
                props.states.length +
                ', minmax(0, 1fr))',
        }"
    >
        <div>{{ t("board.view.storyGroup") }}</div>
        <div
            v-for="state in props.states"
            :key="state.id"
            class="flex flex-col items-center gap-1"
        >
            <span>{{ state.name }}</span>
            <span
                v-if="state.wipLimit"
                data-testid="board.wip-badge"
                class="inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-semibold"
                :class="
                    (ticketCountByState?.[state.id] ?? 0) >= state.wipLimit
                        ? 'bg-destructive/20 text-destructive'
                        : 'bg-muted text-muted-foreground'
                "
                :title="
                    state.wipEnforcement
                        ? 'WIP limit enforced — moves blocked at limit'
                        : 'WIP limit (soft)'
                "
            >
                {{ ticketCountByState?.[state.id] ?? 0 }}/{{ state.wipLimit }}
                <span v-if="state.wipEnforcement" class="ml-0.5">🔒</span>
            </span>
        </div>
    </div>
</template>
