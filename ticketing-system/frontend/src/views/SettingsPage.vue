<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import { Button } from "@/components/ui/button";
import { useAdminStore } from "@/stores/admin";
import { useBoardStore } from "@/stores/board";
import { useSessionStore } from "@/stores/session";
import {
    createAdminUser,
    exportProjectReportingSnapshot,
    getProjectAiTriageSettings,
    listProjectCapacitySettings,
    replaceProjectCapacitySettings,
    syncUsersFromIdentityProvider,
    updateProject,
    updateProjectAiTriageSettings,
} from "@/lib/api";
import { useI18n } from "@/lib/i18n";
import type {
    AiTriageSettings,
    CapacitySetting,
    Group,
    ProjectRole,
    ReportingExportFormat,
    WebhookDelivery,
    WebhookEvent,
    WebhookResponse,
    WorkflowState,
} from "@/lib/api";

const props = defineProps<{ projectId: string }>();

const router = useRouter();
const adminStore = useAdminStore();
const boardStore = useBoardStore();
const sessionStore = useSessionStore();
const { t } = useI18n();

const settingsTab = ref<
    "projects" | "users" | "webhooks" | "workflow" | "reporting" | "sprints"
>(
    "projects",
);
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
const showCreateGroup = ref(false);
const syncingUsers = ref(false);
const creatingUser = ref(false);
const newAdminUser = ref({
    username: "",
    email: "",
    firstName: "",
    lastName: "",
    password: "",
});

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

// Workflow editor local state
const workflowStates = computed(() => boardStore.workflowEditorStates);
const workflowLoading = computed(() => boardStore.workflowEditorLoading);
const workflowSaving = computed(() => boardStore.workflowEditorSaving);
const workflowError = computed(() => boardStore.workflowEditorError);
const workflowValidationError = ref("");
const dragIndex = ref<number | null>(null);
const reportingSummary = computed(() => boardStore.projectReportingSummary);
const reportingLoading = computed(() => boardStore.projectReportingLoading);
const reportingFrom = ref("");
const reportingTo = ref("");
const reportingExporting = ref<ReportingExportFormat | null>(null);
const aiTriageSettings = ref<AiTriageSettings>({ enabled: false });
const aiTriageSettingsLoading = ref(false);
const sprintDurationDays = ref<number | null>(null);
const sprintDurationSaving = ref(false);
const capacitySettings = ref<CapacitySetting[]>([]);
const capacitySettingsLoading = ref(false);
const capacitySettingsSaving = ref(false);
const newCapacityRow = ref({ scope: "team" as "team" | "user", label: "", capacity: 0 });

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

const deliveryWebhookId = ref<string | null>(null);
const webhookDeliveries = computed(() => boardStore.webhookDeliveries);
const webhookDeliveriesLoading = computed(
    () => boardStore.webhookDeliveriesLoading,
);
const expandedDeliveryId = ref<string | null>(null);

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
const canCreateAdminUser = computed(
    () =>
        newAdminUser.value.username.trim().length >= 3 &&
        newAdminUser.value.email.trim().length > 0 &&
        newAdminUser.value.password.length >= 8,
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

const loadAiTriageSettings = async (projectId?: string) => {
    const id = projectId || selectedProjectId.value;
    if (!id) return;
    aiTriageSettingsLoading.value = true;
    try {
        aiTriageSettings.value = await getProjectAiTriageSettings(id);
    } catch (err) {
        if (handleAuthError(err)) return;
        aiTriageSettings.value = { enabled: false };
    } finally {
        aiTriageSettingsLoading.value = false;
    }
};

const loadSprintSettings = async () => {
    const id = selectedProjectId.value;
    if (!id) return;
    capacitySettingsLoading.value = true;
    try {
        const project = adminStore.projects.find((p) => p.id === id);
        sprintDurationDays.value = project?.defaultSprintDurationDays ?? null;
        const res = await listProjectCapacitySettings(id);
        capacitySettings.value = res.items;
    } catch (err) {
        if (handleAuthError(err)) return;
    } finally {
        capacitySettingsLoading.value = false;
    }
};

const saveSprintDuration = async () => {
    const id = selectedProjectId.value;
    if (!id) return;
    sprintDurationSaving.value = true;
    try {
        await updateProject(id, {
            defaultSprintDurationDays: sprintDurationDays.value,
        });
        setNotice("success", "Sprint duration saved.");
    } catch (err) {
        if (!handleAuthError(err)) setNotice("error", "Failed to save sprint duration.");
    } finally {
        sprintDurationSaving.value = false;
    }
};

const addCapacityRow = () => {
    if (!newCapacityRow.value.label.trim()) return;
    capacitySettings.value = [
        ...capacitySettings.value,
        {
            id: crypto.randomUUID(),
            projectId: selectedProjectId.value,
            scope: newCapacityRow.value.scope,
            label: newCapacityRow.value.label.trim(),
            capacity: newCapacityRow.value.capacity,
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
        },
    ];
    newCapacityRow.value = { scope: "team", label: "", capacity: 0 };
};

const removeCapacityRow = (index: number) => {
    capacitySettings.value = capacitySettings.value.filter((_, i) => i !== index);
};

const saveCapacitySettings = async () => {
    const id = selectedProjectId.value;
    if (!id) return;
    capacitySettingsSaving.value = true;
    try {
        const res = await replaceProjectCapacitySettings(id, {
            items: capacitySettings.value.map((s) => ({
                scope: s.scope,
                label: s.label,
                capacity: s.capacity,
            })),
        });
        capacitySettings.value = res.items;
        setNotice("success", "Capacity settings saved.");
    } catch (err) {
        if (!handleAuthError(err)) setNotice("error", "Failed to save capacity settings.");
    } finally {
        capacitySettingsSaving.value = false;
    }
};

const toggleAiTriage = async (value: boolean) => {
    const id = selectedProjectId.value;
    if (!id) return;
    aiTriageSettingsLoading.value = true;
    try {
        aiTriageSettings.value = await updateProjectAiTriageSettings(id, {
            enabled: value,
        });
        setNotice("success", `AI triage ${value ? "enabled" : "disabled"}.`);
    } catch (err) {
        if (!handleAuthError(err)) {
            setNotice("error", "Unable to update AI triage setting.");
        }
    } finally {
        aiTriageSettingsLoading.value = false;
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

const toggleDeliveryHistory = async (hook: WebhookResponse) => {
    if (deliveryWebhookId.value === hook.id) {
        deliveryWebhookId.value = null;
        boardStore.webhookDeliveries = [];
        expandedDeliveryId.value = null;
        return;
    }
    deliveryWebhookId.value = hook.id;
    expandedDeliveryId.value = null;
    const id = selectedProjectId.value;
    if (!id) return;
    try {
        await boardStore.loadWebhookDeliveries(id, hook.id);
    } catch (err) {
        handleAuthError(err);
    }
};

const toggleDeliveryDetail = (delivery: WebhookDelivery) => {
    expandedDeliveryId.value =
        expandedDeliveryId.value === delivery.id ? null : delivery.id;
};

const timeAgo = (dateStr: string): string => {
    const diff = Date.now() - new Date(dateStr).getTime();
    const seconds = Math.floor(diff / 1000);
    if (seconds < 60) return `${seconds}s ago`;
    const minutes = Math.floor(seconds / 60);
    if (minutes < 60) return `${minutes}m ago`;
    const hours = Math.floor(minutes / 60);
    if (hours < 24) return `${hours}h ago`;
    const days = Math.floor(hours / 24);
    return `${days}d ago`;
};

const toISODate = (date: Date) => {
    const year = date.getUTCFullYear();
    const month = String(date.getUTCMonth() + 1).padStart(2, "0");
    const day = String(date.getUTCDate()).padStart(2, "0");
    return `${year}-${month}-${day}`;
};

const loadReporting = async () => {
    if (!selectedProjectId.value) return;
    if (!reportingFrom.value || !reportingTo.value) {
        const end = new Date();
        const start = new Date(end);
        start.setUTCDate(end.getUTCDate() - 13);
        reportingFrom.value = toISODate(start);
        reportingTo.value = toISODate(end);
    }
    try {
        await boardStore.loadProjectReportingSummary(selectedProjectId.value, {
            from: reportingFrom.value,
            to: reportingTo.value,
        });
    } catch (err) {
        handleAuthError(err);
    }
};

const downloadBlob = (blob: Blob, filename: string) => {
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.URL.revokeObjectURL(url);
};

const exportReporting = async (format: ReportingExportFormat) => {
    if (!selectedProjectId.value) return;
    if (!reportingFrom.value || !reportingTo.value) {
        await loadReporting();
    }
    reportingExporting.value = format;
    try {
        const file = await exportProjectReportingSnapshot(selectedProjectId.value, {
            from: reportingFrom.value,
            to: reportingTo.value,
            format,
        });
        downloadBlob(file.blob, file.filename);
        setNotice(
            "success",
            format === "csv"
                ? "Reporting CSV exported."
                : "Reporting JSON exported.",
        );
    } catch (err) {
        if (!handleAuthError(err)) {
            setNotice("error", "Unable to export reporting snapshot.");
        }
    } finally {
        reportingExporting.value = null;
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

const syncUsersSubmit = async () => {
    if (syncingUsers.value) return;
    syncingUsers.value = true;
    try {
        const result = await syncUsersFromIdentityProvider();
        setNotice(
            "success",
            `Synced ${result.synced} of ${result.total} users from identity provider.`,
        );
    } catch (err) {
        if (!handleAuthError(err)) {
            setNotice("error", "Unable to sync users from identity provider.");
        }
    } finally {
        syncingUsers.value = false;
    }
};

const createAdminUserSubmit = async () => {
    if (!canCreateAdminUser.value || creatingUser.value) return;
    creatingUser.value = true;
    try {
        const created = await createAdminUser({
            username: newAdminUser.value.username.trim(),
            email: newAdminUser.value.email.trim(),
            firstName: newAdminUser.value.firstName.trim() || undefined,
            lastName: newAdminUser.value.lastName.trim() || undefined,
            password: newAdminUser.value.password,
        });
        newAdminUser.value = {
            username: "",
            email: "",
            firstName: "",
            lastName: "",
            password: "",
        };
        setNotice(
            "success",
            `User ${created.name} created successfully. You can add them to groups now.`,
        );
    } catch (err) {
        if (!handleAuthError(err)) {
            setNotice("error", "Unable to create user.");
        }
    } finally {
        creatingUser.value = false;
    }
};

// Workflow editor methods
const loadWorkflow = async () => {
    if (!selectedProjectId.value) return;
    await boardStore.loadWorkflowEditor(selectedProjectId.value);
};

const addWorkflowState = () => {
    const nextOrder = workflowStates.value.length + 1;
    const isFirst = workflowStates.value.length === 0;
    boardStore.workflowEditorStates = [
        ...workflowStates.value,
        {
            id: `new-${Date.now()}`,
            projectId: selectedProjectId.value,
            name: "",
            order: nextOrder,
            isDefault: isFirst,
            isClosed: false,
        } as WorkflowState,
    ];
    workflowValidationError.value = "";
};

const removeWorkflowState = (index: number) => {
    const state = workflowStates.value[index];
    if (!state) return;
    const isExisting = state.id && !state.id.startsWith("new-");
    if (isExisting && !window.confirm(`Delete state "${state.name}"? Tickets in this state may be affected.`)) {
        return;
    }
    const updated = [...workflowStates.value];
    const wasDefault = updated[index]?.isDefault;
    updated.splice(index, 1);
    if (wasDefault && updated.length > 0) {
        updated[0]!.isDefault = true;
    }
    boardStore.workflowEditorStates = updated;
    workflowValidationError.value = "";
};

const setDefaultState = (index: number) => {
    boardStore.workflowEditorStates = workflowStates.value.map((s, i) => ({
        ...s,
        isDefault: i === index,
    }));
    workflowValidationError.value = "";
};

const updateStateName = (index: number, name: string) => {
    const updated = [...workflowStates.value];
    if (updated[index]) {
        updated[index] = { ...updated[index]!, name };
    }
    boardStore.workflowEditorStates = updated;
    workflowValidationError.value = "";
};

const toggleStateClosed = (index: number) => {
    const updated = [...workflowStates.value];
    if (updated[index]) {
        updated[index] = { ...updated[index]!, isClosed: !updated[index]!.isClosed };
    }
    boardStore.workflowEditorStates = updated;
};

const onDragStart = (index: number) => {
    dragIndex.value = index;
};

const onDragOver = (event: DragEvent, index: number) => {
    event.preventDefault();
    if (dragIndex.value === null || dragIndex.value === index) return;
    const updated = [...workflowStates.value];
    const [moved] = updated.splice(dragIndex.value, 1);
    if (moved) {
        updated.splice(index, 0, moved);
    }
    boardStore.workflowEditorStates = updated;
    dragIndex.value = index;
};

const onDragEnd = () => {
    dragIndex.value = null;
};

const validateWorkflow = (): string | null => {
    const states = workflowStates.value;
    if (states.length === 0) return "At least one state is required.";
    const emptyNames = states.some((s) => !s.name.trim());
    if (emptyNames) return "All states must have a name.";
    const names = states.map((s) => s.name.trim().toLowerCase());
    const uniqueNames = new Set(names);
    if (uniqueNames.size !== names.length) return "State names must be unique.";
    const defaultCount = states.filter((s) => s.isDefault).length;
    if (defaultCount !== 1) return "Exactly one state must be set as default.";
    return null;
};

const saveWorkflow = async () => {
    const error = validateWorkflow();
    if (error) {
        workflowValidationError.value = error;
        return;
    }
    workflowValidationError.value = "";
    if (!selectedProjectId.value) return;
    try {
        await boardStore.saveWorkflowEditor(selectedProjectId.value);
        setNotice("success", "Workflow saved.");
    } catch (err) {
        if (!handleAuthError(err)) {
            setNotice("error", "Unable to save workflow.");
        }
    }
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
        await loadAiTriageSettings(selectedProjectId.value);
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
        await loadAiTriageSettings(value);
        if (settingsTab.value === "reporting") {
            await loadReporting();
        }
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
                {{ t("settings.title") }}
            </p>
            <h2 class="text-2xl font-semibold">{{ t("settings.subtitle") }}</h2>
        </div>
        <div class="flex items-center gap-2">
            <Button
                variant="ghost"
                size="sm"
                :disabled="settingsTab === 'projects'"
                @click="settingsTab = 'projects'"
            >
                {{ t("settings.tab.projects") }}
            </Button>
            <Button
                data-testid="settings.users-tab"
                variant="ghost"
                size="sm"
                :disabled="settingsTab === 'users'"
                @click="settingsTab = 'users'"
            >
                {{ t("settings.tab.users") }}
            </Button>
            <Button
                variant="ghost"
                size="sm"
                :disabled="settingsTab === 'webhooks'"
                @click="settingsTab = 'webhooks'"
            >
                {{ t("settings.tab.webhooks") }}
            </Button>
            <Button
                variant="ghost"
                size="sm"
                :disabled="settingsTab === 'workflow'"
                @click="settingsTab = 'workflow'; loadWorkflow()"
            >
                {{ t("settings.tab.workflow") }}
            </Button>
            <Button
                data-testid="settings.reporting-tab"
                variant="ghost"
                size="sm"
                :disabled="settingsTab === 'reporting'"
                @click="settingsTab = 'reporting'; loadReporting()"
            >
                {{ t("settings.tab.reporting") }}
            </Button>
            <Button
                data-testid="settings.sprints-tab"
                variant="ghost"
                size="sm"
                :disabled="settingsTab === 'sprints'"
                @click="settingsTab = 'sprints'; loadSprintSettings()"
            >
                Sprints
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
                <label
                    class="mt-2 inline-flex items-center gap-2 text-xs text-muted-foreground"
                >
                    <input
                        data-testid="settings.ai-triage-toggle"
                        type="checkbox"
                        :checked="aiTriageSettings.enabled"
                        :disabled="!selectedProjectId || aiTriageSettingsLoading"
                        @change="
                            toggleAiTriage(
                                ($event.target as HTMLInputElement).checked,
                            )
                        "
                    />
                    Enable AI triage copilot for this project
                </label>
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

                <!-- Create Group Collapsible -->
                <div
                    class="rounded-xl border border-border bg-background px-3 py-2 text-xs"
                >
                    <button
                        type="button"
                        class="w-full text-left text-xs font-semibold text-muted-foreground hover:text-foreground transition"
                        @click="showCreateGroup = !showCreateGroup"
                    >
                        {{ showCreateGroup ? '&minus; Create new group' : '+ Create new group' }}
                    </button>
                    <div v-if="showCreateGroup" class="mt-3 grid gap-2">
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
                </div>

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
                        <div
                            class="rounded-lg border border-border bg-muted/20 px-3 py-2"
                        >
                            <div
                                class="flex flex-wrap items-center justify-between gap-2"
                            >
                                <p class="text-[11px] text-muted-foreground">
                                    New users missing from search? Sync users
                                    from identity provider first.
                                </p>
                                <Button
                                    variant="outline"
                                    size="sm"
                                    data-testid="settings.sync-users-button"
                                    :disabled="syncingUsers"
                                    @click="syncUsersSubmit"
                                >
                                    {{
                                        syncingUsers
                                            ? "Syncing..."
                                            : "Sync users"
                                    }}
                                </Button>
                            </div>
                        </div>

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
        v-if="settingsTab === 'users'"
        class="grid gap-6 lg:grid-cols-[1.1fr_0.9fr]"
    >
        <div class="rounded-3xl border border-border bg-card/80 p-6 shadow-sm">
            <p class="text-xs uppercase tracking-[0.3em] text-muted-foreground">
                Users
            </p>
            <h2 class="mt-2 text-2xl font-semibold">Create user account</h2>
            <p class="mt-2 text-sm text-muted-foreground">
                Create a user in the identity provider and sync it into the
                app directory in one step.
            </p>
            <div class="mt-6 grid gap-4">
                <div class="grid gap-2 sm:grid-cols-2">
                    <div>
                        <label class="text-xs font-semibold text-muted-foreground"
                            >Username</label
                        >
                        <input
                            v-model="newAdminUser.username"
                            data-testid="settings.user-create-username-input"
                            type="text"
                            placeholder="jane.doe"
                            class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                        />
                    </div>
                    <div>
                        <label class="text-xs font-semibold text-muted-foreground"
                            >Email</label
                        >
                        <input
                            v-model="newAdminUser.email"
                            data-testid="settings.user-create-email-input"
                            type="email"
                            placeholder="jane.doe@example.com"
                            class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                        />
                    </div>
                </div>
                <div class="grid gap-2 sm:grid-cols-2">
                    <div>
                        <label class="text-xs font-semibold text-muted-foreground"
                            >First name (optional)</label
                        >
                        <input
                            v-model="newAdminUser.firstName"
                            data-testid="settings.user-create-firstname-input"
                            type="text"
                            placeholder="Jane"
                            class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                        />
                    </div>
                    <div>
                        <label class="text-xs font-semibold text-muted-foreground"
                            >Last name (optional)</label
                        >
                        <input
                            v-model="newAdminUser.lastName"
                            data-testid="settings.user-create-lastname-input"
                            type="text"
                            placeholder="Doe"
                            class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                        />
                    </div>
                </div>
                <div>
                    <label class="text-xs font-semibold text-muted-foreground"
                        >Initial password</label
                    >
                    <input
                        v-model="newAdminUser.password"
                        data-testid="settings.user-create-password-input"
                        type="password"
                        placeholder="Minimum 8 characters"
                        class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                    />
                </div>
                <div class="flex items-center justify-end">
                    <Button
                        data-testid="settings.user-create-submit-button"
                        :disabled="!canCreateAdminUser || creatingUser"
                        @click="createAdminUserSubmit"
                    >
                        {{ creatingUser ? "Creating..." : "Create user" }}
                    </Button>
                </div>
            </div>
        </div>
        <div class="rounded-3xl border border-border bg-card/80 p-6 shadow-sm">
            <h3 class="text-sm font-semibold text-foreground">
                Next steps
            </h3>
            <ul class="mt-3 space-y-2 text-xs text-muted-foreground">
                <li>
                    1. Create user account here.
                </li>
                <li>
                    2. Go to <span class="font-semibold">Projects</span> tab
                    and add the user to a group.
                </li>
                <li>
                    3. Assign group role per project.
                </li>
            </ul>
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
                                variant="outline"
                                size="sm"
                                :disabled="webhookLoading"
                                @click="toggleDeliveryHistory(hook)"
                            >
                                {{
                                    deliveryWebhookId === hook.id
                                        ? "Hide history"
                                        : "History"
                                }}
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

                    <!-- Delivery History Panel -->
                    <div
                        v-if="deliveryWebhookId === hook.id"
                        class="mt-4 rounded-xl border border-border bg-muted/30 p-3"
                    >
                        <p
                            class="mb-2 text-xs font-semibold text-muted-foreground uppercase tracking-[0.15em]"
                        >
                            Delivery history
                        </p>
                        <div
                            v-if="webhookDeliveriesLoading"
                            class="text-xs text-muted-foreground"
                        >
                            Loading deliveries...
                        </div>
                        <div
                            v-else-if="webhookDeliveries.length === 0"
                            class="text-xs text-muted-foreground"
                        >
                            No delivery attempts recorded yet.
                        </div>
                        <div v-else class="space-y-1">
                            <div
                                v-for="delivery in webhookDeliveries"
                                :key="delivery.id"
                                class="rounded-lg border border-border bg-background"
                            >
                                <button
                                    type="button"
                                    class="flex w-full items-center gap-3 px-3 py-2 text-left text-xs hover:bg-muted/50 transition"
                                    @click="toggleDeliveryDetail(delivery)"
                                >
                                    <span
                                        class="inline-block h-2 w-2 rounded-full"
                                        :class="
                                            delivery.delivered
                                                ? 'bg-emerald-500'
                                                : 'bg-red-500'
                                        "
                                    ></span>
                                    <span class="font-mono text-muted-foreground">
                                        {{ delivery.event }}
                                    </span>
                                    <span class="text-muted-foreground">
                                        #{{ delivery.attempt }}
                                    </span>
                                    <span
                                        v-if="delivery.statusCode"
                                        class="font-mono"
                                        :class="
                                            delivery.delivered
                                                ? 'text-emerald-400'
                                                : 'text-red-400'
                                        "
                                    >
                                        {{ delivery.statusCode }}
                                    </span>
                                    <span class="text-muted-foreground">
                                        {{ delivery.durationMs }}ms
                                    </span>
                                    <span
                                        class="ml-auto text-muted-foreground"
                                    >
                                        {{ timeAgo(delivery.createdAt) }}
                                    </span>
                                </button>
                                <div
                                    v-if="expandedDeliveryId === delivery.id"
                                    class="border-t border-border px-3 py-2 text-xs space-y-1"
                                >
                                    <div v-if="delivery.error">
                                        <span
                                            class="font-semibold text-red-400"
                                            >Error:</span
                                        >
                                        <span class="text-muted-foreground ml-1">
                                            {{ delivery.error }}
                                        </span>
                                    </div>
                                    <div v-if="delivery.responseBody">
                                        <span
                                            class="font-semibold text-muted-foreground"
                                            >Response:</span
                                        >
                                        <pre
                                            class="mt-1 max-h-32 overflow-auto rounded bg-muted p-2 text-[11px] text-muted-foreground"
                                            >{{ delivery.responseBody }}</pre
                                        >
                                    </div>
                                    <div
                                        v-if="
                                            !delivery.error &&
                                            !delivery.responseBody
                                        "
                                        class="text-muted-foreground"
                                    >
                                        No additional details.
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </section>

    <!-- Workflow Editor Tab -->
    <section
        v-if="settingsTab === 'workflow'"
        class="rounded-3xl border border-border bg-card/80 p-6 shadow-sm"
    >
        <div class="flex items-center justify-between">
            <div>
                <p
                    class="text-xs uppercase tracking-[0.3em] text-muted-foreground"
                >
                    Workflow
                </p>
                <h2 class="text-2xl font-semibold">Manage states</h2>
                <p class="mt-2 text-sm text-muted-foreground">
                    Add, rename, reorder, and delete workflow states. Drag rows
                    to reorder.
                </p>
                <p class="mt-1 text-xs text-muted-foreground">
                    Target project: {{ selectedProjectLabel }}
                </p>
            </div>
            <div class="flex items-center gap-2">
                <Button
                    variant="ghost"
                    size="sm"
                    :disabled="workflowLoading"
                    @click="loadWorkflow"
                >
                    {{ workflowLoading ? "Loading..." : "Reload" }}
                </Button>
            </div>
        </div>

        <div
            v-if="workflowValidationError || workflowError"
            data-testid="workflow.error"
            class="mt-4 rounded-2xl border px-4 py-3 text-sm border-destructive/40 bg-destructive/10 text-destructive"
        >
            {{ workflowValidationError || workflowError }}
        </div>

        <div
            v-if="workflowLoading"
            class="mt-6 text-sm text-muted-foreground"
        >
            Loading workflow states...
        </div>

        <div v-else class="mt-6 space-y-2">
            <div
                class="grid grid-cols-[auto_1fr_auto_auto_auto] items-center gap-3 px-3 text-[11px] uppercase tracking-[0.15em] text-muted-foreground font-semibold"
            >
                <span class="w-6"></span>
                <span>Name</span>
                <span class="text-center w-16">Default</span>
                <span class="text-center w-16">Closed</span>
                <span class="w-16"></span>
            </div>

            <div
                v-for="(state, index) in workflowStates"
                :key="state.id || index"
                data-testid="workflow.state-row"
                class="grid grid-cols-[auto_1fr_auto_auto_auto] items-center gap-3 rounded-xl border border-border bg-background px-3 py-2 transition"
                :class="dragIndex === index ? 'opacity-50 border-primary' : ''"
                draggable="true"
                @dragstart="onDragStart(index)"
                @dragover="onDragOver($event, index)"
                @dragend="onDragEnd"
            >
                <span
                    class="w-6 cursor-grab text-center text-muted-foreground select-none"
                    title="Drag to reorder"
                >
                    &#8801;
                </span>

                <input
                    data-testid="workflow.state-name-input"
                    type="text"
                    :value="state.name"
                    placeholder="State name"
                    class="w-full rounded-lg border border-input bg-background px-2 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                    @input="
                        updateStateName(
                            index,
                            ($event.target as HTMLInputElement).value,
                        )
                    "
                />

                <label
                    class="flex items-center justify-center w-16"
                    title="Default state for new tickets"
                >
                    <input
                        data-testid="workflow.state-default-radio"
                        type="radio"
                        name="workflow-default-state"
                        :checked="state.isDefault"
                        class="h-4 w-4 border-border text-primary focus:ring-ring"
                        @change="setDefaultState(index)"
                    />
                </label>

                <label
                    class="flex items-center justify-center w-16"
                    title="Mark as closed/resolved state"
                >
                    <input
                        data-testid="workflow.state-closed-checkbox"
                        type="checkbox"
                        :checked="state.isClosed"
                        class="h-4 w-4 rounded border-border text-primary focus:ring-ring"
                        @change="toggleStateClosed(index)"
                    />
                </label>

                <div class="flex justify-end w-16">
                    <Button
                        data-testid="workflow.state-delete-button"
                        variant="ghost"
                        size="sm"
                        :disabled="workflowStates.length <= 1"
                        @click="removeWorkflowState(index)"
                    >
                        Delete
                    </Button>
                </div>
            </div>

            <div
                v-if="workflowStates.length === 0"
                class="rounded-xl border border-dashed border-border bg-muted/30 p-4 text-center text-sm text-muted-foreground"
            >
                No workflow states. Add one to get started.
            </div>
        </div>

        <div class="mt-6 flex items-center gap-3">
            <Button
                data-testid="workflow.add-state-button"
                variant="outline"
                size="sm"
                @click="addWorkflowState"
            >
                Add state
            </Button>
            <Button
                data-testid="workflow.save-button"
                size="sm"
                :disabled="
                    workflowSaving ||
                    workflowStates.length === 0 ||
                    !selectedProjectId
                "
                @click="saveWorkflow"
            >
                {{ workflowSaving ? "Saving..." : "Save workflow" }}
            </Button>
        </div>
    </section>

    <section
        v-if="settingsTab === 'reporting'"
        data-testid="reporting.view"
        class="rounded-3xl border border-border bg-card/80 p-6 shadow-sm"
    >
        <div class="flex items-center justify-between">
            <div>
                <p class="text-xs uppercase tracking-[0.3em] text-muted-foreground">
                    Reporting
                </p>
                <h2 class="text-2xl font-semibold">Project summary</h2>
                <p class="mt-2 text-sm text-muted-foreground">
                    Read-only throughput, cycle-time, and open-by-state trend.
                </p>
                <p class="mt-1 text-xs text-muted-foreground">
                    Target project: {{ selectedProjectLabel }}
                </p>
            </div>
            <div class="flex items-center gap-2">
                <input
                    data-testid="reporting.from-input"
                    v-model="reportingFrom"
                    type="date"
                    class="rounded-lg border border-input bg-background px-2 py-1.5 text-xs"
                />
                <input
                    data-testid="reporting.to-input"
                    v-model="reportingTo"
                    type="date"
                    class="rounded-lg border border-input bg-background px-2 py-1.5 text-xs"
                />
                <Button
                    data-testid="reporting.reload-button"
                    variant="ghost"
                    size="sm"
                    :disabled="reportingLoading"
                    @click="loadReporting"
                >
                    {{ reportingLoading ? "Loading..." : "Reload" }}
                </Button>
                <Button
                    data-testid="reporting.export-json-button"
                    variant="outline"
                    size="sm"
                    :disabled="reportingLoading || reportingExporting !== null"
                    @click="exportReporting('json')"
                >
                    {{
                        reportingExporting === "json"
                            ? "Exporting JSON..."
                            : "Export JSON"
                    }}
                </Button>
                <Button
                    data-testid="reporting.export-csv-button"
                    variant="outline"
                    size="sm"
                    :disabled="reportingLoading || reportingExporting !== null"
                    @click="exportReporting('csv')"
                >
                    {{
                        reportingExporting === "csv"
                            ? "Exporting CSV..."
                            : "Export CSV"
                    }}
                </Button>
            </div>
        </div>

        <div v-if="reportingLoading" class="mt-6 text-sm text-muted-foreground">
            Loading reporting summary...
        </div>

        <div
            v-else-if="!reportingSummary"
            class="mt-6 rounded-2xl border border-border bg-muted/20 px-4 py-3 text-sm text-muted-foreground"
        >
            No reporting data available.
        </div>

        <div v-else class="mt-6 space-y-5">
            <div class="grid gap-4 md:grid-cols-3">
                <div class="rounded-2xl border border-border bg-background px-4 py-3">
                    <p class="text-[10px] uppercase tracking-[0.2em] text-muted-foreground">Range</p>
                    <p class="mt-2 text-sm font-semibold">
                        {{ reportingSummary.from }} to {{ reportingSummary.to }}
                    </p>
                </div>
                <div class="rounded-2xl border border-border bg-background px-4 py-3">
                    <p class="text-[10px] uppercase tracking-[0.2em] text-muted-foreground">Avg Cycle Time</p>
                    <p class="mt-2 text-2xl font-semibold">
                        {{ Number(reportingSummary.averageCycleTimeHours || 0).toFixed(1) }}h
                    </p>
                </div>
                <div class="rounded-2xl border border-border bg-background px-4 py-3">
                    <p class="text-[10px] uppercase tracking-[0.2em] text-muted-foreground">Days</p>
                    <p class="mt-2 text-2xl font-semibold">
                        {{ reportingSummary.throughputByDay.length }}
                    </p>
                </div>
            </div>

            <div class="rounded-2xl border border-border bg-background px-4 py-3">
                <p class="text-xs font-semibold text-foreground">Throughput by Day</p>
                <div
                    data-testid="reporting.throughput-list"
                    class="mt-3 max-h-56 overflow-y-auto space-y-1.5"
                >
                    <div
                        v-for="point in reportingSummary.throughputByDay"
                        :key="point.date"
                        class="flex items-center justify-between rounded-lg border border-border/70 px-3 py-1.5 text-xs"
                    >
                        <span class="text-muted-foreground">{{ point.date }}</span>
                        <span class="font-semibold text-foreground">{{ point.value }}</span>
                    </div>
                </div>
            </div>

            <div class="rounded-2xl border border-border bg-background px-4 py-3">
                <p class="text-xs font-semibold text-foreground">Open by State (Daily)</p>
                <div
                    data-testid="reporting.open-by-state-list"
                    class="mt-3 max-h-64 overflow-y-auto space-y-2"
                >
                    <div
                        v-for="point in reportingSummary.openByState"
                        :key="point.date"
                        class="rounded-lg border border-border/70 px-3 py-2 text-xs"
                    >
                        <p class="font-semibold text-muted-foreground">{{ point.date }}</p>
                        <div class="mt-1 flex flex-wrap gap-2">
                            <span
                                v-for="count in point.counts"
                                :key="count.label"
                                class="rounded-md border border-border bg-muted/30 px-2 py-0.5"
                            >
                                {{ count.label }}: {{ count.value }}
                            </span>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </section>

    <section
        v-if="settingsTab === 'sprints'"
        data-testid="settings.sprints-view"
        class="rounded-3xl border border-border bg-card/80 p-6 shadow-sm"
    >
        <div>
            <p class="text-xs uppercase tracking-[0.3em] text-muted-foreground">Sprint Configuration</p>
            <h3 class="mt-1 text-lg font-semibold">Sprint Duration & Capacity</h3>
        </div>

        <div v-if="capacitySettingsLoading" class="mt-6 text-sm text-muted-foreground">Loading...</div>

        <template v-else>
            <div class="mt-6 rounded-2xl border border-border bg-background p-4">
                <label class="text-xs font-semibold text-muted-foreground">Default Sprint Duration (days)</label>
                <div class="mt-2 flex items-center gap-3">
                    <input
                        data-testid="settings.sprint_duration_input"
                        v-model.number="sprintDurationDays"
                        type="number"
                        min="1"
                        placeholder="e.g. 14"
                        class="w-32 rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                    />
                    <Button
                        size="sm"
                        :disabled="sprintDurationSaving"
                        @click="saveSprintDuration"
                    >
                        {{ sprintDurationSaving ? "Saving..." : "Save" }}
                    </Button>
                </div>
            </div>

            <div class="mt-6 rounded-2xl border border-border bg-background p-4">
                <label class="text-xs font-semibold text-muted-foreground">Capacity Settings</label>
                <div data-testid="settings.capacity_table" class="mt-3 space-y-2">
                    <div
                        v-for="(setting, index) in capacitySettings"
                        :key="setting.id || index"
                        class="flex items-center gap-2 rounded-xl border border-border bg-card px-3 py-2 text-xs"
                    >
                        <span class="rounded bg-muted px-1.5 py-0.5 text-[10px] font-semibold uppercase">{{ setting.scope }}</span>
                        <span class="font-semibold text-foreground">{{ setting.label }}</span>
                        <span class="ml-auto text-muted-foreground">{{ setting.capacity }} SP/sprint</span>
                        <button
                            type="button"
                            class="text-[10px] text-destructive hover:text-destructive/80 transition ml-2"
                            @click="removeCapacityRow(index)"
                        >
                            Remove
                        </button>
                    </div>
                    <p v-if="!capacitySettings.length" class="text-[10px] text-muted-foreground">No capacity settings yet.</p>
                </div>
                <div class="mt-3 flex items-center gap-2">
                    <select
                        v-model="newCapacityRow.scope"
                        class="rounded-xl border border-input bg-background px-2 py-1.5 text-xs"
                    >
                        <option value="team">team</option>
                        <option value="user">user</option>
                    </select>
                    <input
                        v-model="newCapacityRow.label"
                        type="text"
                        placeholder="Label"
                        class="flex-1 rounded-xl border border-input bg-background px-2 py-1.5 text-xs"
                    />
                    <input
                        v-model.number="newCapacityRow.capacity"
                        type="number"
                        min="0"
                        placeholder="SP"
                        class="w-20 rounded-xl border border-input bg-background px-2 py-1.5 text-xs"
                    />
                    <Button variant="outline" size="sm" @click="addCapacityRow">Add</Button>
                </div>
                <div class="mt-3">
                    <Button
                        size="sm"
                        :disabled="capacitySettingsSaving"
                        @click="saveCapacitySettings"
                    >
                        {{ capacitySettingsSaving ? "Saving..." : "Save Capacity Settings" }}
                    </Button>
                </div>
            </div>
        </template>
    </section>
</template>
