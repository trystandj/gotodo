# CLAUDE.md

This file gives Claude Code guidance for working in this repository.

## Project

GoTodo is a task tracker written in Go (module name `gotodo`), based on the roadmap.sh Task Tracker project. It has two modes:

- **Interactive TUI** (no arguments): a full-screen Bubble Tea app in `ui/app.go`.
- **Command mode** (`add`, `update`, `delete`, `mark-in-progress`, `mark-done`, `list [status]`): handled by the `switch` in `main.go`.

## Layout

- `main.go`: entry point and CLI argument parsing.
- `functions/functions.go`: the `Task` model, the `Status` constants (`todo`, `in-progress`, `done`) and all persistence (read, modify, write the whole JSON file on every operation).
- Task data lives in `<os.UserConfigDir()>/gotodo/tasks.json`, which `$GOTODO_FILE` overrides. When that file doesn't exist, the app falls back to reading the legacy `./functions/task.json`, and the next save copies those tasks to the new location. Saves write a `.tmp` file and rename it into place.
- `ui/app.go`: the TUI model, including the menu, list, input and confirm screens. Styling uses lipgloss, and column widths use go-runewidth.

## Commands

```bash
go build -o todo .     # build (on Windows, use -o todo.exe)
go run .               # run the interactive TUI
go run . list          # run a CLI command
go vet ./...           # static checks
gofmt -l .             # list unformatted files
```

There are no tests yet. If you add them, put them in `functions/functions_test.go` and point `filename` at a temp file (`t.TempDir()`).

## Conventions

- Task IDs are strings holding integers. `NextId` returns the highest remaining ID plus one, so if you delete the highest-numbered task, its ID gets reused.
- Timestamps are RFC3339 with a local offset. Write them with `functions.Now()` and read them with `functions.ParseTime()`, which also accepts the legacy format (local time with a fake `Z`).
- Status values are the typed `functions.Status` constants. Don't use raw strings.
- The UI calls the `functions` package directly and calls `reload()` after every change. Keep all file I/O in `functions`.
- Use `fmt.Errorf("...: %w", err)` to wrap errors.
- Run `gofmt` on Go code before you finish.
- Task files are user data. When testing, set `GOTODO_FILE` to a scratch path so the real files aren't touched.
