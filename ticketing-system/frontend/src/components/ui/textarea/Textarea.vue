<script setup lang="ts">
import { cn } from "@/lib/utils";

type Props = {
  modelValue?: string;
  placeholder?: string;
  rows?: number;
  disabled?: boolean;
  class?: string;
};

const props = withDefaults(defineProps<Props>(), {
  rows: 3,
});

const emit = defineEmits<{
  (e: "update:modelValue", value: string): void;
}>();
</script>

<template>
  <textarea
    :value="props.modelValue"
    :placeholder="props.placeholder"
    :rows="props.rows"
    :disabled="props.disabled"
    :class="
      cn(
        'w-full rounded-xl border border-input bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring disabled:cursor-not-allowed disabled:opacity-50',
        props.class,
      )
    "
    @input="emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"
  />
</template>
