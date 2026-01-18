import { KeycloakAuth } from "./auth.js";

export interface TicketComment {
  id: string;
  ticketId: string;
  authorId: string;
  authorName: string;
  message: string;
  createdAt: string;
}

export interface WorkflowState {
  id: string;
  projectId: string;
  name: string;
  order: number;
  isDefault: boolean;
  isClosed: boolean;
}

export interface UserSummary {
  id: string;
  name: string;
  email?: string;
}

export interface Ticket {
  id: string;
  key: string;
  number: number;
  type: "feature" | "bug";
  projectId: string;
  projectKey: string;
  storyId: string | null;
  story?: {
    id: string;
    projectId: string;
    title: string;
    description: string;
    createdAt: string;
    updatedAt: string;
  };
  title: string;
  description: string;
  stateId: string;
  state: WorkflowState;
  assigneeId: string | null;
  assignee?: UserSummary;
  priority: "low" | "medium" | "high" | "urgent";
  position: number;
  createdAt: string;
  updatedAt: string;
}

export interface Project {
  id: string;
  key: string;
  name: string;
  description: string;
  createdAt: string;
  updatedAt: string;
}

export interface BoardResponse {
  project: Project;
  states: WorkflowState[];
  tickets: Ticket[];
}

export class TicketingAPIClient {
  private baseUrl: string;
  private auth: KeycloakAuth;

  constructor(auth: KeycloakAuth, baseUrl?: string) {
    this.auth = auth;
    this.baseUrl = baseUrl || process.env.TICKETING_API_BASE_URL || "http://ticketing-api:8080";
  }

  private async request<T>(path: string, options: RequestInit = {}): Promise<T> {
    const token = await this.auth.getAccessToken();
    const url = `${this.baseUrl}${path}`;

    const response = await fetch(url, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
        ...(options.headers || {}),
      },
    });

    if (!response.ok) {
      const errorBody = await response.text();
      throw new Error(
        `API request failed: ${response.status} ${response.statusText} - ${errorBody}`
      );
    }

    return response.json() as Promise<T>;
  }

  async getTicket(id: string): Promise<Ticket> {
    return this.request<Ticket>(`/tickets/${id}`);
  }

  async listTickets(
    projectId: string,
    options?: {
      stateId?: string;
      assigneeId?: string;
      query?: string;
      limit?: number;
      offset?: number;
    }
  ): Promise<{ items: Ticket[]; total: number }> {
    const params = new URLSearchParams();
    if (options?.stateId) params.append("stateId", options.stateId);
    if (options?.assigneeId) params.append("assigneeId", options.assigneeId);
    if (options?.query) params.append("q", options.query);
    if (options?.limit) params.append("limit", options.limit.toString());
    if (options?.offset) params.append("offset", options.offset.toString());

    const query = params.toString();
    const path = `/projects/${projectId}/tickets${query ? `?${query}` : ""}`;

    return this.request<{ items: Ticket[]; total: number }>(path);
  }

  async getComments(ticketId: string): Promise<TicketComment[]> {
    const response = await this.request<{ items: TicketComment[] }>(
      `/tickets/${ticketId}/comments`
    );
    return response.items;
  }

  async addComment(ticketId: string, message: string): Promise<TicketComment> {
    return this.request<TicketComment>(`/tickets/${ticketId}/comments`, {
      method: "POST",
      body: JSON.stringify({ message }),
    });
  }

  async updateTicket(
    id: string,
    update: Partial<Ticket>
  ): Promise<Ticket> {
    return this.request<Ticket>(`/tickets/${id}`, {
      method: "PATCH",
      body: JSON.stringify(update),
    });
  }

  async getWorkflow(projectId: string): Promise<WorkflowState[]> {
    const response = await this.request<{ states: WorkflowState[] }>(
      `/projects/${projectId}/workflow`
    );
    return response.states;
  }

  async getBoard(projectId: string): Promise<BoardResponse> {
    return this.request<BoardResponse>(`/projects/${projectId}/board`);
  }

  async listProjects(): Promise<Project[]> {
    const response = await this.request<{ items: Project[] }>("/projects");
    return response.items;
  }
}

export function createAPIClient(auth: KeycloakAuth): TicketingAPIClient {
  return new TicketingAPIClient(auth);
}
