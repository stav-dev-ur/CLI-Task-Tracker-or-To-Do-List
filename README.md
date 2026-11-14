# CLI Task Tracker

A dependency-free Go command-line tool that stores tasks in a local JSON file.

## Requirements

Go 1.21 or later.

## Usage

Run commands from this folder:

```text
go run . add "Buy groceries"
go run . list
go run . complete 1
go run . delete 1
```

By default, tasks are saved in `tasks.json` in the current directory. Use
`--file` before the command to select another JSON file:

```text
go run . --file C:\Users\you\tasks.json add "Call Sam"
```

Task IDs are shown by `list` and by the `add` command. Completed tasks remain
in the list marked with `[x]`. Build a standalone executable with `go build`.
