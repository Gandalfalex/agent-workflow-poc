<script setup lang="ts">
import { Button } from "@/components/ui/button";
import { MarkdownEditor } from "@/components/ui/markdown-editor";
import { computed, ref } from "vue";
import { useI18n } from "@/lib/i18n";
import type {
    Story,
    TicketPriority,
    TicketType,
    WorkflowState,
    GroupMember,
    AiTriageSuggestion,
} from "@/lib/api";

type NewTicketForm = {
    title: string;
    description: string;
    priority: TicketPriority;
    type: TicketType;
    storyId: string;
    assignee: string;
    stateId: string;
};

const props = defineProps<{
    show: boolean;
    ticket: NewTicketForm;
    states: WorkflowState[];
    stories: Story[];
    priorities: TicketPriority[];
    ticketTypes: TicketType[];
    canSubmit: boolean;
    groupMembers: GroupMember[];
    aiTriageEnabled: boolean;
    aiTriageLoading: boolean;
    aiTriageBusy: boolean;
    aiTriageError: string;
    aiTriageSuggestion: AiTriageSuggestion | null;
    aiFieldSelection: {
        summary: boolean;
        priority: boolean;
        state: boolean;
        assignee: boolean;
    };
}>();

const emit = defineEmits<{
    (e: "update:ticket", value: NewTicketForm): void;
    (e: "close"): void;
    (e: "create"): void;
    (e: "request-ai-triage"): void;
    (e: "toggle-ai-field", field: "summary" | "priority" | "state" | "assignee", value: boolean): void;
}>();

const assigneeSearch = ref("");
const showAssigneeDropdown = ref(false);
const { t } = useI18n();

const updateField = (patch: Partial<NewTicketForm>) => {
    emit("update:ticket", { ...props.ticket, ...patch });
};

const fuzzyScore = (query: string, text: string): number => {
    const q = query.toLowerCase();
    const t = text.toLowerCase();
    if (t === q) return 1000;
    if (t.startsWith(q)) return 500;
    const consecutiveIndex = t.indexOf(q);
    if (consecutiveIndex !== -1) return 300 + (100 - consecutiveIndex);

    let score = 0;
    let queryIdx = 0;
    for (let i = 0; i < t.length && queryIdx < q.length; i++) {
        if (t[i] === q[queryIdx]) {
            score += 10 - i * 0.1;
            queryIdx++;
        }
    }
    return queryIdx === q.length ? score : 0;
};

const filteredMembers = computed(() => {
    if (!assigneeSearch.value.trim()) {
        return props.groupMembers.slice(0, 5);
    }

    const query = assigneeSearch.value.trim();
    return props.groupMembers
        .map((member) => ({
            member,
            score: Math.max(
                fuzzyScore(query, member.user?.email || ""),
                fuzzyScore(query, member.user?.name || ""),
            ),
        }))
        .filter((item) => item.score > 0)
        .sort((a, b) => b.score - a.score)
        .slice(0, 5)
        .map((item) => item.member);
});

const assignMember = (memberId: string, memberName: string) => {
    updateField({ assignee: memberId });
    assigneeSearch.value = memberName;
    showAssigneeDropdown.value = false;
};

const handleAssigneeBlur = () => {
    setTimeout(() => {
        showAssigneeDropdown.value = false;
    }, 200);
};

const closeModal = (event?: Event) => {
    event?.preventDefault();
    event?.stopPropagation();
    emit("close");
};
</script>

<template>
    <div
        v-if="props.show"
        data-testid="new-ticket.modal"
        class="fixed inset-0 z-[120] flex items-center justify-center bg-black/30 px-6"
        @click.self="emit('close')"
    >
        <div
            class="w-full max-w-lg rounded-3xl border border-border bg-card p-6 shadow-xl"
        >
            <div class="flex items-center justify-between">
                <div>
                    <p
                        class="text-xs uppercase tracking-[0.3em] text-muted-foreground"
                    >
                        {{ t("newTicket.title") }}
                    </p>
                    <h2 class="text-2xl font-semibold">{{ t("newTicket.subtitle") }}</h2>
                </div>
                <button
                    type="button"
                    data-testid="new-ticket.close-button"
                    class="inline-flex items-center justify-center rounded-md px-3 py-2 text-sm font-medium transition hover:bg-muted"
                    @click="closeModal"
                >
                    {{ t("common.close") }}
                </button>
            </div>
            <div class="mt-6 space-y-4">
                <div>
                    <div class="flex items-center justify-between">
                        <label class="text-xs font-semibold text-muted-foreground"
                            >{{ t("newTicket.aiTriage") }}</label
                        >
                        <button
                            v-if="props.aiTriageEnabled"
                            data-testid="new-ticket.ai-suggest-button"
                            type="button"
                            class="rounded-lg border border-border bg-background px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[0.12em] text-muted-foreground transition hover:border-foreground hover:text-foreground disabled:opacity-50"
                            :disabled="props.aiTriageBusy"
                            @click="emit('request-ai-triage')"
                        >
                            {{
                                props.aiTriageBusy
                                    ? t("newTicket.suggesting")
                                    : t("newTicket.suggest")
                            }}
                        </button>
                    </div>
                    <p
                        v-if="!props.aiTriageEnabled && !props.aiTriageLoading"
                        class="mt-2 text-[11px] text-muted-foreground"
                    >
                        {{ t("newTicket.aiDisabled") }}
                    </p>
                    <p
                        v-if="props.aiTriageError"
                        class="mt-2 text-[11px] text-destructive"
                    >
                        {{ props.aiTriageError }}
                    </p>
                    <div
                        v-if="props.aiTriageSuggestion"
                        data-testid="new-ticket.ai-suggestion-panel"
                        class="mt-2 rounded-xl border border-border/80 bg-background/70 p-3 text-xs"
                    >
                        <p class="text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
                            Model {{ props.aiTriageSuggestion.model }} · {{ props.aiTriageSuggestion.promptVersion }}
                        </p>
                        <div class="mt-2 space-y-2">
                            <label class="flex items-center gap-2">
                                <input
                                    data-testid="new-ticket.ai-apply-summary"
                                    type="checkbox"
                                    :checked="props.aiFieldSelection.summary"
                                    @change="emit('toggle-ai-field', 'summary', ($event.target as HTMLInputElement).checked)"
                                />
                                <span>{{ t("newTicket.applySummary") }}</span>
                            </label>
                            <label class="flex items-center gap-2">
                                <input
                                    data-testid="new-ticket.ai-apply-priority"
                                    type="checkbox"
                                    :checked="props.aiFieldSelection.priority"
                                    @change="emit('toggle-ai-field', 'priority', ($event.target as HTMLInputElement).checked)"
                                />
                                <span>{{ t("newTicket.applyPriority") }}</span>
                            </label>
                            <label class="flex items-center gap-2">
                                <input
                                    data-testid="new-ticket.ai-apply-state"
                                    type="checkbox"
                                    :checked="props.aiFieldSelection.state"
                                    @change="emit('toggle-ai-field', 'state', ($event.target as HTMLInputElement).checked)"
                                />
                                <span>{{ t("newTicket.applyState") }}</span>
                            </label>
                            <label class="flex items-center gap-2">
                                <input
                                    data-testid="new-ticket.ai-apply-assignee"
                                    type="checkbox"
                                    :checked="props.aiFieldSelection.assignee"
                                    @change="emit('toggle-ai-field', 'assignee', ($event.target as HTMLInputElement).checked)"
                                />
                                <span>{{ t("newTicket.applyAssignee") }}</span>
                            </label>
                        </div>
                    </div>
                </div>
                <div>
                    <label class="text-xs font-semibold text-muted-foreground"
                        >{{ t("newTicket.fieldTitle") }}</label
                    >
                    <input
                        data-testid="new-ticket.title-input"
                        :value="props.ticket.title"
                        type="text"
                        :placeholder="t('newTicket.shortSummaryPlaceholder')"
                        class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                        @input="
                            updateField({
                                title: ($event.target as HTMLInputElement)
                                    .value,
                            })
                        "
                    />
                </div>
                <div>
                    <label class="text-xs font-semibold text-muted-foreground"
                        >{{ t("newTicket.fieldDescription") }}</label
                    >
                    <MarkdownEditor
                        :model-value="props.ticket.description"
                        @update:model-value="updateField({ description: $event })"
                        :rows="3"
                        :placeholder="t('newTicket.descriptionPlaceholder')"
                        data-testid="new-ticket.description-input"
                        compact
                    />
                </div>
                <div class="grid gap-4 sm:grid-cols-2">
                    <div>
                        <label
                            class="text-xs font-semibold text-muted-foreground"
                            >{{ t("newTicket.fieldType") }}</label
                        >
                        <select
                            data-testid="new-ticket.type-select"
                            :value="props.ticket.type"
                            class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                            @change="
                                updateField({
                                    type: ($event.target as HTMLSelectElement)
                                        .value as TicketType,
                                })
                            "
                        >
                            <option
                                v-for="type in props.ticketTypes"
                                :key="type"
                                :value="type"
                            >
                                {{ type }}
                            </option>
                        </select>
                    </div>
                    <div>
                        <label
                            class="text-xs font-semibold text-muted-foreground"
                            >{{ t("newTicket.fieldPriority") }}</label
                        >
                        <select
                            data-testid="new-ticket.priority-select"
                            :value="props.ticket.priority"
                            class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                            @change="
                                updateField({
                                    priority: (
                                        $event.target as HTMLSelectElement
                                    ).value as TicketPriority,
                                })
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
                    </div>
                </div>
                <div class="grid gap-4 sm:grid-cols-2">
                    <div>
                        <label
                            class="text-xs font-semibold text-muted-foreground"
                            >{{ t("newTicket.fieldStory") }}</label
                        >
                        <select
                            data-testid="new-ticket.story-select"
                            :value="props.ticket.storyId"
                            class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                            @change="
                                updateField({
                                    storyId: (
                                        $event.target as HTMLSelectElement
                                    ).value,
                                })
                            "
                        >
                            <option value="" disabled>{{ t("newTicket.selectStory") }}</option>
                            <option
                                v-for="story in props.stories"
                                :key="story.id"
                                :value="story.id"
                            >
                                {{ story.title }}
                            </option>
                        </select>
                    </div>
                    <div>
                        <label
                            class="text-xs font-semibold text-muted-foreground"
                            >{{ t("newTicket.fieldState") }}</label
                        >
                        <select
                            data-testid="new-ticket.state-select"
                            :value="props.ticket.stateId"
                            class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                            @change="
                                updateField({
                                    stateId: (
                                        $event.target as HTMLSelectElement
                                    ).value,
                                })
                            "
                        >
                            <option
                                v-for="state in props.states"
                                :key="state.id"
                                :value="state.id"
                            >
                                {{ state.name }}
                            </option>
                        </select>
                    </div>
                </div>
                <div>
                    <label class="text-xs font-semibold text-muted-foreground"
                        >{{ t("newTicket.fieldAssignee") }}</label
                    >
                    <div class="relative mt-2">
                        <input
                            data-testid="new-ticket.assignee-input"
                            v-model="assigneeSearch"
                            type="text"
                            :placeholder="t('newTicket.assigneeSearchPlaceholder')"
                            class="w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                            @focus="showAssigneeDropdown = true"
                            @blur="handleAssigneeBlur"
                        />
                        <div
                            v-if="
                                showAssigneeDropdown &&
                                filteredMembers.length > 0
                            "
                            class="absolute top-full left-0 right-0 mt-2 z-50 rounded-xl border border-border bg-card/95 shadow-lg max-h-60 overflow-y-auto"
                        >
                            <div
                                v-for="member in filteredMembers"
                                :key="member.userId"
                                class="px-3 py-2 hover:bg-primary/10 cursor-pointer border-b border-border last:border-b-0 text-sm"
                                @click="
                                    assignMember(
                                        member.userId,
                                        member.user?.name || member.userId,
                                    )
                                "
                            >
                                <div class="font-semibold text-foreground">
                                    {{ member.user?.name || member.userId }}
                                </div>
                                <div class="text-xs text-muted-foreground">
                                    {{ member.user?.email }}
                                </div>
                            </div>
                        </div>
                        <div
                            v-else-if="
                                showAssigneeDropdown &&
                                assigneeSearch.trim() &&
                                filteredMembers.length === 0
                            "
                            class="absolute top-full left-0 right-0 mt-2 z-50 rounded-xl border border-border bg-card/95 shadow-lg px-3 py-2 text-xs text-muted-foreground"
                        >
                            {{
                                t("newTicket.noUsersFound", {
                                    query: assigneeSearch,
                                })
                            }}
                        </div>
                        <div
                            v-else-if="
                                showAssigneeDropdown &&
                                props.groupMembers.length === 0
                            "
                            class="absolute top-full left-0 right-0 mt-2 z-50 rounded-xl border border-border bg-card/95 shadow-lg px-3 py-2 text-xs text-muted-foreground"
                        >
                            {{ t("newTicket.noGroupMembers") }}
                        </div>
                    </div>
                    <div
                        v-if="props.ticket.assignee"
                        class="mt-2 inline-flex items-center gap-2 rounded-lg bg-primary/10 px-2 py-1 text-xs"
                    >
                        <span class="text-foreground font-semibold">
                            {{
                                props.groupMembers.find(
                                    (m) => m.userId === props.ticket.assignee,
                                )?.user?.name || props.ticket.assignee
                            }}
                        </span>
                        <button
                            type="button"
                            class="text-muted-foreground hover:text-foreground"
                            @click="updateField({ assignee: '' })"
                        >
                            ✕
                        </button>
                    </div>
                </div>
            </div>
            <div class="mt-6 flex items-center justify-end gap-3">
                <Button
                    data-testid="new-ticket.cancel-button"
                    variant="outline"
                    @click="closeModal"
                >
                    {{ t("common.cancel") }}
                </Button>
                <Button
                    data-testid="new-ticket.create-button"
                    :disabled="!props.canSubmit"
                    @click="emit('create')"
                >
                    {{ t("newTicket.create") }}
                </Button>
            </div>
        </div>
    </div>
</template>
