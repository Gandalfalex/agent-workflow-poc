import type { components } from "./api.schema";

export type WorkflowState = components["schemas"]["WorkflowState"];
export type WorkflowStateInput = components["schemas"]["WorkflowStateInput"];
export type WorkflowResponse = components["schemas"]["WorkflowResponse"];
export type TicketResponse = components["schemas"]["Ticket"];
export type TicketPriority = components["schemas"]["TicketPriority"];
export type TicketType = components["schemas"]["TicketType"];
export type TicketIncidentSeverity =
  components["schemas"]["TicketIncidentSeverity"];
export type BoardResponse = components["schemas"]["BoardResponse"];
export type TicketCreateRequest = components["schemas"]["TicketCreateRequest"];
export type TicketUpdateRequest = components["schemas"]["TicketUpdateRequest"];
export type BulkTicketAction = components["schemas"]["BulkTicketAction"];
export type BulkTicketOperationRequest =
  components["schemas"]["BulkTicketOperationRequest"];
export type BulkTicketOperationResult =
  components["schemas"]["BulkTicketOperationResult"];
export type BulkTicketOperationResponse =
  components["schemas"]["BulkTicketOperationResponse"];
export type DependencyRelationType =
  components["schemas"]["DependencyRelationType"];
export type TicketDependency = components["schemas"]["TicketDependency"];
export type TicketDependencyCreateRequest =
  components["schemas"]["TicketDependencyCreateRequest"];
export type TicketDependencyListResponse =
  components["schemas"]["TicketDependencyListResponse"];
export type TicketDependencyGraphNode =
  components["schemas"]["TicketDependencyGraphNode"];
export type TicketDependencyGraphEdge =
  components["schemas"]["TicketDependencyGraphEdge"];
export type TicketDependencyGraphResponse =
  components["schemas"]["TicketDependencyGraphResponse"];
export type Story = components["schemas"]["Story"];
export type StoryListResponse = components["schemas"]["StoryListResponse"];
export type StoryCreateRequest = components["schemas"]["StoryCreateRequest"];
export type StoryUpdateRequest = components["schemas"]["StoryUpdateRequest"];
export type TicketComment = components["schemas"]["TicketComment"];
export type TicketCommentListResponse =
  components["schemas"]["TicketCommentListResponse"];
export type TicketCommentCreateRequest =
  components["schemas"]["TicketCommentCreateRequest"];
export type TicketActivity = components["schemas"]["TicketActivity"];
export type TicketActivityListResponse =
  components["schemas"]["TicketActivityListResponse"];
export type IncidentTimelineItem = components["schemas"]["IncidentTimelineItem"];
export type IncidentTimelineResponse =
  components["schemas"]["IncidentTimelineResponse"];
export type AuthUser = components["schemas"]["User"];
export type UserSummary = components["schemas"]["UserSummary"];
export type UserListResponse = components["schemas"]["UserListResponse"];
export type SyncUsersResponse = components["schemas"]["SyncUsersResponse"];
export type AdminUserCreateRequest =
  components["schemas"]["AdminUserCreateRequest"];
export type LoginResponse = components["schemas"]["AuthLoginResponse"];
export type WebhookResponse = components["schemas"]["Webhook"];
export type WebhookCreateRequest =
  components["schemas"]["WebhookCreateRequest"];
export type WebhookUpdateRequest =
  components["schemas"]["WebhookUpdateRequest"];
export type WebhookListResponse = components["schemas"]["WebhookListResponse"];
export type WebhookTestRequest = components["schemas"]["WebhookTestRequest"];
export type WebhookTestResponse = components["schemas"]["WebhookTestResponse"];
export type WebhookEvent = components["schemas"]["WebhookEvent"];
export type WebhookDelivery = components["schemas"]["WebhookDelivery"];
export type WebhookDeliveryListResponse =
  components["schemas"]["WebhookDeliveryListResponse"];
export type Attachment = components["schemas"]["Attachment"];
export type AttachmentListResponse =
  components["schemas"]["AttachmentListResponse"];
export type Project = components["schemas"]["Project"];
export type ProjectRole = components["schemas"]["ProjectRole"];
export type ProjectListResponse = components["schemas"]["ProjectListResponse"];
export type ProjectCreateRequest =
  components["schemas"]["ProjectCreateRequest"];
export type ProjectUpdateRequest =
  components["schemas"]["ProjectUpdateRequest"];
export type Group = components["schemas"]["Group"];
export type GroupListResponse = components["schemas"]["GroupListResponse"];
export type GroupCreateRequest = components["schemas"]["GroupCreateRequest"];
export type GroupUpdateRequest = components["schemas"]["GroupUpdateRequest"];
export type GroupMember = components["schemas"]["GroupMember"];
export type GroupMemberListResponse =
  components["schemas"]["GroupMemberListResponse"];
export type GroupMemberCreateRequest =
  components["schemas"]["GroupMemberCreateRequest"];
export type ProjectGroup = components["schemas"]["ProjectGroup"];
export type ProjectGroupListResponse =
  components["schemas"]["ProjectGroupListResponse"];
export type ProjectGroupCreateRequest =
  components["schemas"]["ProjectGroupCreateRequest"];
export type StatCount = components["schemas"]["StatCount"];
export type ProjectStats = components["schemas"]["ProjectStats"];
export type ProjectActivity = components["schemas"]["ProjectActivity"];
export type ProjectActivityListResponse =
  components["schemas"]["ProjectActivityListResponse"];
export type DateValuePoint = components["schemas"]["DateValuePoint"];
export type StateOpenPoint = components["schemas"]["StateOpenPoint"];
export type ProjectReportingSummary =
  components["schemas"]["ProjectReportingSummary"];
export type ProjectReportingExportJson =
  components["schemas"]["ProjectReportingExportJson"];
export type Sprint = components["schemas"]["Sprint"];
export type SprintListResponse = components["schemas"]["SprintListResponse"];
export type SprintCreateRequest = components["schemas"]["SprintCreateRequest"];
export type CapacitySetting = components["schemas"]["CapacitySetting"];
export type CapacitySettingInput =
  components["schemas"]["CapacitySettingInput"];
export type CapacitySettingsResponse =
  components["schemas"]["CapacitySettingsResponse"];
export type CapacitySettingsReplaceRequest =
  components["schemas"]["CapacitySettingsReplaceRequest"];
export type SprintForecastSummary =
  components["schemas"]["SprintForecastSummary"];
export type AiTriageField = components["schemas"]["AiTriageField"];
export type AiTriageConfidence = components["schemas"]["AiTriageConfidence"];
export type AiTriageSettings = components["schemas"]["AiTriageSettings"];
export type AiTriageSettingsUpdateRequest =
  components["schemas"]["AiTriageSettingsUpdateRequest"];
export type AiTriageSuggestion = components["schemas"]["AiTriageSuggestion"];
export type AiTriageSuggestionCreateRequest =
  components["schemas"]["AiTriageSuggestionCreateRequest"];
export type AiTriageSuggestionDecision =
  components["schemas"]["AiTriageSuggestionDecision"];
export type AiTriageSuggestionDecisionRequest =
  components["schemas"]["AiTriageSuggestionDecisionRequest"];
export type ProjectLiveEventType =
  components["schemas"]["ProjectLiveEventType"];
export type ProjectLiveEvent = components["schemas"]["ProjectLiveEvent"];
export type Notification = components["schemas"]["Notification"];
export type NotificationListResponse =
  components["schemas"]["NotificationListResponse"];
export type NotificationPreferences =
  components["schemas"]["NotificationPreferences"];
export type NotificationPreferencesUpdateRequest =
  components["schemas"]["NotificationPreferencesUpdateRequest"];
export type NotificationUnreadCountResponse =
  components["schemas"]["NotificationUnreadCountResponse"];
export type NotificationMarkAllResponse =
  components["schemas"]["NotificationMarkAllResponse"];
export type ProjectGroupUpdateRequest =
  components["schemas"]["ProjectGroupUpdateRequest"];
export type BoardFilter = components["schemas"]["BoardFilter"];
export type BoardFilterPreset = components["schemas"]["BoardFilterPreset"];
export type BoardFilterPresetCreateRequest =
  components["schemas"]["BoardFilterPresetCreateRequest"];
export type BoardFilterPresetUpdateRequest =
  components["schemas"]["BoardFilterPresetUpdateRequest"];
export type BoardFilterPresetListResponse =
  components["schemas"]["BoardFilterPresetListResponse"];
export type TimeEntry = components["schemas"]["TimeEntry"];
export type TimeEntryCreateRequest =
  components["schemas"]["TimeEntryCreateRequest"];
export type TimeEntryListResponse =
  components["schemas"]["TimeEntryListResponse"];

const API_BASE = (
  import.meta.env.VITE_API_BASE ||
  (import.meta.env.BASE_URL || "/").replace(/\/$/, "") + "/rest/v1"
).replace(/\/$/, "");

const buildApiUrl = (path: string, base = API_BASE) => `${base}${path}`;

const shouldRetryWithoutBase = (res: Response) =>
  !!API_BASE && (res.status === 404 || res.status === 405);
const DEFAULT_PROJECT_ID = import.meta.env.VITE_PROJECT_ID || "";

function resolveProjectId(projectId?: string): string {
  const id = projectId || DEFAULT_PROJECT_ID;
  if (!id) {
    throw new Error(
      "Missing project id. Set VITE_PROJECT_ID or pass projectId.",
    );
  }
  return id;
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const bases = API_BASE ? [API_BASE, ""] : [API_BASE];

  for (const base of bases) {
    let res: Response;
    try {
      res = await fetch(buildApiUrl(path, base), {
        credentials: "include",
        headers: {
          Accept: "application/json",
          "Content-Type": "application/json",
          ...(options.headers || {}),
        },
        ...options,
      });
    } catch (cause) {
      if (base === API_BASE && API_BASE) {
        continue;
      }
      const error = new Error(`request_failed:${path}`);
      (
        error as Error & {
          status?: number;
          cause?: unknown;
        }
      ).status = 0;
      (error as Error & { cause?: unknown }).cause = cause;
      throw error;
    }

    if (!res.ok) {
      if (base === API_BASE && shouldRetryWithoutBase(res)) {
        continue;
      }
      const message = await res.text().catch(() => "");
      const error = new Error(message || res.statusText);
      (error as Error & { status?: number }).status = res.status;
      throw error;
    }

    if (res.status === 204) {
      return undefined as T;
    }

    try {
      return (await res.json()) as T;
    } catch {
      // When frontend and API base paths are misaligned, the frontend shell can
      // be returned as HTML with 200. Retry once against root API path.
      if (base === API_BASE && API_BASE) {
        continue;
      }
      const error = new Error("invalid_api_response");
      (error as Error & { status?: number }).status = res.status;
      throw error;
    }
  }

  const error = new Error(`request_failed:${path}`);
  (error as Error & { status?: number }).status = 0;
  throw error;
}

async function requestText(path: string, options: RequestInit = {}): Promise<string> {
  let res = await fetch(buildApiUrl(path), {
    credentials: "include",
    headers: {
      Accept: "text/plain, */*",
      ...(options.headers || {}),
    },
    ...options,
  });

  if (!res.ok && shouldRetryWithoutBase(res)) {
    res = await fetch(buildApiUrl(path, ""), {
      credentials: "include",
      headers: {
        Accept: "text/plain, */*",
        ...(options.headers || {}),
      },
      ...options,
    });
  }

  if (!res.ok) {
    const message = await res.text().catch(() => "");
    const error = new Error(message || res.statusText);
    (error as Error & { status?: number }).status = res.status;
    throw error;
  }
  return res.text();
}

export type ReportingExportFormat = "json" | "csv";

export type ReportingExportFile = {
  blob: Blob;
  contentType: string;
  filename: string;
};

export async function getBoard(projectId?: string): Promise<BoardResponse> {
  const id = resolveProjectId(projectId);
  return request<BoardResponse>(`/projects/${id}/board`);
}

export async function listProjects(): Promise<ProjectListResponse> {
  return request<ProjectListResponse>("/projects");
}

export async function createProject(
  payload: ProjectCreateRequest,
): Promise<Project> {
  return request<Project>("/projects", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function updateProject(
  projectId: string,
  payload: ProjectUpdateRequest,
): Promise<Project> {
  return request<Project>(`/projects/${projectId}`, {
    method: "PATCH",
    body: JSON.stringify(payload),
  });
}

export async function deleteProject(projectId: string): Promise<void> {
  await request<void>(`/projects/${projectId}`, {
    method: "DELETE",
  });
}

export async function listGroups(): Promise<GroupListResponse> {
  return request<GroupListResponse>("/groups");
}

export async function createGroup(payload: GroupCreateRequest): Promise<Group> {
  return request<Group>("/groups", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function updateGroup(
  groupId: string,
  payload: GroupUpdateRequest,
): Promise<Group> {
  return request<Group>(`/groups/${groupId}`, {
    method: "PATCH",
    body: JSON.stringify(payload),
  });
}

export async function deleteGroup(groupId: string): Promise<void> {
  await request<void>(`/groups/${groupId}`, {
    method: "DELETE",
  });
}

export async function listGroupMembers(
  groupId: string,
): Promise<GroupMemberListResponse> {
  return request<GroupMemberListResponse>(`/groups/${groupId}/members`);
}

export async function addGroupMember(
  groupId: string,
  payload: GroupMemberCreateRequest,
): Promise<GroupMember> {
  return request<GroupMember>(`/groups/${groupId}/members`, {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function deleteGroupMember(
  groupId: string,
  userId: string,
): Promise<void> {
  await request<void>(`/groups/${groupId}/members/${userId}`, {
    method: "DELETE",
  });
}

export async function listProjectGroups(
  projectId?: string,
): Promise<ProjectGroupListResponse> {
  const id = resolveProjectId(projectId);
  return request<ProjectGroupListResponse>(`/projects/${id}/groups`);
}

export async function addProjectGroup(
  projectId: string | undefined,
  payload: ProjectGroupCreateRequest,
): Promise<ProjectGroup> {
  const id = resolveProjectId(projectId);
  return request<ProjectGroup>(`/projects/${id}/groups`, {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function updateProjectGroup(
  projectId: string | undefined,
  groupId: string,
  payload: ProjectGroupUpdateRequest,
): Promise<ProjectGroup> {
  const id = resolveProjectId(projectId);
  return request<ProjectGroup>(`/projects/${id}/groups/${groupId}`, {
    method: "PATCH",
    body: JSON.stringify(payload),
  });
}

export async function deleteProjectGroup(
  projectId: string | undefined,
  groupId: string,
): Promise<void> {
  const id = resolveProjectId(projectId);
  await request<void>(`/projects/${id}/groups/${groupId}`, {
    method: "DELETE",
  });
}

export async function updateWorkflow(
  projectId: string | undefined,
  states: WorkflowStateInput[],
): Promise<WorkflowResponse> {
  const id = resolveProjectId(projectId);
  return request<WorkflowResponse>(`/projects/${id}/workflow`, {
    method: "PUT",
    body: JSON.stringify({ states }),
  });
}

export async function login(
  identifier: string,
  password: string,
): Promise<LoginResponse> {
  return request<LoginResponse>("/auth/login", {
    method: "POST",
    body: JSON.stringify({ identifier, password }),
  });
}

export async function logout(): Promise<void> {
  await request<void>("/auth/logout", {
    method: "POST",
  });
}

export async function getMe(): Promise<AuthUser> {
  return request<AuthUser>("/auth/me");
}

export async function listUsers(query?: string): Promise<UserListResponse> {
  const params = query ? `?q=${encodeURIComponent(query)}` : "";
  return request<UserListResponse>(`/users${params}`);
}

export async function syncUsersFromIdentityProvider(): Promise<SyncUsersResponse> {
  return request<SyncUsersResponse>("/admin/sync-users", {
    method: "POST",
  });
}

export async function createAdminUser(
  payload: AdminUserCreateRequest,
): Promise<UserSummary> {
  return request<UserSummary>("/admin/users", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function listStories(
  projectId?: string,
): Promise<StoryListResponse> {
  const id = resolveProjectId(projectId);
  return request<StoryListResponse>(`/projects/${id}/stories`);
}

export async function createStory(
  projectId: string | undefined,
  payload: StoryCreateRequest,
): Promise<Story> {
  const id = resolveProjectId(projectId);
  return request<Story>(`/projects/${id}/stories`, {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function updateStory(
  id: string,
  payload: StoryUpdateRequest,
): Promise<Story> {
  return request<Story>(`/stories/${id}`, {
    method: "PATCH",
    body: JSON.stringify(payload),
  });
}

export async function deleteStory(id: string): Promise<void> {
  await request<void>(`/stories/${id}`, {
    method: "DELETE",
  });
}

export async function listTicketActivities(
  ticketId: string,
): Promise<TicketActivityListResponse> {
  return request<TicketActivityListResponse>(
    `/tickets/${ticketId}/activities`,
  );
}

export async function listTicketIncidentTimeline(
  ticketId: string,
): Promise<IncidentTimelineResponse> {
  return request<IncidentTimelineResponse>(`/tickets/${ticketId}/incident-timeline`);
}

export async function getTicketIncidentPostmortem(
  ticketId: string,
): Promise<string> {
  return requestText(`/tickets/${ticketId}/incident-postmortem`);
}

export async function listTicketComments(
  ticketId: string,
): Promise<TicketCommentListResponse> {
  return request<TicketCommentListResponse>(`/tickets/${ticketId}/comments`);
}

export async function addTicketComment(
  ticketId: string,
  payload: TicketCommentCreateRequest,
): Promise<TicketComment> {
  return request<TicketComment>(`/tickets/${ticketId}/comments`, {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function deleteTicketComment(
  ticketId: string,
  commentId: string,
): Promise<void> {
  await request<void>(`/tickets/${ticketId}/comments/${commentId}`, {
    method: "DELETE",
  });
}

export async function createTicket(
  projectId: string | undefined,
  payload: TicketCreateRequest,
): Promise<TicketResponse> {
  const id = resolveProjectId(projectId);
  return request<TicketResponse>(`/projects/${id}/tickets`, {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function updateTicket(
  id: string,
  payload: TicketUpdateRequest,
): Promise<TicketResponse> {
  return request<TicketResponse>(`/tickets/${id}`, {
    method: "PATCH",
    body: JSON.stringify(payload),
  });
}

export async function deleteTicket(id: string): Promise<void> {
  await request<void>(`/tickets/${id}`, {
    method: "DELETE",
  });
}

export async function listTicketDependencies(
  ticketId: string,
): Promise<TicketDependencyListResponse> {
  return request<TicketDependencyListResponse>(`/tickets/${ticketId}/dependencies`);
}

export async function createTicketDependency(
  ticketId: string,
  payload: TicketDependencyCreateRequest,
): Promise<TicketDependency> {
  return request<TicketDependency>(`/tickets/${ticketId}/dependencies`, {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function deleteTicketDependency(
  ticketId: string,
  dependencyId: string,
): Promise<void> {
  await request<void>(`/tickets/${ticketId}/dependencies/${dependencyId}`, {
    method: "DELETE",
  });
}

export async function getProjectDependencyGraph(
  projectId: string,
  opts: { rootTicketId?: string; depth?: number } = {},
): Promise<TicketDependencyGraphResponse> {
  const id = resolveProjectId(projectId);
  const q = new URLSearchParams();
  if (opts.rootTicketId) q.set("rootTicketId", opts.rootTicketId);
  if (typeof opts.depth === "number") q.set("depth", String(opts.depth));
  const suffix = q.toString() ? `?${q.toString()}` : "";
  return request<TicketDependencyGraphResponse>(
    `/projects/${id}/dependency-graph${suffix}`,
  );
}

export async function bulkTicketOperation(
  projectId: string | undefined,
  payload: BulkTicketOperationRequest,
): Promise<BulkTicketOperationResponse> {
  const id = resolveProjectId(projectId);
  return request<BulkTicketOperationResponse>(`/projects/${id}/tickets/bulk`, {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function listWebhooks(
  projectId?: string,
): Promise<WebhookListResponse> {
  const id = resolveProjectId(projectId);
  return request<WebhookListResponse>(`/projects/${id}/webhooks`);
}

export async function createWebhook(
  projectId: string | undefined,
  payload: WebhookCreateRequest,
): Promise<WebhookResponse> {
  const id = resolveProjectId(projectId);
  return request<WebhookResponse>(`/projects/${id}/webhooks`, {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function updateWebhook(
  projectId: string | undefined,
  id: string,
  payload: WebhookUpdateRequest,
): Promise<WebhookResponse> {
  const project = resolveProjectId(projectId);
  return request<WebhookResponse>(`/projects/${project}/webhooks/${id}`, {
    method: "PATCH",
    body: JSON.stringify(payload),
  });
}

export async function deleteWebhook(
  projectId: string | undefined,
  id: string,
): Promise<void> {
  const project = resolveProjectId(projectId);
  await request<void>(`/projects/${project}/webhooks/${id}`, {
    method: "DELETE",
  });
}

export async function testWebhook(
  projectId: string | undefined,
  id: string,
  payload: WebhookTestRequest,
): Promise<WebhookTestResponse> {
  const project = resolveProjectId(projectId);
  return request<WebhookTestResponse>(
    `/projects/${project}/webhooks/${id}/test`,
    {
      method: "POST",
      body: JSON.stringify(payload),
    },
  );
}

export async function listWebhookDeliveries(
  projectId: string | undefined,
  webhookId: string,
): Promise<WebhookDeliveryListResponse> {
  const project = resolveProjectId(projectId);
  return request<WebhookDeliveryListResponse>(
    `/projects/${project}/webhooks/${webhookId}/deliveries`,
  );
}

export async function getWorkflow(
  projectId: string,
): Promise<WorkflowResponse> {
  const id = resolveProjectId(projectId);
  return request<WorkflowResponse>(`/projects/${id}/workflow`);
}

export async function getMyProjectRole(
  projectId: string,
): Promise<{ role: ProjectRole }> {
  const id = resolveProjectId(projectId);
  return request<{ role: ProjectRole }>(`/projects/${id}/my-role`);
}

export async function listBoardFilterPresets(
  projectId: string,
): Promise<BoardFilterPresetListResponse> {
  const id = resolveProjectId(projectId);
  return request<BoardFilterPresetListResponse>(`/projects/${id}/board-filters`);
}

export async function createBoardFilterPreset(
  projectId: string,
  payload: BoardFilterPresetCreateRequest,
): Promise<BoardFilterPreset> {
  const id = resolveProjectId(projectId);
  return request<BoardFilterPreset>(`/projects/${id}/board-filters`, {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function updateBoardFilterPreset(
  projectId: string,
  presetId: string,
  payload: BoardFilterPresetUpdateRequest,
): Promise<BoardFilterPreset> {
  const id = resolveProjectId(projectId);
  return request<BoardFilterPreset>(
    `/projects/${id}/board-filters/${presetId}`,
    {
      method: "PATCH",
      body: JSON.stringify(payload),
    },
  );
}

export async function deleteBoardFilterPreset(
  projectId: string,
  presetId: string,
): Promise<void> {
  const id = resolveProjectId(projectId);
  await request<void>(`/projects/${id}/board-filters/${presetId}`, {
    method: "DELETE",
  });
}

export async function getSharedBoardFilterPreset(
  projectId: string,
  token: string,
): Promise<BoardFilterPreset> {
  const id = resolveProjectId(projectId);
  return request<BoardFilterPreset>(
    `/projects/${id}/board-filters/shared/${encodeURIComponent(token)}`,
  );
}

export async function getProjectStats(
  projectId: string,
): Promise<ProjectStats> {
  const id = resolveProjectId(projectId);
  return request<ProjectStats>(`/projects/${id}/stats`);
}

export async function getProjectActivities(
  projectId: string,
  limit = 20,
): Promise<ProjectActivityListResponse> {
  const id = resolveProjectId(projectId);
  return request<ProjectActivityListResponse>(
    `/projects/${id}/activities?limit=${limit}`,
  );
}

export async function getProjectReportingSummary(
  projectId: string,
  opts: { from?: string; to?: string } = {},
): Promise<ProjectReportingSummary> {
  const id = resolveProjectId(projectId);
  const q = new URLSearchParams();
  if (opts.from) q.set("from", opts.from);
  if (opts.to) q.set("to", opts.to);
  const suffix = q.toString() ? `?${q.toString()}` : "";
  return request<ProjectReportingSummary>(
    `/projects/${id}/reporting/summary${suffix}`,
  );
}

export async function listProjectSprints(
  projectId: string,
): Promise<SprintListResponse> {
  const id = resolveProjectId(projectId);
  return request<SprintListResponse>(`/projects/${id}/sprints`);
}

export async function createProjectSprint(
  projectId: string,
  payload: SprintCreateRequest,
): Promise<Sprint> {
  const id = resolveProjectId(projectId);
  return request<Sprint>(`/projects/${id}/sprints`, {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function addSprintTickets(
  projectId: string,
  sprintId: string,
  ticketIds: string[],
): Promise<Sprint> {
  const id = resolveProjectId(projectId);
  return request<Sprint>(`/projects/${id}/sprints/${sprintId}/tickets`, {
    method: "POST",
    body: JSON.stringify({ ticketIds }),
  });
}

export async function removeSprintTickets(
  projectId: string,
  sprintId: string,
  ticketIds: string[],
): Promise<Sprint> {
  const id = resolveProjectId(projectId);
  return request<Sprint>(`/projects/${id}/sprints/${sprintId}/tickets`, {
    method: "DELETE",
    body: JSON.stringify({ ticketIds }),
  });
}

export async function listProjectCapacitySettings(
  projectId: string,
): Promise<CapacitySettingsResponse> {
  const id = resolveProjectId(projectId);
  return request<CapacitySettingsResponse>(`/projects/${id}/capacity-settings`);
}

export async function replaceProjectCapacitySettings(
  projectId: string,
  payload: CapacitySettingsReplaceRequest,
): Promise<CapacitySettingsResponse> {
  const id = resolveProjectId(projectId);
  return request<CapacitySettingsResponse>(`/projects/${id}/capacity-settings`, {
    method: "PUT",
    body: JSON.stringify(payload),
  });
}

export async function getProjectSprintForecast(
  projectId: string,
  opts: { sprintId?: string; iterations?: number } = {},
): Promise<SprintForecastSummary> {
  const id = resolveProjectId(projectId);
  const q = new URLSearchParams();
  if (opts.sprintId) q.set("sprintId", opts.sprintId);
  if (typeof opts.iterations === "number") {
    q.set("iterations", String(opts.iterations));
  }
  const suffix = q.toString() ? `?${q.toString()}` : "";
  return request<SprintForecastSummary>(
    `/projects/${id}/sprint-forecast${suffix}`,
  );
}

export async function getProjectAiTriageSettings(
  projectId: string,
): Promise<AiTriageSettings> {
  const id = resolveProjectId(projectId);
  return request<AiTriageSettings>(`/projects/${id}/ai-triage/settings`);
}

export async function updateProjectAiTriageSettings(
  projectId: string,
  payload: AiTriageSettingsUpdateRequest,
): Promise<AiTriageSettings> {
  const id = resolveProjectId(projectId);
  return request<AiTriageSettings>(`/projects/${id}/ai-triage/settings`, {
    method: "PATCH",
    body: JSON.stringify(payload),
  });
}

export async function createAiTriageSuggestion(
  projectId: string,
  payload: AiTriageSuggestionCreateRequest,
): Promise<AiTriageSuggestion> {
  const id = resolveProjectId(projectId);
  return request<AiTriageSuggestion>(`/projects/${id}/ai-triage/suggestions`, {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function recordAiTriageSuggestionDecision(
  projectId: string,
  suggestionId: string,
  payload: AiTriageSuggestionDecisionRequest,
): Promise<AiTriageSuggestionDecision> {
  const id = resolveProjectId(projectId);
  return request<AiTriageSuggestionDecision>(
    `/projects/${id}/ai-triage/suggestions/${suggestionId}/decision`,
    {
      method: "POST",
      body: JSON.stringify(payload),
    },
  );
}

export async function exportProjectReportingSnapshot(
  projectId: string,
  opts: { from?: string; to?: string; format?: ReportingExportFormat } = {},
): Promise<ReportingExportFile> {
  const id = resolveProjectId(projectId);
  const q = new URLSearchParams();
  if (opts.from) q.set("from", opts.from);
  if (opts.to) q.set("to", opts.to);
  if (opts.format) q.set("format", opts.format);
  const suffix = q.toString() ? `?${q.toString()}` : "";
  const res = await fetch(`${API_BASE}/projects/${id}/reporting/export${suffix}`, {
    credentials: "include",
  });
  if (!res.ok) {
    const message = await res.text().catch(() => "");
    const error = new Error(message || res.statusText);
    (error as Error & { status?: number }).status = res.status;
    throw error;
  }

  const contentType = res.headers.get("Content-Type") || "application/octet-stream";
  const disposition = res.headers.get("Content-Disposition") || "";
  const match = disposition.match(/filename="?([^"]+)"?/i);
  const fallbackExt = opts.format === "csv" ? "csv" : "json";
  const filename = match?.[1] || `project-reporting-export.${fallbackExt}`;

  return {
    blob: await res.blob(),
    contentType,
    filename,
  };
}

export function buildProjectEventsWebSocketUrls(projectId: string): string[] {
  const id = resolveProjectId(projectId);
  const paths = API_BASE
    ? [`${API_BASE}/projects/${id}/events/ws`, `/projects/${id}/events/ws`]
    : [`/projects/${id}/events/ws`];
  const seen = new Set<string>();
  const urls: string[] = [];
  for (const path of paths) {
    const url = new URL(path, window.location.origin);
    url.protocol = url.protocol === "https:" ? "wss:" : "ws:";
    const value = url.toString();
    if (!seen.has(value)) {
      seen.add(value);
      urls.push(value);
    }
  }
  return urls;
}

export async function listNotifications(
  projectId: string,
  opts: { limit?: number; unreadOnly?: boolean } = {},
): Promise<NotificationListResponse> {
  const id = resolveProjectId(projectId);
  const q = new URLSearchParams();
  if (typeof opts.limit === "number") q.set("limit", String(opts.limit));
  if (typeof opts.unreadOnly === "boolean") {
    q.set("unreadOnly", String(opts.unreadOnly));
  }
  const suffix = q.toString() ? `?${q.toString()}` : "";
  return request<NotificationListResponse>(
    `/projects/${id}/notifications${suffix}`,
  );
}

export async function getNotificationUnreadCount(
  projectId: string,
): Promise<NotificationUnreadCountResponse> {
  const id = resolveProjectId(projectId);
  return request<NotificationUnreadCountResponse>(
    `/projects/${id}/notifications/unread-count`,
  );
}

export async function markNotificationRead(
  projectId: string,
  notificationId: string,
): Promise<Notification> {
  const id = resolveProjectId(projectId);
  return request<Notification>(
    `/projects/${id}/notifications/${notificationId}/read`,
    { method: "POST" },
  );
}

export async function markAllNotificationsRead(
  projectId: string,
): Promise<NotificationMarkAllResponse> {
  const id = resolveProjectId(projectId);
  return request<NotificationMarkAllResponse>(
    `/projects/${id}/notifications/read-all`,
    { method: "POST" },
  );
}

export async function getNotificationPreferences(
  projectId: string,
): Promise<NotificationPreferences> {
  const id = resolveProjectId(projectId);
  return request<NotificationPreferences>(
    `/projects/${id}/notification-preferences`,
  );
}

export async function updateNotificationPreferences(
  projectId: string,
  payload: NotificationPreferencesUpdateRequest,
): Promise<NotificationPreferences> {
  const id = resolveProjectId(projectId);
  return request<NotificationPreferences>(
    `/projects/${id}/notification-preferences`,
    { method: "PATCH", body: JSON.stringify(payload) },
  );
}

export async function listTicketAttachments(
  projectId: string,
  ticketId: string,
): Promise<AttachmentListResponse> {
  return request<AttachmentListResponse>(
    `/projects/${projectId}/tickets/${ticketId}/attachments`,
  );
}

export async function uploadTicketAttachment(
  projectId: string,
  ticketId: string,
  file: File,
): Promise<Attachment> {
  const formData = new FormData();
  formData.append("file", file);
  const res = await fetch(
    `${API_BASE}/projects/${projectId}/tickets/${ticketId}/attachments`,
    {
      method: "POST",
      credentials: "include",
      body: formData,
    },
  );
  if (!res.ok) {
    const message = await res.text().catch(() => "");
    const error = new Error(message || res.statusText);
    (error as Error & { status?: number }).status = res.status;
    throw error;
  }
  return (await res.json()) as Attachment;
}

export function downloadTicketAttachmentUrl(
  projectId: string,
  ticketId: string,
  attachmentId: string,
): string {
  return `${API_BASE}/projects/${projectId}/tickets/${ticketId}/attachments/${attachmentId}/download`;
}

export async function deleteTicketAttachment(
  projectId: string,
  ticketId: string,
  attachmentId: string,
): Promise<void> {
  await request<void>(
    `/projects/${projectId}/tickets/${ticketId}/attachments/${attachmentId}`,
    { method: "DELETE" },
  );
}

export async function listTicketTimeEntries(
  projectId: string,
  ticketId: string,
): Promise<TimeEntryListResponse> {
  return request<TimeEntryListResponse>(
    `/projects/${projectId}/tickets/${ticketId}/time-entries`,
  );
}

export async function createTicketTimeEntry(
  projectId: string,
  ticketId: string,
  payload: TimeEntryCreateRequest,
): Promise<TimeEntry> {
  return request<TimeEntry>(
    `/projects/${projectId}/tickets/${ticketId}/time-entries`,
    { method: "POST", body: JSON.stringify(payload) },
  );
}

export async function deleteTicketTimeEntry(
  projectId: string,
  ticketId: string,
  timeEntryId: string,
): Promise<void> {
  await request<void>(
    `/projects/${projectId}/tickets/${ticketId}/time-entries/${timeEntryId}`,
    { method: "DELETE" },
  );
}
