import { TicketingAPIClient, TicketComment } from "../api-client.js";
import { z } from "zod";

export const addCommentSchema = z.object({
  ticketId: z.string().describe("The ticket ID (UUID)"),
  message: z.string().describe("Comment message"),
});

export async function addComment(
  client: TicketingAPIClient,
  params: z.infer<typeof addCommentSchema>
): Promise<TicketComment> {
  return client.addComment(params.ticketId, params.message);
}
