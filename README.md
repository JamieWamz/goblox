# goblox - CLI Task Tracker

> Production-grade task management from your terminal

[![Go Version](https://img.shields.io/badge/Go-1.22-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## Why goblox?

I built goblox to demonstrate production-ready Go skills for remote backend roles. This isn't a student project - it's a portfolio piece showing:

- Clean architecture with separation of concerns
- SQLite with proper migrations and WAL mode
- Comprehensive test coverage (>80%)
- Professional CLI with cobra
- CI/CD with GitHub Actions

## Features

- ✅ Add tasks with priority and due dates
- ✅ List tasks with filtering (status, priority)
- ✅ Update task details
- ✅ Delete/Archive tasks
- ✅ Export to JSON/CSV

## Installation

```bash
go install github.com/JamieWamz/goblox@latest
```

## Usage

```bash
# Add a task
goblox add "Complete project documentation" -p high -d 2024-12-31

# List all tasks
goblox list

# List tasks by status
goblox list -s pending

# List tasks by priority
goblox list -p high

# Update a task
goblox update <id>

# Delete a task
goblox delete <id>
```

## Development

```bash
# Build
make build

# Run tests
make test

# Run with coverage
make coverage
```

## License

MIT License - see [LICENSE](LICENSE) for details.