<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useBoardStore } from "@/stores/board";
import { useSessionStore } from "@/stores/session";
import { useI18n } from "@/lib/i18n";

const props = defineProps<{ projectId: string }>();

const boardStore = useBoardStore();
const sessionStore = useSessionStore();
const { t } = useI18n();

const stats = computed(() => boardStore.dashboardStats);
const loading = computed(() => boardStore.dashboardLoading);
const activities = computed(() => boardStore.projectActivities);
const activitiesLoading = computed(() => boardStore.projectActivitiesLoading);
const dependencyGraph = computed(() => boardStore.ticketDependencyGraph);
const dependencyGraphLoading = computed(
    () => boardStore.ticketDependencyGraphLoading,
);
const sprints = computed(() => boardStore.sprints);
const sprintsLoading = computed(() => boardStore.sprintsLoading);
const capacitySettings = computed(() => boardStore.capacitySettings);
const sprintForecast = computed(() => boardStore.sprintForecastSummary);
const sprintForecastLoading = computed(() => boardStore.sprintForecastLoading);
const selectedSprintId = ref<string>("");
const forecastIterations = ref<number>(250);

const totalTickets = computed(
    () => (stats.value?.totalOpen ?? 0) + (stats.value?.totalClosed ?? 0),
);
const isEmpty = computed(
    () => !loading.value && stats.value && totalTickets.value === 0,
);

const maxByState = computed(() =>
    Math.max(1, ...(stats.value?.byState?.map((s) => s.value) ?? [1])),
);
const maxByPriority = computed(() =>
    Math.max(1, ...(stats.value?.byPriority?.map((s) => s.value) ?? [1])),
);
const maxByType = computed(() =>
    Math.max(1, ...(stats.value?.byType?.map((s) => s.value) ?? [1])),
);
const maxByAssignee = computed(() =>
    Math.max(1, ...(stats.value?.byAssignee?.map((s) => s.value) ?? [1])),
);

const priorityColor: Record<string, string> = {
    urgent: "bg-red-500",
    high: "bg-orange-400",
    medium: "bg-blue-400",
    low: "bg-zinc-500",
};
const priorityTrack: Record<string, string> = {
    urgent: "bg-red-500/15",
    high: "bg-orange-400/15",
    medium: "bg-blue-400/15",
    low: "bg-zinc-500/15",
};

const typeColor: Record<string, string> = {
    feature: "bg-emerald-400",
    bug: "bg-rose-400",
};
const typeTrack: Record<string, string> = {
    feature: "bg-emerald-400/15",
    bug: "bg-rose-400/15",
};

import type { ProjectActivity, FlowHealthStats, Release } from "@/lib/api";
import { getFlowHealth, listReleases } from "@/lib/api";

const relativeTime = (dateStr: string): string => {
    const diffMs = Date.now() - new Date(dateStr).getTime();
    const diffSec = Math.floor(diffMs / 1000);
    const diffMin = Math.floor(diffSec / 60);
    const diffHour = Math.floor(diffMin / 60);
    const diffDay = Math.floor(diffHour / 24);
    if (diffSec < 60) return "just now";
    if (diffMin < 60) return `${diffMin}m ago`;
    if (diffHour < 24) return `${diffHour}h ago`;
    if (diffDay === 1) return "yesterday";
    if (diffDay < 7) return `${diffDay}d ago`;
    return new Date(dateStr).toLocaleDateString();
};

const activityLabel = (a: ProjectActivity): string => {
    switch (a.action) {
        case "state_changed":
            return t("dashboard.activity.stateChanged", { value: a.newValue || "-" });
        case "priority_changed":
            return t("dashboard.activity.priorityChanged", { value: a.newValue || "-" });
        case "assignee_changed":
            return a.newValue
                ? t("dashboard.activity.assignedTo", { value: a.newValue })
                : t("dashboard.activity.removedAssignee");
        case "type_changed":
            return t("dashboard.activity.typeChanged", { value: a.newValue || "-" });
        case "title_changed":
            return t("dashboard.activity.renamed");
        default:
            return a.action;
    }
};

const handleAuthError = (err: unknown) => {
    const error = err as Error & { status?: number };
    if (error.status === 401 || error.status === 403) {
        sessionStore.reset();
        return true;
    }
    return false;
};

const flowHealth = ref<FlowHealthStats | null>(null);
const flowHealthLoading = ref(false);
const releases = ref<Release[]>([]);

const maxThroughput = computed(() =>
    Math.max(1, ...(flowHealth.value?.throughput?.map((d) => d.count) ?? [1])),
);

const loadFlowHealth = async () => {
    if (!props.projectId) return;
    flowHealthLoading.value = true;
    try {
        flowHealth.value = await getFlowHealth(props.projectId);
    } catch {
        // non-critical; silent fail
    } finally {
        flowHealthLoading.value = false;
    }
};

const loadStats = async () => {
    if (!props.projectId) return;
    try {
        const [relRes] = await Promise.all([
            listReleases(props.projectId).catch(() => ({ items: [] as Release[] })),
            boardStore.loadDashboardStats(props.projectId),
            boardStore.loadProjectActivities(props.projectId),
            boardStore.loadDependencyGraph(props.projectId, { depth: 2 }),
            boardStore.loadSprints(props.projectId),
            boardStore.loadCapacitySettings(props.projectId),
            loadFlowHealth(),
        ]);
        releases.value = relRes.items;
        if (!selectedSprintId.value && boardStore.sprints.length > 0) {
            selectedSprintId.value = boardStore.sprints[0]!.id;
        }
        await boardStore.loadSprintForecast(props.projectId, {
            sprintId: selectedSprintId.value || undefined,
            iterations: forecastIterations.value,
        });
    } catch (err) {
        handleAuthError(err);
    }
};

const reloadForecast = async () => {
    if (!props.projectId) return;
    await boardStore.loadSprintForecast(props.projectId, {
        sprintId: selectedSprintId.value || undefined,
        iterations: forecastIterations.value,
    });
};

onMounted(loadStats);
watch(() => props.projectId, loadStats);
watch(selectedSprintId, reloadForecast);
</script>

<template>
    <!-- Loading skeleton -->
    <div v-if="loading" class="flex flex-col gap-5 animate-pulse">
        <!-- Summary cards skeleton -->
        <div class="grid grid-cols-2 gap-4 sm:grid-cols-4">
            <div
                v-for="i in 4"
                :key="i"
                class="h-28 rounded-2xl border border-border bg-card/40 p-5"
            >
                <div class="h-3 bg-muted rounded w-16 mb-4"></div>
                <div class="h-8 bg-muted rounded w-20"></div>
            </div>
        </div>

        <!-- State bars skeleton -->
        <div class="rounded-2xl border border-border bg-card/40 p-5">
            <div class="h-3 bg-muted rounded w-32 mb-6"></div>
            <div class="space-y-3">
                <div v-for="i in 5" :key="i" class="flex items-center gap-3">
                    <div class="h-4 bg-muted rounded w-24"></div>
                    <div class="flex-1 h-5 bg-muted rounded"></div>
                    <div class="h-4 bg-muted rounded w-8"></div>
                </div>
            </div>
        </div>

        <!-- Priority + Type skeleton -->
        <div class="grid grid-cols-2 gap-4">
            <div
                v-for="i in 2"
                :key="i"
                class="rounded-2xl border border-border bg-card/40 p-5"
            >
                <div class="h-3 bg-muted rounded w-24 mb-6"></div>
                <div class="space-y-3">
                    <div v-for="j in 4" :key="j" class="flex items-center gap-3">
                        <div class="h-4 bg-muted rounded w-16"></div>
                        <div class="flex-1 h-5 bg-muted rounded"></div>
                        <div class="h-4 bg-muted rounded w-6"></div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Activity feed skeleton -->
        <div class="rounded-2xl border border-border bg-card/40 p-5">
            <div class="h-3 bg-muted rounded w-28 mb-6"></div>
            <div class="space-y-3">
                <div v-for="i in 5" :key="i" class="flex items-start gap-3">
                    <div class="h-5 w-5 bg-muted rounded-full"></div>
                    <div class="flex-1 h-4 bg-muted rounded"></div>
                    <div class="h-4 bg-muted rounded w-20"></div>
                </div>
            </div>
        </div>
    </div>

    <!-- Empty state -->
    <div
        v-else-if="isEmpty"
        class="flex flex-col items-center justify-center gap-3 rounded-2xl border border-border bg-card/70 px-6 py-20"
    >
        <div
            class="flex h-12 w-12 items-center justify-center rounded-2xl bg-muted text-xl text-muted-foreground"
        >
            &empty;
        </div>
        <p class="text-sm font-medium text-muted-foreground">
            {{ t("dashboard.empty.title") }}
        </p>
        <p class="text-xs text-muted-foreground/60">
            {{ t("dashboard.empty.body") }}
        </p>
    </div>

    <!-- Dashboard content -->
    <div v-else-if="stats" class="flex flex-col gap-5">
        <!-- Summary cards -->
        <div class="grid grid-cols-2 gap-4 sm:grid-cols-3">
            <div class="rounded-2xl border border-border border-l-4 border-l-primary bg-card/70 px-5 py-5">
                <p class="text-[10px] uppercase tracking-[0.25em] text-muted-foreground">
                    {{ t("dashboard.summary.total") }}
                </p>
                <p class="mt-2 text-3xl font-bold tabular-nums">
                    {{ totalTickets }}
                </p>
            </div>
            <div class="rounded-2xl border border-border border-l-4 border-l-blue-400 bg-card/70 px-5 py-5">
                <p class="text-[10px] uppercase tracking-[0.25em] text-muted-foreground">
                    {{ t("dashboard.summary.open") }}
                </p>
                <p class="mt-2 text-3xl font-bold tabular-nums text-blue-400">
                    {{ stats.totalOpen }}
                </p>
            </div>
            <div class="rounded-2xl border border-border border-l-4 border-l-emerald-400 bg-card/70 px-5 py-5">
                <p class="text-[10px] uppercase tracking-[0.25em] text-muted-foreground">
                    {{ t("dashboard.summary.closed") }}
                </p>
                <p class="mt-2 text-3xl font-bold tabular-nums text-emerald-400">
                    {{ stats.totalClosed }}
                </p>
            </div>
            <div class="rounded-2xl border border-border border-l-4 border-l-rose-400 bg-rose-400/5 px-5 py-5">
                <p class="text-[10px] uppercase tracking-[0.25em] text-muted-foreground">
                    {{ t("dashboard.summary.blockedOpen") }}
                </p>
                <p class="mt-2 text-3xl font-bold tabular-nums text-rose-300">
                    {{ stats.blockedOpen }}
                </p>
            </div>
            <div
                data-testid="dashboard.sla-breached-card"
                class="rounded-2xl border border-border border-l-4 border-l-red-500 bg-red-500/5 px-5 py-5"
            >
                <p class="text-[10px] uppercase tracking-[0.25em] text-muted-foreground">
                    SLA Breached
                </p>
                <p class="mt-2 text-3xl font-bold tabular-nums text-red-400">
                    {{ stats.slaBreached }}
                </p>
            </div>
            <div
                data-testid="dashboard.sla-at-risk-card"
                class="rounded-2xl border border-border border-l-4 border-l-amber-400 bg-amber-400/5 px-5 py-5"
            >
                <p class="text-[10px] uppercase tracking-[0.25em] text-muted-foreground">
                    SLA At Risk
                </p>
                <p class="mt-2 text-3xl font-bold tabular-nums text-amber-300">
                    {{ stats.slaAtRisk }}
                </p>
            </div>
        </div>

        <!-- By State -->
        <section
            class="rounded-2xl border border-border bg-card/70 px-5 py-5"
        >
            <p
                class="text-[10px] uppercase tracking-[0.25em] text-muted-foreground"
            >
                {{ t("dashboard.byState") }}
            </p>
            <div class="mt-4 flex flex-col gap-3">
                <div
                    v-for="item in stats.byState"
                    :key="item.label"
                    class="flex items-center gap-3"
                >
                    <span
                        class="w-24 shrink-0 truncate text-xs text-muted-foreground"
                        :title="item.label"
                        >{{ item.label }}</span
                    >
                    <div
                        class="relative h-5 flex-1 overflow-hidden rounded-md bg-primary/10"
                    >
                        <div
                            class="absolute inset-y-0 left-0 rounded-md bg-primary/80 transition-all duration-500"
                            :style="{
                                width:
                                    (item.value / maxByState) * 100 + '%',
                            }"
                        ></div>
                    </div>
                    <span
                        class="w-8 shrink-0 text-right text-xs font-semibold tabular-nums"
                        >{{ item.value }}</span
                    >
                </div>
            </div>
        </section>

        <section data-testid="dashboard.dependency-graph" class="rounded-2xl border border-border bg-card/70 px-5 py-5">
            <div class="flex items-center justify-between">
                <p class="text-[10px] uppercase tracking-[0.25em] text-muted-foreground">
                    {{ t("dashboard.dependencyGraph") }}
                </p>
                <span v-if="!dependencyGraphLoading" class="text-[10px] text-muted-foreground/60">
                    {{ dependencyGraph.nodes.length }} nodes · {{ dependencyGraph.edges.length }} edges
                </span>
            </div>
            <div v-if="dependencyGraphLoading" class="mt-4 flex gap-2 animate-pulse">
                <div class="h-8 w-20 rounded-lg bg-muted"></div>
                <div class="h-8 w-16 rounded-lg bg-muted"></div>
                <div class="h-8 w-20 rounded-lg bg-muted"></div>
            </div>
            <div v-else-if="dependencyGraph.edges.length === 0" class="mt-4 text-xs text-muted-foreground">
                {{ t("dashboard.noDependencies") }}
            </div>
            <div v-else class="mt-4 max-h-64 space-y-2 overflow-y-auto">
                <div
                    v-for="edge in dependencyGraph.edges"
                    :key="edge.id"
                    class="flex items-center gap-2 rounded-xl border border-border/60 bg-background/60 px-3 py-2"
                >
                    <span class="rounded-md border border-primary/30 bg-primary/10 px-2 py-0.5 font-mono text-[11px] font-bold text-primary">
                        {{ dependencyGraph.nodes.find((n) => n.ticket.id === edge.sourceTicketId)?.ticket.key || edge.sourceTicketId }}
                    </span>
                    <span class="flex items-center gap-1 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
                        <span class="h-px w-6 bg-border"></span>
                        {{ edge.relationType }}
                        <span class="h-px w-6 bg-border"></span>
                        <svg class="h-3 w-3 shrink-0" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M10.293 3.293a1 1 0 011.414 0l6 6a1 1 0 010 1.414l-6 6a1 1 0 01-1.414-1.414L14.586 11H3a1 1 0 110-2h11.586l-4.293-4.293a1 1 0 010-1.414z" clip-rule="evenodd"/></svg>
                    </span>
                    <span
                        class="rounded-md border px-2 py-0.5 font-mono text-[11px] font-bold"
                        :class="edge.relationType === 'blocks' ? 'border-rose-500/30 bg-rose-500/10 text-rose-300' : 'border-amber-500/30 bg-amber-500/10 text-amber-300'"
                    >
                        {{ dependencyGraph.nodes.find((n) => n.ticket.id === edge.targetTicketId)?.ticket.key || edge.targetTicketId }}
                    </span>
                    <span
                        v-if="edge.relationType === 'blocks'"
                        class="ml-auto rounded-full border border-rose-500/20 bg-rose-500/10 px-1.5 py-0.5 text-[9px] font-bold uppercase tracking-wider text-rose-400"
                    >blocked</span>
                </div>
            </div>
        </section>

        <!-- Priority + Type row -->
        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <!-- By Priority -->
            <section
                class="rounded-2xl border border-border bg-card/70 px-5 py-5"
            >
                <p
                    class="text-[10px] uppercase tracking-[0.25em] text-muted-foreground"
                >
                    {{ t("dashboard.byPriority") }}
                </p>
                <div class="mt-4 flex flex-col gap-3">
                    <div
                        v-for="item in stats.byPriority"
                        :key="item.label"
                        class="flex items-center gap-3"
                    >
                        <span
                            class="w-16 shrink-0 text-xs capitalize text-muted-foreground"
                            >{{ item.label }}</span
                        >
                        <div
                            class="relative h-5 flex-1 overflow-hidden rounded-md"
                            :class="
                                priorityTrack[item.label] ?? 'bg-primary/10'
                            "
                        >
                            <div
                                class="absolute inset-y-0 left-0 rounded-md transition-all duration-500"
                                :class="
                                    priorityColor[item.label] ??
                                    'bg-primary/80'
                                "
                                :style="{
                                    width:
                                        (item.value / maxByPriority) * 100 +
                                        '%',
                                }"
                            ></div>
                        </div>
                        <span
                            class="w-6 shrink-0 text-right text-xs font-semibold tabular-nums"
                            >{{ item.value }}</span
                        >
                    </div>
                </div>
            </section>

            <!-- By Type -->
            <section
                class="rounded-2xl border border-border bg-card/70 px-5 py-5"
            >
                <p
                    class="text-[10px] uppercase tracking-[0.25em] text-muted-foreground"
                >
                    {{ t("dashboard.byType") }}
                </p>
                <div class="mt-4 flex flex-col gap-3">
                    <div
                        v-for="item in stats.byType"
                        :key="item.label"
                        class="flex items-center gap-3"
                    >
                        <span
                            class="w-16 shrink-0 text-xs capitalize text-muted-foreground"
                            >{{ item.label }}</span
                        >
                        <div
                            class="relative h-5 flex-1 overflow-hidden rounded-md"
                            :class="
                                typeTrack[item.label] ?? 'bg-primary/10'
                            "
                        >
                            <div
                                class="absolute inset-y-0 left-0 rounded-md transition-all duration-500"
                                :class="
                                    typeColor[item.label] ?? 'bg-primary/80'
                                "
                                :style="{
                                    width:
                                        (item.value / maxByType) * 100 +
                                        '%',
                                }"
                            ></div>
                        </div>
                        <span
                            class="w-6 shrink-0 text-right text-xs font-semibold tabular-nums"
                            >{{ item.value }}</span
                        >
                    </div>
                </div>
            </section>
        </div>

        <!-- By Assignee -->
        <section
            v-if="stats.byAssignee.length > 0"
            class="rounded-2xl border border-border bg-card/70 px-5 py-5"
        >
            <p
                class="text-[10px] uppercase tracking-[0.25em] text-muted-foreground"
            >
                    {{ t("dashboard.byAssignee") }}
            </p>
            <div class="mt-4 flex flex-col gap-3">
                <div
                    v-for="item in stats.byAssignee"
                    :key="item.label"
                    class="flex items-center gap-3"
                >
                    <div class="flex w-28 shrink-0 items-center gap-2">
                        <div
                            class="flex h-5 w-5 items-center justify-center rounded-full bg-primary/15 text-[8px] font-bold text-primary"
                        >
                            {{
                                item.label === "Unassigned"
                                    ? "?"
                                    : item.label
                                          .slice(0, 2)
                                          .toUpperCase()
                            }}
                        </div>
                        <span
                            class="truncate text-xs text-muted-foreground"
                            :title="item.label"
                            >{{ item.label }}</span
                        >
                    </div>
                    <div
                        class="relative h-5 flex-1 overflow-hidden rounded-md bg-primary/10"
                    >
                        <div
                            class="absolute inset-y-0 left-0 rounded-md bg-primary/60 transition-all duration-500"
                            :style="{
                                width:
                                    (item.value / maxByAssignee) * 100 +
                                    '%',
                            }"
                        ></div>
                    </div>
                    <span
                        class="w-6 shrink-0 text-right text-xs font-semibold tabular-nums"
                        >{{ item.value }}</span
                    >
                </div>
            </div>
        </section>

        <!-- Recent Activity -->
        <section data-testid="dashboard.recent-activity" class="rounded-2xl border border-border bg-card/70 px-5 py-5">
            <p class="text-[10px] uppercase tracking-[0.25em] text-muted-foreground">
                {{ t("dashboard.recentActivity") }}
            </p>
            <div v-if="activitiesLoading" class="mt-4 space-y-3">
                <div v-for="i in 5" :key="i" class="h-4 rounded bg-muted animate-pulse" />
            </div>
            <div v-else-if="activities.length === 0" class="mt-4 text-xs text-muted-foreground">
                {{ t("dashboard.noActivity") }}
            </div>
            <div v-else class="mt-4 flex flex-col gap-3">
                <div
                    v-for="activity in activities"
                    :key="activity.id"
                    class="flex items-start gap-3 text-xs"
                >
                    <div class="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-primary/15 text-[8px] font-bold text-primary">
                        {{ activity.actorName.slice(0, 2).toUpperCase() }}
                    </div>
                    <div class="flex-1 min-w-0">
                        <span class="font-medium text-foreground">{{ activity.actorName }}</span>
                        <span class="text-muted-foreground">
                            {{ activityLabel(activity) }} {{ t("dashboard.activity.on") }}
                        </span>
                        <span class="font-medium text-foreground">{{ activity.ticketKey }}</span>
                        <span class="text-muted-foreground truncate"> · {{ activity.ticketTitle }}</span>
                    </div>
                    <span class="shrink-0 text-muted-foreground/60">
                        {{ relativeTime(activity.createdAt) }}
                    </span>
                </div>
            </div>
        </section>

        <section data-testid="dashboard.sprint-forecast" class="rounded-2xl border border-border bg-card/70 px-5 py-5">
            <div class="flex flex-wrap items-end gap-3">
                <div class="min-w-52">
                    <p class="text-[10px] uppercase tracking-[0.25em] text-muted-foreground">
                        {{ t("dashboard.sprintForecast") }}
                    </p>
                    <select
                        data-testid="dashboard.sprint-select"
                        class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                        :disabled="sprintsLoading"
                        v-model="selectedSprintId"
                    >
                        <option value="">{{ t("dashboard.latestSprint") }}</option>
                        <option
                            v-for="sprint in sprints"
                            :key="sprint.id"
                            :value="sprint.id"
                        >
                            {{ sprint.name }} ({{ sprint.startDate }} - {{ sprint.endDate }})
                        </option>
                    </select>
                </div>
                <div>
                    <label class="text-[10px] uppercase tracking-[0.18em] text-muted-foreground">{{ t("dashboard.iterations") }}</label>
                    <input
                        data-testid="dashboard.sprint-iterations-input"
                        type="number"
                        min="10"
                        max="5000"
                        class="mt-2 w-28 rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                        v-model.number="forecastIterations"
                    />
                </div>
                <button
                    data-testid="dashboard.sprint-forecast-reload"
                    class="rounded-lg border border-border bg-background px-3 py-2 text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground transition hover:border-foreground hover:text-foreground"
                    @click="reloadForecast"
                >
                    {{ t("dashboard.recalculate") }}
                </button>
            </div>
            <div v-if="sprintForecastLoading" class="mt-4 text-xs text-muted-foreground">
                {{ t("dashboard.runningForecast") }}
            </div>
            <div v-else-if="sprintForecast" class="mt-4 grid grid-cols-2 gap-3 md:grid-cols-4">
                <div data-testid="dashboard.sprint-committed" class="rounded-lg border border-border/70 bg-background/60 px-3 py-3">
                    <p class="text-[10px] uppercase tracking-[0.2em] text-muted-foreground">{{ t("dashboard.committed") }}</p>
                    <p class="mt-1 text-xl font-bold">{{ sprintForecast.committedTickets }}</p>
                </div>
                <div data-testid="dashboard.sprint-projected" class="rounded-lg border border-border/70 bg-background/60 px-3 py-3">
                    <p class="text-[10px] uppercase tracking-[0.2em] text-muted-foreground">{{ t("dashboard.projected") }}</p>
                    <p class="mt-1 text-xl font-bold">{{ sprintForecast.projectedCompletion }}</p>
                </div>
                <div data-testid="dashboard.sprint-capacity" class="rounded-lg border border-border/70 bg-background/60 px-3 py-3">
                    <p class="text-[10px] uppercase tracking-[0.2em] text-muted-foreground">{{ t("dashboard.capacity") }}</p>
                    <p class="mt-1 text-xl font-bold">{{ sprintForecast.capacity }}</p>
                </div>
                <div data-testid="dashboard.sprint-delta" class="rounded-lg border border-border/70 bg-background/60 px-3 py-3">
                    <p class="text-[10px] uppercase tracking-[0.2em] text-muted-foreground">{{ t("dashboard.overCapacity") }}</p>
                    <p class="mt-1 text-xl font-bold">{{ sprintForecast.overCapacityDelta }}</p>
                </div>
                <div class="col-span-2 md:col-span-4 text-xs text-muted-foreground">
                    {{
                        t("dashboard.confidenceLine", {
                            confidence: Math.round(sprintForecast.confidence * 100),
                            iterations: sprintForecast.iterations,
                            capacity: capacitySettings.length,
                        })
                    }}
                </div>
            </div>
            <div v-else class="mt-4 text-xs text-muted-foreground">
                {{ t("dashboard.noForecast") }}
            </div>
        </section>

        <!-- Flow Health Panel -->
        <section
            v-if="flowHealth || flowHealthLoading"
            data-testid="dashboard.flow-health"
            class="rounded-2xl border border-border bg-card/80 p-5 shadow-sm"
        >
            <p class="text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
                Flow Health
            </p>

            <div v-if="flowHealthLoading" class="mt-4 text-xs text-muted-foreground animate-pulse">
                Loading...
            </div>

            <div v-else-if="flowHealth" class="mt-4 space-y-6">
                <!-- WIP Stats -->
                <div v-if="flowHealth.wipStats.length > 0">
                    <p class="mb-2 text-xs font-semibold text-foreground">WIP per State</p>
                    <div class="space-y-2">
                        <div
                            v-for="stat in flowHealth.wipStats"
                            :key="stat.stateId"
                            data-testid="dashboard.wip-stat-row"
                            class="flex items-center gap-3"
                        >
                            <span class="w-28 truncate text-xs text-muted-foreground">{{ stat.stateName }}</span>
                            <div class="flex-1 h-5 rounded-full overflow-hidden bg-muted/50">
                                <div
                                    class="h-5 rounded-full transition-all"
                                    :class="
                                        stat.wipLimit && stat.ticketCount >= stat.wipLimit
                                            ? 'bg-destructive/70'
                                            : 'bg-primary/60'
                                    "
                                    :style="{
                                        width: stat.wipLimit
                                            ? Math.min(100, (stat.ticketCount / stat.wipLimit) * 100) + '%'
                                            : '0%',
                                    }"
                                ></div>
                            </div>
                            <span class="w-20 text-right text-xs font-medium">
                                {{ stat.ticketCount }}
                                <span v-if="stat.wipLimit" class="text-muted-foreground">/{{ stat.wipLimit }}</span>
                            </span>
                            <span
                                v-if="stat.avgAgeHours > 0"
                                class="w-20 text-right text-xs text-muted-foreground"
                            >avg {{ stat.avgAgeHours.toFixed(1) }}h</span>
                        </div>
                    </div>
                </div>

                <!-- Cycle Time -->
                <div v-if="flowHealth.cycleTime.sampleCount > 0">
                    <p class="mb-2 text-xs font-semibold text-foreground">Cycle Time (last 30 days)</p>
                    <div class="grid grid-cols-3 gap-3">
                        <div class="rounded-lg border border-border/70 bg-background/60 px-3 py-2 text-center">
                            <p class="text-[10px] uppercase tracking-[0.15em] text-muted-foreground">P50</p>
                            <p class="mt-1 text-lg font-bold">{{ flowHealth.cycleTime.p50Hours.toFixed(1) }}h</p>
                        </div>
                        <div class="rounded-lg border border-border/70 bg-background/60 px-3 py-2 text-center">
                            <p class="text-[10px] uppercase tracking-[0.15em] text-muted-foreground">P75</p>
                            <p class="mt-1 text-lg font-bold">{{ flowHealth.cycleTime.p75Hours.toFixed(1) }}h</p>
                        </div>
                        <div class="rounded-lg border border-border/70 bg-background/60 px-3 py-2 text-center">
                            <p class="text-[10px] uppercase tracking-[0.15em] text-muted-foreground">P95</p>
                            <p class="mt-1 text-lg font-bold">{{ flowHealth.cycleTime.p95Hours.toFixed(1) }}h</p>
                        </div>
                    </div>
                    <p class="mt-1 text-[10px] text-muted-foreground">{{ flowHealth.cycleTime.sampleCount }} closed tickets</p>
                </div>

                <!-- Throughput -->
                <div v-if="flowHealth.throughput.length > 0">
                    <p class="mb-2 text-xs font-semibold text-foreground">Daily Throughput (last 14 days)</p>
                    <div class="flex items-end gap-1 h-16">
                        <div
                            v-for="day in flowHealth.throughput"
                            :key="day.day"
                            class="flex-1 rounded-t bg-primary/50 hover:bg-primary/70 transition-colors"
                            :style="{ height: Math.round((day.count / maxThroughput) * 100) + '%' }"
                            :title="`${day.day}: ${day.count} tickets`"
                        ></div>
                    </div>
                </div>
            </div>
        </section>

        <!-- Releases Panel -->
        <section
            v-if="releases.length > 0"
            data-testid="dashboard.releases_section"
            class="rounded-2xl border border-border bg-card/80 p-5 shadow-sm"
        >
            <p class="text-[10px] uppercase tracking-[0.2em] font-semibold text-muted-foreground mb-4">
                Releases
            </p>
            <div class="space-y-3">
                <div
                    v-for="rel in releases.filter(r => r.status !== 'archived')"
                    :key="rel.id"
                    class="flex items-center gap-3"
                >
                    <div class="flex-1 min-w-0">
                        <div class="flex items-center gap-2 mb-1">
                            <span class="text-sm font-medium truncate">{{ rel.name }}</span>
                            <span v-if="rel.version" class="text-xs text-muted-foreground">{{ rel.version }}</span>
                            <span
                                class="rounded-full px-2 py-0.5 text-[10px] font-semibold border"
                                :class="{
                                    'bg-blue-500/15 text-blue-300 border-blue-500/25': rel.status === 'planned',
                                    'bg-amber-500/15 text-amber-300 border-amber-500/25': rel.status === 'in_progress',
                                    'bg-emerald-500/15 text-emerald-300 border-emerald-500/25': rel.status === 'released',
                                }"
                            >{{ rel.status }}</span>
                        </div>
                        <div class="h-1.5 rounded-full bg-muted overflow-hidden">
                            <div
                                class="h-full rounded-full transition-all"
                                :class="rel.status === 'released' ? 'bg-emerald-500' : 'bg-primary'"
                                :style="{ width: rel.ticketCount > 0 ? Math.round((rel.closedTicketCount / rel.ticketCount) * 100) + '%' : (rel.status === 'released' ? '100%' : '0%') }"
                            ></div>
                        </div>
                    </div>
                    <span class="text-xs text-muted-foreground tabular-nums shrink-0">
                        {{ rel.closedTicketCount }}/{{ rel.ticketCount }}
                    </span>
                </div>
            </div>
        </section>
    </div>
</template>
