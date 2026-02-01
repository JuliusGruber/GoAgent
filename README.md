# GoAgent

A Claude-powered AI agent written in Go.

This is an implementation of the tutorial [How to Build an Agent](https://ampcode.com/how-to-build-an-agent) from ampcode.com, which demonstrates how to create a functional code-editing AI agent in under 400 lines of Go code.

## Requirements

- Go 1.24+
- Anthropic API key

## Usage

Set your Anthropic API key:

```bash
export ANTHROPIC_API_KEY=your-api-key
```

Run the agent:

```bash
go run .
```

## Self-Development Success

After finishing the tutorial, I used GoAgent to build GoAgent itself. With just four tools, it was able to successfully modify itself and change its color to orange.
