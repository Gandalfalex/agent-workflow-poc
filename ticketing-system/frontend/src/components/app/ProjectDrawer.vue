<script setup lang="ts">
import { Button } from "@/components/ui/button";
import type { Project } from "@/lib/api";

const props = defineProps<{
  modelValue: boolean;
  projectSearch: string;
  projectLoading: boolean;
  filteredProjects: Project[];
  onSelectProject: (projectId: string) => void;
}>();

const emit = defineEmits<{
  (e: "update:modelValue", value: boolean): void;
  (e: "update:projectSearch", value: string): void;
}>();
</script>

<template>
  <div
    v-if="modelValue"
    class="fixed inset-0 z-20 bg-black/50"
    @click.self="emit('update:modelValue', false)"
  >
    <div class="absolute right-0 top-0 h-full w-full max-w-md bg-card p-6 shadow-2xl">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-xs uppercase tracking-[0.3em] text-muted-foreground">
            Projects
          </p>
          <h2 class="text-xl font-semibold">Select board</h2>
        </div>
        <Button variant="ghost" size="sm" @click="emit('update:modelValue', false)">
          Close
        </Button>
      </div>
      <div class="mt-6 grid gap-3">
        <input
          :value="projectSearch"
          type="text"
          placeholder="Search projects"
          class="w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
          @input="emit('update:projectSearch', ($event.target as HTMLInputElement).value)"
        />
        <div v-if="projectLoading" class="text-xs text-muted-foreground">
          Loading projects...
        </div>
        <div
          v-else-if="filteredProjects.length === 0"
          class="rounded-xl border border-border bg-background px-3 py-2 text-xs text-muted-foreground"
        >
          No projects found.
        </div>
        <div
          v-else
          class="overflow-hidden rounded-xl border border-border bg-background"
        >
          <button
            v-for="project in filteredProjects"
            :key="project.id"
            class="flex w-full items-center justify-between border-b border-border px-3 py-2 text-left text-sm last:border-b-0 hover:bg-muted/60"
            @click="props.onSelectProject(project.id)"
          >
            <span class="font-semibold">{{ project.key }}</span>
            <span class="text-xs text-muted-foreground">{{ project.name }}</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
