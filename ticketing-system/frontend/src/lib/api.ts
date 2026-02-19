import type { components } from "./api.schema";

export type WorkflowState = components["schemas"]["WorkflowState"];
export type WorkflowStateInput = components["schemas"]["WorkflowStateInput"];
export type WorkflowResponse = components["schemas"]["WorkflowResponse"];
export type TicketResponse = components["schemas"]["Ticket"];
export type TicketPriority = components["schemas"]["TicketPriority"];
export type TicketType = components["schemas"]["TicketType"];
export type BoardResponse = components["schemas"]["BoardResponse"];
export type TicketCreateRequest = components["schemas"]["TicketCreateRequest"];
export type TicketUpdateRequest = components["schemas"]["TicketUpdateRequest"];
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
export type AuthUser = components["schemas"]["User"];
export type UserSummary = components["schemas"]["UserSummary"];
export type UserListResponse = components["schemas"]["UserListResponse"];
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
export type ProjectGroupUpdateRequest =
  components["schemas"]["ProjectGroupUpdateRequest"];

const API_BASE = (
  import.meta.env.VITE_API_BASE ||
  (import.meta.env.BASE_URL === "/"
    ? ""
    : import.meta.env.BASE_URL.replace(/\/$/, ""))
).replace(/\/$/, "");
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
  const res = await fetch(`${API_BASE}${path}`, {
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...(options.headers || {}),
    },
    ...options,
  });

  if (!res.ok) {
    const message = await res.text().catch(() => "");
    const error = new Error(message || res.statusText);
    (error as Error & { status?: number }).status = res.status;
    throw error;
  }

  if (res.status === 204) {
    return undefined as T;
  }

  return (await res.json()) as T;
}

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
