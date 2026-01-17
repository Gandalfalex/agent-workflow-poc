<script setup lang="ts">
import { Button } from "@/components/ui/button";
import type { Project } from "@/lib/api";

const props = defineProps<{
    activeProjectLabel: string;
    activePage: "board" | "settings";
    currentUserName: string;
    projectLoading: boolean;
    projects: Project[];
    activeProjectId: string;
}>();

const emit = defineEmits<{
    (e: "set-page", value: "board" | "settings"): void;
    (e: "select-project", value: string): void;
    (e: "logout"): void;
    (e: "refresh"): void;
}>();
</script>

<template>
    <header
        class="relative z-10 mx-auto flex w-full max-w-6xl items-center justify-between px-6 py-6"
    >
        <div class="flex items-center gap-3">
            <div class="h-11 w-11 rounded-2xl bg-primary/90 shadow-sm"></div>
            <div>
                <p
                    class="text-xs uppercase tracking-[0.3em] text-muted-foreground"
                >
                    Ops Console
                </p>
                <p class="text-lg font-semibold">
                    {{ props.activeProjectLabel || "Ticketing Workspace" }}
                </p>
            </div>
        </div>
        <div class="flex items-center gap-2">
            <div
                class="flex items-center gap-2 rounded-xl border border-border bg-card/60 px-2 py-1.5"
            >
                <span
                    class="hidden text-[10px] uppercase tracking-[0.2em] text-muted-foreground md:inline-flex"
                >
                    Project
                </span>
                <select
                    class="rounded-lg border border-input bg-background px-2 py-1 text-xs focus:outline-none focus:ring-2 focus:ring-ring"
                    :value="props.activeProjectId"
                    :disabled="
                        props.projectLoading || props.projects.length === 0
                    "
                    @change="
                        emit(
                            'select-project',
                            ($event.target as HTMLSelectElement).value,
                        )
                    "
                >
                    <option value="" disabled>Select project</option>
                    <option
                        v-for="project in props.projects"
                        :key="project.id"
                        :value="project.id"
                    >
                        {{ project.key }} · {{ project.name }}
                    </option>
                </select>
            </div>
            <div
                v-if="props.currentUserName"
                class="hidden flex-col text-right text-xs text-muted-foreground sm:flex"
            >
                <span class="uppercase tracking-[0.2em]">Signed in</span>
                <span class="text-sm font-semibold text-foreground">
                    {{ props.currentUserName }}
                </span>
            </div>
            <Button
                variant="ghost"
                size="sm"
                :disabled="props.activePage === 'board'"
                @click="emit('set-page', 'board')"
            >
                Board
            </Button>
            <Button
                variant="ghost"
                size="sm"
                :disabled="props.activePage === 'settings'"
                @click="emit('set-page', 'settings')"
            >
                Settings
            </Button>
            <Button variant="ghost" size="sm" @click="emit('logout')"
                >Logout</Button
            >
            <Button variant="outline" size="sm" @click="emit('refresh')"
                >Refresh</Button
            >
        </div>
    </header>
</template>
