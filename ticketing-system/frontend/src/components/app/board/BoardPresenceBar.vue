<script setup lang="ts">
import { computed } from "vue";
import type { PresenceEntry } from "@/lib/api";

const props = defineProps<{
    users: PresenceEntry[];
}>();

const MAX_SHOWN = 5;
const shown = computed(() => props.users.slice(0, MAX_SHOWN));
const overflow = computed(() => Math.max(0, props.users.length - MAX_SHOWN));

const initials = (name: string) =>
    name
        .split(" ")
        .map((w) => w[0] ?? "")
        .join("")
        .slice(0, 2)
        .toUpperCase() || "?";

const tooltip = (u: PresenceEntry) =>
    u.view === "ticket" && u.ticketId
        ? `${u.userName} — editing ticket`
        : `${u.userName} — viewing board`;
</script>

<template>
    <div v-if="users.length > 0" class="flex items-center gap-1">
        <span class="text-[10px] uppercase tracking-wider text-muted-foreground/60 mr-1">Now:</span>
        <div
            v-for="u in shown"
            :key="u.userId"
            class="flex h-6 w-6 items-center justify-center rounded-full border-2 border-background bg-primary/20 text-[9px] font-bold text-primary -ml-1 first:ml-0 cursor-default select-none"
            :class="u.view === 'ticket' ? 'ring-1 ring-amber-400/60' : ''"
            :title="tooltip(u)"
        >
            {{ initials(u.userName) }}
        </div>
        <span v-if="overflow > 0" class="text-[10px] text-muted-foreground ml-1">+{{ overflow }}</span>
    </div>
</template>
