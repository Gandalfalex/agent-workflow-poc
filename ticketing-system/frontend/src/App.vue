<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import AppHeader from "@/components/app/AppHeader.vue";
import LoginView from "@/components/app/LoginView.vue";
import { useAdminStore } from "@/stores/admin";
import { useBoardStore } from "@/stores/board";
import { useSessionStore } from "@/stores/session";
import { buildProjectEventsWebSocketUrls } from "@/lib/api";
import type { NotificationPreferences, ProjectLiveEvent } from "@/lib/api";

const adminStore = useAdminStore();
const boardStore = useBoardStore();
const sessionStore = useSessionStore();
const router = useRouter();
const route = useRoute();

const loginForm = ref({
    identifier: "",
    password: "",
});
const loginError = ref("");
const loginBusy = ref(false);

const authStatus = computed(() => sessionStore.status);
const currentUserName = computed(
    () => sessionStore.user?.name || sessionStore.user?.email || "",
);
const projects = computed(() => adminStore.projects);
const projectLoading = computed(() => adminStore.projectStatus === "loading");
const notifications = computed(() => boardStore.notifications);
const notificationsLoading = computed(() => boardStore.notificationsLoading);
const notificationsUnreadCount = computed(
    () => boardStore.notificationsUnreadCount,
);
const notificationPreferences = computed(
    () => boardStore.notificationPreferences,
);
const notificationPreferencesSaving = computed(
    () => boardStore.notificationPreferencesSaving,
);
const inboxOpen = ref(false);

const activeProjectId = computed(() =>
    typeof route.params.projectId === "string" ? route.params.projectId : "",
);
const activeProject = computed(() =>
    projects.value.find((project) => project.id === activeProjectId.value),
);
const activeProjectLabel = computed(() =>
    activeProject.value
        ? `${activeProject.value.key} · ${activeProject.value.name}`
        : "Ticketing Workspace",
);
const activePage = computed<"board" | "dashboard" | "settings">(() =>
    route.name === "settings"
        ? "settings"
        : route.name === "dashboard"
          ? "dashboard"
          : "board",
);

const canLogin = computed(
    () =>
        loginForm.value.identifier.trim().length > 0 &&
        loginForm.value.password.trim().length > 0,
);
const showLogin = computed(() => authStatus.value !== "authenticated");
const isChecking = computed(() => authStatus.value === "checking");

const handleAuthError = (err: unknown) => {
    const error = err as Error & { status?: number };
    if (error.status === 401 || error.status === 403) {
        sessionStore.reset();
        return true;
    }
    return false;
};

const checkSession = async () => {
    loginError.value = "";
    try {
        await sessionStore.checkSession();
        await adminStore.loadProjects();
        if (!activeProjectId.value && projects.value.length > 0) {
            const fallback = projects.value[0]?.id;
            if (fallback) {
                await router.replace({
                    name: "board",
                    params: { projectId: fallback },
                });
            }
        }
    } catch (err) {
        const error = err as Error & { status?: number };
        if (error.status && error.status !== 401 && error.status !== 403) {
            loginError.value = "Backend unavailable. Please try again.";
        }
    }
};

let notificationsTimer: number | null = null;
let projectEventsSocket: WebSocket | null = null;
let wsReconnectTimer: number | null = null;
const wsConnected = ref(false);
const wsFeatureEnabled =
    (import.meta.env.VITE_USE_WS_LIVE_UPDATES || "true").toLowerCase() !==
    "false";
const liveUpdateMode = computed<"ws" | "polling">(() =>
    wsFeatureEnabled && wsConnected.value ? "ws" : "polling",
);

const stopNotificationPolling = () => {
    if (notificationsTimer !== null) {
        window.clearInterval(notificationsTimer);
        notificationsTimer = null;
    }
};

const stopWsReconnectTimer = () => {
    if (wsReconnectTimer !== null) {
        window.clearTimeout(wsReconnectTimer);
        wsReconnectTimer = null;
    }
};

const stopProjectEventsSocket = () => {
    stopWsReconnectTimer();
    if (projectEventsSocket) {
        projectEventsSocket.onopen = null;
        projectEventsSocket.onmessage = null;
        projectEventsSocket.onerror = null;
        projectEventsSocket.onclose = null;
        projectEventsSocket.close();
        projectEventsSocket = null;
    }
    wsConnected.value = false;
};

const startNotificationPolling = () => {
    if (wsFeatureEnabled && wsConnected.value) return;
    stopNotificationPolling();
    notificationsTimer = window.setInterval(async () => {
        if (!activeProjectId.value || showLogin.value) return;
        try {
            await boardStore.loadNotificationUnreadCount(activeProjectId.value);
        } catch (err) {
            handleAuthError(err);
        }
    }, 5000);
};

const handleProjectLiveEvent = async (event: ProjectLiveEvent) => {
    const projectId = activeProjectId.value;
    if (!projectId) return;
    const payload = (event.payload || {}) as Record<string, unknown>;
    switch (event.type) {
        case "notifications.unread_count": {
            const count = payload.count;
            if (typeof count === "number") {
                boardStore.notificationsUnreadCount = count;
                return;
            }
            await boardStore.loadNotificationUnreadCount(projectId);
            return;
        }
        case "notifications.changed":
            await boardStore.loadNotificationUnreadCount(projectId);
            if (inboxOpen.value) {
                await boardStore.loadNotifications(projectId, { limit: 20 });
            }
            return;
        case "heartbeat":
            return;
        case "board.refresh":
            if (activePage.value === "board") {
                await boardStore.loadBoard(projectId);
                await boardStore.loadStories(projectId);
            }
            return;
        case "activity.changed":
            if (activePage.value === "dashboard") {
                await boardStore.loadProjectActivities(projectId);
            }
            return;
        default:
            return;
    }
};

const scheduleWsReconnect = () => {
    if (!wsFeatureEnabled || showLogin.value || !activeProjectId.value) {
        return;
    }
    if (wsReconnectTimer !== null) return;
    wsReconnectTimer = window.setTimeout(() => {
        wsReconnectTimer = null;
        startProjectEventsSocket();
    }, 2500);
};

const startProjectEventsSocket = () => {
    if (!wsFeatureEnabled || showLogin.value || !activeProjectId.value) {
        startNotificationPolling();
        return;
    }

    stopProjectEventsSocket();
    const socketURLs = buildProjectEventsWebSocketUrls(activeProjectId.value);

    const openSocket = (index: number) => {
        if (index >= socketURLs.length) {
            wsConnected.value = false;
            if (showLogin.value) return;
            startNotificationPolling();
            scheduleWsReconnect();
            return;
        }

        const socket = new WebSocket(socketURLs[index]!);
        projectEventsSocket = socket;
        let opened = false;

        socket.onopen = async () => {
            opened = true;
            wsConnected.value = true;
            stopNotificationPolling();
            try {
                await boardStore.loadNotificationUnreadCount(activeProjectId.value);
            } catch (err) {
                handleAuthError(err);
            }
        };

        socket.onmessage = async (message) => {
            try {
                const event = JSON.parse(message.data) as ProjectLiveEvent;
                await handleProjectLiveEvent(event);
            } catch {
                // Ignore malformed or unexpected events.
            }
        };

        socket.onerror = () => {
            socket.close();
        };

        socket.onclose = () => {
            if (!opened && index + 1 < socketURLs.length) {
                openSocket(index + 1);
                return;
            }
            wsConnected.value = false;
            if (showLogin.value) return;
            startNotificationPolling();
            scheduleWsReconnect();
        };
    };

    openSocket(0);
};

const submitLogin = async () => {
    if (!canLogin.value || loginBusy.value) return;
    loginBusy.value = true;
    loginError.value = "";
    try {
        await sessionStore.login(
            loginForm.value.identifier.trim(),
            loginForm.value.password,
        );
        await adminStore.loadProjects();
        if (!activeProjectId.value && projects.value.length > 0) {
            const fallback = projects.value[0]?.id;
            if (fallback) {
                await router.replace({
                    name: "board",
                    params: { projectId: fallback },
                });
            }
        }
    } catch (err) {
        const error = err as Error & { status?: number };
        loginError.value =
            error.status === 401 || error.status === 403
                ? "Invalid credentials."
                : "Unable to sign in.";
    } finally {
        loginBusy.value = false;
    }
};

const performLogout = async () => {
    await sessionStore.logout();
    boardStore.reset();
    adminStore.reset();
    loginForm.value = { identifier: "", password: "" };
    await router.replace({ name: "home" });
};

const selectProject = async (projectId: string) => {
    if (!projectId) return;
    await router.push({
        name: activePage.value,
        params: { projectId },
    });
};

const setPage = async (page: "board" | "dashboard" | "settings") => {
    const targetProjectId = activeProjectId.value || projects.value[0]?.id;
    if (!targetProjectId) {
        await adminStore.loadProjects();
        return;
    }
    await router.push({ name: page, params: { projectId: targetProjectId } });
};

const refreshActive = async () => {
    const projectId = activeProjectId.value;
    if (!projectId) {
        await adminStore.loadProjects();
        return;
    }
    if (activePage.value === "board") {
        try {
            await boardStore.loadCurrentUserRole(projectId);
            await boardStore.loadBoard(projectId);
            await boardStore.loadStories(projectId);
            await boardStore.loadWebhooks(projectId);
            await boardStore.loadNotificationUnreadCount(projectId);
        } catch (err) {
            handleAuthError(err);
        }
        return;
    }
    if (activePage.value === "dashboard") {
        try {
            await boardStore.loadCurrentUserRole(projectId);
            await boardStore.loadDashboardStats(projectId);
        } catch (err) {
            handleAuthError(err);
        }
        return;
    }
    try {
        await boardStore.loadCurrentUserRole(projectId);
    } catch (err) {
        handleAuthError(err);
        return;
    }
    await adminStore.loadProjects();
    await adminStore.loadGroups();
    await adminStore.loadProjectGroups(projectId);
    try {
        await boardStore.loadWebhooks(projectId);
    } catch (err) {
        handleAuthError(err);
    }
};

const openInbox = async () => {
    const projectId = activeProjectId.value;
    if (!projectId) return;
    try {
        await boardStore.loadNotifications(projectId, { limit: 20 });
        await boardStore.loadNotificationPreferences(projectId);
        await boardStore.loadNotificationUnreadCount(projectId);
    } catch (err) {
        handleAuthError(err);
    }
};

const onInboxVisibilityChange = (value: boolean) => {
    inboxOpen.value = value;
};

const markNotificationRead = async (notificationId: string) => {
    const projectId = activeProjectId.value;
    if (!projectId) return;
    try {
        await boardStore.markNotificationRead(projectId, notificationId);
    } catch (err) {
        handleAuthError(err);
    }
};

const markAllNotificationsRead = async () => {
    const projectId = activeProjectId.value;
    if (!projectId) return;
    try {
        await boardStore.markAllNotificationsRead(projectId);
        await boardStore.loadNotifications(projectId, { limit: 20 });
    } catch (err) {
        handleAuthError(err);
    }
};

const updateNotificationPreferences = async (
    payload: Partial<NotificationPreferences>,
) => {
    const projectId = activeProjectId.value;
    if (!projectId) return;
    try {
        await boardStore.updateNotificationPreferences(projectId, payload);
    } catch (err) {
        handleAuthError(err);
    }
};

onMounted(() => {
    checkSession();
    startProjectEventsSocket();
});

onUnmounted(() => {
    stopProjectEventsSocket();
    stopNotificationPolling();
});

watch(
    () => activeProjectId.value,
    async (projectId) => {
        if (!projectId || showLogin.value) return;
        try {
            await boardStore.loadNotificationUnreadCount(projectId);
            await boardStore.loadNotificationPreferences(projectId);
        } catch (err) {
            handleAuthError(err);
        }
        if (wsFeatureEnabled) {
            startProjectEventsSocket();
            return;
        }
        startNotificationPolling();
    },
);

watch(
    () => showLogin.value,
    (value) => {
        if (value) {
            stopProjectEventsSocket();
            stopNotificationPolling();
            return;
        }
        startProjectEventsSocket();
    },
);
</script>

<template>
    <div class="min-h-screen bg-background text-foreground">
        <div class="relative">
            <div
                class="pointer-events-none absolute inset-0 z-0 bg-[radial-gradient(circle_at_top,_rgba(120,160,255,0.18),_transparent_55%)]"
            ></div>
            <div
                class="pointer-events-none absolute inset-0 z-0 opacity-40 [background-size:24px_24px] [background-image:linear-gradient(to_right,rgba(255,255,255,0.04)_1px,transparent_1px),linear-gradient(to_bottom,rgba(255,255,255,0.04)_1px,transparent_1px)]"
            ></div>

            <LoginView
                v-if="showLogin"
                class="relative z-10"
                v-model:identifier="loginForm.identifier"
                v-model:password="loginForm.password"
                :login-error="loginError"
                :login-busy="loginBusy"
                :is-checking="isChecking"
                :can-login="canLogin"
                @submit="submitLogin"
            />

            <div v-else class="relative z-10">
                <AppHeader
                    :active-project-label="activeProjectLabel"
                    :active-page="activePage"
                    :current-user-name="currentUserName"
                    :project-loading="projectLoading"
                    :projects="projects"
                    :active-project-id="activeProjectId"
                    :can-manage-project="boardStore.canManageProject"
                    :notifications="notifications"
                    :notifications-loading="notificationsLoading"
                    :notifications-unread-count="notificationsUnreadCount"
                    :notification-preferences="notificationPreferences"
                    :notification-preferences-saving="notificationPreferencesSaving"
                    :live-update-mode="liveUpdateMode"
                    @set-page="setPage"
                    @select-project="selectProject"
                    @logout="performLogout"
                    @refresh="refreshActive"
                    @open-inbox="openInbox"
                    @inbox-visibility-change="onInboxVisibilityChange"
                    @mark-notification-read="markNotificationRead"
                    @mark-all-notifications-read="markAllNotificationsRead"
                    @update-notification-preferences="
                        updateNotificationPreferences
                    "
                />

                <main
                    class="relative flex w-full flex-col gap-8 px-6 pb-20"
                >
                    <RouterView />
                </main>
            </div>
        </div>
    </div>
</template>
