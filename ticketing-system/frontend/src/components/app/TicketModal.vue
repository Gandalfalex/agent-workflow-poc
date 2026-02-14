<script setup lang="ts">
import { Button } from "@/components/ui/button";
import { marked } from "marked";
import { ref } from "vue";
import type {
    TicketComment,
    TicketPriority,
    TicketType,
    WorkflowState,
} from "@/lib/api";

type TicketEditor = {
    title: string;
    description: string;
    priority: TicketPriority;
    stateId: string;
    type: TicketType;
};

const props = defineProps<{
    show: boolean;
    ticketKey: string;
    editor: TicketEditor;
    states: WorkflowState[];
    priorities: TicketPriority[];
    ticketTypes: TicketType[];
    ticketSaving: boolean;
    ticketError: string;
    comments: TicketComment[];
    commentDraft: string;
    commentSaving: boolean;
    commentError: string;
    currentUserId?: string;
}>();

const emit = defineEmits<{
    (e: "update:editor", value: TicketEditor): void;
    (e: "update:commentDraft", value: string): void;
    (e: "close"): void;
    (e: "save"): void;
    (e: "delete"): void;
    (e: "add-comment"): void;
}>();

const showDescriptionPreview = ref(false);
const menuOpen = ref(false);

const updateEditor = (patch: Partial<TicketEditor>) => {
    emit("update:editor", { ...props.editor, ...patch });
};

const isCurrentUser = (userId: string) => {
    return props.currentUserId === userId;
};

const priorityColor = (priority: string) => {
    switch (priority) {
        case "urgent":
            return "bg-red-500/20 text-red-300 border-red-500/30";
        case "high":
            return "bg-orange-500/20 text-orange-300 border-orange-500/30";
        case "medium":
            return "bg-amber-500/20 text-amber-300 border-amber-500/30";
        case "low":
        default:
            return "bg-slate-500/15 text-slate-400 border-slate-500/20";
    }
};
</script>

<template>
    <div
        v-if="props.show"
        class="fixed inset-0 z-30 flex items-center justify-center bg-black/50 px-6"
        @click.self="emit('close')"
    >
        <div
            class="flex w-full max-h-[92vh] max-w-5xl flex-col rounded-3xl border border-border bg-card shadow-xl"
        >
            <!-- Header: fixed at top -->
            <div class="flex items-center justify-between px-6 pt-5 pb-4 border-b border-border flex-shrink-0">
                <div class="flex items-center gap-4">
                    <div>
                        <p
                            class="text-[10px] uppercase tracking-[0.3em] text-muted-foreground"
                        >
                            Ticket
                        </p>
                        <h2 class="text-xl font-semibold">
                            {{ props.ticketKey }}
                        </h2>
                    </div>
                    <span
                        class="rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase border"
                        :class="priorityColor(props.editor.priority)"
                    >
                        {{ props.editor.priority }}
                    </span>
                </div>
                <div class="flex items-center gap-2">
                    <Button variant="ghost" size="sm" @click="emit('close')">
                        Close
                    </Button>
                    <div class="relative">
                        <button
                            class="rounded-full border border-border bg-background px-2 py-1 text-lg font-semibold text-muted-foreground transition hover:border-foreground hover:text-foreground cursor-pointer"
                            aria-label="Ticket actions"
                            @click.stop="menuOpen = !menuOpen"
                        >
                            &#x22EE;
                        </button>
                        <div
                            v-if="menuOpen"
                            class="dropdown-backdrop"
                            @click="menuOpen = false"
                        ></div>
                        <div
                            v-if="menuOpen"
                            class="absolute right-0 top-full mt-2 w-40 rounded-2xl border border-border bg-card/95 backdrop-blur p-2 text-xs z-50 shadow-lg"
                        >
                            <Button
                                variant="outline"
                                size="sm"
                                class="w-full border-destructive/30 text-destructive hover:bg-destructive/5"
                                :disabled="props.ticketSaving"
                                @click.stop="menuOpen = false; emit('delete')"
                            >
                                Delete ticket
                            </Button>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Body: scrollable two-column layout -->
            <div
                class="flex-1 min-h-0 grid gap-0 lg:grid-cols-[1fr_1fr]"
            >
                <!-- Left: Form fields (independently scrollable) -->
                <div class="overflow-y-auto p-6 border-r border-border">
                    <div class="grid gap-4">
                        <div>
                            <label
                                class="text-xs font-semibold text-muted-foreground"
                                >Title</label
                            >
                            <input
                                :value="props.editor.title"
                                type="text"
                                class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                                @input="
                                    updateEditor({
                                        title: ($event.target as HTMLInputElement)
                                            .value,
                                    })
                                "
                            />
                        </div>
                        <div>
                            <div class="flex items-center justify-between">
                                <label
                                    class="text-xs font-semibold text-muted-foreground"
                                    >Description</label
                                >
                                <button
                                    type="button"
                                    class="text-xs text-muted-foreground hover:text-foreground transition"
                                    @click="
                                        showDescriptionPreview =
                                            !showDescriptionPreview
                                    "
                                >
                                    {{
                                        showDescriptionPreview ? "Edit" : "Preview"
                                    }}
                                </button>
                            </div>
                            <div
                                v-if="showDescriptionPreview"
                                class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm prose prose-sm dark:prose-invert max-w-none min-h-[160px]"
                                v-html="marked(props.editor.description || '')"
                            ></div>
                            <textarea
                                v-else
                                :value="props.editor.description"
                                rows="7"
                                placeholder="Describe the ticket... (supports **bold**, *italic*, `code`)"
                                class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring resize-none"
                                @input="
                                    updateEditor({
                                        description: (
                                            $event.target as HTMLTextAreaElement
                                        ).value,
                                    })
                                "
                            ></textarea>
                        </div>
                        <div class="grid gap-4 sm:grid-cols-3">
                            <div>
                                <label
                                    class="text-xs font-semibold text-muted-foreground"
                                    >Type</label
                                >
                                <select
                                    :value="props.editor.type"
                                    class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                                    @change="
                                        updateEditor({
                                            type: (
                                                $event.target as HTMLSelectElement
                                            ).value as TicketType,
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
                                    :value="props.editor.priority"
                                    class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                                    @change="
                                        updateEditor({
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
                            <div>
                                <label
                                    class="text-xs font-semibold text-muted-foreground"
                                    >State</label
                                >
                                <select
                                    :value="props.editor.stateId"
                                    class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                                    @change="
                                        updateEditor({
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
                        <div
                            v-if="props.ticketError"
                            class="rounded-2xl border border-border bg-secondary/60 px-3 py-2 text-xs"
                        >
                            {{ props.ticketError }}
                        </div>
                    </div>
                </div>

                <!-- Right: Comments (independently scrollable) -->
                <div class="flex flex-col min-h-0 overflow-hidden">
                    <div
                        class="flex items-center justify-between px-6 py-3 flex-shrink-0 border-b border-border"
                    >
                        <span class="text-sm font-semibold text-foreground"
                            >Comments</span
                        >
                        <span
                            v-if="props.commentSaving"
                            class="text-xs text-muted-foreground"
                            >Saving...</span
                        >
                    </div>

                    <div
                        v-if="props.comments.length"
                        class="flex-1 space-y-2 overflow-y-auto px-6 py-3 min-h-0"
                    >
                        <div
                            v-for="comment in props.comments"
                            :key="comment.id"
                            :class="[
                                'rounded-xl px-4 py-3 max-w-[85%]',
                                isCurrentUser(comment.authorId)
                                    ? 'ml-auto bg-primary/10 border border-primary/30'
                                    : 'bg-card border border-border',
                            ]"
                        >
                            <div class="flex items-center justify-between mb-2">
                                <span class="text-xs font-semibold">{{
                                    comment.authorName
                                }}</span>
                                <span
                                    class="text-[10px] text-muted-foreground ml-2"
                                >
                                    {{
                                        new Date(
                                            comment.createdAt,
                                        ).toLocaleString()
                                    }}
                                </span>
                            </div>
                            <div
                                class="text-xs text-foreground prose prose-sm dark:prose-invert max-w-none"
                                v-html="marked(comment.message)"
                            ></div>
                        </div>
                    </div>
                    <div
                        v-else
                        class="flex-1 flex items-center justify-center text-xs text-muted-foreground min-h-0"
                    >
                        No comments yet.
                    </div>

                    <div class="border-t border-border px-6 py-3 flex-shrink-0">
                        <label
                            class="text-[10px] font-semibold text-muted-foreground block mb-2"
                            >Add comment (Markdown)</label
                        >
                        <textarea
                            :value="props.commentDraft"
                            rows="2"
                            placeholder="Progress, blockers, notes..."
                            class="w-full rounded-xl border border-input bg-background px-3 py-2 text-xs focus:outline-none focus:ring-2 focus:ring-ring resize-none"
                            @input="
                                emit(
                                    'update:commentDraft',
                                    ($event.target as HTMLTextAreaElement)
                                        .value,
                                )
                            "
                        ></textarea>
                        <div class="mt-2 flex items-center gap-3">
                            <Button
                                size="sm"
                                :disabled="
                                    props.commentSaving ||
                                    !props.commentDraft.trim().length
                                "
                                @click="emit('add-comment')"
                            >
                                {{
                                    props.commentSaving
                                        ? "Posting..."
                                        : "Post"
                                }}
                            </Button>
                            <span
                                v-if="props.commentError"
                                class="text-xs text-destructive"
                                >{{ props.commentError }}</span
                            >
                        </div>
                    </div>
                </div>
            </div>

            <!-- Footer: fixed at bottom -->
            <div class="flex items-center justify-end gap-2 px-6 py-4 border-t border-border flex-shrink-0">
                <Button variant="ghost" size="sm" @click="emit('close')">
                    Cancel
                </Button>
                <Button
                    size="sm"
                    :disabled="props.ticketSaving"
                    @click="emit('save')"
                >
                    {{ props.ticketSaving ? "Saving..." : "Save changes" }}
                </Button>
            </div>
        </div>
    </div>
</template>
