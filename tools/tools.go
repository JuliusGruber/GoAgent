package tools

import "goagent/agent"

func GetAllTools() []agent.ToolDefinition {
	return []agent.ToolDefinition{
		ReadFileDefinition,
		ListFilesDefinition,
		EditFileDefinition,
		CreateFileDefinition,
		GhDefinition,
	}
}
