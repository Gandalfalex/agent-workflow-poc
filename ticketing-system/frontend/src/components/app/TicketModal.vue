<script setup lang="ts">
import { Button } from "@/components/ui/button";
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
</script>

<template>
    <div
        v-if="props.show"
        class="fixed inset-0 z-30 flex items-center justify-center bg-black/50 px-6"
        @click.self="emit('close')"
    >
        <div
            class="w-full max-w-2xl rounded-3xl border border-border bg-card p-6 shadow-xl"
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
                <Button variant="ghost" size="sm" @click="emit('close')"
                    >Close</Button
                >
            </div>
            <div class="mt-6 grid gap-6 lg:grid-cols-[1.4fr_0.9fr]">
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
                    <div class="flex items-center justify-between gap-2">
                        <Button
                            variant="outline"
                            size="sm"
                            class="border-destructive/40 text-destructive hover:bg-destructive/10"
                            :disabled="props.ticketSaving"
                            @click="emit('delete')"
                        >
                            Delete ticket
                        </Button>
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
                    class="rounded-2xl border border-border bg-background px-4 py-4 text-xs text-muted-foreground"
                >
                    <div class="flex items-center justify-between">
                        <span class="text-xs font-semibold text-foreground"
                            >Comments</span
                        >
                        <span v-if="props.commentSaving">Saving...</span>
                    </div>
                    <div
                        v-if="props.comments.length"
                        class="mt-4 space-y-4 border-l border-border pl-4"
                    >
                        <div
                            v-for="comment in props.comments"
                            :key="comment.id"
                            class="relative rounded-xl border border-border bg-card px-3 py-2"
                        >
                            <span
                                class="absolute -left-6 top-4 h-2 w-2 rounded-full bg-primary/60"
                            ></span>
                            <div class="flex items-center justify-between">
                                <span class="text-[11px] font-semibold">{{
                                    comment.authorName
                                }}</span>
                                <span class="text-[11px] text-muted-foreground">
                                    {{
                                        new Date(
                                            comment.createdAt,
                                        ).toLocaleString()
                                    }}
                                </span>
                            </div>
                            <p class="mt-2 text-xs text-muted-foreground">
                                {{ comment.message }}
                            </p>
                        </div>
                    </div>
                    <div v-else class="mt-3 text-xs text-muted-foreground">
                        No comments yet.
                    </div>
                    <div class="mt-4">
                        <label
                            class="text-[11px] font-semibold text-muted-foreground"
                            >Add comment</label
                        >
                        <textarea
                            :value="props.commentDraft"
                            rows="3"
                            placeholder="Share progress or blockers"
                            class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-xs focus:outline-none focus:ring-2 focus:ring-ring"
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
                                Add comment
                            </Button>
                            <span v-if="props.commentError">{{
                                props.commentError
                            }}</span>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>
