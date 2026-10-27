# GoTodo

A simple command-line tool to track and manage your tasks, inspired by the [Task Tracker project on roadmap.sh][task-tracker].

## Features

- Add, update, and delete tasks from the command line
- Mark tasks as *todo*, *in-progress*, or *done*
- List all tasks or filter by status
- A full-screen interactive mode
- Tasks are stored in a local JSON file for persistence

## Getting started

### Prerequisites

- Go installed on your system (`go version`)

### Installation

From the project folder:

```bash
go build -o todo .       # macOS / Linux
go build -o todo.exe .   # Windows
```

### Interactive mode

Run the app with no arguments to open it as a full-screen terminal app:

```bash
./todo
```

You'll get a menu to pick from. Everything happens in one session, so there's no need to re-run the app.

| Screen | Keys |
| --- | --- |
| Menu | `↑/↓` move, `enter` select, `1-6` jump, `q` quit |
| Task list | `enter` cycle status, `1` todo, `2` in progress, `3` done, `a` add, `e` edit, `d` delete, `tab` switch filter, `esc` back to menu, `q` quit |
| Add / edit | `enter` save, `esc` cancel |

### Command mode

You can also run one-shot commands:

```bash
# Add a new task (quotes are optional)
./todo add "Buy groceries"

# Update a task
./todo update 1 "Buy groceries and cook dinner"

# Delete a task
./todo delete 1

# Mark a task as in progress
./todo mark-in-progress 1

# Mark a task as done
./todo mark-done 1

# List all tasks
./todo list

# List tasks by status
./todo list todo
./todo list in-progress
./todo list done

# Show help
./todo help
```

## Where tasks are stored

Tasks are saved to `tasks.json` in a `gotodo` folder inside your user config directory, so the app works from any folder:

| OS | Path |
| --- | --- |
| Windows | `%AppData%\gotodo\tasks.json` |
| macOS | `~/Library/Application Support/gotodo/tasks.json` |
| Linux | `~/.config/gotodo/tasks.json` |

To use a different file, set the `GOTODO_FILE` environment variable. If you have tasks from an older version in `functions/task.json`, run the app once from the project folder. It reads them and copies them to the new location the next time you make a change.

Each task includes:

- `id`: Unique identifier
- `description`: Task details
- `status`: `todo`, `in-progress`, or `done`
- `createdAt`: Creation timestamp
- `updatedAt`: Last update timestamp

## Why this project?

This CLI helps you practice Go programming, file handling, and user input processing. It's a practical tool for personal productivity and a good learning project for beginners and intermediate developers.

## Contributing

Pull requests and suggestions are welcome!

## References

- [Task Tracker project on roadmap.sh][task-tracker]
- [Community solutions][solutions]

[task-tracker]: https://roadmap.sh/projects/task-tracker
[solutions]: https://roadmap.sh/projects/task-tracker/solutions
