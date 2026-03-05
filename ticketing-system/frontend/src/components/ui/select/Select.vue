<script setup lang="ts">
import { cn } from "@/lib/utils";

type Props = {
  modelValue?: string;
  disabled?: boolean;
  class?: string;
};

const props = defineProps<Props>();

const emit = defineEmits<{
  (e: "update:modelValue", value: string): void;
}>();
</script>

<template>
  <select
    :value="props.modelValue"
    :disabled="props.disabled"
    :class="
      cn(
        'w-full rounded-xl border border-input bg-background px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring disabled:cursor-not-allowed disabled:opacity-50',
        props.class,
      )
    "
    @change="emit('update:modelValue', ($event.target as HTMLSelectElement).value)"
  >
    <slot />
  </select>
</template>
