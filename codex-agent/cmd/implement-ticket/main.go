package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Ticket struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Priority    string `json:"priority"`
	StateID     string `json:"stateId"`
}

type ImplementationResult struct {
	Success       bool     `json:"success"`
	TicketKey     string   `json:"ticketKey"`
	WorkspacePath string   `json:"workspacePath"`
	Branch        string   `json:"branch"`
	Summary       string   `json:"summary"`
	FilesChanged  []string `json:"filesChanged,omitempty"`
	TestsRun      bool     `json:"testsRun,omitempty"`
	TestsPassed   bool     `json:"testsPassed,omitempty"`
	CommitSha     string   `json:"commitSha,omitempty"`
	NextSteps     []string `json:"nextSteps,omitempty"`
	Error         string   `json:"error,omitempty"`
}

func main() {
	ticketID := flag.String("ticket", "", "Ticket UUID or key (e.g., PROJ-001)")
	repoPath := flag.String("repo", os.Getenv("REPO_PATH"), "Path to repository")
	workspaceRoot := flag.String("workspace", os.Getenv("WORKSPACE_ROOT"), "Root directory for worktrees")
	flag.Parse()

	if *ticketID == "" {
		fmt.Fprintf(os.Stderr, "Usage: implement-ticket --ticket <UUID|KEY> [--repo <PATH>] [--workspace <PATH>]\n")
		os.Exit(1)
	}

	if *repoPath == "" {
		*repoPath = "."
	}

	if *workspaceRoot == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			*workspaceRoot = "/tmp/worktrees"
		} else {
			*workspaceRoot = filepath.Join(home, "worktrees")
		}
	}

	result := implementTicket(*ticketID, *repoPath, *workspaceRoot)

	// Output result as JSON
	jsonResult, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(jsonResult))

	if !result.Success {
		os.Exit(1)
	}
}

func implementTicket(ticketID, repoPath, workspaceRoot string) ImplementationResult {
	fmt.Fprintf(os.Stderr, "[Implement] Starting implementation for ticket: %s\n", ticketID)

	// Step 1: Fetch ticket details from API
	fmt.Fprintf(os.Stderr, "[Implement] Step 1: Fetching ticket details...\n")
	ticket, err := getTicket(ticketID)
	if err != nil {
		return ImplementationResult{
			Success:   false,
			TicketKey: ticketID,
			Summary:   "Implementation failed",
			Error:     fmt.Sprintf("Failed to fetch ticket: %v", err),
		}
	}

	fmt.Fprintf(os.Stderr, "[Implement] Resolved ticket: %s\n", ticket.Key)

	// Step 2: Validate ticket type
	if ticket.Type != "feature" {
		return ImplementationResult{
			Success:   false,
			TicketKey: ticket.Key,
			Summary:   fmt.Sprintf("Ticket is a %s, not a feature", ticket.Type),
			Error:     "Invalid ticket type",
		}
	}

	// Step 3: Create worktree
	fmt.Fprintf(os.Stderr, "[Implement] Step 2: Creating git worktree...\n")
	worktreeInfo, err := createWorktree(ticket.Key, repoPath, workspaceRoot)
	if err != nil {
		return ImplementationResult{
			Success:   false,
			TicketKey: ticket.Key,
			Summary:   "Failed to create worktree",
			Error:     err.Error(),
		}
	}

	fmt.Fprintf(os.Stderr, "[Implement] Worktree created at: %s\n", worktreeInfo.Path)

	// Step 4: Spawn subagent to implement feature
	fmt.Fprintf(os.Stderr, "[Implement] Step 3: Spawning Claude subagent...\n")
	result, err := spawnSubagent(ticket, worktreeInfo)
	if err != nil {
		return ImplementationResult{
			Success:       false,
			TicketKey:     ticket.Key,
			WorkspacePath: worktreeInfo.Path,
			Branch:        worktreeInfo.Branch,
			Summary:       "Subagent failed",
			Error:         err.Error(),
		}
	}

	// Step 5: Update ticket state
	fmt.Fprintf(os.Stderr, "[Implement] Step 4: Updating ticket state...\n")
	if err := updateTicketState(ticket.ID, "In Review"); err != nil {
		fmt.Fprintf(os.Stderr, "[Implement] Warning: Could not update ticket state: %v\n", err)
	}

	// Step 6: Add comment with summary
	fmt.Fprintf(os.Stderr, "[Implement] Step 5: Adding implementation summary comment...\n")
	comment := formatImplementationComment(result)
	if err := addTicketComment(ticket.ID, comment); err != nil {
		fmt.Fprintf(os.Stderr, "[Implement] Warning: Could not add comment: %v\n", err)
	}

	fmt.Fprintf(os.Stderr, "[Implement] Completed: SUCCESS\n")

	return ImplementationResult{
		Success:       true,
		TicketKey:     ticket.Key,
		WorkspacePath: worktreeInfo.Path,
		Branch:        worktreeInfo.Branch,
		Summary:       result.Summary,
		FilesChanged:  result.FilesChanged,
		TestsRun:      result.TestsRun,
		TestsPassed:   result.TestsPassed,
		CommitSha:     result.CommitSha,
		NextSteps:     result.NextSteps,
	}
}

type WorktreeInfo struct {
	Path   string
	Branch string
}

func createWorktree(ticketKey, repoPath, workspaceRoot string) (WorktreeInfo, error) {
	// Ensure workspace root exists
	if err := os.MkdirAll(workspaceRoot, 0755); err != nil {
		return WorktreeInfo{}, fmt.Errorf("failed to create workspace root: %w", err)
	}

	branch := fmt.Sprintf("feature/%s", ticketKey)
	worktreePath := filepath.Join(workspaceRoot, ticketKey)

	// Create worktree
	cmd := exec.Command("git", "worktree", "add", worktreePath, "-b", branch)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return WorktreeInfo{}, fmt.Errorf("failed to create worktree: %s", string(output))
	}

	return WorktreeInfo{
		Path:   worktreePath,
		Branch: branch,
	}, nil
}

type SubagentResult struct {
	Summary      string
	FilesChanged []string
	TestsRun     bool
	TestsPassed  bool
	CommitSha    string
	NextSteps    []string
}

func spawnSubagent(ticket *Ticket, worktreeInfo WorktreeInfo) (SubagentResult, error) {
	// Create prompt for Claude
	prompt := fmt.Sprintf(`Implement the following feature in this workspace:

Ticket: %s
Title: %s
Description: %s
Priority: %s

Guidelines:
1. Read existing code to understand patterns
2. Implement the feature following project conventions
3. Add appropriate tests
4. Run tests to verify functionality
5. Commit changes with message: "%s: %s"

When complete, output a JSON summary with:
{
  "success": true/false,
  "summary": "What was implemented",
  "filesChanged": ["list", "of", "files"],
  "testsRun": true/false,
  "testsPassed": true/false,
  "commitSha": "abc123",
  "nextSteps": ["any manual steps needed"]
}
`, ticket.Key, ticket.Title, ticket.Description, ticket.Priority, ticket.Key, ticket.Title)

	// Spawn Claude subagent
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "claude", "--working-directory", worktreeInfo.Path, "--execute", prompt)
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return SubagentResult{}, fmt.Errorf("subagent failed: %w", err)
	}

	// Parse result JSON from output
	var result SubagentResult
	if err := json.Unmarshal(output, &result); err != nil {
		// If no JSON found, extract from output
		result.Summary = string(output)
	}

	return result, nil
}

func getTicket(ticketID string) (*Ticket, error) {
	apiURL := os.Getenv("TICKETING_API_BASE_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}

	url := fmt.Sprintf("%s/api/tickets/%s", apiURL, ticketID)

	req, _ := http.NewRequest("GET", url, nil)
	// Try to use session cookie if available
	if cookie := os.Getenv("TICKETING_SESSION"); cookie != "" {
		req.AddCookie(&http.Cookie{
			Name:  "ticketing_session",
			Value: cookie,
		})
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var ticket Ticket
	if err := json.NewDecoder(resp.Body).Decode(&ticket); err != nil {
		return nil, err
	}

	return &ticket, nil
}

func updateTicketState(ticketID, stateName string) error {
	apiURL := os.Getenv("TICKETING_API_BASE_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}

	url := fmt.Sprintf("%s/api/tickets/%s/state", apiURL, ticketID)

	payload := map[string]string{"stateName": stateName}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("PATCH", url, strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")

	if cookie := os.Getenv("TICKETING_SESSION"); cookie != "" {
		req.AddCookie(&http.Cookie{
			Name:  "ticketing_session",
			Value: cookie,
		})
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

func addTicketComment(ticketID, comment string) error {
	apiURL := os.Getenv("TICKETING_API_BASE_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}

	url := fmt.Sprintf("%s/api/tickets/%s/comments", apiURL, ticketID)

	payload := map[string]string{"message": comment}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", url, strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")

	if cookie := os.Getenv("TICKETING_SESSION"); cookie != "" {
		req.AddCookie(&http.Cookie{
			Name:  "ticketing_session",
			Value: cookie,
		})
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

func formatImplementationComment(result SubagentResult) string {
	return fmt.Sprintf(`## Implementation Complete ✅

**Summary:** %s

**Files Changed:** %s

**Tests:** %v (Passed: %v)

**Commit:** %s

**Next Steps:**
%s
`,
		result.Summary,
		strings.Join(result.FilesChanged, ", "),
		result.TestsRun,
		result.TestsPassed,
		result.CommitSha,
		strings.Join(result.NextSteps, "\n"),
	)
}
