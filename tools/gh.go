package tools

import (
	"encoding/json"
	"os/exec"
	"strings"

	"goagent/agent"
)

type GhInput struct {
	Args string `json:"args" jsonschema_description:"The arguments to pass to the gh command (e.g., 'issue list', 'pr view 123', 'repo view')"`
}

var GhDefinition = agent.ToolDefinition{
	Name:        "run_gh",
	Description: "Run GitHub CLI (gh) commands to interact with GitHub. Can list/create issues, view/create PRs, manage repos, etc. The gh CLI must be authenticated. Examples: 'issue list', 'pr view 123', 'repo view owner/repo'.",
	InputSchema: GenerateSchema[GhInput](),
	Function:    RunGh,
}

func RunGh(input json.RawMessage) (string, error) {
	parsed, err := ParseInput[GhInput](input)
	if err != nil {
		return "", err
	}

	args := strings.Fields(parsed.Args)
	cmd := exec.Command("gh", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), err
	}
	return string(output), nil
}
