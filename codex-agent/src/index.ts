import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import {
  ListToolsRequestSchema,
  CallToolRequestSchema,
} from "@modelcontextprotocol/sdk/types.js";
import { createAuth } from "./auth.js";
import { createAPIClient } from "./api-client.js";
import {
  getTicket,
  getTicketSchema,
  listTickets,
  listTicketsSchema,
  searchTickets,
  searchTicketsSchema,
} from "./tools/tickets.js";
import { addComment, addCommentSchema } from "./tools/comments.js";
import {
  updateTicketState,
  updateTicketStateSchema,
  getProjectWorkflow,
  getProjectWorkflowSchema,
} from "./tools/workflow.js";
import {
  implementTicket,
  implementTicketSchema,
} from "./tools/implementation.js";

const server = new Server(
  {
    name: "ticketing-mcp",
    version: "1.0.0",
  },
  {
    capabilities: {
      tools: {},
    },
  },
);

// Initialize auth and API client
const auth = createAuth();
const apiClient = createAPIClient(auth);

// Tool definitions
const tools = [
  {
    name: "get_ticket",
    description:
      "Get a specific ticket with all its details including comments",
    inputSchema: getTicketSchema,
  },
  {
    name: "list_tickets",
    description: "List tickets in a project with optional filtering",
    inputSchema: listTicketsSchema,
  },
  {
    name: "search_tickets",
    description:
      "Search for tickets across projects or within a specific project",
    inputSchema: searchTicketsSchema,
  },
  {
    name: "add_comment",
    description: "Add a comment to a ticket",
    inputSchema: addCommentSchema,
  },
  {
    name: "update_ticket_state",
    description:
      "Update a ticket's state/status. Can use stateId (UUID) or stateName (e.g., 'Done', 'In Review')",
    inputSchema: updateTicketStateSchema,
  },
  {
    name: "get_project_workflow",
    description: "Get all workflow states available in a project",
    inputSchema: getProjectWorkflowSchema,
  },
  {
    name: "implement_ticket",
    description:
      "Spawn a subagent to automatically implement a feature from a ticket. Creates isolated workspace, implements feature, runs tests, and updates ticket state.",
    inputSchema: implementTicketSchema,
  },
];

// List tools handler
server.setRequestHandler(ListToolsRequestSchema, async () => {
  return {
    tools: tools.map((tool) => ({
      name: tool.name,
      description: tool.description,
      inputSchema: tool.inputSchema as any,
    })),
  };
});

// Call tool handler
server.setRequestHandler(CallToolRequestSchema, async (request) => {
  const toolName = request.params.name;
  const args = request.params.arguments;

  try {
    let result: any;

    switch (toolName) {
      case "get_ticket":
        result = await getTicket(apiClient, getTicketSchema.parse(args));
        break;

      case "list_tickets":
        result = await listTickets(apiClient, listTicketsSchema.parse(args));
        break;

      case "search_tickets":
        result = await searchTickets(
          apiClient,
          searchTicketsSchema.parse(args),
        );
        break;

      case "add_comment":
        result = await addComment(apiClient, addCommentSchema.parse(args));
        break;

      case "update_ticket_state":
        result = await updateTicketState(
          apiClient,
          updateTicketStateSchema.parse(args),
        );
        break;

      case "get_project_workflow":
        result = await getProjectWorkflow(
          apiClient,
          getProjectWorkflowSchema.parse(args),
        );
        break;

      case "implement_ticket":
        result = await implementTicket(
          apiClient,
          implementTicketSchema.parse(args),
        );
        break;

      default:
        throw new Error(`Unknown tool: ${toolName}`);
    }

    return {
      content: [
        {
          type: "text",
          text: JSON.stringify(result, null, 2),
        },
      ],
    };
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : String(error);
    return {
      content: [
        {
          type: "text",
          text: `Error: ${errorMessage}`,
          isError: true,
        },
      ],
    };
  }
});

// Start server
async function main() {
  const transport = new StdioServerTransport();
  await server.connect(transport);
  console.error("Ticketing MCP server started");
}

main().catch((error) => {
  console.error("Failed to start server:", error);
  process.exit(1);
});
