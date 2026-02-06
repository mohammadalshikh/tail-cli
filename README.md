# Tail CLI

A CLI tool for analyzing log files to detect P99 latency spikes and explain root causes using AI

## Features
- P99 latency analysis using Top-K selection with min-heap
- AI root cause analysis
- Interactive TUI for visualizing outliers
- Single binary distribution

## Requirements
- Go 1.25+
- OpenAI API key (optional with `--no-ai` flag)

## Installation

```bash
go install github.com/mohammadalshikh/tail-cli@latest
```

## Usage

```bash
# Interactive TUI mode (default)
export OPENAI_API_KEY="sk-..."
tail-cli analyze --file logs/app.log

# Analyze top 10 outliers
tail-cli analyze -f logs/app.log -t 10

# Plain text output (for scripts/CI/CD)
tail-cli analyze -f logs/app.log --no-tui

# Skip AI analysis (no API key)
tail-cli analyze -f logs/app.log --no-ai

# Get help
tail-cli help
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--file` | `-f` | *required* | Path to the log file |
| `--top` | `-t` | `5` | Number of top outliers to analyze |
| `--no-tui` | | `false` | Plain text output |
| `--no-ai` | | `false` | Skip AI analysis (no API key) |

### Log Format

Logs must be JSON with `latency_ms` field:

```json
{"level":"info","latency_ms":120,"msg":"GET /api/users"}
{"level":"error","latency_ms":5500,"msg":"DB Timeout"}
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