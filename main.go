package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"goagent/agent"
)

var scanner = bufio.NewScanner(os.Stdin)

func getUserMessage() (string, bool) {
	if !scanner.Scan() {
		return "", false
	}
	return scanner.Text(), true
}

func main() {
	client := anthropic.NewClient()
	claudeAgent := agent.NewAgent(&client, getUserMessage)
	err := claudeAgent.Run(context.TODO())
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
}
