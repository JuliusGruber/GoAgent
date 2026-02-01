package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
)

type Agent struct {
	client         *anthropic.Client
	getUserMessage func() (string, bool)
	tools          []ToolDefinition
}

func NewAgent(client *anthropic.Client, getUserMessage func() (string, bool), tools []ToolDefinition) *Agent {
	return &Agent{
		client:         client,
		getUserMessage: getUserMessage,
		tools:          tools,
	}
}

func (a *Agent) RunConversationLoop(ctx context.Context) error {
	conversation := []anthropic.MessageParam{}

	fmt.Println("Chat with Claude (use 'ctrl-c' to quit)")

	for {
		fmt.Print("\u001b[92mYou\u001b[0m: ")
		userInput, ok := a.getUserMessage()
		if !ok {
			break
		}

		userMessage := anthropic.NewUserMessage(anthropic.NewTextBlock(userInput))
		conversation = append(conversation, userMessage)

		// Inner tool loop: keeps calling Claude until no more tool_use blocks are returned.
		// This allows Claude to chain multiple tool calls autonomously without user input.
		for {
			message, err := a.runInference(ctx, conversation)
			if err != nil {
				return err
			}
			conversation = append(conversation, message.ToParam())

			// Check if Claude wants to use any tools
			toolResults := []anthropic.ContentBlockParamUnion{}
			for _, content := range message.Content {
				switch content.Type {
				case "text":
					fmt.Printf("\u001b[38;5;208mClaude\u001b[0m: %s\n", content.Text)
				case "tool_use":
					result := a.executeTool(content.ID, content.Name, content.Input)
					toolResults = append(toolResults, result)
				}
			}

			// If no tools were used, we're done with this turn
			if len(toolResults) == 0 {
				break
			}

			// Add tool results and let Claude continue
			conversation = append(conversation, anthropic.NewUserMessage(toolResults...))
		}
	}

	return nil
}

func (a *Agent) executeTool(toolID, toolName string, input json.RawMessage) anthropic.ContentBlockParamUnion {
	// Find the tool by name
	var toolDef *ToolDefinition
	for _, t := range a.tools {
		if t.Name == toolName {
			toolDef = &t
			break
		}
	}

	if toolDef == nil {
		return anthropic.NewToolResultBlock(toolID, fmt.Sprintf("tool %q not found", toolName), true)
	}

	// Execute the tool
	result, err := toolDef.Function(input)
	if err != nil {
		return anthropic.NewToolResultBlock(toolID, fmt.Sprintf("error: %s", err.Error()), true)
	}

	return anthropic.NewToolResultBlock(toolID, result, false)
}

func (a *Agent) runInference(ctx context.Context, conversation []anthropic.MessageParam) (*anthropic.Message, error) {
	anthropicTools := []anthropic.ToolUnionParam{}
	for _, tool := range a.tools {
		anthropicTools = append(anthropicTools, anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        tool.Name,
				Description: anthropic.String(tool.Description),
				InputSchema: tool.InputSchema,
			},
		})
	}

	message, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_20250514,
		MaxTokens: int64(1024),
		Messages:  conversation,
		Tools:     anthropicTools,
	})
	return message, err
}
