<script setup lang="ts">
import { ref } from "vue";
import { useI18n } from "@/lib/i18n";
import type {
    BoardFilterPreset,
    TicketPriority,
    TicketType,
    WorkflowState,
} from "@/lib/api";

const props = defineProps<{
    boardSearch: string;
    activePresetId: string;
    hasActiveFilter: boolean;
    showFilterPanel: boolean;
    showShortcutHelp: boolean;
    showPresetEditor: boolean;
    filterStateId: string;
    filterAssigneeId: string;
    filterPriority: string;
    filterType: string;
    filterBlocked: boolean;
    presetName: string;
    presetBusy: boolean;
    presetMessage: string;
    boardFilterPresetsLoading: boolean;
    boardFilterPresetsError: string;
    states: WorkflowState[];
    assigneeOptions: Array<{ id: string; name: string }>;
    priorities: TicketPriority[];
    ticketTypes: TicketType[];
    boardFilterPresets: BoardFilterPreset[];
}>();

const emit = defineEmits<{
    (e: "update:boardSearch", value: string): void;
    (e: "update:activePresetId", value: string): void;
    (e: "toggle-filter-panel"): void;
    (e: "toggle-shortcut-help"): void;
    (e: "clear-filters"): void;
    (e: "preset-change"): void;
    (e: "update:filterStateId", value: string): void;
    (e: "update:filterAssigneeId", value: string): void;
    (e: "update:filterPriority", value: string): void;
    (e: "update:filterType", value: string): void;
    (e: "update:filterBlocked", value: boolean): void;
    (e: "open-preset-editor"): void;
    (e: "update:presetName", value: string): void;
    (e: "save-preset-from-editor"): void;
    (e: "cancel-preset-editor"): void;
    (e: "rename-preset"): void;
    (e: "delete-preset"): void;
    (e: "share-preset"): void;
}>();

const { t } = useI18n();
const searchInput = ref<HTMLInputElement | null>(null);

const focusSearch = () => {
    searchInput.value?.focus();
    searchInput.value?.select();
};

defineExpose({ focusSearch });
</script>

<template>
    <section class="rounded-2xl border border-border bg-card/70 px-4 py-3">
        <div class="flex flex-wrap items-center gap-2">
            <div class="relative min-w-[220px] flex-1">
                <input
                    ref="searchInput"
                    :value="props.boardSearch"
                    data-testid="board.filter-search-input"
                    type="text"
                    :placeholder="t('board.search.placeholder')"
                    class="w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                    @input="
                        emit(
                            'update:boardSearch',
                            ($event.target as HTMLInputElement).value,
                        )
                    "
                />
            </div>
            <select
                :value="props.activePresetId"
                data-testid="board.preset-select"
                class="rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                @change="
                    emit(
                        'update:activePresetId',
                        ($event.target as HTMLSelectElement).value,
                    );
                    emit('preset-change');
                "
            >
                <option value="">{{ t("board.toolbar.views") }}</option>
                <option
                    v-for="preset in props.boardFilterPresets"
                    :key="preset.id"
                    :value="preset.id"
                >
                    {{ preset.name }}
                </option>
            </select>
            <button
                class="rounded-lg border border-border bg-background px-3 py-2 text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground transition hover:border-foreground hover:text-foreground"
                @click="emit('toggle-filter-panel')"
            >
                {{
                    props.showFilterPanel
                        ? t("board.toolbar.hideFilters")
                        : t("board.toolbar.filters")
                }}
            </button>
            <button
                class="rounded-lg border border-border bg-background px-2.5 py-2 text-xs font-semibold text-muted-foreground transition hover:border-foreground hover:text-foreground"
                @click="emit('toggle-shortcut-help')"
            >
                ?
            </button>
            <button
                v-if="props.hasActiveFilter"
                data-testid="board.filter-clear-button"
                class="whitespace-nowrap rounded-lg border border-border bg-background px-2.5 py-2 text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground transition hover:border-foreground hover:text-foreground"
                @click="emit('clear-filters')"
            >
                {{ t("common.clear") }}
            </button>
        </div>

        <div
            v-if="props.showShortcutHelp"
            class="mt-2 grid grid-cols-1 gap-1 rounded-xl border border-border/70 bg-background/50 px-3 py-2 text-xs text-muted-foreground md:grid-cols-2"
        >
            <p><span class="font-mono text-foreground">/</span> {{ t("board.shortcuts.search") }}</p>
            <p><span class="font-mono text-foreground">N</span> {{ t("board.shortcuts.newTicket") }}</p>
        </div>

        <div
            v-if="props.showFilterPanel"
            class="mt-2 space-y-2 rounded-xl border border-border/70 bg-background/40 p-3"
        >
            <div class="grid grid-cols-1 gap-2 md:grid-cols-5">
                <select
                    :value="props.filterStateId"
                    data-testid="board.filter-state-select"
                    class="rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                    @change="
                        emit(
                            'update:filterStateId',
                            ($event.target as HTMLSelectElement).value,
                        )
                    "
                >
                    <option value="">{{ t("board.filter.allStates") }}</option>
                    <option v-for="state in props.states" :key="state.id" :value="state.id">
                        {{ state.name }}
                    </option>
                </select>
                <select
                    :value="props.filterAssigneeId"
                    data-testid="board.filter-assignee-select"
                    class="rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                    @change="
                        emit(
                            'update:filterAssigneeId',
                            ($event.target as HTMLSelectElement).value,
                        )
                    "
                >
                    <option value="">{{ t("board.filter.allAssignees") }}</option>
                    <option
                        v-for="assignee in props.assigneeOptions"
                        :key="assignee.id"
                        :value="assignee.id"
                    >
                        {{ assignee.name }}
                    </option>
                </select>
                <select
                    :value="props.filterPriority"
                    data-testid="board.filter-priority-select"
                    class="rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                    @change="
                        emit(
                            'update:filterPriority',
                            ($event.target as HTMLSelectElement).value,
                        )
                    "
                >
                    <option value="">{{ t("board.filter.allPriorities") }}</option>
                    <option
                        v-for="priority in props.priorities"
                        :key="priority"
                        :value="priority"
                    >
                        {{ priority }}
                    </option>
                </select>
                <select
                    :value="props.filterType"
                    data-testid="board.filter-type-select"
                    class="rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                    @change="
                        emit(
                            'update:filterType',
                            ($event.target as HTMLSelectElement).value,
                        )
                    "
                >
                    <option value="">{{ t("board.filter.allTypes") }}</option>
                    <option
                        v-for="ticketType in props.ticketTypes"
                        :key="ticketType"
                        :value="ticketType"
                    >
                        {{ ticketType }}
                    </option>
                </select>
                <label
                    class="flex items-center gap-2 rounded-xl border border-input bg-background px-3 py-2 text-sm"
                >
                    <input
                        :checked="props.filterBlocked"
                        data-testid="board.filter-blocked-checkbox"
                        type="checkbox"
                        @change="
                            emit(
                                'update:filterBlocked',
                                ($event.target as HTMLInputElement).checked,
                            )
                        "
                    />
                    {{ t("board.filter.blockedOnly") }}
                </label>
            </div>

            <div class="flex flex-wrap items-center gap-2">
                <input
                    :value="props.presetName"
                    type="text"
                    data-testid="board.preset-name-input"
                    :placeholder="t('board.filter.presetName')"
                    class="rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                    @input="
                        emit(
                            'update:presetName',
                            ($event.target as HTMLInputElement).value,
                        )
                    "
                />
                <button
                    data-testid="board.preset-save-button"
                    class="rounded-lg border border-border bg-background px-3 py-2 text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground transition hover:border-foreground hover:text-foreground"
                    :disabled="props.presetBusy"
                    @click="emit('save-preset-from-editor')"
                >
                    {{ t("board.preset.save") }}
                </button>
                <button
                    data-testid="board.preset-rename-button"
                    class="rounded-lg border border-border bg-background px-3 py-2 text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground transition hover:border-foreground hover:text-foreground disabled:opacity-50"
                    :disabled="props.presetBusy || !props.activePresetId"
                    @click="emit('rename-preset')"
                >
                    {{ t("board.preset.rename") }}
                </button>
                <button
                    data-testid="board.preset-delete-button"
                    class="rounded-lg border border-border bg-background px-3 py-2 text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground transition hover:border-foreground hover:text-foreground disabled:opacity-50"
                    :disabled="props.presetBusy || !props.activePresetId"
                    @click="emit('delete-preset')"
                >
                    {{ t("board.preset.delete") }}
                </button>
                <button
                    data-testid="board.preset-share-button"
                    class="rounded-lg border border-border bg-background px-3 py-2 text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground transition hover:border-foreground hover:text-foreground disabled:opacity-50"
                    :disabled="props.presetBusy || !props.activePresetId"
                    @click="emit('share-preset')"
                >
                    {{ t("board.preset.share") }}
                </button>
            </div>
            <p
                v-if="
                    props.presetMessage ||
                    props.boardFilterPresetsLoading ||
                    props.boardFilterPresetsError
                "
                class="text-xs text-muted-foreground"
            >
                {{
                    props.presetMessage ||
                    (props.boardFilterPresetsLoading
                        ? t("board.filter.loadingPresets")
                        : props.boardFilterPresetsError)
                }}
            </p>
        </div>
    </section>
</template>
