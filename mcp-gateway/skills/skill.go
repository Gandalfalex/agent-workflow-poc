package skills

// Skill is an MCP-aligned capability exposed to agents.
type Skill interface {
	Name() string
	Description() string
	Execute(input map[string]any) (map[string]any, error)
}