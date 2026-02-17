<script setup lang="ts">
import { computed, onMounted, watch } from "vue";
import { useBoardStore } from "@/stores/board";
import { useSessionStore } from "@/stores/session";

const props = defineProps<{ projectId: string }>();

const boardStore = useBoardStore();
const sessionStore = useSessionStore();

const stats = computed(() => boardStore.dashboardStats);
const loading = computed(() => boardStore.dashboardLoading);

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
        await boardStore.loadDashboardStats(props.projectId);
    } catch (err) {
        handleAuthError(err);
    }
};

onMounted(loadStats);
watch(() => props.projectId, loadStats);
</script>

<template>
    <!-- Loading skeleton -->
    <div v-if="loading" class="flex flex-col gap-5 animate-pulse">
        <div class="grid grid-cols-3 gap-4">
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
    </div>
</template>
