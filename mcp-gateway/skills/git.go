package skills

import (
	"os/exec"
)

type GitDiffSkill struct{}

func (g GitDiffSkill) Name() string {
	return "git.diff"
}

func (g GitDiffSkill) Description() string {
	return "Returns git diff for a given repository path."
}

func (g GitDiffSkill) Execute(input map[string]any) (map[string]any, error) {
	repo := input["repo"].(string)
	cmd := exec.Command("git", "-C", repo, "diff")
	out, err := cmd.CombinedOutput()
	return map[string]any{"diff": string(out)}, err
}