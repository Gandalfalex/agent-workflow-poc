<script setup lang="ts">
import { cn } from "@/lib/utils";

type Props = {
  show: boolean;
  maxWidth?: string;
  zIndex?: string;
  class?: string;
  testId?: string;
};

const props = withDefaults(defineProps<Props>(), {
  maxWidth: "max-w-lg",
  zIndex: "z-[120]",
});

const emit = defineEmits<{
  (e: "close"): void;
}>();
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition-opacity duration-200 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition-opacity duration-150 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="props.show"
        :data-testid="props.testId"
        :class="
          cn(
            'fixed inset-0 flex items-center justify-center bg-black/65 backdrop-blur-[2px] px-6',
            props.zIndex,
          )
        "
        @click.self="emit('close')"
      >
        <div
          :class="
            cn(
              'modal-panel w-full rounded-3xl border border-border bg-card p-6 shadow-xl',
              props.maxWidth,
              props.class,
            )
          "
        >
          <slot />
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.modal-panel {
  animation: modal-scale-in 200ms cubic-bezier(0, 0, 0.2, 1) both;
}

@keyframes modal-scale-in {
  from {
    transform: scale(0.95) translateY(8px);
    opacity: 0;
  }
  to {
    transform: scale(1) translateY(0);
    opacity: 1;
  }
}
</style>
