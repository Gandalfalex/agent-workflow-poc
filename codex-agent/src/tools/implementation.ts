import { TicketingAPIClient, Ticket, WorkflowState } from "../api-client.js";
import { z } from "zod";
import { spawnSubagent, SubagentResult } from "../utils/subagent.js";
import { renderTemplate } from "../utils/template.js";
import * as fs from "fs";
import * as path from "path";
import * as child_process from "child_process";
import * as util from "util";

const exec = util.promisify(child_process.exec);

/**
 * Parse timeout value - supports both milliseconds (number/string) and duration strings (30m, 1h, etc.)
 */
function parseTimeout(value: string): number {
  // If it's a pure number (milliseconds), use it directly
  const asNumber = parseInt(value, 10);
  if (!isNaN(asNumber) && value === asNumber.toString()) {
    return asNumber;
  }

  // Parse duration string (e.g., "30m", "1h", "90s")
  const match = value.match(/^(\d+(?:\.\d+)?)(ms|s|m|h)$/);
  if (!match) {
    console.error(
      `[Implement] Invalid timeout format: ${value}, using default 30m`,
    );
    return 30 * 60 * 1000; // 30 minutes default
  }

  const amount = parseFloat(match[1]);
  const unit = match[2];

  switch (unit) {
    case "ms":
      return amount;
    case "s":
      return amount * 1000;
    case "m":
      return amount * 60 * 1000;
    case "h":
      return amount * 60 * 60 * 1000;
    default:
      return 30 * 60 * 1000;
  }
}

export const implementTicketSchema = z.object({
  ticketId: z.string().describe("Ticket UUID or key (e.g., 'PROJ-001')"),
  workspaceRoot: z
    .string()
    .optional()
    .describe("Root directory for worktrees (default: ~/worktrees)"),
  autoCommit: z
    .boolean()
    .optional()
    .default(true)
    .describe("Auto-commit changes"),
  autoPush: z
    .boolean()
    .optional()
    .default(false)
    .describe("Auto-push to remote"),
  repoPath: z
    .string()
    .optional()
    .describe("Path to repository for worktree creation"),
});

export interface ImplementationOutput {
  success: boolean;
  ticketKey: string;
  workspacePath: string;
  branch: string;
  summary: string;
  filesChanged?: string[];
  testsRun?: boolean;
  testsPassed?: boolean;
  commitSha?: string;
  nextSteps?: string[];
  error?: string;
}

/**
 * Main implementation tool - orchestrates feature implementation
 */
export async function implementTicket(
  client: TicketingAPIClient,
  params: z.infer<typeof implementTicketSchema>,
): Promise<ImplementationOutput> {
  console.error(
    `[Implement] Starting implementation for ticket: ${params.ticketId}`,
  );

  try {
    // Step 1: Resolve and fetch ticket
    console.error("[Implement] Step 1: Fetching ticket details...");
    const ticket = await resolveTicket(client, params.ticketId);
    console.error(`[Implement] Resolved ticket: ${ticket.key}`);

    // Step 2: Validate ticket type
    if (ticket.type !== "feature") {
      return {
        success: false,
        ticketKey: ticket.key,
        workspacePath: "",
        branch: "",
        summary: `Ticket is a ${ticket.type}, not a feature. Use different workflow for bugs.`,
        error: "Invalid ticket type",
      };
    }

    // Step 3: Fetch comments and full context
    console.error("[Implement] Step 2: Fetching ticket comments...");
    const comments = await client.getComments(ticket.id);

    // Step 4: Create worktree
    console.error("[Implement] Step 3: Creating git worktree...");
    const worktreeRoot =
      params.workspaceRoot ||
      path.join(process.env.HOME || "/tmp", "worktrees");
    const repoPath = params.repoPath || process.env.REPO_PATH || ".";

    const worktreeInfo = await createWorktree(
      ticket.key,
      repoPath,
      worktreeRoot,
    );

    if (!worktreeInfo.success) {
      return {
        success: false,
        ticketKey: ticket.key,
        workspacePath: "",
        branch: "",
        summary: "Failed to create worktree",
        error: worktreeInfo.error,
      };
    }

    const { workspacePath, branch } = worktreeInfo;
    console.error(
      `[Implement] Created worktree at: ${workspacePath} (branch: ${branch})`,
    );

    // Step 5: Generate prompt
    console.error("[Implement] Step 4: Generating subagent prompt...");
    const prompt = generatePrompt(ticket, comments, {
      workspacePath,
      branch,
      repoRoot: worktreeInfo.repoRoot,
    });

    // Step 6: Spawn subagent
    console.error("[Implement] Step 5: Spawning subagent...");
    const subagentTimeout = parseTimeout(process.env.SUBAGENT_TIMEOUT || "30m"); // 30 min default
    const subagentResult = await spawnSubagent({
      workspacePath,
      prompt,
      timeout: subagentTimeout,
    });

    // Step 7: Update ticket and add comment
    console.error("[Implement] Step 6: Updating ticket state...");
    const updateSuccess = await updateTicketAfterImplementation(
      client,
      ticket,
      subagentResult,
      { workspacePath, branch },
    );

    if (!updateSuccess) {
      console.error(
        "[Implement] Warning: Failed to update ticket state/comment",
      );
    }

    // Step 8: Return result
    const output: ImplementationOutput = {
      success: subagentResult.success,
      ticketKey: ticket.key,
      workspacePath,
      branch,
      summary: subagentResult.summary || "Implementation completed",
      filesChanged: subagentResult.filesChanged,
      testsRun: subagentResult.testsRun,
      testsPassed: subagentResult.testsPassed,
      commitSha: subagentResult.commitSha,
      nextSteps: subagentResult.nextSteps,
      error: subagentResult.error,
    };

    console.error(
      `[Implement] Completed: ${output.success ? "SUCCESS" : "FAILED"}`,
    );
    return output;
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : String(error);
    console.error(`[Implement] Error: ${errorMessage}`);

    return {
      success: false,
      ticketKey: params.ticketId,
      workspacePath: "",
      branch: "",
      summary: "Implementation failed",
      error: errorMessage,
    };
  }
}

/**
 * Resolve ticket by UUID or key
 */
async function resolveTicket(
  client: TicketingAPIClient,
  ticketId: string,
): Promise<Ticket> {
  // Try as UUID first
  try {
    // Check if it looks like a UUID
    if (
      ticketId.match(
        /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i,
      )
    ) {
      return await client.getTicket(ticketId);
    }
  } catch (error) {
    console.error(`[Implement] Not a valid UUID: ${ticketId}`);
  }

  // Try as ticket key - fetch all projects and search
  const projects = await client.listProjects();

  for (const project of projects) {
    try {
      // Search in each project for matching key
      const result = await client.listTickets(project.id, { query: ticketId });
      const match = result.items.find((t) => t.key === ticketId);
      if (match) {
        return match;
      }
    } catch (error) {
      // Continue searching other projects
    }
  }

  throw new Error(`Ticket not found: ${ticketId} (tried UUID and key formats)`);
}

/**
 * Create git worktree for the ticket
 */
async function createWorktree(
  ticketKey: string,
  repoPath: string,
  worktreeRoot: string,
): Promise<{
  success: boolean;
  workspacePath: string;
  branch: string;
  repoRoot: string;
  error?: string;
}> {
  try {
    // Ensure repo path exists
    if (!fs.existsSync(repoPath)) {
      return {
        success: false,
        workspacePath: "",
        branch: "",
        repoRoot: "",
        error: `Repository path does not exist: ${repoPath}`,
      };
    }

    // Get the scripts directory
    const scriptDir = path.join(
      import.meta.url.replace("file://", ""),
      "..",
      "..",
      "scripts",
    );
    const scriptPath = path.join(scriptDir, "create-worktree.sh");

    if (!fs.existsSync(scriptPath)) {
      return {
        success: false,
        workspacePath: "",
        branch: "",
        repoRoot: "",
        error: `Script not found: ${scriptPath}`,
      };
    }

    // Run the worktree creation script
    const { stdout, stderr } = await exec(
      `bash "${scriptPath}" "${ticketKey}" "${repoPath}" "${worktreeRoot}"`,
      { maxBuffer: 10 * 1024 * 1024 },
    );

    if (stderr) {
      console.error(`[Worktree] ${stderr}`);
    }

    // Parse JSON output
    const result = JSON.parse(stdout);

    if (!result.success) {
      return {
        success: false,
        workspacePath: "",
        branch: "",
        repoRoot: "",
        error: "Script returned success: false",
      };
    }

    return {
      success: true,
      workspacePath: result.worktreePath,
      branch: result.branch,
      repoRoot: result.repoRoot,
    };
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : String(error);
    return {
      success: false,
      workspacePath: "",
      branch: "",
      repoRoot: "",
      error: `Failed to create worktree: ${errorMessage}`,
    };
  }
}

/**
 * Generate prompt for subagent
 */
function generatePrompt(
  ticket: Ticket,
  comments: any[],
  context: { workspacePath: string; branch: string; repoRoot: string },
): string {
  const commentsList = comments
    .map(
      (c) =>
        `**${c.authorName}** (${new Date(c.createdAt).toLocaleString()}):\n> ${c.message}`,
    )
    .join("\n\n");

  const template = fs.readFileSync(
    path.join(
      import.meta.url.replace("file://", ""),
      "..",
      "..",
      "prompts",
      "implement-feature.md",
    ),
    "utf-8",
  );

  return renderTemplate(template, {
    ticketKey: ticket.key,
    title: ticket.title,
    type: ticket.type,
    priority: ticket.priority,
    status: ticket.state.name,
    description: ticket.description || "No description provided",
    story: ticket.story
      ? {
          title: ticket.story.title,
          description: ticket.story.description,
        }
      : null,
    comments: commentsList || "No comments",
    workspacePath: context.workspacePath,
    branch: context.branch,
    repoRoot: context.repoRoot,
  });
}

/**
 * Update ticket state and add completion comment
 */
async function updateTicketAfterImplementation(
  client: TicketingAPIClient,
  ticket: Ticket,
  result: SubagentResult,
  context: { workspacePath: string; branch: string },
): Promise<boolean> {
  try {
    // Get available states to find "In Review" state
    const states = await client.getWorkflow(ticket.projectId);
    const reviewState = states.find(
      (s) =>
        s.name.toLowerCase() === "in review" ||
        s.name.toLowerCase() === "review" ||
        s.name.toLowerCase() === "in_review",
    );

    if (reviewState && process.env.AUTO_UPDATE_STATE !== "false") {
      // Update ticket state to review
      await client.updateTicket(ticket.id, { stateId: reviewState.id });
      console.error(`[Implement] Updated ticket state to: ${reviewState.name}`);
    }

    // Add comment with implementation summary
    const comment = buildCompletionComment(result, context);
    await client.addComment(ticket.id, comment);
    console.error("[Implement] Added completion comment to ticket");

    return true;
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : String(error);
    console.error(`[Implement] Failed to update ticket: ${errorMessage}`);
    return false;
  }
}

/**
 * Build completion comment for ticket
 */
function buildCompletionComment(
  result: SubagentResult,
  context: { workspacePath: string; branch: string },
): string {
  const status = result.success
    ? "✅ Implementation Complete"
    : "⚠️ Implementation Incomplete";

  let filesSection = "";
  if (result.filesChanged && result.filesChanged.length > 0) {
    filesSection = `\n**Files Changed:**\n${result.filesChanged.map((f) => `- ${f}`).join("\n")}`;
  }

  let testsSection = "";
  if (result.testsRun) {
    testsSection = `\n**Tests:** ${result.testsPassed ? "✅ Passed" : "⚠️ Failed - Review Required"}`;
  }

  let commitSection = "";
  if (result.commitSha) {
    commitSection = `\n**Commit:** \`${result.commitSha.substring(0, 7)}\``;
  }

  let nextStepsSection = "";
  if (result.nextSteps && result.nextSteps.length > 0) {
    nextStepsSection = `\n**Next Steps:**\n${result.nextSteps.map((s) => `- ${s}`).join("\n")}`;
  }

  return `${status}

**Branch:** \`${context.branch}\`
**Workspace:** \`${context.workspacePath}\`

${result.summary}${filesSection}${testsSection}${commitSection}${nextStepsSection}

---
*Automatically generated by feature implementation agent*`;
}
