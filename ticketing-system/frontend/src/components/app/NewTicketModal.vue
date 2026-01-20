<script setup lang="ts">
import { Button } from "@/components/ui/button";
import { computed, ref } from "vue";
import type {
    Story,
    TicketPriority,
    TicketType,
    WorkflowState,
    GroupMember,
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
}>();

const emit = defineEmits<{
    (e: "update:ticket", value: NewTicketForm): void;
    (e: "close"): void;
    (e: "create"): void;
}>();

const assigneeSearch = ref("");
const showAssigneeDropdown = ref(false);

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
</script>

<template>
    <div
        v-if="props.show"
        class="fixed inset-0 z-20 flex items-center justify-center bg-black/30 px-6"
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
                        New ticket
                    </p>
                    <h2 class="text-2xl font-semibold">Create a card</h2>
                </div>
                <Button variant="ghost" size="sm" @click="emit('close')"
                    >Close</Button
                >
            </div>
            <div class="mt-6 space-y-4">
                <div>
                    <label class="text-xs font-semibold text-muted-foreground"
                        >Title</label
                    >
                    <input
                        :value="props.ticket.title"
                        type="text"
                        placeholder="Short summary"
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
                        >Description</label
                    >
                    <textarea
                        :value="props.ticket.description"
                        rows="3"
                        placeholder="What needs to happen?"
                        class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                        @input="
                            updateField({
                                description: (
                                    $event.target as HTMLTextAreaElement
                                ).value,
                            })
                        "
                    ></textarea>
                </div>
                <div class="grid gap-4 sm:grid-cols-2">
                    <div>
                        <label
                            class="text-xs font-semibold text-muted-foreground"
                            >Type</label
                        >
                        <select
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
                            >Priority</label
                        >
                        <select
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
                            >Story</label
                        >
                        <select
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
                            <option value="">None</option>
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
                            >State</label
                        >
                        <select
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
                        >Assignee</label
                    >
                    <div class="relative mt-2">
                        <input
                            v-model="assigneeSearch"
                            type="text"
                            placeholder="Search from group members..."
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
                            No users found matching "{{ assigneeSearch }}"
                        </div>
                        <div
                            v-else-if="
                                showAssigneeDropdown &&
                                props.groupMembers.length === 0
                            "
                            class="absolute top-full left-0 right-0 mt-2 z-50 rounded-xl border border-border bg-card/95 shadow-lg px-3 py-2 text-xs text-muted-foreground"
                        >
                            No group members available
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
                <Button variant="outline" @click="emit('close')">Cancel</Button>
                <Button :disabled="!props.canSubmit" @click="emit('create')">
                    Create ticket
                </Button>
            </div>
        </div>
    </div>
</template>
