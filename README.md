# Tail CLI

A CLI tool for analyzing log files to detect P99 latency spikes and explain root causes using AI

## Features
- P99 latency analysis using Top-K selection with min-heap
- AI root cause analysis
- Interactive TUI for visualizing outliers
- Single binary distribution

## Requirements
- Go 1.25+
- OpenAI API key

## Installation

```bash
go install github.com/mohammadalshikh/tail-cli@latest
```

## Usage

```bash
# Set your OpenAI API key
export OPENAI_API_KEY="sk-..."

# Analyze a log file
tail-cli analyze --file logs/app.log
```

## Development

```bash
# Set your OpenAI API key
export OPENAI_API_KEY="sk-..."

# Run locally
go run main.go analyze --file test.log

# Build binary
go build -o tail-cli
```