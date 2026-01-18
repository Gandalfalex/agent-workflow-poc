# Feature Implementation Task

## Ticket Information

**Ticket ID:** {{ticketKey}}
**Title:** {{title}}
**Type:** {{type}}
**Priority:** {{priority}}
**Status:** {{status}}

### Description
{{description}}

{{#story}}
## Story (Epic Context)
**Story Title:** {{story.title}}
{{story.description}}
{{/story}}

## Ticket Discussion & Comments

{{#comments}}
{{#items}}
**{{authorName}}** ({{createdAt}}):
> {{message}}

{{/items}}
{{/comments}}

## Your Task

Implement the feature described above in this isolated workspace.

### Guidelines

1. **Code Understanding**
   - Read existing code to understand project structure and patterns
   - Look for similar implementations to follow conventions
   - Check for existing tests to understand testing patterns

2. **Implementation**
   - Implement the feature according to the description
   - Follow the project's code style and conventions
   - Make changes in logical, focused commits
   - Do NOT modify git remotes or push code

3. **Testing**
   - Write tests for new functionality
   - Run existing test suite to ensure nothing breaks
   - Verify all tests pass before completing

4. **Commits**
   - Make clear, descriptive commit messages
   - Format: `{{ticketKey}}: Brief description of changes`
   - Keep commits focused and logical
   - Example: `PROJ-001: Add user profile page component`

### Technical Details

- **Working Directory:** {{workspacePath}}
- **Branch:** {{branch}}
- **Repository:** {{repoRoot}}
- **Ticket Key:** {{ticketKey}}

### Success Criteria

✅ Feature is fully implemented as described
✅ Code follows project conventions and style
✅ Tests are written for new functionality
✅ All tests pass (both new and existing)
✅ Changes are committed with clear messages
✅ No uncommitted changes remain

### Important Constraints

- ⚠️ Work ONLY in {{workspacePath}}
- ⚠️ Branch commits to {{branch}} only
- ⚠️ Do NOT push to remote - commits stay local
- ⚠️ Do NOT modify remote configuration
- ⚠️ Complete within reasonable time (start with most critical parts)

## When Complete

Provide a JSON response with implementation details:

\`\`\`json
{
  "success": true,
  "summary": "What was implemented (1-2 sentences)",
  "filesChanged": ["file1.ts", "file2.ts", "..."],
  "testsRun": true,
  "testsPassed": true,
  "commitSha": "abc123def456",
  "nextSteps": ["Any manual review steps", "Additional testing needed"]
}
\`\`\`

## If Implementation Fails

If you cannot complete the implementation, provide:

\`\`\`json
{
  "success": false,
  "summary": "Why implementation failed",
  "nextSteps": ["What needs to be done manually"]
}
\`\`\`

---

**Begin implementation now.**
