# GoAgent Architecture

GoAgent is a CLI-based conversational agent that uses the Anthropic Claude API with tool-calling capabilities. This document describes the system architecture and design decisions.

## High-Level Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                              main.go                                │
│  - Entry point                                                      │
│  - API client initialization                                        │
│  - User input handling (stdin)                                      │
└─────────────────────────────────────────────────────────────────────┘
                                   │
                                   ▼
┌─────────────────────────────────────────────────────────────────────┐
│                           agent/agent.go                            │
│  - Conversation loop                                                │
│  - Tool execution loop                                              │
│  - API inference calls                                              │
└─────────────────────────────────────────────────────────────────────┘
                                   │
                                   ▼
┌─────────────────────────────────────────────────────────────────────┐
│                             tools/*                                 │
│  - Tool definitions (read_file, etc.)                               │
│  - JSON schema generation                                           │
│  - Tool implementations                                             │
└─────────────────────────────────────────────────────────────────────┘
```

## Package Structure

```
goagent/
├── main.go              # Entry point, stdin handling
├── agent/
│   ├── agent.go         # Agent struct, conversation loop
│   └── tool.go          # ToolDefinition type
└── tools/
    ├── tools.go         # Tool registry (GetAllTools)
    ├── json.go          # Generic schema/parsing helpers
    └── read_file.go     # ReadFile tool implementation
```

## Core Components

### 1. Agent (`agent/agent.go`)

The `Agent` struct is the central orchestrator:

```go
type Agent struct {
    client         *anthropic.Client
    getUserMessage func() (string, bool)
    tools          []ToolDefinition
}
```

**Key design decisions:**

- **Dependency injection**: The `getUserMessage` function is injected, allowing flexible input sources (stdin, tests, etc.)
- **Tool registry**: Tools are passed in at construction, making the agent agnostic to specific tool implementations

### 2. Conversation Loop

The `RunConversationLoop` method implements a **nested loop architecture**:

```
┌──────────────────────────────────────────────────────────────────────┐
│  Outer Loop: User Input                                              │
│    for {                                                             │
│      read user input                                                 │
│      ┌────────────────────────────────────────────────────────────┐  │
│      │  Inner Loop: Tool Execution                                │  │
│      │    for {                                                   │  │
│      │      call Claude API                                       │  │
│      │      process response blocks                               │  │
│      │      if no tool_use → break                                │  │
│      │      execute tools, send results                           │  │
│      │    }                                                       │  │
│      └────────────────────────────────────────────────────────────┘  │
│    }                                                                 │
└──────────────────────────────────────────────────────────────────────┘
```

**Why nested loops?**

Claude's tool-calling API requires a request-response pattern:
1. Claude returns `tool_use` blocks requesting tool execution
2. Client executes tools and sends `tool_result` blocks
3. Claude processes results and may request more tools

The inner loop allows Claude to chain multiple tool calls autonomously before returning control to the user.

### 3. Tool Definition (`agent/tool.go`)

```go
type ToolDefinition struct {
    Name        string
    Description string
    InputSchema anthropic.ToolInputSchemaParam
    Function    func(input json.RawMessage) (string, error)
}
```

**Design principles:**

- **Self-describing**: Each tool carries its own schema for the API
- **Uniform interface**: All tools accept `json.RawMessage` and return `(string, error)`
- **Decoupled from agent**: Tools don't know about the Agent; they're pure functions

### 4. Tool Registry (`tools/tools.go`)

```go
func GetAllTools() []agent.ToolDefinition {
    return []agent.ToolDefinition{
        ReadFileDefinition,
    }
}
```

Central registry pattern - add new tools here to make them available to the agent.

### 5. Schema Generation (`tools/json.go`)

Uses Go generics for type-safe schema generation:

```go
func GenerateSchema[T any]() anthropic.ToolInputSchemaParam
func ParseInput[T any](input json.RawMessage) (T, error)
```

**Benefits:**
- Compile-time type safety
- Schema derived from struct tags (e.g., `jsonschema_description`)
- No manual JSON schema authoring

## Message Flow

```
User: "What's in main.go?"
         │
         ▼
┌─────────────────────┐
│ Create user message │
└─────────────────────┘
         │
         ▼
┌─────────────────────┐
│  runInference()     │──────────────────────────────────┐
└─────────────────────┘                                  │
         │                                               │
         ▼                                               │
┌─────────────────────┐                                  │
│ Claude returns:     │                                  │
│ [tool_use:          │                                  │
│   read_file("...")]│                                  │
└─────────────────────┘                                  │
         │                                               │
         ▼                                               │
┌─────────────────────┐                                  │
│  executeTool()      │                                  │
│  → reads file       │                                  │
│  → returns content  │                                  │
└─────────────────────┘                                  │
         │                                               │
         ▼                                               │
┌─────────────────────┐                                  │
│ Append tool_result  │                                  │
│ as user message     │──────────────────────────────────┘
└─────────────────────┘        (loop continues)
         │
         ▼
┌─────────────────────┐
│ Claude returns:     │
│ [text: "The file    │
│  contains..."]      │
└─────────────────────┘
         │
         ▼
┌─────────────────────┐
│ No tools → break    │
│ Print text response │
│ Back to user prompt │
└─────────────────────┘
```

## Error Handling Strategy

| Error Source | Handling |
|--------------|----------|
| Tool not found | Return `tool_result` with `isError: true` |
| Tool execution fails | Return `tool_result` with `isError: true` |
| API error | Propagate up, exit conversation loop |
| User EOF (ctrl-c) | Break loop, clean exit |

Claude receives tool errors and can decide how to proceed (retry, inform user, try alternative).

## Adding New Tools

1. Create `tools/new_tool.go`:

```go
type NewToolInput struct {
    Param string `json:"param" jsonschema_description:"Description"`
}

var NewToolDefinition = agent.ToolDefinition{
    Name:        "new_tool",
    Description: "What this tool does",
    InputSchema: GenerateSchema[NewToolInput](),
    Function:    NewToolFunc,
}

func NewToolFunc(input json.RawMessage) (string, error) {
    parsed, err := ParseInput[NewToolInput](input)
    if err != nil {
        return "", err
    }
    // Implementation
    return "result", nil
}
```

2. Register in `tools/tools.go`:

```go
func GetAllTools() []agent.ToolDefinition {
    return []agent.ToolDefinition{
        ReadFileDefinition,
        NewToolDefinition,  // Add here
    }
}
```

## Future Considerations

- **Streaming responses**: Currently waits for complete response; could stream text as it arrives
- **Parallel tool execution**: Multiple `tool_use` blocks could be executed concurrently
- **Tool result caching**: Cache results for idempotent tools
- **Conversation persistence**: Save/restore conversation state
- **System prompts**: Add configurable system prompts for agent behavior
