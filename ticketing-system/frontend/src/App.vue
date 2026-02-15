<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import AppHeader from "@/components/app/AppHeader.vue";
import LoginView from "@/components/app/LoginView.vue";
import { useAdminStore } from "@/stores/admin";
import { useBoardStore } from "@/stores/board";
import { useSessionStore } from "@/stores/session";

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
const activePage = computed<"board" | "settings">(() =>
    route.name === "settings" ? "settings" : "board",
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

const setPage = async (page: "board" | "settings") => {
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
            await boardStore.loadBoard(projectId);
            await boardStore.loadStories(projectId);
            await boardStore.loadWebhooks(projectId);
        } catch (err) {
            handleAuthError(err);
        }
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

onMounted(() => {
    checkSession();
});
</script>

<template>
    <div class="min-h-screen bg-background text-foreground">
        <div class="relative">
            <div
                class="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_top,_rgba(120,160,255,0.18),_transparent_55%)]"
            ></div>
            <div
                class="pointer-events-none absolute inset-0 opacity-40 [background-size:24px_24px] [background-image:linear-gradient(to_right,rgba(255,255,255,0.04)_1px,transparent_1px),linear-gradient(to_bottom,rgba(255,255,255,0.04)_1px,transparent_1px)]"
            ></div>

            <LoginView
                v-if="showLogin"
                v-model:identifier="loginForm.identifier"
                v-model:password="loginForm.password"
                :login-error="loginError"
                :login-busy="loginBusy"
                :is-checking="isChecking"
                :can-login="canLogin"
                @submit="submitLogin"
            />

            <div v-else>
                <AppHeader
                    :active-project-label="activeProjectLabel"
                    :active-page="activePage"
                    :current-user-name="currentUserName"
                    :project-loading="projectLoading"
                    :projects="projects"
                    :active-project-id="activeProjectId"
                    @set-page="setPage"
                    @select-project="selectProject"
                    @logout="performLogout"
                    @refresh="refreshActive"
                />

                <main
                    class="relative z-10 mx-auto flex w-full max-w-6xl flex-col gap-8 px-6 pb-20"
                >
                    <RouterView />
                </main>
            </div>
        </div>
    </div>
</template>
