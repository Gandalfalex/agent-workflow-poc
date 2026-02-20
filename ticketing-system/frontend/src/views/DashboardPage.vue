<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useBoardStore } from "@/stores/board";
import { useSessionStore } from "@/stores/session";

const props = defineProps<{ projectId: string }>();

const boardStore = useBoardStore();
const sessionStore = useSessionStore();

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

import type { ProjectActivity } from "@/lib/api";

const activityLabel = (a: ProjectActivity): string => {
    switch (a.action) {
        case "state_changed":
            return `changed state to ${a.newValue}`;
        case "priority_changed":
            return `changed priority to ${a.newValue}`;
        case "assignee_changed":
            return a.newValue ? `assigned to ${a.newValue}` : "removed assignee";
        case "type_changed":
            return `changed type to ${a.newValue}`;
        case "title_changed":
            return "renamed ticket";
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

const loadStats = async () => {
    if (!props.projectId) return;
    try {
        await Promise.all([
            boardStore.loadDashboardStats(props.projectId),
            boardStore.loadProjectActivities(props.projectId),
            boardStore.loadDependencyGraph(props.projectId, { depth: 2 }),
            boardStore.loadSprints(props.projectId),
            boardStore.loadCapacitySettings(props.projectId),
        ]);
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
        <div class="grid grid-cols-4 gap-4">
            <div
                v-for="i in 3"
                :key="i"
                class="h-28 rounded-2xl border border-border bg-card/40"
            ></div>
        </div>
        <div
            class="h-56 rounded-2xl border border-border bg-card/40"
        ></div>
        <div class="grid grid-cols-2 gap-4">
            <div
                v-for="i in 2"
                :key="i"
                class="h-44 rounded-2xl border border-border bg-card/40"
            ></div>
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
            No tickets yet
        </p>
        <p class="text-xs text-muted-foreground/60">
            Create tickets on the board to see project statistics here.
        </p>
    </div>

    <!-- Dashboard content -->
    <div v-else-if="stats" class="flex flex-col gap-5">
        <!-- Summary cards -->
        <div class="grid grid-cols-3 gap-4">
            <div
                class="rounded-2xl border border-border bg-card/70 px-5 py-5"
            >
                <p
                    class="text-[10px] uppercase tracking-[0.25em] text-muted-foreground"
                >
                    Total
                </p>
                <p class="mt-2 text-3xl font-bold tabular-nums">
                    {{ totalTickets }}
                </p>
            </div>
            <div
                class="rounded-2xl border border-border bg-card/70 px-5 py-5"
            >
                <p
                    class="text-[10px] uppercase tracking-[0.25em] text-muted-foreground"
                >
                    Open
                </p>
                <p class="mt-2 text-3xl font-bold tabular-nums text-blue-400">
                    {{ stats.totalOpen }}
                </p>
            </div>
            <div
                class="rounded-2xl border border-border bg-card/70 px-5 py-5"
            >
                <p
                    class="text-[10px] uppercase tracking-[0.25em] text-muted-foreground"
                >
                    Closed
                </p>
                <p
                    class="mt-2 text-3xl font-bold tabular-nums text-emerald-400"
                >
                    {{ stats.totalClosed }}
                </p>
            </div>
            <div
                class="rounded-2xl border border-border bg-card/70 px-5 py-5"
            >
                <p
                    class="text-[10px] uppercase tracking-[0.25em] text-muted-foreground"
                >
                    Blocked Open
                </p>
                <p class="mt-2 text-3xl font-bold tabular-nums text-rose-300">
                    {{ stats.blockedOpen }}
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
                By State
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
            <p class="text-[10px] uppercase tracking-[0.25em] text-muted-foreground">
                Dependency Graph
            </p>
            <div v-if="dependencyGraphLoading" class="mt-4 text-xs text-muted-foreground">
                Loading dependency graph...
            </div>
            <div v-else class="mt-4 space-y-2">
                <p class="text-xs text-muted-foreground">
                    {{ dependencyGraph.nodes.length }} nodes · {{ dependencyGraph.edges.length }} edges (up to 2-hop expansion)
                </p>
                <div
                    v-if="dependencyGraph.edges.length === 0"
                    class="text-xs text-muted-foreground"
                >
                    No dependencies yet.
                </div>
                <div v-else class="max-h-56 space-y-1 overflow-y-auto">
                    <div
                        v-for="edge in dependencyGraph.edges"
                        :key="edge.id"
                        class="rounded-lg border border-border/70 bg-background/70 px-3 py-2 text-xs"
                    >
                        <span class="font-medium text-foreground">
                            {{
                                dependencyGraph.nodes.find((n) => n.ticket.id === edge.sourceTicketId)?.ticket.key ||
                                edge.sourceTicketId
                            }}
                        </span>
                        <span class="mx-2 text-muted-foreground">{{ edge.relationType }}</span>
                        <span class="font-medium text-foreground">
                            {{
                                dependencyGraph.nodes.find((n) => n.ticket.id === edge.targetTicketId)?.ticket.key ||
                                edge.targetTicketId
                            }}
                        </span>
                    </div>
                </div>
            </div>
        </section>

        <!-- Priority + Type row -->
        <div class="grid grid-cols-2 gap-4">
            <!-- By Priority -->
            <section
                class="rounded-2xl border border-border bg-card/70 px-5 py-5"
            >
                <p
                    class="text-[10px] uppercase tracking-[0.25em] text-muted-foreground"
                >
                    By Priority
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
                    By Type
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
                By Assignee
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
                Recent Activity
            </p>
            <div v-if="activitiesLoading" class="mt-4 space-y-3">
                <div v-for="i in 5" :key="i" class="h-4 rounded bg-muted animate-pulse" />
            </div>
            <div v-else-if="activities.length === 0" class="mt-4 text-xs text-muted-foreground">
                No activity yet.
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
                        <span class="text-muted-foreground"> {{ activityLabel(activity) }} on </span>
                        <span class="font-medium text-foreground">{{ activity.ticketKey }}</span>
                        <span class="text-muted-foreground truncate"> · {{ activity.ticketTitle }}</span>
                    </div>
                    <span class="shrink-0 text-muted-foreground/60">
                        {{ new Date(activity.createdAt).toLocaleString() }}
                    </span>
                </div>
            </div>
        </section>

        <section data-testid="dashboard.sprint-forecast" class="rounded-2xl border border-border bg-card/70 px-5 py-5">
            <div class="flex flex-wrap items-end gap-3">
                <div class="min-w-52">
                    <p class="text-[10px] uppercase tracking-[0.25em] text-muted-foreground">
                        Sprint Forecast
                    </p>
                    <select
                        data-testid="dashboard.sprint-select"
                        class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                        :disabled="sprintsLoading"
                        v-model="selectedSprintId"
                    >
                        <option value="">Latest sprint</option>
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
                    <label class="text-[10px] uppercase tracking-[0.18em] text-muted-foreground">Iterations</label>
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
                    Recalculate
                </button>
            </div>
            <div v-if="sprintForecastLoading" class="mt-4 text-xs text-muted-foreground">
                Running forecast simulation...
            </div>
            <div v-else-if="sprintForecast" class="mt-4 grid grid-cols-2 gap-3 md:grid-cols-4">
                <div data-testid="dashboard.sprint-committed" class="rounded-lg border border-border/70 bg-background/60 px-3 py-3">
                    <p class="text-[10px] uppercase tracking-[0.2em] text-muted-foreground">Committed</p>
                    <p class="mt-1 text-xl font-bold">{{ sprintForecast.committedTickets }}</p>
                </div>
                <div data-testid="dashboard.sprint-projected" class="rounded-lg border border-border/70 bg-background/60 px-3 py-3">
                    <p class="text-[10px] uppercase tracking-[0.2em] text-muted-foreground">Projected</p>
                    <p class="mt-1 text-xl font-bold">{{ sprintForecast.projectedCompletion }}</p>
                </div>
                <div data-testid="dashboard.sprint-capacity" class="rounded-lg border border-border/70 bg-background/60 px-3 py-3">
                    <p class="text-[10px] uppercase tracking-[0.2em] text-muted-foreground">Capacity</p>
                    <p class="mt-1 text-xl font-bold">{{ sprintForecast.capacity }}</p>
                </div>
                <div data-testid="dashboard.sprint-delta" class="rounded-lg border border-border/70 bg-background/60 px-3 py-3">
                    <p class="text-[10px] uppercase tracking-[0.2em] text-muted-foreground">Over Capacity</p>
                    <p class="mt-1 text-xl font-bold">{{ sprintForecast.overCapacityDelta }}</p>
                </div>
                <div class="col-span-2 md:col-span-4 text-xs text-muted-foreground">
                    Confidence: {{ Math.round(sprintForecast.confidence * 100) }}% · Iterations: {{ sprintForecast.iterations }} · Capacity entries: {{ capacitySettings.length }}
                </div>
            </div>
            <div v-else class="mt-4 text-xs text-muted-foreground">
                No sprint forecast available yet.
            </div>
        </section>
    </div>
</template>
