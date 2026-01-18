import { TicketingAPIClient, Ticket, WorkflowState } from "../api-client.js";
import { z } from "zod";

export const updateTicketStateSchema = z.union([
  z.object({
    ticketId: z.string().describe("The ticket ID (UUID)"),
    stateId: z.string().describe("The state ID (UUID)"),
  }),
  z.object({
    ticketId: z.string().describe("The ticket ID (UUID)"),
    stateName: z.string().describe("The state name (e.g., 'Todo', 'In Review', 'Done')"),
  }),
]);

export const getProjectWorkflowSchema = z.object({
  projectId: z.string().describe("The project ID (UUID)"),
});

export async function updateTicketState(
  client: TicketingAPIClient,
  params: z.infer<typeof updateTicketStateSchema>
): Promise<Ticket> {
  // Get current ticket to find its project
  const ticket = await client.getTicket(params.ticketId);
  const projectId = ticket.projectId;

  let stateId: string;

  // If stateName is provided, resolve it to stateId
  if ("stateName" in params && params.stateName) {
    const states = await client.getWorkflow(projectId);
    const state = states.find(
      (s) => s.name.toLowerCase() === params.stateName.toLowerCase()
    );

    if (!state) {
      const availableStates = states.map((s) => s.name).join(", ");
      throw new Error(
        `State "${params.stateName}" not found. Available states: ${availableStates}`
      );
    }

    stateId = state.id;
  } else {
    stateId = (params as any).stateId;
  }

  // Update the ticket with new state
  return client.updateTicket(params.ticketId, {
    stateId,
  });
}

export async function getProjectWorkflow(
  client: TicketingAPIClient,
  params: z.infer<typeof getProjectWorkflowSchema>
): Promise<WorkflowState[]> {
  return client.getWorkflow(params.projectId);
}
