package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"goagent/agent"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

func main() {
	apiKey, ok := getAnthropicAPIKey()
	if !ok {
		return
	}
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	claudeAgent := agent.NewAgent(&client, getUserMessage)
	err := claudeAgent.RunConversationLoop(context.TODO())
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
}

func getAnthropicAPIKey() (string, bool) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: ANTHROPIC_API_KEY environment variable is not set")
		return "", false
	}
	return apiKey, true
}

var scanner = bufio.NewScanner(os.Stdin)

func getUserMessage() (string, bool) {
	if !scanner.Scan() {
		return "", false
	}
	return scanner.Text(), true
}
