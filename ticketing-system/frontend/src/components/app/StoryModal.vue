<script setup lang="ts">
import { Button } from "@/components/ui/button";
import { Modal } from "@/components/ui/modal";
import { MarkdownEditor } from "@/components/ui/markdown-editor";
import { useI18n } from "@/lib/i18n";

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

const { t } = useI18n();

const updateStory = (patch: Partial<StoryDraft>) => {
  emit("update:story", { ...props.story, ...patch });
};
</script>

<template>
  <Modal :show="props.show" test-id="story.modal" @close="emit('close')">
    <div class="flex items-center justify-between">
      <div>
        <p class="text-xs uppercase tracking-[0.3em] text-muted-foreground">
          {{ t("story.modal.label") }}
        </p>
        <h2 class="text-2xl font-semibold">{{ t("story.modal.heading") }}</h2>
      </div>
      <Button variant="ghost" size="sm" @click="emit('close')">{{ t("common.close") }}</Button>
    </div>
    <div class="mt-6 space-y-4">
      <div>
        <label class="text-xs font-semibold text-muted-foreground">{{ t("story.modal.titleLabel") }}</label>
        <input
          data-testid="story.title-input"
          :value="props.story.title"
          type="text"
          :placeholder="t('story.modal.titlePlaceholder')"
          class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
          @input="updateStory({ title: ($event.target as HTMLInputElement).value })"
        />
      </div>
      <div>
        <label class="text-xs font-semibold text-muted-foreground">{{ t("story.modal.descriptionLabel") }}</label>
        <MarkdownEditor
          :model-value="props.story.description"
          @update:model-value="updateStory({ description: $event })"
          :rows="4"
          :placeholder="t('story.modal.descriptionPlaceholder')"
          data-testid="story.description-input"
          compact
        />
      </div>
      <div>
        <label class="text-xs font-semibold text-muted-foreground">{{ t("story.storyPoints") }}</label>
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
          {{ props.storySaving ? t("story.modal.creating") : t("story.modal.create") }}
        </Button>
        <span v-if="props.storyError" class="text-xs">{{ props.storyError }}</span>
      </div>
    </div>
  </Modal>
</template>
