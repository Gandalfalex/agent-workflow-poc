<script setup lang="ts">
import { Button } from "@/components/ui/button";
import type { Story, TicketPriority, TicketType, WorkflowState } from "@/lib/api";

type NewTicketForm = {
  title: string;
  description: string;
  priority: TicketPriority;
  type: TicketType;
  storyId: string;
  assignee: string;
  stateId: string;
};

const props = defineProps<{
  show: boolean;
  ticket: NewTicketForm;
  states: WorkflowState[];
  stories: Story[];
  priorities: TicketPriority[];
  ticketTypes: TicketType[];
  canSubmit: boolean;
}>();

const emit = defineEmits<{
  (e: "update:ticket", value: NewTicketForm): void;
  (e: "close"): void;
  (e: "create"): void;
}>();

const updateField = (patch: Partial<NewTicketForm>) => {
  emit("update:ticket", { ...props.ticket, ...patch });
};
</script>

<template>
  <div
    v-if="props.show"
    class="fixed inset-0 z-20 flex items-center justify-center bg-black/30 px-6"
    @click.self="emit('close')"
  >
    <div class="w-full max-w-lg rounded-3xl border border-border bg-card p-6 shadow-xl">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-xs uppercase tracking-[0.3em] text-muted-foreground">
            New ticket
          </p>
          <h2 class="text-2xl font-semibold">Create a card</h2>
        </div>
        <Button variant="ghost" size="sm" @click="emit('close')">Close</Button>
      </div>
      <div class="mt-6 space-y-4">
        <div>
          <label class="text-xs font-semibold text-muted-foreground">Title</label>
          <input
            :value="props.ticket.title"
            type="text"
            placeholder="Short summary"
            class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            @input="updateField({ title: ($event.target as HTMLInputElement).value })"
          />
        </div>
        <div>
          <label class="text-xs font-semibold text-muted-foreground"
            >Description</label
          >
          <textarea
            :value="props.ticket.description"
            rows="3"
            placeholder="What needs to happen?"
            class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            @input="updateField({ description: ($event.target as HTMLTextAreaElement).value })"
          ></textarea>
        </div>
        <div class="grid gap-4 sm:grid-cols-2">
          <div>
            <label class="text-xs font-semibold text-muted-foreground">Type</label>
            <select
              :value="props.ticket.type"
              class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              @change="updateField({ type: ($event.target as HTMLSelectElement).value as TicketType })"
            >
              <option v-for="type in props.ticketTypes" :key="type" :value="type">
                {{ type }}
              </option>
            </select>
          </div>
          <div>
            <label class="text-xs font-semibold text-muted-foreground"
              >Priority</label
            >
            <select
              :value="props.ticket.priority"
              class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              @change="updateField({ priority: ($event.target as HTMLSelectElement).value as TicketPriority })"
            >
              <option
                v-for="priority in props.priorities"
                :key="priority"
                :value="priority"
              >
                {{ priority }}
              </option>
            </select>
          </div>
        </div>
        <div class="grid gap-4 sm:grid-cols-2">
          <div>
            <label class="text-xs font-semibold text-muted-foreground">Story</label>
            <select
              :value="props.ticket.storyId"
              class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              @change="updateField({ storyId: ($event.target as HTMLSelectElement).value })"
            >
              <option value="">None</option>
              <option v-for="story in props.stories" :key="story.id" :value="story.id">
                {{ story.title }}
              </option>
            </select>
          </div>
          <div>
            <label class="text-xs font-semibold text-muted-foreground">State</label>
            <select
              :value="props.ticket.stateId"
              class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              @change="updateField({ stateId: ($event.target as HTMLSelectElement).value })"
            >
              <option v-for="state in props.states" :key="state.id" :value="state.id">
                {{ state.name }}
              </option>
            </select>
          </div>
        </div>
        <div>
          <label class="text-xs font-semibold text-muted-foreground">Assignee</label>
          <input
            :value="props.ticket.assignee"
            type="text"
            placeholder="Name"
            class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            @input="updateField({ assignee: ($event.target as HTMLInputElement).value })"
          />
        </div>
      </div>
      <div class="mt-6 flex items-center justify-end gap-3">
        <Button variant="outline" @click="emit('close')">Cancel</Button>
        <Button :disabled="!props.canSubmit" @click="emit('create')">
          Create ticket
        </Button>
      </div>
    </div>
  </div>
</template>
