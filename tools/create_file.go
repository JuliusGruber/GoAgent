package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"goagent/agent"
)

type CreateFileInput struct {
	Path    string `json:"path" jsonschema_description:"The path to the file to create"`
	Content string `json:"content" jsonschema_description:"The content to write to the file"`
}

var CreateFileDefinition = agent.ToolDefinition{
	Name:        "create_file",
	Description: "Create a new file with the given content at the specified path. Creates parent directories if needed.",
	InputSchema: GenerateSchema[CreateFileInput](),
	Function:    CreateFile,
}

func CreateFile(input json.RawMessage) (string, error) {
	parsed, err := ParseInput[CreateFileInput](input)
	if err != nil {
		return "", err
	}
	return createFile(parsed.Path, parsed.Content)
}

func createFile(filePath, content string) (string, error) {
	dir := filepath.Dir(filePath)
	if dir != "." && dir != "" {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create directory: %w", err)
		}
	}

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}

	return fmt.Sprintf("Successfully created file %s", filePath), nil
}
