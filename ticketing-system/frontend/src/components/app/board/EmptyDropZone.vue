<script setup lang="ts">
import { useI18n } from "@/lib/i18n";

defineProps<{ active?: boolean }>();

const { t } = useI18n();
</script>

<template>
    <div
        class="group flex min-h-[100px] flex-col items-center justify-center rounded-xl border-2 border-dashed px-4 py-6 text-center transition-all duration-150"
        :class="
            active
                ? 'border-primary/70 bg-primary/10 scale-[1.02]'
                : 'border-border/25 bg-background/5 hover:border-primary/35 hover:bg-primary/5'
        "
    >
        <svg
            class="mb-2 h-8 w-8 transition"
            :class="active ? 'text-primary/60' : 'text-muted-foreground/20 group-hover:text-muted-foreground/50'"
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
            class="text-[10px] font-semibold uppercase tracking-wider transition"
            :class="active ? 'text-primary/80' : 'text-muted-foreground/40 group-hover:text-muted-foreground'"
        >
            {{ t("board.view.dropHere") }}
        </p>
    </div>
</template>
