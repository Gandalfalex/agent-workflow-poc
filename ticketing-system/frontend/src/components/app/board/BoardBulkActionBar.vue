<script setup lang="ts">
import { useI18n } from "@/lib/i18n";
import type { BulkTicketAction, TicketPriority, WorkflowState } from "@/lib/api";

const props = defineProps<{
    selectedCount: number;
    bulkAction: BulkTicketAction;
    bulkStateId: string;
    bulkAssigneeId: string;
    bulkPriority: TicketPriority;
    bulkBusy: boolean;
    bulkMessage: string;
    states: WorkflowState[];
    priorities: TicketPriority[];
    assigneeOptions: Array<{ id: string; name: string }>;
}>();

const emit = defineEmits<{
    (e: "update:bulkAction", value: BulkTicketAction): void;
    (e: "update:bulkStateId", value: string): void;
    (e: "update:bulkAssigneeId", value: string): void;
    (e: "update:bulkPriority", value: TicketPriority): void;
    (e: "apply"): void;
}>();

const { t } = useI18n();
</script>

<template>
    <section
        class="fixed inset-x-4 bottom-4 z-50 rounded-2xl border border-border bg-card/95 p-3 shadow-2xl backdrop-blur"
    >
        <div class="flex flex-wrap items-center gap-2">
            <span
                data-testid="board.bulk-message"
                class="rounded-full bg-primary/15 px-3 py-1 text-xs font-semibold text-primary"
            >
                {{ t("board.view.selectedCount", { count: props.selectedCount }) }}
            </span>
            <select
                :value="props.bulkAction"
                data-testid="board.bulk-action-select"
                class="rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                @change="
                    emit(
                        'update:bulkAction',
                        ($event.target as HTMLSelectElement).value as BulkTicketAction,
                    )
                "
            >
                <option value="move_state">{{ t("board.bulk.moveState") }}</option>
                <option value="assign">{{ t("board.bulk.assignUser") }}</option>
                <option value="set_priority">{{ t("board.bulk.setPriority") }}</option>
                <option value="delete">{{ t("board.bulk.deleteTickets") }}</option>
            </select>
            <select
                v-if="props.bulkAction === 'move_state'"
                :value="props.bulkStateId"
                data-testid="board.bulk-state-select"
                class="rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                @change="
                    emit(
                        'update:bulkStateId',
                        ($event.target as HTMLSelectElement).value,
                    )
                "
            >
                <option value="">{{ t("board.bulk.selectState") }}</option>
                <option v-for="state in props.states" :key="state.id" :value="state.id">
                    {{ state.name }}
                </option>
            </select>
            <select
                v-else-if="props.bulkAction === 'assign'"
                :value="props.bulkAssigneeId"
                data-testid="board.bulk-assignee-select"
                class="rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                @change="
                    emit(
                        'update:bulkAssigneeId',
                        ($event.target as HTMLSelectElement).value,
                    )
                "
            >
                <option value="">{{ t("board.bulk.selectAssignee") }}</option>
                <option
                    v-for="assignee in props.assigneeOptions"
                    :key="assignee.id"
                    :value="assignee.id"
                >
                    {{ assignee.name }}
                </option>
            </select>
            <select
                v-else-if="props.bulkAction === 'set_priority'"
                :value="props.bulkPriority"
                data-testid="board.bulk-priority-select"
                class="rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                @change="
                    emit(
                        'update:bulkPriority',
                        ($event.target as HTMLSelectElement).value as TicketPriority,
                    )
                "
            >
                <option
                    v-for="priority in props.priorities"
                    :key="priority"
                    :value="priority"
                >
                    {{ priority }}
                </option>
            </select>
            <div
                v-else
                class="rounded-xl border border-border bg-background px-3 py-2 text-sm text-muted-foreground"
            >
                {{ t("board.bulk.deleteHint") }}
            </div>
            <button
                data-testid="board.bulk-apply-button"
                class="rounded-lg border border-border bg-background px-3 py-2 text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground transition hover:border-foreground hover:text-foreground disabled:opacity-50"
                :disabled="props.bulkBusy"
                @click="emit('apply')"
            >
                {{ props.bulkBusy ? t("board.bulk.applying") : t("board.bulk.apply") }}
            </button>
        </div>
        <p v-if="props.bulkMessage" class="mt-2 text-xs text-muted-foreground">
            {{ props.bulkMessage }}
        </p>
    </section>
</template>
