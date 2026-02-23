import { defineStore } from "pinia";
import {
  addTicketComment,
  bulkTicketOperation as apiBulkTicketOperation,
  createStory,
  createTicket,
  deleteStory,
  deleteTicket,
  deleteTicketAttachment,
  deleteTicketDependency as apiDeleteTicketDependency,
  deleteWebhook,
  getBoard,
  getProjectDependencyGraph as apiGetProjectDependencyGraph,
  getMyProjectRole,
  getNotificationPreferences as apiGetNotificationPreferences,
  getNotificationUnreadCount,
  getProjectReportingSummary as apiGetProjectReportingSummary,
  getProjectSprintForecast as apiGetProjectSprintForecast,
  getProjectStats,
  getWorkflow,
  listProjectCapacitySettings as apiListProjectCapacitySettings,
  listProjectSprints as apiListProjectSprints,
  listStories,
  listTicketDependencies as apiListTicketDependencies,
  listTicketAttachments,
  getProjectActivities,
  listBoardFilterPresets,
  listNotifications,
  listTicketActivities,
  listTicketComments,
  listWebhookDeliveries,
  listWebhooks,
  markAllNotificationsRead as apiMarkAllNotificationsRead,
  markNotificationRead as apiMarkNotificationRead,
  testWebhook,
  updateTicket,
  replaceProjectCapacitySettings as apiReplaceProjectCapacitySettings,
  updateBoardFilterPreset as apiUpdateBoardFilterPreset,
  updateNotificationPreferences as apiUpdateNotificationPreferences,
  updateWorkflow,
  uploadTicketAttachment,
  createTicketDependency as apiCreateTicketDependency,
  createProjectSprint as apiCreateProjectSprint,
  addSprintTickets as apiAddSprintTickets,
  removeSprintTickets as apiRemoveSprintTickets,
  createBoardFilterPreset as apiCreateBoardFilterPreset,
  deleteBoardFilterPreset as apiDeleteBoardFilterPreset,
  getSharedBoardFilterPreset,
  listTicketTimeEntries as apiListTicketTimeEntries,
  createTicketTimeEntry as apiCreateTicketTimeEntry,
  deleteTicketTimeEntry as apiDeleteTicketTimeEntry,
  type Attachment,
  type BoardFilter,
  type BoardFilterPreset,
  type BulkTicketOperationRequest,
  type BulkTicketOperationResponse,
  type Notification,
  type NotificationPreferences,
  type ProjectRole,
  type Story,
  type ProjectActivity,
  type TicketActivity,
  type TicketComment,
  type TicketCreateRequest,
  type TicketPriority,
  type TicketResponse,
  type TicketDependency,
  type TicketDependencyCreateRequest,
  type TicketDependencyGraphResponse,
  type TicketUpdateRequest,
  type ProjectStats,
  type ProjectReportingSummary,
  type Sprint,
  type SprintCreateRequest,
  type SprintForecastSummary,
  type CapacitySetting,
  type CapacitySettingInput,
  type WebhookDelivery,
  type WebhookEvent,
  type TimeEntry,
  type TimeEntryCreateRequest,
  type WebhookResponse,
  type WorkflowState,
  type WorkflowStateInput,
  createWebhook as apiCreateWebhook,
  updateWebhook as apiUpdateWebhook,
} from "@/lib/api";

type ApiMode = "live" | "demo";

const demoProjectId = "00000000-0000-0000-0000-000000000000";
const demoProjectKey = "DEMO";

const demoStates: WorkflowState[] = [
  {
    id: "bfc9e1a1-3cd9-4e1d-9f1e-000000000001",
    projectId: demoProjectId,
    name: "Backlog",
    order: 1,
    isDefault: true,
    isClosed: false,
  },
  {
    id: "bfc9e1a1-3cd9-4e1d-9f1e-000000000002",
    projectId: demoProjectId,
    name: "In Progress",
    order: 2,
    isDefault: false,
    isClosed: false,
  },
  {
    id: "bfc9e1a1-3cd9-4e1d-9f1e-000000000003",
    projectId: demoProjectId,
    name: "Review",
    order: 3,
    isDefault: false,
    isClosed: false,
  },
  {
    id: "bfc9e1a1-3cd9-4e1d-9f1e-000000000004",
    projectId: demoProjectId,
    name: "Done",
    order: 4,
    isDefault: false,
    isClosed: true,
  },
];

const demoStateBacklog = demoStates[0]!;
const demoStateInProgress = demoStates[1]!;
const demoStateReview = demoStates[2]!;
const demoStateDone = demoStates[3]!;

const demoAssignees = ["Ari", "Nova", "Jules", "Sam", "Ira", "Mika", "Rene"];
const demoStoryTitles = ["Checkout Reliability", "Mobile Experience", "Platform Foundations"];
const demoTicketTitles = [
  "Reduce API latency spikes",
  "Fix flaky webhook retries",
  "Improve board keyboard navigation",
  "Stabilize attachment uploads",
  "Refine dashboard trend chart",
  "Add validation for workflow edits",
  "Harden notification preference sync",
  "Improve ticket modal performance",
  "Polish dependency graph layout",
  "Fix stale board filter badge",
  "Tighten markdown preview spacing",
  "Improve login error feedback",
];

const rand = (max: number) => Math.floor(Math.random() * max);

function makeDemoData(): { stories: Story[]; tickets: TicketResponse[] } {
  const now = new Date().toISOString();
  const stories: Story[] = demoStoryTitles.map((title, index) => ({
    id: `demo-story-${index + 1}`,
    projectId: demoProjectId,
    title,
    description: `Demo storyline for ${title.toLowerCase()}.`,
    createdAt: now,
    updatedAt: now,
  }));

  const statePool = [demoStateBacklog, demoStateInProgress, demoStateReview, demoStateDone];
  const priorityPool: TicketPriority[] = ["low", "medium", "high", "urgent"];
  const typePool: Array<"feature" | "bug"> = ["feature", "bug"];
  const tickets: TicketResponse[] = [];
  let ticketNumber = 100;

  stories.forEach((story) => {
    const ticketCount = 3 + rand(3); // 3..5 per story
    for (let i = 0; i < ticketCount; i += 1) {
      const baseTitle = demoTicketTitles[rand(demoTicketTitles.length)]!;
      const title = `${baseTitle} (${story.title.split(" ")[0]})`;
      ticketNumber += 1;
      const state = statePool[rand(statePool.length)]!;
      const assigneeName = demoAssignees[rand(demoAssignees.length)]!;
      const assigneeID = `demo-user-${assigneeName.toLowerCase()}`;
      tickets.push({
        id: `demo-ticket-${ticketNumber}`,
        key: `${demoProjectKey}-${ticketNumber}`,
        number: ticketNumber,
        type: typePool[rand(typePool.length)]!,
        projectId: demoProjectId,
        projectKey: demoProjectKey,
        storyId: story.id,
        story,
        title,
        description: `Demo ticket for ${story.title}.`,
        stateId: state.id,
        state,
        priority: priorityPool[rand(priorityPool.length)]!,
        incidentEnabled: false,
        position: i + 1,
        blockedByCount: 0,
        isBlocked: false,
        createdAt: now,
        updatedAt: now,
        assigneeId: assigneeID,
        assignee: { id: assigneeID, name: assigneeName },
      });
    }
  });

  return { stories, tickets };
}

const isAuthError = (err: unknown) => {
  const error = err as Error & { status?: number };
  return error.status === 401 || error.status === 403;
};

export const useBoardStore = defineStore("board", {
  state: () => ({
    states: [] as WorkflowState[],
    tickets: [] as TicketResponse[],
    stories: [] as Story[],
    webhooks: [] as WebhookResponse[],
    ticketComments: [] as TicketComment[],
    ticketActivities: [] as TicketActivity[],
    apiMode: "live" as ApiMode,
    loading: true,
    errorMessage: "",
    workflowSetupBusy: false,
    workflowSetupError: "",
    webhookLoading: false,
    webhookError: "",
    webhookTestStatus: {} as Record<
      string,
      { message: string; success: boolean }
    >,
    storyLoading: false,
    storyError: "",
    commentSaving: false,
    commentError: "",
    ticketAttachments: [] as Attachment[],
    ticketDependencies: [] as TicketDependency[],
    ticketDependenciesLoading: false,
    ticketDependencyGraph: {
      nodes: [],
      edges: [],
    } as TicketDependencyGraphResponse,
    ticketDependencyGraphLoading: false,
    attachmentUploading: false,
    attachmentError: "",
    webhookDeliveries: [] as WebhookDelivery[],
    webhookDeliveriesLoading: false,
    dashboardStats: null as ProjectStats | null,
    dashboardLoading: false,
    projectActivities: [] as ProjectActivity[],
    projectActivitiesLoading: false,
    projectReportingSummary: null as ProjectReportingSummary | null,
    projectReportingLoading: false,
    sprints: [] as Sprint[],
    sprintsLoading: false,
    capacitySettings: [] as CapacitySetting[],
    capacitySettingsLoading: false,
    sprintForecastSummary: null as SprintForecastSummary | null,
    sprintForecastLoading: false,
    currentUserRole: null as ProjectRole | null,
    workflowEditorStates: [] as WorkflowState[],
    workflowEditorLoading: false,
    workflowEditorSaving: false,
    workflowEditorError: "",
    boardFilterPresets: [] as BoardFilterPreset[],
    boardFilterPresetsLoading: false,
    boardFilterPresetsError: "",
    notifications: [] as Notification[],
    notificationsLoading: false,
    notificationsUnreadCount: 0,
    notificationPreferences: {
      mentionEnabled: true,
      assignmentEnabled: true,
    } as NotificationPreferences,
    notificationPreferencesSaving: false,
    ticketTimeEntries: [] as TimeEntry[],
    ticketTimeEntriesTotalMinutes: 0,
    ticketTimeEntriesLoading: false,
  }),
  getters: {
    canEditTickets(): boolean {
      const role = this.currentUserRole;
      return role === "admin" || role === "contributor";
    },
    canManageProject(): boolean {
      return this.currentUserRole === "admin";
    },
  },
  actions: {
    reset() {
      this.states = [];
      this.tickets = [];
      this.stories = [];
      this.webhooks = [];
      this.ticketComments = [];
      this.ticketActivities = [];
      this.apiMode = "live";
      this.loading = true;
      this.errorMessage = "";
      this.workflowSetupBusy = false;
      this.workflowSetupError = "";
      this.webhookLoading = false;
      this.webhookError = "";
      this.webhookTestStatus = {};
      this.storyLoading = false;
      this.storyError = "";
      this.commentSaving = false;
      this.commentError = "";
      this.ticketAttachments = [];
      this.ticketDependencies = [];
      this.ticketDependenciesLoading = false;
      this.ticketDependencyGraph = { nodes: [], edges: [] };
      this.ticketDependencyGraphLoading = false;
      this.attachmentUploading = false;
      this.attachmentError = "";
      this.webhookDeliveries = [];
      this.webhookDeliveriesLoading = false;
      this.dashboardStats = null;
      this.dashboardLoading = false;
      this.projectActivities = [];
      this.projectActivitiesLoading = false;
      this.projectReportingSummary = null;
      this.projectReportingLoading = false;
      this.sprints = [];
      this.sprintsLoading = false;
      this.capacitySettings = [];
      this.capacitySettingsLoading = false;
      this.sprintForecastSummary = null;
      this.sprintForecastLoading = false;
      this.currentUserRole = null;
      this.workflowEditorStates = [];
      this.workflowEditorLoading = false;
      this.workflowEditorSaving = false;
      this.workflowEditorError = "";
      this.boardFilterPresets = [];
      this.boardFilterPresetsLoading = false;
      this.boardFilterPresetsError = "";
      this.notifications = [];
      this.notificationsLoading = false;
      this.notificationsUnreadCount = 0;
      this.notificationPreferences = {
        mentionEnabled: true,
        assignmentEnabled: true,
      };
      this.notificationPreferencesSaving = false;
      this.ticketTimeEntries = [];
      this.ticketTimeEntriesTotalMinutes = 0;
      this.ticketTimeEntriesLoading = false;
    },
    applyDemo() {
      const demo = makeDemoData();
      this.apiMode = "demo";
      this.states = demoStates;
      this.tickets = demo.tickets;
      this.stories = demo.stories;
      this.loading = false;
    },
    clearComments() {
      this.ticketComments = [];
      this.commentError = "";
      this.ticketActivities = [];
      this.ticketAttachments = [];
      this.ticketDependencies = [];
      this.ticketDependencyGraph = { nodes: [], edges: [] };
      this.attachmentError = "";
    },
    async loadBoard(projectId: string) {
      this.loading = true;
      this.errorMessage = "";
      try {
        const board = await getBoard(projectId);
        this.states = board.states;
        this.tickets = board.tickets;
        this.apiMode = "live";
      } catch (err) {
        if (isAuthError(err)) {
          this.loading = false;
          throw err;
        }
        const error = err as Error & { status?: number };
        if (typeof error.status === "number") {
          this.errorMessage = `Unable to load board (HTTP ${error.status}).`;
          this.loading = false;
          return;
        }
        this.errorMessage = "Unable to load board.";
        this.apiMode = "live";
        this.states = [];
        this.tickets = [];
        this.stories = [];
        if (error.message) {
          this.errorMessage = `Unable to load board. ${error.message}`;
        }
        this.loading = false;
        return;
      }
      this.loading = false;
    },
    async initializeWorkflow(projectId: string, states: WorkflowStateInput[]) {
      if (this.workflowSetupBusy) return;
      this.workflowSetupBusy = true;
      this.workflowSetupError = "";
      try {
        const result = await updateWorkflow(projectId, states);
        this.states = result.states;
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.workflowSetupError = "Unable to create workflow.";
      } finally {
        this.workflowSetupBusy = false;
      }
    },
    async loadStories(projectId: string) {
      if (this.apiMode === "demo") {
        if (this.stories.length === 0) {
          const demo = makeDemoData();
          this.tickets = demo.tickets;
          this.stories = demo.stories;
        }
        return;
      }
      this.storyLoading = true;
      this.storyError = "";
      try {
        const list = await listStories(projectId);
        this.stories = list.items;
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.stories = [];
        this.storyError = "Unable to load stories.";
      } finally {
        this.storyLoading = false;
      }
    },
    async createStory(
      projectId: string,
      payload: { title: string; description?: string; storyPoints?: number | null },
    ) {
      this.storyError = "";
      try {
        if (this.apiMode === "demo") {
          const created: Story = {
            id: `demo-story-${Date.now()}`,
            projectId: demoProjectId,
            title: payload.title,
            description: payload.description,
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
          };
          this.stories = [created, ...this.stories];
          return created;
        }
        const created = await createStory(projectId, payload);
        this.stories = [created, ...this.stories];
        return created;
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.storyError = "Unable to create story.";
        throw err;
      }
    },
    async removeStory(storyId: string) {
      this.storyError = "";
      try {
        if (this.apiMode === "demo") {
          this.stories = this.stories.filter((story) => story.id !== storyId);
          this.tickets = this.tickets.filter(
            (ticket) => ticket.storyId !== storyId,
          );
          return;
        }
        await deleteStory(storyId);
        this.stories = this.stories.filter((story) => story.id !== storyId);
        this.tickets = this.tickets.filter(
          (ticket) => ticket.storyId !== storyId,
        );
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.storyError = "Unable to delete story.";
        throw err;
      }
    },
    async loadWebhooks(projectId: string) {
      this.webhookLoading = true;
      this.webhookError = "";
      try {
        const list = await listWebhooks(projectId);
        this.webhooks = list.items;
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.webhookError = "Unable to load webhooks.";
      } finally {
        this.webhookLoading = false;
      }
    },
    async createWebhook(
      projectId: string,
      payload: {
        url: string;
        events: WebhookEvent[];
        secret?: string;
        enabled?: boolean;
      },
    ) {
      this.webhookError = "";
      try {
        const created = await apiCreateWebhook(projectId, payload);
        this.webhooks = [created, ...this.webhooks];
        return created;
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.webhookError = "Unable to create webhook.";
        throw err;
      }
    },
    async updateWebhook(
      projectId: string,
      id: string,
      payload: {
        url?: string;
        events?: WebhookEvent[];
        secret?: string;
        enabled?: boolean;
      },
    ) {
      this.webhookError = "";
      try {
        const updated = await apiUpdateWebhook(projectId, id, payload);
        this.webhooks = this.webhooks.map((hook) =>
          hook.id === updated.id ? updated : hook,
        );
        return updated;
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.webhookError = "Unable to update webhook.";
        throw err;
      }
    },
    async deleteWebhook(projectId: string, id: string) {
      this.webhookError = "";
      try {
        await deleteWebhook(projectId, id);
        this.webhooks = this.webhooks.filter((hook) => hook.id !== id);
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.webhookError = "Unable to delete webhook.";
        throw err;
      }
    },
    async testWebhook(
      projectId: string,
      id: string,
      payload: { event: WebhookEvent },
    ) {
      this.webhookTestStatus = { ...this.webhookTestStatus };
      try {
        const result = await testWebhook(projectId, id, payload);
        this.webhookTestStatus[id] = {
          message: result.responseBody || "Delivered",
          success: result.delivered,
        };
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.webhookTestStatus[id] = {
          message: "Delivery failed",
          success: false,
        };
      }
    },
    async loadWebhookDeliveries(projectId: string, webhookId: string) {
      this.webhookDeliveriesLoading = true;
      this.webhookDeliveries = [];
      try {
        const list = await listWebhookDeliveries(projectId, webhookId);
        this.webhookDeliveries = list.items;
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.webhookDeliveries = [];
      } finally {
        this.webhookDeliveriesLoading = false;
      }
    },
    async loadDashboardStats(projectId: string) {
      this.dashboardLoading = true;
      try {
        this.dashboardStats = await getProjectStats(projectId);
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.dashboardStats = null;
      } finally {
        this.dashboardLoading = false;
      }
    },
    async loadProjectActivities(projectId: string) {
      this.projectActivitiesLoading = true;
      try {
        const result = await getProjectActivities(projectId);
        this.projectActivities = result.items;
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.projectActivities = [];
      } finally {
        this.projectActivitiesLoading = false;
      }
    },
    async loadProjectReportingSummary(
      projectId: string,
      opts: { from?: string; to?: string } = {},
    ) {
      this.projectReportingLoading = true;
      try {
        this.projectReportingSummary = await apiGetProjectReportingSummary(
          projectId,
          opts,
        );
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.projectReportingSummary = null;
      } finally {
        this.projectReportingLoading = false;
      }
    },
    async loadSprints(projectId: string) {
      this.sprintsLoading = true;
      try {
        const result = await apiListProjectSprints(projectId);
        this.sprints = result.items;
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.sprints = [];
      } finally {
        this.sprintsLoading = false;
      }
    },
    async createSprint(projectId: string, payload: SprintCreateRequest) {
      const created = await apiCreateProjectSprint(projectId, payload);
      this.sprints = [created, ...this.sprints.filter((s) => s.id !== created.id)];
      return created;
    },
    async addSprintTickets(projectId: string, sprintId: string, ticketIds: string[]) {
      const updated = await apiAddSprintTickets(projectId, sprintId, ticketIds);
      this.sprints = this.sprints.map((s) => (s.id === updated.id ? updated : s));
      return updated;
    },
    async removeSprintTickets(projectId: string, sprintId: string, ticketIds: string[]) {
      const updated = await apiRemoveSprintTickets(projectId, sprintId, ticketIds);
      this.sprints = this.sprints.map((s) => (s.id === updated.id ? updated : s));
      return updated;
    },
    async loadCapacitySettings(projectId: string) {
      this.capacitySettingsLoading = true;
      try {
        const result = await apiListProjectCapacitySettings(projectId);
        this.capacitySettings = result.items;
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.capacitySettings = [];
      } finally {
        this.capacitySettingsLoading = false;
      }
    },
    async replaceCapacitySettings(projectId: string, items: CapacitySettingInput[]) {
      const result = await apiReplaceProjectCapacitySettings(projectId, { items });
      this.capacitySettings = result.items;
      return result.items;
    },
    async loadSprintForecast(
      projectId: string,
      opts: { sprintId?: string; iterations?: number } = {},
    ) {
      this.sprintForecastLoading = true;
      try {
        this.sprintForecastSummary = await apiGetProjectSprintForecast(projectId, opts);
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.sprintForecastSummary = null;
      } finally {
        this.sprintForecastLoading = false;
      }
    },
    async loadCurrentUserRole(projectId: string) {
      try {
        const result = await getMyProjectRole(projectId);
        this.currentUserRole = result.role;
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.currentUserRole = null;
      }
    },
    async loadBoardFilterPresets(projectId: string) {
      this.boardFilterPresetsLoading = true;
      this.boardFilterPresetsError = "";
      try {
        const result = await listBoardFilterPresets(projectId);
        this.boardFilterPresets = result.items;
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.boardFilterPresets = [];
        this.boardFilterPresetsError = "Unable to load filter presets.";
      } finally {
        this.boardFilterPresetsLoading = false;
      }
    },
    async createBoardFilterPreset(
      projectId: string,
      payload: { name: string; filters: BoardFilter; generateShareToken?: boolean },
    ) {
      this.boardFilterPresetsError = "";
      try {
        const created = await apiCreateBoardFilterPreset(projectId, payload);
        this.boardFilterPresets = [created, ...this.boardFilterPresets];
        return created;
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.boardFilterPresetsError = "Unable to create filter preset.";
        throw err;
      }
    },
    async updateBoardFilterPreset(
      projectId: string,
      presetId: string,
      payload: { name?: string; filters?: BoardFilter; generateShareToken?: boolean },
    ) {
      this.boardFilterPresetsError = "";
      try {
        const updated = await apiUpdateBoardFilterPreset(projectId, presetId, payload);
        this.boardFilterPresets = this.boardFilterPresets.map((preset) =>
          preset.id === updated.id ? updated : preset,
        );
        return updated;
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.boardFilterPresetsError = "Unable to update filter preset.";
        throw err;
      }
    },
    async deleteBoardFilterPreset(projectId: string, presetId: string) {
      this.boardFilterPresetsError = "";
      try {
        await apiDeleteBoardFilterPreset(projectId, presetId);
        this.boardFilterPresets = this.boardFilterPresets.filter(
          (preset) => preset.id !== presetId,
        );
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.boardFilterPresetsError = "Unable to delete filter preset.";
        throw err;
      }
    },
    async resolveSharedBoardFilterPreset(projectId: string, token: string) {
      return getSharedBoardFilterPreset(projectId, token);
    },
    async loadNotifications(
      projectId: string,
      opts: { limit?: number; unreadOnly?: boolean } = {},
    ) {
      this.notificationsLoading = true;
      try {
        const result = await listNotifications(projectId, opts);
        this.notifications = result.items;
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.notifications = [];
      } finally {
        this.notificationsLoading = false;
      }
    },
    async loadNotificationUnreadCount(projectId: string) {
      try {
        const result = await getNotificationUnreadCount(projectId);
        this.notificationsUnreadCount = result.count;
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.notificationsUnreadCount = 0;
      }
    },
    async markNotificationRead(projectId: string, notificationId: string) {
      const updated = await apiMarkNotificationRead(projectId, notificationId);
      this.notifications = this.notifications.map((item) =>
        item.id === updated.id ? updated : item,
      );
      await this.loadNotificationUnreadCount(projectId);
      return updated;
    },
    async markAllNotificationsRead(projectId: string) {
      await apiMarkAllNotificationsRead(projectId);
      this.notifications = this.notifications.map((item) => ({
        ...item,
        readAt: item.readAt || new Date().toISOString(),
      }));
      this.notificationsUnreadCount = 0;
    },
    async loadNotificationPreferences(projectId: string) {
      try {
        this.notificationPreferences =
          await apiGetNotificationPreferences(projectId);
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.notificationPreferences = {
          mentionEnabled: true,
          assignmentEnabled: true,
        };
      }
    },
    async updateNotificationPreferences(
      projectId: string,
      payload: Partial<NotificationPreferences>,
    ) {
      this.notificationPreferencesSaving = true;
      try {
        this.notificationPreferences = await apiUpdateNotificationPreferences(
          projectId,
          payload,
        );
      } finally {
        this.notificationPreferencesSaving = false;
      }
    },
    async createTicket(projectId: string, payload: TicketCreateRequest) {
      try {
        if (this.apiMode === "demo") {
          const id = `demo-${Date.now()}`;
          const number = Math.floor(Math.random() * 900 + 100);
          const state =
            this.states.find((item) => item.id === payload.stateId) ||
            this.states[0];
          const resolvedState = state ?? demoStateBacklog;
          const story =
            this.stories.find((item) => item.id === payload.storyId) ||
            undefined;
          const created: TicketResponse = {
            id,
            key: `${demoProjectKey}-${number}`,
            number,
            type: payload.type || "feature",
            projectId: demoProjectId,
            projectKey: demoProjectKey,
            storyId: payload.storyId,
            story,
            title: payload.title,
            description: payload.description || "",
            stateId: resolvedState.id,
            state: resolvedState,
            priority: payload.priority || "medium",
            incidentEnabled: payload.incidentEnabled ?? false,
            position: 0,
            blockedByCount: 0,
            isBlocked: false,
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
          };
          this.tickets = [created, ...this.tickets];
          return created;
        }
        const created = await createTicket(projectId, payload);
        this.tickets = [created, ...this.tickets];
        return created;
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.errorMessage = "Unable to create ticket.";
        throw err;
      }
    },
    async updateTicket(id: string, payload: TicketUpdateRequest) {
      if (this.apiMode === "demo") {
        this.tickets = this.tickets.map((ticket) =>
          ticket.id === id ? { ...ticket, ...payload } : ticket,
        );
        const updated = this.tickets.find((ticket) => ticket.id === id);
        if (!updated) {
          throw new Error("Ticket not found.");
        }
        return updated;
      }
      const updated = await updateTicket(id, payload);
      this.tickets = this.tickets.map((ticket) =>
        ticket.id === updated.id ? updated : ticket,
      );
      return updated;
    },
    async removeTicket(id: string) {
      this.errorMessage = "";
      try {
        if (this.apiMode === "demo") {
          this.tickets = this.tickets.filter((ticket) => ticket.id !== id);
          this.ticketComments = [];
          return;
        }
        await deleteTicket(id);
        this.tickets = this.tickets.filter((ticket) => ticket.id !== id);
        this.ticketComments = [];
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.errorMessage = "Unable to delete ticket.";
        throw err;
      }
    },
    async bulkTicketOperation(
      projectId: string,
      payload: BulkTicketOperationRequest,
    ): Promise<BulkTicketOperationResponse> {
      this.errorMessage = "";
      if (this.apiMode === "demo") {
        const results = payload.ticketIds.map((ticketId) => ({
          ticketId,
          success: true,
        }));
        if (payload.action === "delete") {
          this.tickets = this.tickets.filter(
            (ticket) => !payload.ticketIds.includes(ticket.id),
          );
        } else if (payload.action === "move_state" && payload.stateId) {
          this.tickets = this.tickets.map((ticket) =>
            payload.ticketIds.includes(ticket.id)
              ? { ...ticket, stateId: payload.stateId as string }
              : ticket,
          );
        } else if (payload.action === "assign" && payload.assigneeId) {
          const assignee = this.tickets.find(
            (ticket) => ticket.assignee?.id === payload.assigneeId,
          )?.assignee;
          this.tickets = this.tickets.map((ticket) =>
            payload.ticketIds.includes(ticket.id)
              ? { ...ticket, assigneeId: payload.assigneeId, assignee }
              : ticket,
          );
        } else if (payload.action === "set_priority" && payload.priority) {
          this.tickets = this.tickets.map((ticket) =>
            payload.ticketIds.includes(ticket.id)
              ? { ...ticket, priority: payload.priority as TicketPriority }
              : ticket,
          );
        }
        return {
          action: payload.action,
          total: payload.ticketIds.length,
          successCount: payload.ticketIds.length,
          errorCount: 0,
          results,
        };
      }
      return apiBulkTicketOperation(projectId, payload);
    },
    async loadTicketComments(ticketId: string) {
      this.commentError = "";
      try {
        const list = await listTicketComments(ticketId);
        this.ticketComments = list.items;
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.ticketComments = [];
        this.commentError = "Unable to load comments.";
      }
    },
    async loadTicketActivities(ticketId: string) {
      try {
        const list = await listTicketActivities(ticketId);
        this.ticketActivities = list.items;
      } catch {
        this.ticketActivities = [];
      }
    },
    async addTicketComment(
      ticketId: string,
      message: string,
      authorName: string,
    ) {
      this.commentSaving = true;
      this.commentError = "";
      try {
        if (this.apiMode === "demo") {
          const comment: TicketComment = {
            id: `demo-comment-${Date.now()}`,
            ticketId,
            authorId: "demo-user",
            authorName,
            message,
            createdAt: new Date().toISOString(),
          };
          this.ticketComments = [...this.ticketComments, comment];
          return comment;
        }
        await addTicketComment(ticketId, { message });
        await this.loadTicketComments(ticketId);
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.commentError = "Unable to add comment.";
        throw err;
      } finally {
        this.commentSaving = false;
      }
      return null;
    },
    async loadTicketAttachments(projectId: string, ticketId: string) {
      this.attachmentError = "";
      try {
        const list = await listTicketAttachments(projectId, ticketId);
        this.ticketAttachments = list.items;
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.ticketAttachments = [];
        this.attachmentError = "Unable to load attachments.";
      }
    },
    async loadTicketDependencies(ticketId: string) {
      this.ticketDependenciesLoading = true;
      try {
        const list = await apiListTicketDependencies(ticketId);
        this.ticketDependencies = list.items;
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.ticketDependencies = [];
      } finally {
        this.ticketDependenciesLoading = false;
      }
    },
    async createTicketDependency(
      ticketId: string,
      payload: TicketDependencyCreateRequest,
    ) {
      const created = await apiCreateTicketDependency(ticketId, payload);
      this.ticketDependencies = [created, ...this.ticketDependencies];
      return created;
    },
    async removeTicketDependency(ticketId: string, dependencyId: string) {
      await apiDeleteTicketDependency(ticketId, dependencyId);
      this.ticketDependencies = this.ticketDependencies.filter(
        (item) => item.id !== dependencyId,
      );
    },
    async loadDependencyGraph(
      projectId: string,
      opts: { rootTicketId?: string; depth?: number } = {},
    ) {
      this.ticketDependencyGraphLoading = true;
      try {
        this.ticketDependencyGraph = await apiGetProjectDependencyGraph(
          projectId,
          opts,
        );
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.ticketDependencyGraph = { nodes: [], edges: [] };
      } finally {
        this.ticketDependencyGraphLoading = false;
      }
    },
    async uploadAttachment(projectId: string, ticketId: string, file: File) {
      this.attachmentUploading = true;
      this.attachmentError = "";
      try {
        await uploadTicketAttachment(projectId, ticketId, file);
        await this.loadTicketAttachments(projectId, ticketId);
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.attachmentError = "Unable to upload file.";
        throw err;
      } finally {
        this.attachmentUploading = false;
      }
    },
    async removeAttachment(
      projectId: string,
      ticketId: string,
      attachmentId: string,
    ) {
      this.attachmentError = "";
      try {
        await deleteTicketAttachment(projectId, ticketId, attachmentId);
        this.ticketAttachments = this.ticketAttachments.filter(
          (a) => a.id !== attachmentId,
        );
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.attachmentError = "Unable to delete attachment.";
        throw err;
      }
    },
    async loadWorkflowEditor(projectId: string) {
      this.workflowEditorLoading = true;
      this.workflowEditorError = "";
      try {
        const result = await getWorkflow(projectId);
        this.workflowEditorStates = result.states.map((s) => ({ ...s }));
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.workflowEditorError = "Unable to load workflow states.";
      } finally {
        this.workflowEditorLoading = false;
      }
    },
    async saveWorkflowEditor(projectId: string) {
      this.workflowEditorSaving = true;
      this.workflowEditorError = "";
      try {
        const inputs: WorkflowStateInput[] = this.workflowEditorStates.map(
          (s, i) => {
            const input: WorkflowStateInput = {
              name: s.name,
              order: i + 1,
              isDefault: s.isDefault,
              isClosed: s.isClosed,
            };
            // New client-side rows use temporary ids ("new-*"); omit them.
            if (s.id && !s.id.startsWith("new-")) {
              input.id = s.id;
            }
            return input;
          },
        );
        const result = await updateWorkflow(projectId, inputs);
        this.workflowEditorStates = result.states.map((s) => ({ ...s }));
        this.states = result.states;
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.workflowEditorError = "Unable to save workflow.";
        throw err;
      } finally {
        this.workflowEditorSaving = false;
      }
    },
    async loadTicketTimeEntries(projectId: string, ticketId: string) {
      this.ticketTimeEntriesLoading = true;
      try {
        const result = await apiListTicketTimeEntries(projectId, ticketId);
        this.ticketTimeEntries = result.items;
        this.ticketTimeEntriesTotalMinutes = result.totalMinutes;
      } catch (err) {
        if (isAuthError(err)) {
          throw err;
        }
        this.ticketTimeEntries = [];
        this.ticketTimeEntriesTotalMinutes = 0;
      } finally {
        this.ticketTimeEntriesLoading = false;
      }
    },
    async createTicketTimeEntry(
      projectId: string,
      ticketId: string,
      payload: TimeEntryCreateRequest,
    ) {
      const created = await apiCreateTicketTimeEntry(projectId, ticketId, payload);
      this.ticketTimeEntries = [created, ...this.ticketTimeEntries];
      this.ticketTimeEntriesTotalMinutes += created.minutes;
      return created;
    },
    async removeTicketTimeEntry(
      projectId: string,
      ticketId: string,
      timeEntryId: string,
    ) {
      const entry = this.ticketTimeEntries.find((e) => e.id === timeEntryId);
      await apiDeleteTicketTimeEntry(projectId, ticketId, timeEntryId);
      this.ticketTimeEntries = this.ticketTimeEntries.filter(
        (e) => e.id !== timeEntryId,
      );
      if (entry) {
        this.ticketTimeEntriesTotalMinutes -= entry.minutes;
      }
    },
  },
});
