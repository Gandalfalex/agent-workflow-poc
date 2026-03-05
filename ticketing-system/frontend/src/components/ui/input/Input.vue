<script setup lang="ts">
import { cn } from "@/lib/utils";

type Props = {
  modelValue?: string | number | null;
  type?: string;
  placeholder?: string;
  disabled?: boolean;
  class?: string;
};

const props = withDefaults(defineProps<Props>(), {
  type: "text",
});

const emit = defineEmits<{
  (e: "update:modelValue", value: string): void;
}>();
</script>

<template>
  <input
    :type="props.type"
    :value="props.modelValue"
    :placeholder="props.placeholder"
    :disabled="props.disabled"
    :class="
      cn(
        'w-full rounded-xl border border-input bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring disabled:cursor-not-allowed disabled:opacity-50',
        props.class,
      )
    "
    @input="emit('update:modelValue', ($event.target as HTMLInputElement).value)"
  />
</template>
