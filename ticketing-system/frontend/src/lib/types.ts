import type { TicketResponse } from "./api";

export type StoryRow = {
  id: string;
  title: string;
  description?: string;
  ticketsByState: Record<string, TicketResponse[]>;
};
