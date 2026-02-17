<script setup lang="ts">
import { computed, ref, nextTick } from "vue";
import { marked } from "marked";
import {
    Bold,
    Italic,
    Code,
    Link,
    List,
    ListOrdered,
    Quote,
    Heading2,
    Eye,
    Pencil,
} from "lucide-vue-next";
import {
    toolbarActions,
    handleTab,
    type ToolbarAction,
    type ActionResult,
} from "./toolbar-actions";

const props = withDefaults(
    defineProps<{
        modelValue: string;
        rows?: number;
        placeholder?: string;
        showPreview?: boolean;
        compact?: boolean;
        disabled?: boolean;
        dataTestid?: string;
    }>(),
    {
        rows: 5,
        placeholder: "",
        showPreview: true,
        compact: false,
        disabled: false,
        dataTestid: undefined,
    },
);

const emit = defineEmits<{
    (e: "update:modelValue", value: string): void;
}>();

const textareaRef = ref<HTMLTextAreaElement | null>(null);
const previewing = ref(false);

const iconMap: Record<string, unknown> = {
    Bold,
    Italic,
    Code,
    Link,
    List,
    ListOrdered,
    Quote,
    Heading2,
};

const visibleActions = computed(() => {
    if (props.compact) {
        return toolbarActions.filter((a) => a.compact);
    }
    return toolbarActions;
});

const renderedHtml = computed(() => {
    return marked(props.modelValue || "");
});

function applyResult(result: ActionResult) {
    emit("update:modelValue", result.value);
    nextTick(() => {
        const el = textareaRef.value;
        if (el) {
            el.focus();
            el.setSelectionRange(result.start, result.end);
        }
    });
}

function applyAction(action: ToolbarAction) {
    const el = textareaRef.value;
    if (!el) return;
    applyResult(action.apply(el));
}

function handleKeydown(event: KeyboardEvent) {
    const el = textareaRef.value;
    if (!el) return;

    // Tab / Shift+Tab
    if (event.key === "Tab") {
        event.preventDefault();
        applyResult(handleTab(el, event.shiftKey));
        return;
    }

    // Ctrl/Cmd + shortcut key
    if (event.ctrlKey || event.metaKey) {
        const action = toolbarActions.find(
            (a) => a.shortcutKey === event.key.toLowerCase(),
        );
        if (action) {
            event.preventDefault();
            applyResult(action.apply(el));
        }
    }
}
</script>

<template>
    <div class="mt-2">
        <!-- Toolbar -->
        <div
            class="flex items-center gap-0.5 rounded-t-xl border border-b-0 border-input bg-muted/30 px-1.5 py-1"
        >
            <button
                v-for="action in visibleActions"
                :key="action.id"
                type="button"
                :title="
                    action.label +
                    (action.shortcutKey
                        ? ` (Ctrl+${action.shortcutKey.toUpperCase()})`
                        : '')
                "
                class="rounded-lg p-1.5 text-muted-foreground transition hover:bg-background hover:text-foreground"
                :disabled="props.disabled || previewing"
                @click="applyAction(action)"
            >
                <component
                    :is="iconMap[action.icon]"
                    :size="14"
                    :stroke-width="2"
                />
            </button>
            <span class="flex-1" />
            <button
                v-if="props.showPreview"
                type="button"
                class="flex items-center gap-1 rounded-lg px-2 py-1 text-[10px] font-semibold text-muted-foreground transition hover:bg-background hover:text-foreground"
                @click="previewing = !previewing"
            >
                <component
                    :is="previewing ? Pencil : Eye"
                    :size="12"
                    :stroke-width="2"
                />
                {{ previewing ? "Edit" : "Preview" }}
            </button>
        </div>

        <!-- Textarea (edit mode) -->
        <textarea
            v-if="!previewing"
            ref="textareaRef"
            :data-testid="props.dataTestid"
            :value="props.modelValue"
            :rows="props.rows"
            :placeholder="props.placeholder"
            :disabled="props.disabled"
            class="w-full rounded-b-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring resize-none"
            @input="
                emit(
                    'update:modelValue',
                    ($event.target as HTMLTextAreaElement).value,
                )
            "
            @keydown="handleKeydown"
        ></textarea>

        <!-- Preview (preview mode) -->
        <div
            v-else
            class="w-full rounded-b-xl border border-input bg-background px-3 py-2 text-sm prose prose-sm dark:prose-invert max-w-none"
            :style="{ minHeight: props.rows * 1.5 + 1 + 'rem' }"
            v-html="renderedHtml"
        ></div>
    </div>
</template>
