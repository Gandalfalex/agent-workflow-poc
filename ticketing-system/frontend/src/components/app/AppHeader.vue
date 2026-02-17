<script setup lang="ts">
import { Button } from "@/components/ui/button";
import type { Project } from "@/lib/api";

const props = defineProps<{
    activeProjectLabel: string;
    activePage: "board" | "dashboard" | "settings";
    currentUserName: string;
    projectLoading: boolean;
    projects: Project[];
    activeProjectId: string;
    canManageProject: boolean;
}>();

const emit = defineEmits<{
    (e: "set-page", value: "board" | "dashboard" | "settings"): void;
    (e: "select-project", value: string): void;
    (e: "logout"): void;
    (e: "refresh"): void;
}>();
</script>

<template>
    <header
        class="relative z-10 mx-auto flex w-full max-w-6xl items-center justify-between px-6 py-5"
    >
        <!-- Left: Brand + project name -->
        <div class="flex items-center gap-3">
            <div class="h-10 w-10 rounded-2xl bg-primary/90 shadow-sm"></div>
            <div>
                <p
                    class="text-[10px] uppercase tracking-[0.3em] text-muted-foreground"
                >
                    Ops Console
                </p>
                <p class="text-base font-semibold">
                    {{ props.activeProjectLabel || "Ticketing Workspace" }}
                </p>
            </div>
        </div>

        <!-- Right: Nav + controls -->
        <div class="flex items-center gap-3">
            <!-- Navigation tabs -->
            <nav
                class="flex items-center rounded-xl border border-border bg-card/60 p-1"
            >
                <button
                    data-testid="nav.board-tab"
                    class="rounded-lg px-3 py-1.5 text-xs font-semibold transition"
                    :class="
                        props.activePage === 'board'
                            ? 'bg-primary text-primary-foreground shadow-sm'
                            : 'text-muted-foreground hover:text-foreground'
                    "
                    @click="emit('set-page', 'board')"
                >
                    Board
                </button>
                <button
                    data-testid="nav.dashboard-tab"
                    class="rounded-lg px-3 py-1.5 text-xs font-semibold transition"
                    :class="
                        props.activePage === 'dashboard'
                            ? 'bg-primary text-primary-foreground shadow-sm'
                            : 'text-muted-foreground hover:text-foreground'
                    "
                    @click="emit('set-page', 'dashboard')"
                >
                    Dashboard
                </button>
                <button
                    v-if="props.canManageProject"
                    data-testid="nav.settings-tab"
                    class="rounded-lg px-3 py-1.5 text-xs font-semibold transition"
                    :class="
                        props.activePage === 'settings'
                            ? 'bg-primary text-primary-foreground shadow-sm'
                            : 'text-muted-foreground hover:text-foreground'
                    "
                    @click="emit('set-page', 'settings')"
                >
                    Settings
                </button>
            </nav>

            <!-- Project selector -->
            <select
                data-testid="nav.project-select"
                class="rounded-lg border border-input bg-background px-2 py-1.5 text-xs focus:outline-none focus:ring-2 focus:ring-ring"
                :value="props.activeProjectId"
                :disabled="props.projectLoading || props.projects.length === 0"
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

            <!-- User + actions -->
            <div
                v-if="props.currentUserName"
                class="hidden items-center gap-2 sm:flex"
            >
                <div
                    class="flex h-7 w-7 items-center justify-center rounded-full bg-primary/15 text-[10px] font-bold text-primary"
                    :title="props.currentUserName"
                >
                    {{ props.currentUserName.slice(0, 2).toUpperCase() }}
                </div>
            </div>
            <Button
                data-testid="nav.refresh-button"
                variant="ghost"
                size="sm"
                @click="emit('refresh')"
                >&#x21BB;</Button
            >
            <Button
                data-testid="nav.logout-button"
                variant="ghost"
                size="sm"
                @click="emit('logout')"
                >Logout</Button
            >
        </div>
    </header>
</template>
