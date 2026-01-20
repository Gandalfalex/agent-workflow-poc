package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// --- Data Types ---

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type WorkflowState struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	Name      string `json:"name"`
	Order     int    `json:"order"`
	IsDefault bool   `json:"isDefault"`
	IsClosed  bool   `json:"isClosed"`
}

type Story struct {
	ID          string `json:"id"`
	ProjectID   string `json:"projectId"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type Ticket struct {
	ID          string        `json:"id"`
	Key         string        `json:"key"`
	Number      int           `json:"number"`
	Type        string        `json:"type"`
	ProjectID   string        `json:"projectId"`
	ProjectKey  string        `json:"projectKey"`
	StoryID     *string       `json:"storyId"`
	Story       *Story        `json:"story,omitempty"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	StateID     string        `json:"stateId"`
	State       WorkflowState `json:"state"`
	Priority    string        `json:"priority"`
	CreatedAt   string        `json:"createdAt"`
	UpdatedAt   string        `json:"updatedAt"`
}

type TicketComment struct {
	ID         string `json:"id"`
	TicketID   string `json:"ticketId"`
	AuthorID   string `json:"authorId"`
	AuthorName string `json:"authorName"`
	Message    string `json:"message"`
	CreatedAt  string `json:"createdAt"`
}

type Project struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
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

// --- Keycloak Auth ---

type KeycloakAuth struct {
	baseURL      string
	realm        string
	clientID     string
	username     string
	password     string
	accessToken  string
	refreshToken string
	expiresAt    time.Time
	httpClient   *http.Client
}

func NewKeycloakAuth() (*KeycloakAuth, error) {
	baseURL := os.Getenv("KEYCLOAK_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8081"
	}

	realm := os.Getenv("KEYCLOAK_REALM")
	if realm == "" {
		realm = "ticketing"
	}

	clientID := os.Getenv("KEYCLOAK_CLIENT_ID")
	if clientID == "" {
		clientID = "myclient"
	}

	username := os.Getenv("KEYCLOAK_USERNAME")
	password := os.Getenv("KEYCLOAK_PASSWORD")

	if username == "" || password == "" {
		return nil, fmt.Errorf("missing KEYCLOAK_USERNAME and/or KEYCLOAK_PASSWORD environment variables")
	}

	return &KeycloakAuth{
		baseURL:    baseURL,
		realm:      realm,
		clientID:   clientID,
		username:   username,
		password:   password,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (k *KeycloakAuth) GetAccessToken() (string, error) {
	// Return cached token if still valid (with 30s buffer)
	if k.accessToken != "" && time.Now().Add(30*time.Second).Before(k.expiresAt) {
		return k.accessToken, nil
	}

	// Try to refresh if we have a refresh token
	if k.refreshToken != "" {
		token, err := k.refreshAccessToken()
		if err == nil {
			return token, nil
		}
		// Fall through to password grant on refresh failure
	}

	// Get new token using password grant
	return k.getNewToken()
}

func (k *KeycloakAuth) getNewToken() (string, error) {
	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", k.baseURL, k.realm)

	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("client_id", k.clientID)
	data.Set("username", k.username)
	data.Set("password", k.password)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to authenticate with Keycloak: %s", string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	k.accessToken = tokenResp.AccessToken
	k.refreshToken = tokenResp.RefreshToken
	k.expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return k.accessToken, nil
}

func (k *KeycloakAuth) refreshAccessToken() (string, error) {
	if k.refreshToken == "" {
		return "", fmt.Errorf("no refresh token available")
	}

	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", k.baseURL, k.realm)

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", k.clientID)
	data.Set("refresh_token", k.refreshToken)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		k.accessToken = ""
		k.refreshToken = ""
		return "", fmt.Errorf("token refresh failed")
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	k.accessToken = tokenResp.AccessToken
	if tokenResp.RefreshToken != "" {
		k.refreshToken = tokenResp.RefreshToken
	}
	k.expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return k.accessToken, nil
}

// --- API Client ---

type APIClient struct {
	baseURL    string
	auth       *KeycloakAuth
	httpClient *http.Client
}

func NewAPIClient(auth *KeycloakAuth) *APIClient {
	baseURL := os.Getenv("TICKETING_API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	return &APIClient{
		baseURL:    baseURL,
		auth:       auth,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *APIClient) request(method, path string, body interface{}) ([]byte, error) {
	token, err := c.auth.GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (c *APIClient) GetTicket(id string) (*Ticket, error) {
	body, err := c.request("GET", "/tickets/"+id, nil)
	if err != nil {
		return nil, err
	}

	var ticket Ticket
	if err := json.Unmarshal(body, &ticket); err != nil {
		return nil, err
	}

	return &ticket, nil
}

func (c *APIClient) ListProjects() ([]Project, error) {
	body, err := c.request("GET", "/projects", nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Items []Project `json:"items"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return response.Items, nil
}

func (c *APIClient) ListTickets(projectID string, query string) ([]Ticket, error) {
	path := "/projects/" + projectID + "/tickets"
	if query != "" {
		path += "?q=" + url.QueryEscape(query)
	}

	body, err := c.request("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Items []Ticket `json:"items"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return response.Items, nil
}

func (c *APIClient) GetComments(ticketID string) ([]TicketComment, error) {
	body, err := c.request("GET", "/tickets/"+ticketID+"/comments", nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Items []TicketComment `json:"items"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return response.Items, nil
}

func (c *APIClient) GetWorkflow(projectID string) ([]WorkflowState, error) {
	body, err := c.request("GET", "/projects/"+projectID+"/workflow", nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		States []WorkflowState `json:"states"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return response.States, nil
}

func (c *APIClient) UpdateTicket(ticketID string, update map[string]interface{}) error {
	_, err := c.request("PATCH", "/tickets/"+ticketID, update)
	return err
}

func (c *APIClient) AddComment(ticketID, message string) error {
	_, err := c.request("POST", "/tickets/"+ticketID+"/comments", map[string]string{
		"message": message,
	})
	return err
}

// --- Ticket Resolution ---

var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func (c *APIClient) ResolveTicket(ticketID string) (*Ticket, error) {
	// Try as UUID first
	if uuidRegex.MatchString(ticketID) {
		ticket, err := c.GetTicket(ticketID)
		if err == nil {
			return ticket, nil
		}
		fmt.Fprintf(os.Stderr, "[Implement] Not a valid UUID or ticket not found: %s\n", ticketID)
	}

	// Try as ticket key - search all projects
	projects, err := c.ListProjects()
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	for _, project := range projects {
		tickets, err := c.ListTickets(project.ID, ticketID)
		if err != nil {
			continue
		}

		for _, ticket := range tickets {
			if ticket.Key == ticketID {
				return &ticket, nil
			}
		}
	}

	return nil, fmt.Errorf("ticket not found: %s (tried UUID and key formats)", ticketID)
}

// --- Worktree Management ---

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

	// Check if worktree already exists
	if _, err := os.Stat(worktreePath); err == nil {
		fmt.Fprintf(os.Stderr, "[Implement] Worktree already exists at: %s\n", worktreePath)
		return WorktreeInfo{
			Path:   worktreePath,
			Branch: branch,
		}, nil
	}

	// Check if branch exists
	checkBranch := exec.Command("git", "rev-parse", "--verify", branch)
	checkBranch.Dir = repoPath
	branchExists := checkBranch.Run() == nil

	var cmd *exec.Cmd
	if branchExists {
		// Use existing branch
		cmd = exec.Command("git", "worktree", "add", worktreePath, branch)
	} else {
		// Create new branch
		cmd = exec.Command("git", "worktree", "add", worktreePath, "-b", branch)
	}
	cmd.Dir = repoPath

	if output, err := cmd.CombinedOutput(); err != nil {
		return WorktreeInfo{}, fmt.Errorf("failed to create worktree: %s", string(output))
	}

	return WorktreeInfo{
		Path:   worktreePath,
		Branch: branch,
	}, nil
}

// --- Subagent ---

type SubagentResult struct {
	Success      bool     `json:"success"`
	Summary      string   `json:"summary"`
	FilesChanged []string `json:"filesChanged"`
	TestsRun     bool     `json:"testsRun"`
	TestsPassed  bool     `json:"testsPassed"`
	CommitSha    string   `json:"commitSha"`
	NextSteps    []string `json:"nextSteps"`
	Error        string   `json:"error"`
}

func generatePrompt(ticket *Ticket, comments []TicketComment, worktreeInfo WorktreeInfo) string {
	// Format comments
	var commentsList string
	if len(comments) > 0 {
		var parts []string
		for _, c := range comments {
			parts = append(parts, fmt.Sprintf("**%s** (%s):\n> %s", c.AuthorName, c.CreatedAt, c.Message))
		}
		commentsList = strings.Join(parts, "\n\n")
	} else {
		commentsList = "No comments"
	}

	// Format story context if available
	var storyContext string
	if ticket.Story != nil {
		storyContext = fmt.Sprintf(`
## Parent Story
**Title:** %s
**Description:** %s
`, ticket.Story.Title, ticket.Story.Description)
	}

	return fmt.Sprintf(`# Feature Implementation Task

## Ticket Details
- **Key:** %s
- **Title:** %s
- **Type:** %s
- **Priority:** %s
- **Status:** %s

## Description
%s
%s
## Discussion/Comments
%s

## Workspace
- **Path:** %s
- **Branch:** %s

## Instructions

1. **Understand the Context**: Read existing code to understand patterns and conventions used in this project.

2. **Implement the Feature**:
   - Follow the project's coding style and conventions
   - Keep changes focused and minimal
   - Add appropriate error handling

3. **Write Tests**:
   - Add unit tests for new functionality
   - Ensure existing tests still pass

4. **Verify**:
   - Run the test suite
   - Check for linting issues if applicable

5. **Commit**:
   - Commit your changes with message: "%s: %s"
   - Include only relevant files

## Output Format

When complete, output a JSON object with:
{
  "success": true,
  "summary": "Brief description of what was implemented",
  "filesChanged": ["path/to/file1.ts", "path/to/file2.ts"],
  "testsRun": true,
  "testsPassed": true,
  "commitSha": "abc1234",
  "nextSteps": ["Any manual steps or follow-up items"]
}

If implementation fails:
{
  "success": false,
  "summary": "What was attempted",
  "error": "Description of what went wrong",
  "nextSteps": ["Suggested remediation steps"]
}
`,
		ticket.Key,
		ticket.Title,
		ticket.Type,
		ticket.Priority,
		ticket.State.Name,
		ticket.Description,
		storyContext,
		commentsList,
		worktreeInfo.Path,
		worktreeInfo.Branch,
		ticket.Key,
		ticket.Title,
	)
}

func spawnSubagent(ticket *Ticket, comments []TicketComment, worktreeInfo WorktreeInfo) (SubagentResult, error) {
	prompt := generatePrompt(ticket, comments, worktreeInfo)

	// Spawn Claude subagent
	timeout := 30 * time.Minute
	if envTimeout := os.Getenv("SUBAGENT_TIMEOUT"); envTimeout != "" {
		if parsed, err := time.ParseDuration(envTimeout); err == nil {
			timeout = parsed
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "claude", "--dangerously-skip-permissions", "--output-format", "json", "--print", prompt)
	cmd.Dir = worktreeInfo.Path
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return SubagentResult{}, fmt.Errorf("subagent failed: %w", err)
	}

	// Try to extract JSON from the output
	result := parseSubagentOutput(output)
	return result, nil
}

func parseSubagentOutput(output []byte) SubagentResult {
	// First try to parse the Claude JSON output format
	var claudeOutput struct {
		Result string `json:"result"`
	}
	if err := json.Unmarshal(output, &claudeOutput); err == nil && claudeOutput.Result != "" {
		// Try to find JSON in the result text
		if extracted := extractJSON([]byte(claudeOutput.Result)); extracted != nil {
			var result SubagentResult
			if err := json.Unmarshal(extracted, &result); err == nil {
				return result
			}
		}
		// If no JSON found, use the result as summary
		return SubagentResult{
			Success: true,
			Summary: claudeOutput.Result,
		}
	}

	// Try to find JSON object in raw output
	if extracted := extractJSON(output); extracted != nil {
		var result SubagentResult
		if err := json.Unmarshal(extracted, &result); err == nil {
			return result
		}
	}

	// Fall back to using raw output as summary
	return SubagentResult{
		Success: true,
		Summary: string(output),
	}
}

func extractJSON(data []byte) []byte {
	// Find the first { and last } to extract JSON object
	start := bytes.IndexByte(data, '{')
	end := bytes.LastIndexByte(data, '}')

	if start == -1 || end == -1 || end <= start {
		return nil
	}

	return data[start : end+1]
}

// --- Comment Formatting ---

func formatImplementationComment(result SubagentResult, worktreeInfo WorktreeInfo) string {
	status := "✅ Implementation Complete"
	if !result.Success {
		status = "⚠️ Implementation Incomplete"
	}

	var parts []string
	parts = append(parts, fmt.Sprintf("## %s\n", status))
	parts = append(parts, fmt.Sprintf("**Branch:** `%s`\n", worktreeInfo.Branch))
	parts = append(parts, fmt.Sprintf("**Workspace:** `%s`\n", worktreeInfo.Path))
	parts = append(parts, fmt.Sprintf("\n%s\n", result.Summary))

	if len(result.FilesChanged) > 0 {
		parts = append(parts, "\n**Files Changed:**\n")
		for _, f := range result.FilesChanged {
			parts = append(parts, fmt.Sprintf("- %s\n", f))
		}
	}

	if result.TestsRun {
		testStatus := "✅ Passed"
		if !result.TestsPassed {
			testStatus = "⚠️ Failed - Review Required"
		}
		parts = append(parts, fmt.Sprintf("\n**Tests:** %s\n", testStatus))
	}

	if result.CommitSha != "" {
		sha := result.CommitSha
		if len(sha) > 7 {
			sha = sha[:7]
		}
		parts = append(parts, fmt.Sprintf("\n**Commit:** `%s`\n", sha))
	}

	if len(result.NextSteps) > 0 {
		parts = append(parts, "\n**Next Steps:**\n")
		for _, s := range result.NextSteps {
			parts = append(parts, fmt.Sprintf("- %s\n", s))
		}
	}

	if result.Error != "" {
		parts = append(parts, fmt.Sprintf("\n**Error:** %s\n", result.Error))
	}

	parts = append(parts, "\n---\n*Automatically generated by feature implementation agent*")

	return strings.Join(parts, "")
}

// --- Main Implementation Flow ---

func implementTicket(client *APIClient, ticketID, repoPath, workspaceRoot string) ImplementationResult {
	fmt.Fprintf(os.Stderr, "[Implement] Starting implementation for ticket: %s\n", ticketID)

	// Step 1: Resolve and fetch ticket
	fmt.Fprintf(os.Stderr, "[Implement] Step 1: Resolving ticket...\n")
	ticket, err := client.ResolveTicket(ticketID)
	if err != nil {
		return ImplementationResult{
			Success:   false,
			TicketKey: ticketID,
			Summary:   "Failed to resolve ticket",
			Error:     err.Error(),
		}
	}
	fmt.Fprintf(os.Stderr, "[Implement] Resolved ticket: %s - %s\n", ticket.Key, ticket.Title)

	// Step 2: Validate ticket type
	if ticket.Type != "feature" {
		return ImplementationResult{
			Success:   false,
			TicketKey: ticket.Key,
			Summary:   fmt.Sprintf("Ticket is a %s, not a feature. Use different workflow for bugs.", ticket.Type),
			Error:     "Invalid ticket type",
		}
	}

	// Step 3: Fetch comments for context
	fmt.Fprintf(os.Stderr, "[Implement] Step 2: Fetching ticket comments...\n")
	comments, err := client.GetComments(ticket.ID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Implement] Warning: Could not fetch comments: %v\n", err)
		comments = []TicketComment{}
	}
	fmt.Fprintf(os.Stderr, "[Implement] Found %d comments\n", len(comments))

	// Step 4: Create worktree
	fmt.Fprintf(os.Stderr, "[Implement] Step 3: Creating git worktree...\n")
	worktreeInfo, err := createWorktree(ticket.Key, repoPath, workspaceRoot)
	if err != nil {
		return ImplementationResult{
			Success:   false,
			TicketKey: ticket.Key,
			Summary:   "Failed to create worktree",
			Error:     err.Error(),
		}
	}
	fmt.Fprintf(os.Stderr, "[Implement] Worktree ready at: %s (branch: %s)\n", worktreeInfo.Path, worktreeInfo.Branch)

	// Step 5: Spawn subagent to implement feature
	fmt.Fprintf(os.Stderr, "[Implement] Step 4: Spawning Claude subagent...\n")
	subagentResult, err := spawnSubagent(ticket, comments, worktreeInfo)
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

	// Step 6: Update ticket state to "In Review"
	fmt.Fprintf(os.Stderr, "[Implement] Step 5: Updating ticket state...\n")
	if os.Getenv("AUTO_UPDATE_STATE") != "false" {
		states, err := client.GetWorkflow(ticket.ProjectID)
		if err == nil {
			for _, state := range states {
				nameLower := strings.ToLower(state.Name)
				if nameLower == "in review" || nameLower == "review" || nameLower == "in_review" {
					if err := client.UpdateTicket(ticket.ID, map[string]interface{}{"stateId": state.ID}); err != nil {
						fmt.Fprintf(os.Stderr, "[Implement] Warning: Could not update ticket state: %v\n", err)
					} else {
						fmt.Fprintf(os.Stderr, "[Implement] Updated ticket state to: %s\n", state.Name)
					}
					break
				}
			}
		}
	}

	// Step 7: Add comment with summary
	fmt.Fprintf(os.Stderr, "[Implement] Step 6: Adding implementation summary comment...\n")
	comment := formatImplementationComment(subagentResult, worktreeInfo)
	if err := client.AddComment(ticket.ID, comment); err != nil {
		fmt.Fprintf(os.Stderr, "[Implement] Warning: Could not add comment: %v\n", err)
	} else {
		fmt.Fprintf(os.Stderr, "[Implement] Added completion comment to ticket\n")
	}

	fmt.Fprintf(os.Stderr, "[Implement] Completed: %s\n", map[bool]string{true: "SUCCESS", false: "FAILED"}[subagentResult.Success])

	return ImplementationResult{
		Success:       subagentResult.Success,
		TicketKey:     ticket.Key,
		WorkspacePath: worktreeInfo.Path,
		Branch:        worktreeInfo.Branch,
		Summary:       subagentResult.Summary,
		FilesChanged:  subagentResult.FilesChanged,
		TestsRun:      subagentResult.TestsRun,
		TestsPassed:   subagentResult.TestsPassed,
		CommitSha:     subagentResult.CommitSha,
		NextSteps:     subagentResult.NextSteps,
		Error:         subagentResult.Error,
	}
}

func main() {
	ticketID := flag.String("ticket", "", "Ticket UUID or key (e.g., PROJ-001)")
	repoPath := flag.String("repo", os.Getenv("REPO_PATH"), "Path to repository")
	workspaceRoot := flag.String("workspace", os.Getenv("WORKSPACE_ROOT"), "Root directory for worktrees")
	flag.Parse()

	if *ticketID == "" {
		fmt.Fprintf(os.Stderr, "Usage: implement-ticket --ticket <UUID|KEY> [--repo <PATH>] [--workspace <PATH>]\n")
		fmt.Fprintf(os.Stderr, "\nEnvironment variables:\n")
		fmt.Fprintf(os.Stderr, "  KEYCLOAK_BASE_URL    Keycloak server URL (default: http://localhost:8081)\n")
		fmt.Fprintf(os.Stderr, "  KEYCLOAK_REALM       Keycloak realm (default: ticketing)\n")
		fmt.Fprintf(os.Stderr, "  KEYCLOAK_CLIENT_ID   Keycloak client ID (default: myclient)\n")
		fmt.Fprintf(os.Stderr, "  KEYCLOAK_USERNAME    Keycloak username (required)\n")
		fmt.Fprintf(os.Stderr, "  KEYCLOAK_PASSWORD    Keycloak password (required)\n")
		fmt.Fprintf(os.Stderr, "  TICKETING_API_BASE_URL  Ticketing API URL (default: http://localhost:8080)\n")
		fmt.Fprintf(os.Stderr, "  REPO_PATH            Repository path (default: current directory)\n")
		fmt.Fprintf(os.Stderr, "  WORKSPACE_ROOT       Worktree root directory (default: ~/worktrees)\n")
		fmt.Fprintf(os.Stderr, "  AUTO_UPDATE_STATE    Update ticket state after implementation (default: true)\n")
		fmt.Fprintf(os.Stderr, "  SUBAGENT_TIMEOUT     Subagent timeout duration (default: 30m)\n")
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

	// Initialize Keycloak auth
	auth, err := NewKeycloakAuth()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Create API client
	client := NewAPIClient(auth)

	// Run implementation
	result := implementTicket(client, *ticketID, *repoPath, *workspaceRoot)

	// Output result as JSON
	jsonResult, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(jsonResult))

	if !result.Success {
		os.Exit(1)
	}
}
