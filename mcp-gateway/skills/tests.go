package skills

import (
	"os/exec"
)

type GoTestSkill struct{}

func (t GoTestSkill) Name() string {
	return "go.test"
}

func (t GoTestSkill) Description() string {
	return "Runs go test ./... in a repository."
}

func (t GoTestSkill) Execute(input map[string]any) (map[string]any, error) {
	repo := input["repo"].(string)
	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = repo
	out, err := cmd.CombinedOutput()
	return map[string]any{"output": string(out)}, err
}