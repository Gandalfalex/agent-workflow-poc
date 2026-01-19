<script setup lang="ts">
import { Button } from "@/components/ui/button";
import { marked } from "marked";
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

const updateEditor = (patch: Partial<TicketEditor>) => {
    emit("update:editor", { ...props.editor, ...patch });
};

const isCurrentUser = (userId: string) => {
    return props.currentUserId === userId;
};
</script>

<template>
    <div
        v-if="props.show"
        class="fixed inset-0 z-30 flex items-center justify-center bg-black/50 px-6"
        @click.self="emit('close')"
    >
        <div
            class="w-full max-h-[90vh] max-w-4xl overflow-y-auto rounded-3xl border border-border bg-card p-6 shadow-xl"
        >
            <div class="flex items-center justify-between">
                <div>
                    <p
                        class="text-xs uppercase tracking-[0.3em] text-muted-foreground"
                    >
                        Ticket
                    </p>
                    <h2 class="text-2xl font-semibold">
                        {{ props.ticketKey }}
                    </h2>
                </div>
                <details class="relative">
                    <summary
                        class="list-none rounded-full border border-border bg-background px-2 py-1 text-lg font-semibold text-muted-foreground transition hover:border-foreground hover:text-foreground cursor-pointer"
                        aria-label="Ticket actions"
                    >
                        ⋮
                    </summary>
                    <div
                        class="absolute right-0 top-full mt-2 w-40 rounded-2xl border border-border bg-card/95 backdrop-blur p-2 text-xs z-50 shadow-lg"
                    >
                        <Button
                            variant="outline"
                            size="sm"
                            class="w-full border-destructive/30 text-destructive hover:bg-destructive/5"
                            :disabled="props.ticketSaving"
                            @click.stop="emit('delete')"
                        >
                            Delete ticket
                        </Button>
                        <div class="border-t border-border my-2"></div>
                        <Button
                            variant="ghost"
                            size="sm"
                            class="w-full justify-start text-muted-foreground hover:text-foreground"
                            @click.stop="emit('close')"
                        >
                            Close
                        </Button>
                    </div>
                </details>
            </div>
            <div
                class="mt-6 grid gap-6 lg:grid-cols-[1fr_1fr] max-h-[calc(90vh-200px)]"
            >
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
                        <label
                            class="text-xs font-semibold text-muted-foreground"
                            >Description</label
                        >
                        <textarea
                            :value="props.editor.description"
                            rows="4"
                            class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                            @input="
                                updateEditor({
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
                    <div
                        v-if="props.ticketError"
                        class="rounded-2xl border border-border bg-secondary/60 px-3 py-2 text-xs"
                    >
                        {{ props.ticketError }}
                    </div>
                    <div class="flex justify-end gap-2">
                        <Button
                            variant="ghost"
                            size="sm"
                            @click="emit('close')"
                        >
                            Cancel
                        </Button>
                        <Button
                            size="sm"
                            :disabled="props.ticketSaving"
                            @click="emit('save')"
                        >
                            {{
                                props.ticketSaving
                                    ? "Saving..."
                                    : "Save changes"
                            }}
                        </Button>
                    </div>
                </div>
                <div
                    class="rounded-2xl border border-border bg-background px-6 py-6 text-xs text-muted-foreground flex flex-col h-full"
                >
                    <div
                        class="flex items-center justify-between mb-4 flex-shrink-0"
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
                        class="flex-1 space-y-3 overflow-y-auto mb-4 pr-2 min-h-0"
                    >
                        <div
                            v-for="comment in props.comments"
                            :key="comment.id"
                            :class="[
                                'rounded-xl px-4 py-3 max-w-[80%]',
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
                                    class="text-xs text-muted-foreground ml-2"
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
                        class="flex-1 flex items-center justify-center text-xs text-muted-foreground mb-4 min-h-0"
                    >
                        No comments yet.
                    </div>

                    <div class="border-t border-border pt-4 flex-shrink-0">
                        <label
                            class="text-xs font-semibold text-muted-foreground block mb-2"
                            >Add comment (Markdown supported)</label
                        >
                        <textarea
                            :value="props.commentDraft"
                            rows="3"
                            placeholder="Share progress or blockers... (supports **bold**, *italic*, `code`, etc.)"
                            class="w-full rounded-xl border border-input bg-background px-3 py-2 text-xs focus:outline-none focus:ring-2 focus:ring-ring resize-none"
                            @input="
                                emit(
                                    'update:commentDraft',
                                    ($event.target as HTMLTextAreaElement)
                                        .value,
                                )
                            "
                        ></textarea>
                        <div class="mt-3 flex items-center gap-3">
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
                                        : "Add comment"
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
        </div>
    </div>
</template>
