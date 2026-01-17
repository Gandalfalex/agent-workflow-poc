import { defineStore } from "pinia";
import {
  addTicketComment,
  createStory,
  createTicket,
  deleteStory,
  deleteTicket,
  deleteWebhook,
  getBoard,
  listStories,
  listTicketComments,
  listWebhooks,
  testWebhook,
  updateTicket,
  updateWorkflow,
  type Story,
  type TicketComment,
  type TicketCreateRequest,
  type TicketResponse,
  type TicketUpdateRequest,
  type WebhookEvent,
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

const demoTickets: TicketResponse[] = [
  {
    id: "bfc9e1a1-3cd9-4e1d-9f1e-000000000101",
    key: `${demoProjectKey}-101`,
    number: 101,
    type: "feature",
    projectId: demoProjectId,
    projectKey: demoProjectKey,
    title: "Webhook signing for ticket.created",
    description: "Include HMAC signature and retry policy.",
    stateId: demoStateBacklog.id,
    state: demoStateBacklog,
    priority: "high",
    position: 1,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
    assignee: { id: "user-1", name: "Ari" },
  },
  {
    id: "bfc9e1a1-3cd9-4e1d-9f1e-000000000102",
    key: `${demoProjectKey}-102`,
    number: 102,
    type: "feature",
    projectId: demoProjectId,
    projectKey: demoProjectKey,
    title: "Drag-and-drop keyboard support",
    description: "Add arrow move + enter to drop.",
    stateId: demoStateInProgress.id,
    state: demoStateInProgress,
    priority: "medium",
    position: 1,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
    assignee: { id: "user-2", name: "Nova" },
  },
  {
    id: "bfc9e1a1-3cd9-4e1d-9f1e-000000000103",
    key: `${demoProjectKey}-103`,
    number: 103,
    type: "bug",
    projectId: demoProjectId,
    projectKey: demoProjectKey,
    title: "Ticket detail drawer",
    description: "Inline editing for title, description, and assignee.",
    stateId: demoStateReview.id,
    state: demoStateReview,
    priority: "low",
    position: 1,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
    assignee: { id: "user-3", name: "Jules" },
  },
  {
    id: "bfc9e1a1-3cd9-4e1d-9f1e-000000000104",
    key: `${demoProjectKey}-104`,
    number: 104,
    type: "feature",
    projectId: demoProjectId,
    projectKey: demoProjectKey,
    title: "OpenAPI schema sync",
    description: "Auto-generate clients from spec.",
    stateId: demoStateDone.id,
    state: demoStateDone,
    priority: "high",
    position: 1,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
    assignee: { id: "user-4", name: "Sam" },
  },
];

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
  }),
  actions: {
    reset() {
      this.states = [];
      this.tickets = [];
      this.stories = [];
      this.webhooks = [];
      this.ticketComments = [];
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
    },
    applyDemo() {
      this.apiMode = "demo";
      this.states = demoStates;
      this.tickets = demoTickets;
      this.loading = false;
      this.stories = [];
    },
    clearComments() {
      this.ticketComments = [];
      this.commentError = "";
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
        this.errorMessage = "Backend unavailable. Showing demo data.";
        this.applyDemo();
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
        this.stories = [];
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
      payload: { title: string; description?: string },
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
            storyId: payload.storyId || undefined,
            story,
            title: payload.title,
            description: payload.description || "",
            stateId: resolvedState.id,
            state: resolvedState,
            priority: payload.priority || "medium",
            position: 0,
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
  },
});
