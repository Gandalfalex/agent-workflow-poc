import type { TicketResponse } from "./api";

export type StoryRow = {
  id: string;
  title: string;
  description?: string;
  storyPoints?: number | null;
  ticketsByState: Record<string, TicketResponse[]>;
};
