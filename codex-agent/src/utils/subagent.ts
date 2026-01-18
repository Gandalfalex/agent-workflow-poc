import * as fs from "fs";
import * as path from "path";
import * as child_process from "child_process";
import * as util from "util";

const exec = util.promisify(child_process.exec);

export interface SubagentOptions {
  workspacePath: string;
  prompt: string;
  timeout?: number; // milliseconds, default 30 minutes
}

export interface SubagentResult {
  success: boolean;
  output: string;
  summary?: string;
  filesChanged?: string[];
  testsRun?: boolean;
  testsPassed?: boolean;
  commitSha?: string;
  nextSteps?: string[];
  error?: string;
}

/**
 * Spawn a subagent using the existing Claude CLI to implement a feature
 * No API key required - uses the Claude instance already running on the server
 */
export async function spawnSubagent(
  options: SubagentOptions,
): Promise<SubagentResult> {
  const {
    workspacePath,
    prompt,
    timeout = 30 * 60 * 1000, // 30 minutes
  } = options;

  // Validate workspace exists
  if (!fs.existsSync(workspacePath)) {
    return {
      success: false,
      output: "",
      error: `Workspace path does not exist: ${workspacePath}`,
    };
  }

  try {
    // Build system message with workspace context
    const systemMessage = buildSystemMessage(workspacePath);

    // Combine system message and prompt
    const fullPrompt = `${systemMessage}\n\n---\n\n${prompt}`;

    // Create temporary file for prompt (to avoid shell escaping issues)
    const promptFile = path.join(workspacePath, ".claude-prompt.txt");
    fs.writeFileSync(promptFile, fullPrompt);

    try {
      console.error(
        `[Subagent] Starting implementation task in ${workspacePath}`,
      );

      // Call Claude CLI directly - no API key needed
      // The claude command uses the existing authenticated session
      const { stdout, stderr } = await exec(
        `cd "${workspacePath}" && claude --read-file .claude-prompt.txt --output-format text`,
        {
          timeout: timeout,
          maxBuffer: 10 * 1024 * 1024, // 10MB buffer for large outputs
        },
      );

      console.error("[Subagent] Response received from Claude");

      const output = stdout || stderr;

      // Try to parse JSON result from the response
      const result = parseSubagentOutput(output);

      return {
        success: result.success || false,
        output: output,
        summary: result.summary,
        filesChanged: result.filesChanged,
        testsRun: result.testsRun,
        testsPassed: result.testsPassed,
        commitSha: result.commitSha,
        nextSteps: result.nextSteps,
      };
    } finally {
      // Clean up prompt file
      try {
        fs.unlinkSync(promptFile);
      } catch (e) {
        // Ignore cleanup errors
      }
    }
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : String(error);
    console.error(`[Subagent] Error: ${errorMessage}`);

    return {
      success: false,
      output: "",
      error: `Subagent failed: ${errorMessage}`,
    };
  }
}

/**
 * Build system message with workspace context and instructions
 */
function buildSystemMessage(workspacePath: string): string {
  let contextFiles: string[] = [];

  try {
    // List files in workspace to understand structure
    const files = fs.readdirSync(workspacePath);
    contextFiles = files.slice(0, 10); // First 10 files
  } catch (error) {
    // Ignore if we can't read directory
  }

  return `You are an expert software engineer implementing features in a code repository.

You are working in the following directory: ${workspacePath}

Your task is to:
1. Understand the feature request
2. Examine the existing code structure and patterns
3. Implement the feature according to the specifications
4. Write appropriate tests
5. Ensure all tests pass
6. Commit your changes with a clear commit message

Important guidelines:
- Follow the existing code style and conventions
- Write clean, maintainable code
- Add tests for new functionality
- Make small, focused commits
- Use git to track changes: git add, git commit
- Do NOT push to remote - only commit locally
- Do NOT modify the git remote configuration

When you are finished with the implementation, respond with a JSON block like this:
\`\`\`json
{
  "success": true,
  "summary": "Brief description of what was implemented",
  "filesChanged": ["file1.ts", "file2.ts", "test.ts"],
  "testsRun": true,
  "testsPassed": true,
  "commitSha": "abc123def456",
  "nextSteps": []
}
\`\`\`

If implementation fails, include "success": false with an error summary.

Current workspace contents (first 10 items):
${contextFiles.join("\n")}

Begin implementation now.`;
}

/**
 * Parse the subagent output to extract JSON result
 */
function parseSubagentOutput(output: string): Partial<SubagentResult> {
  try {
    // Look for JSON block in the output
    const jsonMatch = output.match(/```json\n([\s\S]*?)\n```/);
    if (jsonMatch && jsonMatch[1]) {
      const parsed = JSON.parse(jsonMatch[1]);
      return parsed;
    }

    // Try raw JSON parsing in case there's no code block
    const directMatch = output.match(/\{[\s\S]*"success"[\s\S]*\}/);
    if (directMatch) {
      const parsed = JSON.parse(directMatch[0]);
      return parsed;
    }

    // If no JSON found, check for success indicators in text
    if (
      output.toLowerCase().includes("success") ||
      output.toLowerCase().includes("complete")
    ) {
      return {
        success: true,
        summary: "Implementation completed (no JSON summary provided)",
      };
    }

    return {
      success: false,
      summary: "Unable to parse subagent output",
    };
  } catch (error) {
    console.error(
      `[Subagent] Failed to parse JSON output: ${error instanceof Error ? error.message : String(error)}`,
    );
    return {
      success: false,
      summary: "Output parsing failed",
    };
  }
}
