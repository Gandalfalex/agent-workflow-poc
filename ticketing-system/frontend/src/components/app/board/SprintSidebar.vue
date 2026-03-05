<script setup lang="ts">
import { computed, ref } from "vue";
import { useI18n } from "@/lib/i18n";
import { Modal } from "@/components/ui/modal";
import type { Sprint, SprintCompleteRequest } from "@/lib/api";

const props = defineProps<{
    sprints: Sprint[];
    loading: boolean;
    projectId: string;
    ticketKeyMap?: Record<string, string>;
}>();

const emit = defineEmits<{
    (e: "start-sprint", sprintId: string): void;
    (e: "complete-sprint", sprintId: string, payload?: SprintCompleteRequest): void;
}>();

const { t } = useI18n();

const showCompleteDialog = ref(false);
const completingSprintId = ref("");
const selectedMoveTicketIds = ref<string[]>([]);
const incompleteTicketIds = ref<string[]>([]);

const activeSprints = computed(() =>
    props.sprints.filter((s) => s.status === "active"),
);
const plannedSprints = computed(() =>
    props.sprints.filter((s) => s.status === "planned"),
);
const completedSprints = computed(() =>
    props.sprints.filter((s) => s.status === "completed"),
);

const showCompleted = ref(false);

const activeSprint = computed(() => activeSprints.value[0] ?? null);
const nextPlannedSprint = computed(() => plannedSprints.value[0] ?? null);

const activeProgress = computed(() => {
    const s = activeSprint.value;
    if (!s || s.committedTickets === 0) return 0;
    const total = s.committedTickets;
    const incomplete = (s.ticketIds ?? []).length;
    const done = total - incomplete;
    return Math.round((done / total) * 100);
});

const openCompleteDialog = (sprintId: string) => {
    completingSprintId.value = sprintId;
    const sprint = props.sprints.find((s) => s.id === sprintId);
    incompleteTicketIds.value = sprint?.ticketIds ?? [];
    selectedMoveTicketIds.value = [...incompleteTicketIds.value];
    showCompleteDialog.value = true;
};

const closeCompleteDialog = () => {
    showCompleteDialog.value = false;
    completingSprintId.value = "";
    selectedMoveTicketIds.value = [];
    incompleteTicketIds.value = [];
};

const toggleTicket = (ticketId: string) => {
    const idx = selectedMoveTicketIds.value.indexOf(ticketId);
    if (idx >= 0) {
        selectedMoveTicketIds.value.splice(idx, 1);
    } else {
        selectedMoveTicketIds.value.push(ticketId);
    }
};

const selectAll = () => {
    selectedMoveTicketIds.value = [...incompleteTicketIds.value];
};

const deselectAll = () => {
    selectedMoveTicketIds.value = [];
};

const confirmComplete = () => {
    const payload: SprintCompleteRequest | undefined =
        selectedMoveTicketIds.value.length > 0
            ? { moveTicketIds: selectedMoveTicketIds.value }
            : undefined;
    emit("complete-sprint", completingSprintId.value, payload);
    closeCompleteDialog();
};

const formatDate = (dateStr: string) => {
    try {
        return new Date(dateStr).toLocaleDateString();
    } catch {
        return dateStr;
    }
};
</script>

<template>
    <aside
        data-testid="board.sprint-sidebar"
        class="w-72 shrink-0 border-l border-border bg-secondary/30 p-4 overflow-y-auto"
    >
        <h3 class="text-sm font-semibold mb-3">{{ t("sprint.sidebar.title") }}</h3>

        <div v-if="loading" class="text-xs text-muted-foreground">
            {{ t("common.loading") }}
        </div>

        <!-- Active Sprint -->
        <section data-testid="sprint.active-section" class="mb-4" v-if="activeSprint">
            <h4 class="text-xs font-medium text-muted-foreground uppercase mb-2">
                {{ t("sprint.sidebar.active") }}
            </h4>
            <div class="rounded-lg border border-border bg-background p-3">
                <div class="font-medium text-sm">{{ activeSprint.name }}</div>
                <div class="text-xs text-muted-foreground mt-1">
                    {{ formatDate(activeSprint.startDate) }} – {{ formatDate(activeSprint.endDate) }}
                </div>
                <div class="mt-2">
                    <div class="flex justify-between text-xs mb-1">
                        <span>{{ t("sprint.sidebar.progress") }}</span>
                        <span>{{ activeSprint.committedTickets }} {{ t("sprint.sidebar.tickets") }}</span>
                    </div>
                    <div class="w-full h-2 bg-muted rounded-full overflow-hidden">
                        <div
                            class="h-full bg-primary rounded-full transition-all"
                            :style="{ width: activeProgress + '%' }"
                        />
                    </div>
                </div>
                <button
                    data-testid="sprint.complete-button"
                    class="mt-3 w-full rounded-md bg-primary px-3 py-1.5 text-xs font-medium text-primary-foreground hover:bg-primary/90"
                    @click="openCompleteDialog(activeSprint.id)"
                >
                    {{ t("sprint.sidebar.completeSprint") }}
                </button>
            </div>
        </section>

        <!-- Planned Sprints -->
        <section data-testid="sprint.planned-section" class="mb-4" v-if="plannedSprints.length > 0">
            <h4 class="text-xs font-medium text-muted-foreground uppercase mb-2">
                {{ t("sprint.sidebar.planned") }}
            </h4>
            <div
                v-for="sprint in plannedSprints"
                :key="sprint.id"
                class="rounded-lg border border-border bg-background p-3 mb-2"
            >
                <div class="font-medium text-sm">{{ sprint.name }}</div>
                <div class="text-xs text-muted-foreground mt-1">
                    {{ formatDate(sprint.startDate) }} – {{ formatDate(sprint.endDate) }}
                    · {{ sprint.committedTickets }} {{ t("sprint.sidebar.tickets") }}
                </div>
                <button
                    data-testid="sprint.start-button"
                    class="mt-2 w-full rounded-md border border-border px-3 py-1.5 text-xs font-medium hover:bg-muted"
                    :disabled="!!activeSprint"
                    @click="emit('start-sprint', sprint.id)"
                >
                    {{ t("sprint.sidebar.startSprint") }}
                </button>
            </div>
        </section>

        <!-- Completed Sprints -->
        <section data-testid="sprint.completed-section" v-if="completedSprints.length > 0">
            <button
                class="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground mb-2"
                @click="showCompleted = !showCompleted"
            >
                <span>{{ showCompleted ? "▼" : "▶" }}</span>
                {{ t("sprint.sidebar.completed") }} ({{ completedSprints.length }})
            </button>
            <template v-if="showCompleted">
                <div
                    v-for="sprint in completedSprints"
                    :key="sprint.id"
                    class="rounded-lg border border-border bg-background/50 p-3 mb-2 opacity-60"
                >
                    <div class="font-medium text-sm">{{ sprint.name }}</div>
                    <div class="text-xs text-muted-foreground mt-1">
                        {{ formatDate(sprint.startDate) }} – {{ formatDate(sprint.endDate) }}
                        · {{ sprint.committedTickets }} {{ t("sprint.sidebar.tickets") }}
                    </div>
                </div>
            </template>
        </section>

        <div
            v-if="!loading && sprints.length === 0"
            class="text-xs text-muted-foreground"
        >
            {{ t("sprint.sidebar.noSprints") }}
        </div>

        <!-- Complete Sprint Dialog -->
        <Modal :show="showCompleteDialog" test-id="sprint.complete-dialog" max-width="max-w-sm" z-index="z-50" @close="closeCompleteDialog">
                <div>
                    <h3 class="font-semibold mb-3">{{ t("sprint.complete.title") }}</h3>

                    <div v-if="incompleteTicketIds.length > 0" class="mb-3">
                        <p class="text-sm mb-2">{{ t("sprint.complete.movePrompt") }}</p>
                        <div v-if="nextPlannedSprint" class="text-xs text-muted-foreground mb-2">
                            {{ t("sprint.complete.nextSprint") }}: {{ nextPlannedSprint.name }}
                        </div>
                        <div v-else class="text-xs text-yellow-600 mb-2">
                            {{ t("sprint.complete.noNextSprint") }}
                        </div>

                        <div class="flex gap-2 mb-2">
                            <button
                                data-testid="sprint.select-all-button"
                                class="text-xs underline"
                                @click="selectAll"
                            >
                                {{ t("sprint.complete.selectAll") }}
                            </button>
                            <button
                                data-testid="sprint.deselect-all-button"
                                class="text-xs underline"
                                @click="deselectAll"
                            >
                                {{ t("sprint.complete.deselectAll") }}
                            </button>
                        </div>

                        <div class="max-h-48 overflow-y-auto space-y-1">
                            <label
                                v-for="ticketId in incompleteTicketIds"
                                :key="ticketId"
                                data-testid="sprint.complete-ticket-checkbox"
                                class="flex items-center gap-2 text-sm cursor-pointer rounded px-2 py-1 hover:bg-muted"
                            >
                                <input
                                    type="checkbox"
                                    :checked="selectedMoveTicketIds.includes(ticketId)"
                                    @change="toggleTicket(ticketId)"
                                />
                                <span class="truncate text-xs">{{ props.ticketKeyMap?.[ticketId] ?? ticketId.slice(0, 8) + '…' }}</span>
                            </label>
                        </div>
                    </div>

                    <div v-else class="text-sm text-muted-foreground mb-3">
                        {{ t("sprint.complete.allDone") }}
                    </div>

                    <div class="flex justify-end gap-2">
                        <button
                            data-testid="sprint.complete-cancel-button"
                            class="rounded-md border border-border px-3 py-1.5 text-xs hover:bg-muted"
                            @click="closeCompleteDialog"
                        >
                            {{ t("common.cancel") }}
                        </button>
                        <button
                            data-testid="sprint.complete-confirm-button"
                            class="rounded-md bg-primary px-3 py-1.5 text-xs font-medium text-primary-foreground hover:bg-primary/90"
                            @click="confirmComplete"
                        >
                            {{ t("sprint.sidebar.completeSprint") }}
                        </button>
                    </div>
                </div>
        </Modal>
    </aside>
</template>
