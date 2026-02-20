<script setup lang="ts">
import { Button } from "@/components/ui/button";
import { ref } from "vue";
import type { Notification, NotificationPreferences, Project } from "@/lib/api";
import { useI18n } from "@/lib/i18n";

const props = defineProps<{
    activeProjectLabel: string;
    activePage: "board" | "dashboard" | "settings";
    currentUserName: string;
    projectLoading: boolean;
    projects: Project[];
    activeProjectId: string;
    canManageProject: boolean;
    notifications: Notification[];
    notificationsLoading: boolean;
    notificationsUnreadCount: number;
    notificationPreferences: NotificationPreferences;
    notificationPreferencesSaving: boolean;
    liveUpdateMode: "ws" | "polling";
}>();

const emit = defineEmits<{
    (e: "set-page", value: "board" | "dashboard" | "settings"): void;
    (e: "select-project", value: string): void;
    (e: "logout"): void;
    (e: "refresh"): void;
    (e: "open-inbox"): void;
    (e: "inbox-visibility-change", value: boolean): void;
    (e: "mark-notification-read", value: string): void;
    (e: "mark-all-notifications-read"): void;
    (e: "update-notification-preferences", value: Partial<NotificationPreferences>): void;
}>();

const showInbox = ref(false);
const { locale, setLocale, t } = useI18n();

const toggleInbox = () => {
    showInbox.value = !showInbox.value;
    emit("inbox-visibility-change", showInbox.value);
    if (showInbox.value) {
        emit("open-inbox");
    }
};

const closeInbox = () => {
    if (!showInbox.value) return;
    showInbox.value = false;
    emit("inbox-visibility-change", false);
};
</script>

<template>
    <header
        class="relative z-40 mx-auto flex w-full max-w-6xl items-center justify-between px-6 py-5"
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
                    {{ t("header.board") }}
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
                    {{ t("header.dashboard") }}
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
                    {{ t("header.settings") }}
                </button>
            </nav>

            <select
                class="rounded-lg border border-input bg-background px-2 py-1.5 text-xs focus:outline-none focus:ring-2 focus:ring-ring"
                :value="locale"
                @change="setLocale(($event.target as HTMLSelectElement).value as 'en' | 'de')"
            >
                <option value="en">EN</option>
                <option value="de">DE</option>
            </select>

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
                <option value="" disabled>{{ t("header.selectProject") }}</option>
                <option
                    v-for="project in props.projects"
                    :key="project.id"
                    :value="project.id"
                >
                    {{ project.key }} · {{ project.name }}
                </option>
            </select>

            <!-- User + actions -->
            <div class="relative z-[80]">
                <button
                    data-testid="nav.inbox-button"
                    class="relative rounded-lg border border-border bg-background px-2.5 py-1.5 text-xs font-semibold text-muted-foreground transition hover:border-foreground hover:text-foreground"
                    @click="toggleInbox"
                >
                    {{ t("header.inbox") }}
                    <span
                        v-if="props.notificationsUnreadCount > 0"
                        class="absolute -right-1.5 -top-1.5 rounded-full bg-primary px-1.5 py-0.5 text-[9px] font-bold text-primary-foreground"
                    >
                        {{ props.notificationsUnreadCount }}
                    </span>
                </button>
                <div
                    v-if="showInbox"
                    class="fixed inset-0 z-[90] bg-slate-950/45 backdrop-blur-[1px]"
                    @click="closeInbox"
                ></div>
                <div
                    v-if="showInbox"
                    data-testid="nav.inbox-panel"
                    class="absolute right-0 z-[100] mt-2 w-[26rem] rounded-2xl border border-slate-700/80 bg-slate-950 p-3 text-slate-100 shadow-[0_28px_70px_-22px_rgba(0,0,0,0.9)] ring-1 ring-white/10"
                    @click.stop
                >
                    <div class="mb-2 flex items-center justify-between">
                        <p class="text-xs font-semibold uppercase tracking-[0.2em] text-slate-300">
                            {{ t("header.notifications") }}
                        </p>
                        <div class="flex items-center gap-1.5">
                            <button
                                data-testid="nav.inbox-mark-all-button"
                                class="rounded border border-slate-700 bg-slate-900 px-2 py-1 text-[10px] font-semibold uppercase tracking-[0.12em] text-slate-200 transition hover:border-slate-500 hover:text-white"
                                @click="emit('mark-all-notifications-read')"
                            >
                                {{ t("header.markAllRead") }}
                            </button>
                            <button
                                class="rounded border border-slate-700 bg-slate-900 px-2 py-1 text-[10px] font-semibold uppercase tracking-[0.12em] text-slate-200 transition hover:border-slate-500 hover:text-white"
                                @click="closeInbox"
                            >
                                {{ t("header.close") }}
                            </button>
                        </div>
                    </div>

                    <div class="mb-2 grid grid-cols-2 gap-2">
                        <label class="flex items-center gap-2 text-[11px] text-slate-300">
                            <input
                                data-testid="nav.inbox-pref-mention"
                                type="checkbox"
                                :checked="props.notificationPreferences.mentionEnabled"
                                :disabled="props.notificationPreferencesSaving"
                                @change="
                                    emit('update-notification-preferences', {
                                        mentionEnabled: ($event.target as HTMLInputElement).checked,
                                    })
                                "
                            />
                            {{ t("header.mentions") }}
                        </label>
                        <label class="flex items-center gap-2 text-[11px] text-slate-300">
                            <input
                                data-testid="nav.inbox-pref-assignment"
                                type="checkbox"
                                :checked="props.notificationPreferences.assignmentEnabled"
                                :disabled="props.notificationPreferencesSaving"
                                @change="
                                    emit('update-notification-preferences', {
                                        assignmentEnabled: ($event.target as HTMLInputElement).checked,
                                    })
                                "
                            />
                            {{ t("header.assignments") }}
                        </label>
                    </div>

                    <div
                        v-if="props.notificationsLoading"
                        class="rounded border border-slate-700 bg-slate-900 px-3 py-2 text-xs text-slate-300"
                    >
                        {{ t("common.loading") }}
                    </div>
                    <div
                        v-else-if="props.notifications.length === 0"
                        class="rounded border border-slate-700 bg-slate-900 px-3 py-2 text-xs text-slate-300"
                    >
                        {{ t("header.noNotifications") }}
                    </div>
                    <div v-else class="max-h-72 space-y-1 overflow-auto">
                        <div
                            v-for="item in props.notifications"
                            :key="item.id"
                            data-testid="nav.inbox-item"
                            class="rounded-xl border px-2.5 py-2 text-xs"
                            :class="
                                item.readAt
                                    ? 'border-slate-700 bg-slate-900'
                                    : 'border-blue-400/50 bg-blue-500/10'
                            "
                        >
                            <p class="font-medium text-slate-100">{{ item.message }}</p>
                            <p class="mt-0.5 text-[10px] text-slate-400">
                                {{ item.ticketKey }} · {{ item.ticketTitle }}
                            </p>
                            <button
                                v-if="!item.readAt"
                                data-testid="nav.inbox-item-read-button"
                                class="mt-1 rounded border border-slate-700 bg-slate-950 px-2 py-1 text-[10px] font-semibold uppercase tracking-[0.1em] text-slate-200 transition hover:border-slate-500 hover:text-white"
                                @click="emit('mark-notification-read', item.id)"
                            >
                                {{ t("header.markRead") }}
                            </button>
                        </div>
                    </div>
                </div>
            </div>
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
            <span
                class="rounded-full border border-border px-2 py-1 text-[10px] font-semibold uppercase tracking-[0.12em] text-muted-foreground"
                :title="
                    props.liveUpdateMode === 'ws'
                        ? t('header.liveWs')
                        : t('header.livePolling')
                "
            >
                {{ props.liveUpdateMode === "ws" ? "WS" : "POLL" }}
            </span>
            <Button
                data-testid="nav.logout-button"
                variant="ghost"
                size="sm"
                @click="emit('logout')"
                >{{ t("header.logout") }}</Button
            >
        </div>
    </header>
</template>
