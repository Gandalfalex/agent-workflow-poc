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
            'w-full rounded-3xl border border-border bg-card p-6 shadow-xl',
            props.maxWidth,
            props.class,
          )
        "
      >
        <slot />
      </div>
    </div>
  </Teleport>
</template>
