<script setup lang="ts">
import { computed } from "vue";
import { RouterLink } from "vue-router";
import { useI18n } from "@/lib/i18n";
import type { ProjectPortfolioEntry } from "@/stores/portfolio";

const props = defineProps<{
    entry: ProjectPortfolioEntry;
}>();

const { t } = useI18n();

type Health = "critical" | "warning" | "healthy";

const health = computed<Health>(() => {
    if (props.entry.urgentOpen > 0) return "critical";
    if (props.entry.blockedOpen > 0) return "warning";
    return "healthy";
});

const borderClass = computed(() => {
    switch (health.value) {
        case "critical": return "border-red-500/40 hover:border-red-500/70";
        case "warning":  return "border-amber-500/40 hover:border-amber-500/70";
        default:         return "border-border hover:border-foreground/30";
    }
});

const healthDotClass = computed(() => {
    switch (health.value) {
        case "critical": return "bg-red-500";
        case "warning":  return "bg-amber-400";
        default:         return "bg-emerald-400";
    }
});

const healthLabelClass = computed(() => {
    switch (health.value) {
        case "critical": return "text-red-400";
        case "warning":  return "text-amber-400";
        default:         return "text-emerald-400";
    }
});

const sprintProgress = computed(() => {
    const { activeSprintCommitted: committed, activeSprintRemaining: remaining } = props.entry;
    if (!committed || committed === 0) return 0;
    const done = committed - remaining;
    return Math.round(Math.max(0, (done / committed)) * 100);
});

const formatDate = (dateStr: string | null | undefined) => {
    if (!dateStr) return "";
    try {
        return new Date(dateStr).toLocaleDateString(undefined, { month: "short", day: "numeric" });
    } catch {
        return dateStr;
    }
};

const boardRoute = computed(() => ({
    name: "board",
    params: { projectId: props.entry.id },
}));

const dashboardRoute = computed(() => ({
    name: "dashboard",
    params: { projectId: props.entry.id },
}));
</script>

<template>
    <div
        data-testid="portfolio.project-card"
        class="group flex flex-col rounded-2xl border bg-card/70 p-5 transition-all duration-200 cursor-pointer hover:bg-card hover:shadow-md"
        :class="borderClass"
    >
        <!-- Header: health indicator + project key -->
        <div class="flex items-center justify-between mb-2.5">
            <div class="flex items-center gap-2">
                <span class="h-2 w-2 rounded-full shrink-0" :class="healthDotClass"></span>
                <span class="font-mono text-[11px] text-muted-foreground tracking-wider">
                    {{ entry.projectKey }}
                </span>
            </div>
            <span
                data-testid="portfolio.project-health"
                class="text-[10px] font-semibold uppercase tracking-[0.15em]"
                :class="healthLabelClass"
            >
                {{ t(`portfolio.health.${health}`) }}
            </span>
        </div>

        <!-- Project name -->
        <h3 class="font-semibold text-sm leading-snug mb-3 line-clamp-2">
            {{ entry.projectName }}
        </h3>

        <!-- Ticket stats row -->
        <div class="flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted-foreground mb-3">
            <span>
                <span class="tabular-nums font-semibold text-foreground">{{ entry.totalOpen }}</span>
                {{ t("portfolio.open") }}
            </span>
            <span v-if="entry.blockedOpen > 0" class="text-amber-400">
                <span class="tabular-nums font-semibold">{{ entry.blockedOpen }}</span>
                {{ t("portfolio.blocked") }}
            </span>
            <span v-if="entry.urgentOpen > 0" class="text-red-400">
                <span class="tabular-nums font-semibold">{{ entry.urgentOpen }}</span>
                {{ t("portfolio.urgent") }}
            </span>
            <span v-if="entry.weeklyThroughput > 0" data-testid="portfolio.throughput-stat" class="text-emerald-400">
                <span class="tabular-nums font-semibold">{{ entry.weeklyThroughput }}</span>
                {{ t("portfolio.closedThisWeek") }}
            </span>
        </div>

        <!-- Active sprint -->
        <div class="mb-3 flex-1">
            <template v-if="entry.activeSprintName">
                <div class="flex items-center justify-between text-xs mb-1.5">
                    <span class="text-muted-foreground truncate max-w-[60%]">
                        ▸ {{ entry.activeSprintName }}
                    </span>
                    <span class="text-muted-foreground tabular-nums shrink-0">
                        <template v-if="entry.activeSprintEndDate">
                            {{ t("portfolio.ends") }} {{ formatDate(entry.activeSprintEndDate) }}
                        </template>
                    </span>
                </div>
                <div class="h-1 rounded-full overflow-hidden bg-muted">
                    <div
                        class="h-full bg-primary rounded-full transition-all duration-700 ease-out"
                        :style="{ width: sprintProgress + '%' }"
                    />
                </div>
                <div class="flex justify-between text-[10px] text-muted-foreground mt-1">
                    <span>{{ sprintProgress }}%</span>
                    <span>{{ entry.activeSprintRemaining }} {{ t("portfolio.remaining") }}</span>
                </div>
            </template>
            <template v-else>
                <span class="text-[11px] text-muted-foreground/60 italic">
                    {{ t("portfolio.noActiveSprint") }}
                </span>
            </template>
        </div>

        <!-- Footer links -->
        <div class="flex items-center justify-between pt-3 border-t border-border/40 mt-auto">
            <RouterLink
                :to="boardRoute"
                class="text-xs text-muted-foreground hover:text-foreground transition-colors font-medium"
                @click.stop
            >
                {{ t("portfolio.goBoard") }} →
            </RouterLink>
            <RouterLink
                :to="dashboardRoute"
                class="text-xs text-muted-foreground hover:text-foreground transition-colors font-medium"
                @click.stop
            >
                {{ t("portfolio.goStats") }} →
            </RouterLink>
        </div>
    </div>
</template>
