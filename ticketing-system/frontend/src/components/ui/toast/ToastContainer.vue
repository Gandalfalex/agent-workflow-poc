<script setup lang="ts">
import { useToast } from "@/composables/useToast";
import { useSwipe } from "@/composables/useSwipe";

const { toasts, dismiss } = useToast();
const { onTouchStart, onTouchEnd } = useSwipe();

const iconFor = (type: string) => {
    switch (type) {
        case "success": return "✓";
        case "error": return "✕";
        case "warning": return "⚠";
        default: return "ℹ";
    }
};
</script>

<template>
    <Teleport to="body">
        <div class="fixed bottom-6 right-6 z-[300] flex flex-col-reverse gap-2 items-end">
            <TransitionGroup
                enter-active-class="transition-all duration-300 ease-out"
                enter-from-class="opacity-0 translate-x-6 scale-95"
                enter-to-class="opacity-100 translate-x-0 scale-100"
                leave-active-class="transition-all duration-200 ease-in"
                leave-from-class="opacity-100 translate-x-0 scale-100"
                leave-to-class="opacity-0 translate-x-8 scale-95"
            >
                <div
                    v-for="toast in toasts"
                    :key="toast.id"
                    class="flex items-start gap-3 rounded-xl border px-4 py-3 shadow-xl min-w-[280px] max-w-sm cursor-pointer backdrop-blur-sm select-none"
                    :class="{
                        'border-emerald-500/40 bg-emerald-950/80 text-emerald-200': toast.type === 'success',
                        'border-rose-500/40 bg-rose-950/80 text-rose-200': toast.type === 'error',
                        'border-amber-500/40 bg-amber-950/80 text-amber-200': toast.type === 'warning',
                        'border-primary/30 bg-card/90 text-foreground': toast.type === 'info',
                    }"
                    @click="dismiss(toast.id)"
                    @touchstart="onTouchStart"
                    @touchend="(e) => onTouchEnd(e, { onSwipeRight: () => dismiss(toast.id) })"
                >
                    <span
                        class="mt-0.5 flex h-4 w-4 shrink-0 items-center justify-center rounded-full text-[10px] font-bold"
                        :class="{
                            'bg-emerald-500/20 text-emerald-400': toast.type === 'success',
                            'bg-rose-500/20 text-rose-400': toast.type === 'error',
                            'bg-amber-500/20 text-amber-400': toast.type === 'warning',
                            'bg-primary/20 text-primary': toast.type === 'info',
                        }"
                    >{{ iconFor(toast.type) }}</span>
                    <span class="flex-1 text-sm font-medium leading-snug">{{ toast.message }}</span>
                    <button
                        class="shrink-0 text-xs opacity-40 hover:opacity-80 transition-opacity"
                        @click.stop="dismiss(toast.id)"
                    >✕</button>
                </div>
            </TransitionGroup>
        </div>
    </Teleport>
</template>
