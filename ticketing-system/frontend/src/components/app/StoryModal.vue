<script setup lang="ts">
import { Button } from "@/components/ui/button";
import { MarkdownEditor } from "@/components/ui/markdown-editor";

type StoryDraft = {
  title: string;
  description: string;
  storyPoints: number | null;
};

const props = defineProps<{
  show: boolean;
  story: StoryDraft;
  canCreate: boolean;
  storySaving: boolean;
  storyError: string;
}>();

const emit = defineEmits<{
  (e: "update:story", value: StoryDraft): void;
  (e: "close"): void;
  (e: "create"): void;
}>();

const updateStory = (patch: Partial<StoryDraft>) => {
  emit("update:story", { ...props.story, ...patch });
};
</script>

<template>
  <div
    v-if="props.show"
    data-testid="story.modal"
    class="fixed inset-0 z-[120] flex items-center justify-center bg-black/65 backdrop-blur-[2px] px-6"
    @click.self="emit('close')"
  >
    <div class="w-full max-w-lg rounded-3xl border border-border bg-card p-6 shadow-xl">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-xs uppercase tracking-[0.3em] text-muted-foreground">
            New story
          </p>
          <h2 class="text-2xl font-semibold">Create a story</h2>
        </div>
        <Button variant="ghost" size="sm" @click="emit('close')">Close</Button>
      </div>
      <div class="mt-6 space-y-4">
        <div>
          <label class="text-xs font-semibold text-muted-foreground">Title</label>
          <input
            data-testid="story.title-input"
            :value="props.story.title"
            type="text"
            placeholder="Story title"
            class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            @input="updateStory({ title: ($event.target as HTMLInputElement).value })"
          />
        </div>
        <div>
          <label class="text-xs font-semibold text-muted-foreground"
            >Description</label
          >
          <MarkdownEditor
            :model-value="props.story.description"
            @update:model-value="updateStory({ description: $event })"
            :rows="4"
            placeholder="Shared goal or summary"
            data-testid="story.description-input"
            compact
          />
        </div>
        <div>
          <label class="text-xs font-semibold text-muted-foreground">Story Points (budget)</label>
          <input
            data-testid="story.story_points_input"
            :value="props.story.storyPoints"
            type="number"
            min="0"
            placeholder="Optional"
            class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            @input="updateStory({ storyPoints: ($event.target as HTMLInputElement).value ? Number(($event.target as HTMLInputElement).value) : null })"
          />
        </div>
        <div class="flex items-center gap-3">
          <Button data-testid="story.create-button" size="sm" :disabled="!props.canCreate || props.storySaving" @click="emit('create')">
            {{ props.storySaving ? "Creating..." : "Create story" }}
          </Button>
          <span v-if="props.storyError" class="text-xs">{{ props.storyError }}</span>
        </div>
      </div>
    </div>
  </div>
</template>
