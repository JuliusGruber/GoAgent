package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"goagent/agent"
)

type EditFileInput struct {
	Path   string `json:"path" jsonschema_description:"The relative path of the file to edit."`
	OldStr string `json:"old_str" jsonschema_description:"The text to replace. If empty and file doesn't exist, creates a new file."`
	NewStr string `json:"new_str" jsonschema_description:"The replacement text."`
}

var EditFileDefinition = agent.ToolDefinition{
	Name:        "edit_file",
	Description: "Edit a file by replacing old_str with new_str. If the file doesn't exist and old_str is empty, creates a new file with new_str as content.",
	InputSchema: GenerateSchema[EditFileInput](),
	Function:    EditFile,
}

func EditFile(input json.RawMessage) (string, error) {
	parsed, err := ParseInput[EditFileInput](input)
	if err != nil {
		return "", err
	}

	if parsed.Path == "" || parsed.OldStr == parsed.NewStr {
		return "", fmt.Errorf("invalid input parameters")
	}

	content, err := os.ReadFile(parsed.Path)
	if err != nil {
		if os.IsNotExist(err) && parsed.OldStr == "" {
			return createNewFile(parsed.Path, parsed.NewStr)
		}
		return "", err
	}

	oldContent := string(content)
	newContent := strings.Replace(oldContent, parsed.OldStr, parsed.NewStr, -1)

	if oldContent == newContent && parsed.OldStr != "" {
		return "", fmt.Errorf("old_str not found in file")
	}

	err = os.WriteFile(parsed.Path, []byte(newContent), 0644)
	if err != nil {
		return "", err
	}

	return "OK", nil
}

func createNewFile(path, content string) (string, error) {
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", err
		}
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", err
	}
	return "OK", nil
}
