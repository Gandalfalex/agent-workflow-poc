<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import { Button } from "@/components/ui/button";
import { useAdminStore } from "@/stores/admin";
import { useBoardStore } from "@/stores/board";
import { useSessionStore } from "@/stores/session";
import type {
    Group,
    ProjectRole,
    WebhookEvent,
    WebhookResponse,
} from "@/lib/api";

const props = defineProps<{ projectId: string }>();

const router = useRouter();
const adminStore = useAdminStore();
const boardStore = useBoardStore();
const sessionStore = useSessionStore();

const settingsTab = ref<"projects" | "webhooks">("projects");
const selectedProjectId = ref(props.projectId || "");
const selectedGroupId = ref("");
const actionNotice = ref<{ tone: "success" | "error"; message: string } | null>(
    null,
);
const newProject = ref({
    key: "",
    name: "",
    description: "",
});
const newGroup = ref({
    name: "",
    description: "",
});
const newProjectGroup = ref({
    groupId: "",
    role: "contributor" as ProjectRole,
});
const newGroupMember = ref({
    userId: "",
});
const userSearchQuery = ref("");

const webhookEvents: WebhookEvent[] = [
    "ticket.created",
    "ticket.updated",
    "ticket.deleted",
    "ticket.state_changed",
];
const newWebhook = ref({
    url: "",
    events: ["ticket.created"] as WebhookEvent[],
    secret: "",
    enabled: true,
});

const projects = computed(() => adminStore.projects);
const groups = computed(() => adminStore.groups);
const projectGroups = computed(() => adminStore.projectGroups);
const groupMembers = computed(() => adminStore.groupMembers);
const userResults = computed(() => adminStore.userResults);
const projectLoading = computed(() => adminStore.projectStatus === "loading");
const groupLoading = computed(() => adminStore.groupStatus === "loading");
const projectGroupLoading = computed(
    () => adminStore.projectGroupStatus === "loading",
);
const groupMemberLoading = computed(
    () => adminStore.groupMemberStatus === "loading",
);
const userSearchLoading = computed(
    () => adminStore.userSearchStatus === "loading",
);
const projectError = computed(() => adminStore.projectError);
const groupError = computed(() => adminStore.groupError);
const projectGroupError = computed(() => adminStore.projectGroupError);
const groupMemberError = computed(() => adminStore.groupMemberError);
const userSearchError = computed(() => adminStore.userSearchError);

const webhooks = computed(() => boardStore.webhooks);
const webhookLoading = computed(() => boardStore.webhookLoading);
const webhookError = computed(() => boardStore.webhookError);
const webhookTestStatus = computed(() => boardStore.webhookTestStatus);
const groupLookup = computed<Record<string, Group>>(() => {
    const map: Record<string, Group> = {};
    groups.value.forEach((group) => {
        map[group.id] = group;
    });
    return map;
});
const selectedProjectLabel = computed(() => {
    const project = projects.value.find(
        (item) => item.id === selectedProjectId.value,
    );
    if (!project) return "No project selected";
    return `${project.key} · ${project.name}`;
});
const selectedGroupLabel = computed(() => {
    const group = groups.value.find(
        (item) => item.id === selectedGroupId.value,
    );
    return group?.name || "No group selected";
});

const canCreateProject = computed(
    () =>
        newProject.value.key.trim().length > 0 &&
        newProject.value.name.trim().length > 0,
);
const canCreateGroup = computed(() => newGroup.value.name.trim().length > 0);
const canAssignGroup = computed(
    () =>
        newProjectGroup.value.groupId.trim().length > 0 &&
        newProjectGroup.value.role.trim().length > 0,
);
const canAddGroupMember = computed(
    () => newGroupMember.value.userId.trim().length > 0,
);
const canCreateWebhook = computed(
    () =>
        newWebhook.value.url.trim().length > 0 &&
        newWebhook.value.events.length > 0,
);

const handleAuthError = (err: unknown) => {
    const error = err as Error & { status?: number };
    if (error.status === 401 || error.status === 403) {
        sessionStore.reset();
        return true;
    }
    return false;
};

const setNotice = (tone: "success" | "error", message: string) => {
    actionNotice.value = { tone, message };
};

const isSubmitShortcut = (event: KeyboardEvent) =>
    event.key === "Enter" && (event.metaKey || event.ctrlKey);

const submitWithShortcut = (
    event: KeyboardEvent,
    submit: () => void | Promise<void>,
) => {
    if (!isSubmitShortcut(event)) return;
    event.preventDefault();
    void submit();
};

const clearSearchState = () => {
    userSearchQuery.value = "";
    adminStore.clearUserResults();
};

const onGlobalKeydown = (event: KeyboardEvent) => {
    if (event.key !== "Escape") return;
    let changed = false;
    if (actionNotice.value) {
        actionNotice.value = null;
        changed = true;
    }
    if (userSearchQuery.value || userResults.value.length > 0) {
        clearSearchState();
        changed = true;
    }
    if (changed) {
        event.preventDefault();
    }
};

const loadProjects = async () => {
    await adminStore.loadProjects();
    if (!selectedProjectId.value && projects.value.length > 0) {
        selectedProjectId.value =
            props.projectId || projects.value[0]?.id || "";
    }
    if (
        selectedProjectId.value &&
        !projects.value.some((item) => item.id === selectedProjectId.value)
    ) {
        selectedProjectId.value = projects.value[0]?.id || "";
    }
};

const loadGroups = async () => {
    await adminStore.loadGroups();
    if (!selectedGroupId.value && groups.value.length > 0) {
        selectedGroupId.value = groups.value[0]?.id ?? "";
    }
};

const loadProjectGroups = async (projectId?: string) => {
    const id = projectId || selectedProjectId.value;
    if (!id) return;
    await adminStore.loadProjectGroups(id);
};

const loadGroupMembers = async () => {
    if (!selectedGroupId.value) {
        adminStore.clearGroupMembers();
        return;
    }
    await adminStore.loadGroupMembers(selectedGroupId.value);
};

const loadWebhooks = async (projectId?: string) => {
    const id = projectId || selectedProjectId.value;
    if (!id) return;
    try {
        await boardStore.loadWebhooks(id);
    } catch (err) {
        handleAuthError(err);
    }
};

const resetNewWebhook = () => {
    newWebhook.value = {
        url: "",
        events: ["ticket.created"],
        secret: "",
        enabled: true,
    };
};

const createWebhookSubmit = async () => {
    if (!canCreateWebhook.value || webhookLoading.value) return;
    const id = selectedProjectId.value;
    if (!id) return;
    try {
        await boardStore.createWebhook(id, {
            url: newWebhook.value.url.trim(),
            events: newWebhook.value.events,
            enabled: newWebhook.value.enabled,
            secret: newWebhook.value.secret.trim() || undefined,
        });
        resetNewWebhook();
        setNotice("success", "Webhook created.");
    } catch (err) {
        if (!handleAuthError(err)) {
            setNotice("error", "Unable to create webhook.");
        }
    }
};

const toggleWebhook = async (hook: WebhookResponse) => {
    const id = selectedProjectId.value;
    if (!id) return;
    try {
        await boardStore.updateWebhook(id, hook.id, {
            enabled: !hook.enabled,
        });
        setNotice(
            "success",
            hook.enabled ? "Webhook disabled." : "Webhook enabled.",
        );
    } catch (err) {
        if (!handleAuthError(err)) {
            setNotice("error", "Unable to update webhook.");
        }
    }
};

const removeWebhook = async (hook: WebhookResponse) => {
    const id = selectedProjectId.value;
    if (!id) return;
    if (!window.confirm(`Remove webhook for ${hook.url}?`)) {
        return;
    }
    try {
        await boardStore.deleteWebhook(id, hook.id);
        setNotice("success", "Webhook removed.");
    } catch (err) {
        if (!handleAuthError(err)) {
            setNotice("error", "Unable to remove webhook.");
        }
    }
};

const sendTestWebhook = async (hook: WebhookResponse) => {
    const id = selectedProjectId.value;
    if (!id) return;
    try {
        await boardStore.testWebhook(id, hook.id, {
            event: hook.events[0] || "ticket.updated",
        });
        setNotice("success", "Test webhook sent.");
    } catch (err) {
        if (!handleAuthError(err)) {
            setNotice("error", "Unable to send test webhook.");
        }
    }
};

const createProjectSubmit = async () => {
    if (!canCreateProject.value || projectLoading.value) return;
    try {
        const created = await adminStore.createProject({
            key: newProject.value.key.trim().toUpperCase(),
            name: newProject.value.name.trim(),
            description: newProject.value.description.trim() || undefined,
        });
        newProject.value = { key: "", name: "", description: "" };
        selectedProjectId.value = created.id;
        setNotice("success", "Project created.");
    } catch (err) {
        if (!handleAuthError(err)) {
            setNotice("error", "Unable to create project.");
        }
    }
};

const createGroupSubmit = async () => {
    if (!canCreateGroup.value || groupLoading.value) return;
    try {
        const created = await adminStore.createGroup({
            name: newGroup.value.name.trim(),
            description: newGroup.value.description.trim() || undefined,
        });
        newGroup.value = { name: "", description: "" };
        if (!selectedGroupId.value) {
            selectedGroupId.value = created.id;
        }
        setNotice("success", "Group created.");
    } catch (err) {
        if (!handleAuthError(err)) {
            setNotice("error", "Unable to create group.");
        }
    }
};

const assignGroupToProject = async () => {
    if (!canAssignGroup.value || projectGroupLoading.value) return;
    if (!selectedProjectId.value) return;
    try {
        await adminStore.assignGroup(
            selectedProjectId.value,
            newProjectGroup.value.groupId,
            newProjectGroup.value.role,
        );
        setNotice("success", "Group assigned to project.");
    } catch (err) {
        if (!handleAuthError(err)) {
            setNotice("error", "Unable to assign group to project.");
        }
    }
};

const updateProjectGroupRole = async (groupId: string, role: ProjectRole) => {
    if (!selectedProjectId.value) return;
    try {
        await adminStore.updateGroupRole(
            selectedProjectId.value,
            groupId,
            role,
        );
        setNotice("success", "Project role updated.");
    } catch (err) {
        if (!handleAuthError(err)) {
            setNotice("error", "Unable to update project role.");
        }
    }
};

const removeGroupFromProject = async (groupId: string) => {
    if (!selectedProjectId.value) return;
    const label = groupLookup.value[groupId]?.name || "this group";
    if (!window.confirm(`Remove ${label} from project access?`)) {
        return;
    }
    try {
        await adminStore.removeGroup(selectedProjectId.value, groupId);
        setNotice("success", "Group removed from project.");
    } catch (err) {
        if (!handleAuthError(err)) {
            setNotice("error", "Unable to remove group from project.");
        }
    }
};

const addMemberToGroup = async (userId?: string) => {
    if (!selectedGroupId.value) return;
    const id = userId || newGroupMember.value.userId;
    if (!id || !id.trim()) return;
    if (groupMemberLoading.value) return;
    try {
        await adminStore.addMember(selectedGroupId.value, id.trim());
        newGroupMember.value.userId = "";
        adminStore.clearUserResults();
        userSearchQuery.value = "";
        setNotice("success", "Member added to group.");
    } catch (err) {
        if (!handleAuthError(err)) {
            setNotice("error", "Unable to add member to group.");
        }
    }
};

const removeMemberFromGroup = async (userId: string) => {
    if (!selectedGroupId.value) return;
    if (!window.confirm("Remove this member from the group?")) {
        return;
    }
    try {
        await adminStore.removeMember(selectedGroupId.value, userId);
        setNotice("success", "Member removed from group.");
    } catch (err) {
        if (!handleAuthError(err)) {
            setNotice("error", "Unable to remove member from group.");
        }
    }
};

const searchUsersSubmit = async () => {
    if (!userSearchQuery.value.trim()) {
        adminStore.clearUserResults();
        return;
    }
    await adminStore.searchUsers(userSearchQuery.value.trim());
    // Apply fuzzy sorting to results
    sortUserResultsByRelevance(userSearchQuery.value.trim());
};

// Fuzzy match scoring for better search results
const fuzzyScore = (query: string, text: string): number => {
    const q = query.toLowerCase();
    const t = text.toLowerCase();

    // Exact match gets highest score
    if (t === q) return 1000;
    if (t.startsWith(q)) return 500;

    // Consecutive characters match
    const consecutiveIndex = t.indexOf(q);
    if (consecutiveIndex !== -1) return 300 + (100 - consecutiveIndex);

    // Fuzzy match score based on character positions
    let score = 0;
    let queryIdx = 0;
    let lastFoundIndex = -1;

    for (let i = 0; i < t.length && queryIdx < q.length; i++) {
        if (t[i] === q[queryIdx]) {
            const gap = i - lastFoundIndex - 1;
            score += Math.max(100 - gap * 10, 10);
            lastFoundIndex = i;
            queryIdx++;
        }
    }

    // If not all characters matched, return 0
    if (queryIdx !== q.length) return 0;

    return score;
};

const sortUserResultsByRelevance = (query: string) => {
    if (adminStore.userResults.length === 0) return;

    const scored = adminStore.userResults.map((user) => ({
        user,
        nameScore: fuzzyScore(query, user.name),
        emailScore: fuzzyScore(query, user.email || ""),
    }));

    scored.sort((a, b) => {
        const aScore = Math.max(a.nameScore, a.emailScore);
        const bScore = Math.max(b.nameScore, b.emailScore);
        return bScore - aScore;
    });

    adminStore.userResults = scored.map((s) => s.user);
};

onMounted(async () => {
    window.addEventListener("keydown", onGlobalKeydown);
    await loadProjects();
    if (props.projectId && props.projectId !== selectedProjectId.value) {
        selectedProjectId.value = props.projectId;
    }
    await loadGroups();
    if (selectedProjectId.value) {
        await loadProjectGroups(selectedProjectId.value);
        await loadWebhooks(selectedProjectId.value);
    }
    if (selectedGroupId.value) {
        await loadGroupMembers();
    }
});

onUnmounted(() => {
    window.removeEventListener("keydown", onGlobalKeydown);
});

watch(
    () => props.projectId,
    async (value) => {
        if (!value) return;
        if (value !== selectedProjectId.value) {
            selectedProjectId.value = value;
        }
        await loadProjectGroups(value);
        await loadWebhooks(value);
    },
);

watch(selectedProjectId, async (value) => {
    if (!value || value === props.projectId) return;
    await router.push({ name: "settings", params: { projectId: value } });
});

watch(selectedGroupId, async () => {
    await loadGroupMembers();
});
</script>

<template>
    <section class="flex flex-wrap items-center justify-between gap-4">
        <div>
            <p class="text-xs uppercase tracking-[0.3em] text-muted-foreground">
                Settings
            </p>
            <h2 class="text-2xl font-semibold">Admin workspace</h2>
        </div>
        <div class="flex items-center gap-2">
            <Button
                variant="ghost"
                size="sm"
                :disabled="settingsTab === 'projects'"
                @click="settingsTab = 'projects'"
            >
                Projects
            </Button>
            <Button
                variant="ghost"
                size="sm"
                :disabled="settingsTab === 'webhooks'"
                @click="settingsTab = 'webhooks'"
            >
                Webhooks
            </Button>
        </div>
    </section>

    <section
        v-if="actionNotice"
        class="rounded-2xl border px-4 py-3 text-sm"
        :class="
            actionNotice.tone === 'success'
                ? 'border-emerald-500/40 bg-emerald-500/10 text-emerald-200'
                : 'border-destructive/40 bg-destructive/10 text-destructive'
        "
    >
        <div class="flex items-center justify-between gap-4">
            <p>{{ actionNotice.message }}</p>
            <Button variant="ghost" size="sm" @click="actionNotice = null">
                Dismiss
            </Button>
        </div>
    </section>

    <section
        v-if="settingsTab === 'projects'"
        class="grid gap-6 lg:grid-cols-[1.1fr_0.9fr]"
    >
        <div class="rounded-3xl border border-border bg-card/80 p-6 shadow-sm">
            <div class="flex items-center justify-between">
                <div>
                    <p
                        class="text-xs uppercase tracking-[0.3em] text-muted-foreground"
                    >
                        Projects
                    </p>
                    <h2 class="text-lg font-semibold">Select board</h2>
                </div>
                <Button
                    variant="ghost"
                    size="sm"
                    :disabled="projectLoading"
                    @click="loadProjects"
                >
                    {{ projectLoading ? "Loading..." : "Reload" }}
                </Button>
            </div>
            <div class="mt-4 grid gap-3">
                <label class="text-xs font-semibold text-muted-foreground"
                    >Project</label
                >
                <select
                    v-model="selectedProjectId"
                    class="w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                    :disabled="projectLoading || projects.length === 0"
                >
                    <option value="" disabled>Select a project</option>
                    <option
                        v-for="project in projects"
                        :key="project.id"
                        :value="project.id"
                    >
                        {{ project.key }} · {{ project.name }}
                    </option>
                </select>
                <p v-if="projectError" class="text-xs text-destructive">
                    {{ projectError }}
                </p>
                <p
                    v-if="!projectLoading && projects.length === 0"
                    class="text-xs text-muted-foreground"
                >
                    No projects yet. Create the first one.
                </p>
                <p v-else class="text-xs text-muted-foreground">
                    Active: {{ selectedProjectLabel }}
                </p>
            </div>
        </div>
        <div class="rounded-3xl border border-border bg-card/80 p-6 shadow-sm">
            <div>
                <p
                    class="text-xs uppercase tracking-[0.3em] text-muted-foreground"
                >
                    New project
                </p>
                <h2 class="text-lg font-semibold">Create a workspace</h2>
            </div>
            <div class="mt-4 grid gap-3">
                <label class="text-xs font-semibold text-muted-foreground"
                    >Project key</label
                >
                <input
                    v-model="newProject.key"
                    type="text"
                    maxlength="4"
                    placeholder="PROJ"
                    class="w-full rounded-xl border border-input bg-background px-3 py-2 text-sm uppercase focus:outline-none focus:ring-2 focus:ring-ring"
                    @keydown="
                        submitWithShortcut(
                            $event as KeyboardEvent,
                            createProjectSubmit,
                        )
                    "
                />
                <label class="text-xs font-semibold text-muted-foreground"
                    >Name</label
                >
                <input
                    v-model="newProject.name"
                    type="text"
                    placeholder="Payments platform"
                    class="w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                    @keydown="
                        submitWithShortcut(
                            $event as KeyboardEvent,
                            createProjectSubmit,
                        )
                    "
                />
                <label class="text-xs font-semibold text-muted-foreground"
                    >Description</label
                >
                <input
                    v-model="newProject.description"
                    type="text"
                    placeholder="Optional summary"
                    class="w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                    @keydown="
                        submitWithShortcut(
                            $event as KeyboardEvent,
                            createProjectSubmit,
                        )
                    "
                />
                <Button
                    size="sm"
                    :disabled="!canCreateProject || projectLoading"
                    @click="createProjectSubmit"
                >
                    Create project
                </Button>
            </div>
        </div>
    </section>

    <section
        v-if="settingsTab === 'projects'"
        class="rounded-3xl border border-border bg-card/80 p-6 shadow-sm"
    >
        <div class="flex items-center justify-between">
            <div>
                <p
                    class="text-xs uppercase tracking-[0.3em] text-muted-foreground"
                >
                    Access
                </p>
                <h3 class="text-xl font-semibold">Groups and members</h3>
                <p class="mt-2 text-sm text-muted-foreground">
                    Create groups, manage membership, and assign to projects.
                </p>
            </div>
            <Button
                variant="ghost"
                size="sm"
                :disabled="groupLoading || projectGroupLoading"
                @click="
                    loadGroups();
                    loadProjectGroups();
                "
            >
                {{
                    groupLoading || projectGroupLoading
                        ? "Loading..."
                        : "Reload"
                }}
            </Button>
        </div>
        <div class="mt-6 grid gap-6 lg:grid-cols-[1fr_1fr]">
            <!-- Left: Group Selection & User Management -->
            <div class="grid gap-4">
                <div class="grid gap-3">
                    <p class="text-xs font-semibold text-muted-foreground">
                        Groups
                    </p>
                    <select
                        v-model="selectedGroupId"
                        class="w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                        :disabled="groupLoading || groups.length === 0"
                    >
                        <option value="" disabled>Select a group</option>
                        <option
                            v-for="group in groups"
                            :key="group.id"
                            :value="group.id"
                        >
                            {{ group.name }}
                        </option>
                    </select>
                    <p v-if="groupError" class="text-xs text-destructive">
                        {{ groupError }}
                    </p>
                    <p class="text-xs text-muted-foreground">
                        Active group: {{ selectedGroupLabel }}
                    </p>
                </div>

                <!-- Create Group Details Submenu -->
                <details
                    class="rounded-xl border border-border bg-background px-3 py-2 text-xs"
                >
                    <summary
                        class="cursor-pointer text-xs font-semibold text-muted-foreground hover:text-foreground"
                    >
                        ➕ Create new group
                    </summary>
                    <div class="mt-3 grid gap-2">
                        <input
                            v-model="newGroup.name"
                            type="text"
                            placeholder="Support team"
                            class="w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                            @keydown="
                                submitWithShortcut(
                                    $event as KeyboardEvent,
                                    createGroupSubmit,
                                )
                            "
                        />
                        <input
                            v-model="newGroup.description"
                            type="text"
                            placeholder="Optional description"
                            class="w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                            @keydown="
                                submitWithShortcut(
                                    $event as KeyboardEvent,
                                    createGroupSubmit,
                                )
                            "
                        />
                        <Button
                            size="sm"
                            :disabled="!canCreateGroup || groupLoading"
                            @click="createGroupSubmit"
                        >
                            Create group
                        </Button>
                    </div>
                </details>

                <!-- User Search & Add -->
                <div class="rounded-xl border border-border bg-background p-4">
                    <p class="mb-3 text-xs font-semibold text-muted-foreground">
                        Add members
                    </p>

                    <div
                        v-if="!selectedGroupId"
                        class="text-xs text-muted-foreground"
                    >
                        Select a group to add members
                    </div>

                    <div v-else class="grid gap-3">
                        <!-- Search input -->
                        <div>
                            <div class="flex gap-2">
                                <input
                                    v-model="userSearchQuery"
                                    type="text"
                                    placeholder="Fuzzy search: name, email (e.g., 'ich', 'admin', 'ich@ich')"
                                    class="flex-1 rounded-lg border border-input bg-background px-2 py-1 text-xs focus:outline-none focus:ring-2 focus:ring-ring"
                                    @keyup.enter="searchUsersSubmit"
                                    @keydown.esc.prevent="clearSearchState"
                                />
                                <Button
                                    variant="outline"
                                    size="sm"
                                    :disabled="userSearchLoading"
                                    @click="searchUsersSubmit"
                                >
                                    {{ userSearchLoading ? "..." : "Search" }}
                                </Button>
                            </div>
                            <p class="mt-1 text-[10px] text-muted-foreground">
                                Press Enter or click Search. Supports fuzzy
                                matching (e.g., 'ich' matches 'ich@ich.ich')
                            </p>
                        </div>

                        <!-- Search Results or Manual Entry -->
                        <div
                            v-if="userResults.length > 0"
                            class="max-h-48 overflow-auto space-y-1"
                        >
                            <div
                                v-for="user in userResults"
                                :key="user.id"
                                class="flex items-center justify-between rounded-lg border border-border bg-muted/50 px-2 py-1.5 text-xs hover:bg-muted"
                            >
                                <div>
                                    <p class="font-semibold">{{ user.name }}</p>
                                    <p
                                        class="text-[11px] text-muted-foreground"
                                    >
                                        {{ user.email || user.id }}
                                    </p>
                                </div>
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    @click="addMemberToGroup(user.id)"
                                >
                                    Add
                                </Button>
                            </div>
                        </div>

                        <!-- Manual ID Entry -->
                        <div class="space-y-2 border-t border-border pt-3">
                            <label
                                class="text-[11px] font-semibold text-muted-foreground"
                            >
                                Or add by User ID
                            </label>
                            <div class="flex gap-2">
                                <input
                                    v-model="newGroupMember.userId"
                                    type="text"
                                    placeholder="User ID"
                                    class="flex-1 rounded-lg border border-input bg-background px-2 py-1 text-xs focus:outline-none focus:ring-2 focus:ring-ring"
                                    @keydown="
                                        submitWithShortcut(
                                            $event as KeyboardEvent,
                                            () =>
                                                addMemberToGroup(
                                                    newGroupMember.userId,
                                                ),
                                        )
                                    "
                                />
                                <Button
                                    size="sm"
                                    :disabled="!canAddGroupMember"
                                    @click="
                                        addMemberToGroup(newGroupMember.userId)
                                    "
                                >
                                    Add
                                </Button>
                            </div>
                        </div>

                        <p
                            v-if="userSearchError"
                            class="text-xs text-destructive"
                        >
                            {{ userSearchError }}
                        </p>
                        <p
                            v-if="groupMemberError"
                            class="text-xs text-destructive"
                        >
                            {{ groupMemberError }}
                        </p>
                    </div>
                </div>

                <!-- Group Members Table -->
                <div
                    v-if="selectedGroupId"
                    class="rounded-xl border border-border bg-background overflow-hidden"
                >
                    <div class="bg-muted px-3 py-2">
                        <p class="text-xs font-semibold text-muted-foreground">
                            Members ({{ groupMembers.length }})
                        </p>
                    </div>
                    <div
                        v-if="groupMemberLoading"
                        class="px-3 py-2 text-xs text-muted-foreground"
                    >
                        Loading...
                    </div>
                    <div
                        v-else-if="groupMembers.length === 0"
                        class="px-3 py-2 text-xs text-muted-foreground"
                    >
                        No members yet
                    </div>
                    <table v-else class="w-full text-xs">
                        <tbody>
                            <tr
                                v-for="member in groupMembers"
                                :key="member.userId"
                                class="border-t border-border hover:bg-muted/50"
                            >
                                <td class="px-3 py-2">
                                    <div class="font-semibold">
                                        {{ member.user?.name || member.userId }}
                                    </div>
                                    <div
                                        class="text-[11px] text-muted-foreground"
                                    >
                                        {{
                                            member.user?.email || member.userId
                                        }}
                                    </div>
                                </td>
                                <td class="px-3 py-2 text-right">
                                    <Button
                                        variant="ghost"
                                        size="sm"
                                        :disabled="groupMemberLoading"
                                        @click="
                                            removeMemberFromGroup(member.userId)
                                        "
                                    >
                                        Remove
                                    </Button>
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>

            <!-- Right: Project Assignment -->
            <div class="grid gap-4">
                <div
                    class="rounded-xl border border-border bg-background overflow-hidden"
                >
                    <div class="bg-muted px-3 py-2">
                        <p class="text-xs font-semibold text-muted-foreground">
                            Project access
                        </p>
                    </div>
                    <div
                        v-if="projectGroupLoading"
                        class="px-3 py-2 text-xs text-muted-foreground"
                    >
                        Loading...
                    </div>
                    <div
                        v-else-if="projectGroups.length === 0"
                        class="px-3 py-2 text-xs text-muted-foreground"
                    >
                        No groups assigned yet
                    </div>
                    <table v-else class="w-full text-xs">
                        <thead class="bg-muted/50">
                            <tr class="text-[11px] uppercase tracking-[0.1em]">
                                <th class="px-3 py-2 text-left">Group</th>
                                <th class="px-3 py-2 text-left">Role</th>
                                <th class="px-3 py-2 text-right">Action</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr
                                v-for="projectGroup in projectGroups"
                                :key="projectGroup.groupId"
                                class="border-t border-border hover:bg-muted/50"
                            >
                                <td class="px-3 py-2">
                                    <div class="font-semibold">
                                        {{
                                            groupLookup[projectGroup.groupId]
                                                ?.name
                                        }}
                                    </div>
                                </td>
                                <td class="px-3 py-2">
                                    <select
                                        class="rounded-md border border-input bg-background px-2 py-1 text-xs"
                                        :value="projectGroup.role"
                                        @change="
                                            updateProjectGroupRole(
                                                projectGroup.groupId,
                                                (
                                                    $event.target as HTMLSelectElement
                                                ).value as ProjectRole,
                                            )
                                        "
                                    >
                                        <option value="admin">Admin</option>
                                        <option value="contributor">
                                            Contributor
                                        </option>
                                        <option value="viewer">Viewer</option>
                                    </select>
                                </td>
                                <td class="px-3 py-2 text-right">
                                    <Button
                                        variant="ghost"
                                        size="sm"
                                        :disabled="projectGroupLoading"
                                        @click="
                                            removeGroupFromProject(
                                                projectGroup.groupId,
                                            )
                                        "
                                    >
                                        Remove
                                    </Button>
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>

                <!-- Assign Group to Project -->
                <div class="rounded-xl border border-border bg-background p-4">
                    <p class="mb-3 text-xs font-semibold text-muted-foreground">
                        Assign group to project
                    </p>
                    <div class="grid gap-2">
                        <select
                            v-model="newProjectGroup.groupId"
                            class="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                            :disabled="groupLoading || groups.length === 0"
                        >
                            <option value="" disabled>Choose group</option>
                            <option
                                v-for="group in groups"
                                :key="group.id"
                                :value="group.id"
                            >
                                {{ group.name }}
                            </option>
                        </select>
                        <select
                            v-model="newProjectGroup.role"
                            class="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                        >
                            <option value="admin">Admin</option>
                            <option value="contributor">Contributor</option>
                            <option value="viewer">Viewer</option>
                        </select>
                        <Button
                            size="sm"
                            :disabled="
                                !selectedProjectId ||
                                !canAssignGroup ||
                                projectGroupLoading
                            "
                            @click="assignGroupToProject"
                        >
                            Add to project
                        </Button>
                        <p
                            v-if="projectGroupError"
                            class="text-xs text-destructive"
                        >
                            {{ projectGroupError }}
                        </p>
                    </div>
                </div>
            </div>
        </div>
    </section>

    <section
        v-if="settingsTab === 'webhooks'"
        class="grid gap-6 lg:grid-cols-[1.1fr_0.9fr]"
    >
        <div class="rounded-3xl border border-border bg-card/80 p-6 shadow-sm">
            <p class="text-xs uppercase tracking-[0.3em] text-muted-foreground">
                Webhooks
            </p>
            <h2 class="mt-2 text-2xl font-semibold">
                Send ticket events to n8n
            </h2>
            <p class="mt-2 text-sm text-muted-foreground">
                Add endpoints to push ticket updates to your automation flows.
            </p>
            <p class="mt-2 text-xs text-muted-foreground">
                Target project: {{ selectedProjectLabel }}
            </p>
            <div class="mt-6 space-y-4">
                <div>
                    <label class="text-xs font-semibold text-muted-foreground"
                        >Endpoint URL</label
                    >
                    <input
                        v-model="newWebhook.url"
                        type="url"
                        placeholder="https://your-n8n-webhook"
                        class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                        @keydown="
                            submitWithShortcut(
                                $event as KeyboardEvent,
                                createWebhookSubmit,
                            )
                        "
                    />
                </div>
                <div>
                    <label class="text-xs font-semibold text-muted-foreground"
                        >Events</label
                    >
                    <div class="mt-2 grid gap-2 sm:grid-cols-2">
                        <label
                            v-for="event in webhookEvents"
                            :key="event"
                            class="flex items-center gap-2 text-xs text-muted-foreground"
                        >
                            <input
                                v-model="newWebhook.events"
                                type="checkbox"
                                :value="event"
                                class="h-4 w-4 rounded border-border text-primary focus:ring-ring"
                            />
                            <span>{{ event }}</span>
                        </label>
                    </div>
                </div>
                <div>
                    <label class="text-xs font-semibold text-muted-foreground"
                        >Signing secret (optional)</label
                    >
                    <input
                        v-model="newWebhook.secret"
                        type="text"
                        placeholder="Shared secret"
                        class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                        @keydown="
                            submitWithShortcut(
                                $event as KeyboardEvent,
                                createWebhookSubmit,
                            )
                        "
                    />
                </div>
                <label
                    class="flex items-center gap-2 text-xs text-muted-foreground"
                >
                    <input
                        v-model="newWebhook.enabled"
                        type="checkbox"
                        class="h-4 w-4 rounded border-border text-primary focus:ring-ring"
                    />
                    Enabled
                </label>
                <div
                    v-if="webhookError"
                    class="rounded-2xl border border-border bg-secondary/60 px-3 py-2 text-xs"
                >
                    {{ webhookError }}
                </div>
                <Button
                    :disabled="
                        !selectedProjectId ||
                        !canCreateWebhook ||
                        webhookLoading
                    "
                    @click="createWebhookSubmit"
                >
                    {{ webhookLoading ? "Saving..." : "Create webhook" }}
                </Button>
            </div>
        </div>

        <div class="rounded-3xl border border-border bg-card/80 p-6 shadow-sm">
            <div class="flex items-center justify-between">
                <div>
                    <p
                        class="text-xs uppercase tracking-[0.3em] text-muted-foreground"
                    >
                        Active hooks
                    </p>
                    <h3 class="mt-2 text-xl font-semibold">
                        {{ webhooks.length }} configured
                    </h3>
                </div>
                <Button variant="outline" size="sm" @click="loadWebhooks">
                    Refresh
                </Button>
            </div>
            <div class="mt-6 space-y-4">
                <div
                    v-if="webhookLoading"
                    class="text-xs text-muted-foreground"
                >
                    Loading webhooks...
                </div>
                <div
                    v-else-if="webhooks.length === 0"
                    class="rounded-2xl border border-border bg-background p-4 text-sm text-muted-foreground"
                >
                    No webhooks yet. Create one to push ticket events to n8n.
                </div>
                <div
                    v-for="hook in webhooks"
                    :key="hook.id"
                    class="rounded-2xl border border-border bg-background p-4"
                >
                    <div
                        class="flex flex-wrap items-center justify-between gap-3"
                    >
                        <div>
                            <p class="text-xs text-muted-foreground">
                                Endpoint
                            </p>
                            <p class="text-sm font-semibold break-all">
                                {{ hook.url }}
                            </p>
                            <p class="mt-1 text-xs text-muted-foreground">
                                Status:
                                {{ hook.enabled ? "Enabled" : "Disabled" }}
                            </p>
                        </div>
                        <div class="flex flex-wrap items-center gap-2">
                            <Button
                                variant="outline"
                                size="sm"
                                :disabled="webhookLoading"
                                @click="toggleWebhook(hook)"
                            >
                                {{ hook.enabled ? "Disable" : "Enable" }}
                            </Button>
                            <Button
                                variant="outline"
                                size="sm"
                                :disabled="webhookLoading"
                                @click="sendTestWebhook(hook)"
                            >
                                Send test
                            </Button>
                            <Button
                                variant="ghost"
                                size="sm"
                                :disabled="webhookLoading"
                                @click="removeWebhook(hook)"
                            >
                                Remove
                            </Button>
                        </div>
                    </div>
                    <div class="mt-3 flex flex-wrap gap-2">
                        <span
                            v-for="event in hook.events"
                            :key="event"
                            class="rounded-full border border-border bg-muted px-3 py-1 text-[11px] uppercase tracking-[0.2em] text-muted-foreground"
                        >
                            {{ event }}
                        </span>
                    </div>
                    <div v-if="webhookTestStatus[hook.id]" class="mt-3 text-xs">
                        {{
                            webhookTestStatus[hook.id]?.message ||
                            (webhookTestStatus[hook.id]?.success
                                ? "Test sent."
                                : "Test failed.")
                        }}
                    </div>
                </div>
            </div>
        </div>
    </section>
</template>
