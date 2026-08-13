# goblox

`goblox` is a local-first task tracker for the terminal. It stores tasks in a
SQLite database and supports priorities, due dates, lifecycle statuses, filtered
lists, and JSON/CSV exports.

[![CI](https://github.com/JamieWamz/goblox/actions/workflows/ci.yml/badge.svg)](https://github.com/JamieWamz/goblox/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.22+-blue.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## Features

- Add, inspect, edit, archive, and permanently delete tasks
- Track `low`, `medium`, or `high` priority
- Track `pending`, `in_progress`, `done`, or `archived` status
- Store optional due dates in `YYYY-MM-DD` format
- Filter lists and exports by priority or status
- Print tables for people and JSON/CSV for scripts
- Use short, unique task ID prefixes in every task command
- Apply embedded, versioned SQLite migrations automatically

## Installation

`goblox` requires Go 1.22 or newer and a C compiler because its SQLite driver
uses CGO.

```bash
go install github.com/JamieWamz/goblox/cmd/goblox@latest
```

To build from a clone instead:

```bash
make build
./bin/goblox --help
```

## Quick start

```bash
# Descriptions may be quoted or supplied as multiple words.
goblox add Prepare release notes --priority high --due 2030-01-02

goblox list
goblox list --status pending --priority high
goblox show <id>

# Flags make updates scriptable. With no update flags, goblox prompts for values.
goblox update <id> --status in_progress
goblox update <id> --description "Publish release notes" --clear-due

goblox archive <id>
goblox delete <id>              # asks for confirmation
goblox delete <id> --force      # intended for automation
```

Table output shows the first eight characters of each task ID. That prefix can
be passed to `show`, `update`, `archive`, and `delete`; goblox rejects the prefix
if it is not unique.

## JSON and CSV

`list` and `show` accept `--format table`, `--format json`, or `--format csv`:

```bash
goblox list --format json
goblox show <id> --format csv
```

Use `export` to write JSON or CSV to standard output or a file:

```bash
goblox export --format json
goblox export --format csv --output tasks.csv
goblox export --format json --status done --output completed.json
```

## Database

The default database is `./goblox.db`. Select a different location with the
persistent `--db` flag:

```bash
goblox --db "$HOME/.local/share/goblox/tasks.db" list
```

SQLite WAL mode and a busy timeout are enabled. Schema migrations are embedded
in the binary, recorded in `schema_migrations`, and applied when a database is
opened. A new database starts empty.

## Development

```bash
make check       # formatting check, vet, race-enabled tests, and build
make test        # race-enabled tests with package coverage
make coverage    # generate coverage.out and coverage.html
make install     # install the command with go install
```

The test suite covers the model, storage, output encoders, validation failures,
interactive command branches, and a complete CLI lifecycle. CI enforces at least
80% total statement coverage.

## License

Released under the [MIT License](LICENSE).
