import { TicketingAPIClient, Ticket } from "../api-client.js";
import { z } from "zod";

export const getTicketSchema = z.object({
  ticketId: z.string().describe("The ticket ID (UUID)"),
});

export const listTicketsSchema = z.object({
  projectId: z.string().describe("The project ID (UUID)"),
  stateId: z.string().optional().describe("Filter by state ID"),
  assigneeId: z.string().optional().describe("Filter by assignee ID"),
  query: z.string().optional().describe("Search query"),
  limit: z.number().optional().default(50).describe("Results limit"),
  offset: z.number().optional().default(0).describe("Results offset"),
});

export const searchTicketsSchema = z.object({
  query: z.string().describe("Search query"),
  projectId: z.string().optional().describe("Project ID to search in"),
});

export async function getTicket(
  client: TicketingAPIClient,
  params: z.infer<typeof getTicketSchema>,
): Promise<Record<string, unknown>> {
  const ticket = await client.getTicket(params.ticketId);

  // Fetch comments for this ticket
  const comments = await client.getComments(params.ticketId);

  return {
    ...ticket,
    comments,
  };
}

export async function listTickets(
  client: TicketingAPIClient,
  params: z.infer<typeof listTicketsSchema>,
): Promise<{ items: Ticket[]; total: number; query?: string }> {
  const result = await client.listTickets(params.projectId, {
    stateId: params.stateId,
    assigneeId: params.assigneeId,
    query: params.query,
    limit: params.limit,
    offset: params.offset,
  });

  return {
    ...result,
    query: params.query,
  };
}

export async function searchTickets(
  client: TicketingAPIClient,
  params: z.infer<typeof searchTicketsSchema>,
): Promise<{ items: Ticket[]; total: number }> {
  // If projectId is provided, search within that project
  if (params.projectId) {
    return client.listTickets(params.projectId, { query: params.query });
  }

  // Otherwise, get all projects and search across them
  const projects = await client.listProjects();
  const allTickets: Ticket[] = [];

  for (const project of projects) {
    const result = await client.listTickets(project.id, {
      query: params.query,
    });
    allTickets.push(...result.items);
  }

  return {
    items: allTickets,
    total: allTickets.length,
  };
}
