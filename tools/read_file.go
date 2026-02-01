package tools

import (
	"encoding/json"
	"os"

	"goagent/agent"
)

type ReadFileInput struct {
	Path string `json:"path" jsonschema_description:"The relative path of a file in the working directory."`
}

var ReadFileDefinition = agent.ToolDefinition{
	Name:        "read_file",
	Description: "Read the contents of a given relative file path. Use this when you want to see what's inside a file. Do not use this with directory names.",
	InputSchema: GenerateSchema[ReadFileInput](),
	Function:    ReadFile,
}

// ReadFile is executed locally by the agent when Claude responds with a tool_use block.
func ReadFile(input json.RawMessage) (string, error) {
	parsed, err := ParseInput[ReadFileInput](input)
	if err != nil {
		return "", err
	}
	content, err := os.ReadFile(parsed.Path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
