<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { usePortfolioStore } from "@/stores/portfolio";
import { useI18n } from "@/lib/i18n";
import { listGroups, type Group } from "@/lib/api";
import ProjectHealthCard from "@/components/app/portfolio/ProjectHealthCard.vue";
import type { ProjectPortfolioEntry } from "@/stores/portfolio";

const portfolioStore = usePortfolioStore();
const { t } = useI18n();

const search = ref("");
const healthFilter = ref<"" | "critical" | "warning" | "healthy">("");
const groupFilter = ref("");
const exportBusy = ref(false);

const groups = ref<Group[]>([]);

const loadWithFilter = () => {
    portfolioStore.load({
        groupId: groupFilter.value || undefined,
    });
};

watch(groupFilter, loadWithFilter);

const loading = computed(() => portfolioStore.loading);
const totals = computed(() => portfolioStore.totals);
const atRiskCount = computed(() => portfolioStore.atRiskCount);

type Health = "critical" | "warning" | "healthy";
const health = (p: ProjectPortfolioEntry): Health => {
    if (p.urgentOpen > 0) return "critical";
    if (p.blockedOpen > 0) return "warning";
    return "healthy";
};

const filteredProjects = computed(() => {
    return portfolioStore.projects
        .filter((p) => !healthFilter.value || health(p) === healthFilter.value)
        .filter((p) => {
            if (!search.value) return true;
            const q = search.value.toLowerCase();
            return (
                p.projectName.toLowerCase().includes(q) ||
                p.projectKey.toLowerCase().includes(q)
            );
        });
});

const noResults = computed(
    () => !loading.value && portfolioStore.projects.length > 0 && filteredProjects.value.length === 0,
);
const isEmpty = computed(
    () => !loading.value && portfolioStore.projects.length === 0,
);

const doExport = async () => {
    if (exportBusy.value) return;
    exportBusy.value = true;
    try {
        await portfolioStore.exportCsv();
    } finally {
        exportBusy.value = false;
    }
};

onMounted(async () => {
    portfolioStore.load();
    try {
        const res = await listGroups();
        groups.value = res.items ?? [];
    } catch {
        // non-critical
    }
});
</script>

<template>
    <div data-testid="portfolio.page" class="w-full">
        <!-- Page header -->
        <div class="flex items-start justify-between mb-8">
            <div>
                <h1 class="text-2xl font-semibold tracking-tight">
                    {{ t("portfolio.title") }}
                </h1>
                <p class="text-sm text-muted-foreground mt-1">
                    {{ t("portfolio.subtitle", { n: totals?.totalProjects ?? 0 }) }}
                </p>
            </div>
            <div class="flex items-center gap-2">
                <button
                    data-testid="portfolio.export-button"
                    class="rounded-lg border border-border bg-card/60 px-3 py-1.5 text-xs font-medium text-muted-foreground transition hover:border-foreground/40 hover:text-foreground disabled:opacity-50"
                    :disabled="exportBusy || loading"
                    @click="doExport"
                >
                    {{ exportBusy ? t("common.loading") : t("portfolio.export") }}
                </button>
                <button
                    class="rounded-lg border border-border bg-card/60 px-3 py-1.5 text-xs font-medium text-muted-foreground transition hover:border-foreground/40 hover:text-foreground"
                    @click="loadWithFilter()"
                >
                    &#x21BB;
                </button>
            </div>
        </div>

        <!-- KPI summary bar (skeleton while loading) -->
        <div class="grid grid-cols-4 gap-4 mb-8">
            <template v-if="loading && !totals">
                <div
                    v-for="i in 4"
                    :key="i"
                    class="rounded-2xl border border-border bg-card/40 h-28 animate-pulse"
                />
            </template>
            <template v-else>
                <!-- Projects -->
                <div
                    data-testid="portfolio.stats-total"
                    class="rounded-2xl border border-border bg-card/70 p-5"
                >
                    <p class="text-[10px] uppercase tracking-[0.25em] text-muted-foreground mb-1">
                        {{ t("portfolio.summary.projects") }}
                    </p>
                    <p class="text-4xl font-semibold tabular-nums">
                        {{ totals?.totalProjects ?? 0 }}
                    </p>
                    <p class="text-xs text-muted-foreground mt-1">
                        {{ t("portfolio.summary.projectsSub", { n: atRiskCount }) }}
                    </p>
                </div>
                <!-- Open -->
                <div
                    data-testid="portfolio.stats-open"
                    class="rounded-2xl border border-border bg-card/70 p-5"
                >
                    <p class="text-[10px] uppercase tracking-[0.25em] text-muted-foreground mb-1">
                        {{ t("portfolio.summary.open") }}
                    </p>
                    <p class="text-4xl font-semibold tabular-nums text-blue-400">
                        {{ totals?.totalOpen ?? 0 }}
                    </p>
                    <p class="text-xs text-muted-foreground mt-1">
                        {{ t("portfolio.summary.openSub") }}
                    </p>
                </div>
                <!-- Blocked -->
                <div
                    data-testid="portfolio.stats-blocked"
                    class="rounded-2xl border border-border bg-card/70 p-5"
                >
                    <p class="text-[10px] uppercase tracking-[0.25em] text-muted-foreground mb-1">
                        {{ t("portfolio.summary.blocked") }}
                    </p>
                    <p class="text-4xl font-semibold tabular-nums text-rose-400">
                        {{ totals?.totalBlocked ?? 0 }}
                    </p>
                    <p class="text-xs text-muted-foreground mt-1">
                        {{ t("portfolio.summary.blockedSub") }}
                    </p>
                </div>
                <!-- At risk -->
                <div
                    data-testid="portfolio.stats-at-risk"
                    class="rounded-2xl border border-border bg-card/70 p-5"
                >
                    <p class="text-[10px] uppercase tracking-[0.25em] text-muted-foreground mb-1">
                        {{ t("portfolio.summary.atRisk") }}
                    </p>
                    <p class="text-4xl font-semibold tabular-nums text-amber-400">
                        {{ atRiskCount }}
                    </p>
                    <p class="text-xs text-muted-foreground mt-1">
                        {{ t("portfolio.summary.atRiskSub") }}
                    </p>
                </div>
            </template>
        </div>

        <!-- Filter bar -->
        <div class="flex items-center justify-between mb-5">
            <h2 class="text-[10px] uppercase tracking-[0.25em] text-muted-foreground font-semibold">
                {{ t("portfolio.projects") }}
            </h2>
            <div class="flex items-center gap-2">
                <input
                    v-model="search"
                    :placeholder="t('portfolio.search')"
                    class="h-8 rounded-lg border border-input bg-background px-3 text-xs focus:outline-none focus:ring-2 focus:ring-ring w-48"
                />
                <select
                    v-model="healthFilter"
                    class="h-8 rounded-lg border border-input bg-background px-2 text-xs focus:outline-none focus:ring-2 focus:ring-ring"
                >
                    <option value="">{{ t("portfolio.filter.all") }}</option>
                    <option value="critical">{{ t("portfolio.health.critical") }}</option>
                    <option value="warning">{{ t("portfolio.health.warning") }}</option>
                    <option value="healthy">{{ t("portfolio.health.healthy") }}</option>
                </select>
                <select
                    v-if="groups.length > 0"
                    v-model="groupFilter"
                    data-testid="portfolio.filter-group-select"
                    class="h-8 rounded-lg border border-input bg-background px-2 text-xs focus:outline-none focus:ring-2 focus:ring-ring"
                >
                    <option value="">{{ t("portfolio.filter.allGroups") }}</option>
                    <option v-for="g in groups" :key="g.id" :value="g.id">{{ g.name }}</option>
                </select>
            </div>
        </div>

        <!-- Project grid -->
        <div class="grid grid-cols-3 gap-4">
            <!-- Loading skeletons -->
            <template v-if="loading && portfolioStore.projects.length === 0">
                <div
                    v-for="i in 6"
                    :key="i"
                    class="rounded-2xl border border-border bg-card/40 h-52 animate-pulse"
                />
            </template>

            <!-- Project cards -->
            <ProjectHealthCard
                v-for="entry in filteredProjects"
                :key="entry.id"
                :entry="entry"
            />

            <!-- No results after filter -->
            <div
                v-if="noResults"
                class="col-span-3 flex flex-col items-center justify-center rounded-2xl border border-border bg-card/40 py-16 text-center"
            >
                <p class="text-sm font-medium text-muted-foreground mb-1">
                    {{ t("portfolio.noMatches") }}
                </p>
                <button
                    class="mt-3 text-xs underline text-muted-foreground hover:text-foreground"
                    @click="search = ''; healthFilter = ''; groupFilter = ''"
                >
                    {{ t("portfolio.clearFilters") }}
                </button>
            </div>

            <!-- Empty workspace -->
            <div
                v-if="isEmpty"
                class="col-span-3 flex flex-col items-center justify-center rounded-2xl border border-border bg-card/40 py-20 text-center"
            >
                <div class="h-12 w-12 rounded-2xl bg-primary/10 flex items-center justify-center mb-4">
                    <span class="text-2xl text-primary/60">◫</span>
                </div>
                <p class="text-sm font-semibold">
                    {{ t("portfolio.noProjects") }}
                </p>
                <p class="text-xs text-muted-foreground mt-1">
                    {{ t("portfolio.noProjectsSub") }}
                </p>
            </div>
        </div>
    </div>
</template>
