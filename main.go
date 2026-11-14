package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Task struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
	CreatedAt   string `json:"created_at"`
}

func loadTasks(path string) ([]Task, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return []Task{}, nil
	}
	if err != nil {
		return nil, err
	}

	var tasks []Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, fmt.Errorf("%s contains invalid JSON: %w", path, err)
	}
	if tasks == nil {
		return nil, fmt.Errorf("%s does not contain a valid task list", path)
	}

	seen := make(map[int]bool, len(tasks))
	for _, task := range tasks {
		if task.ID <= 0 || task.Description == "" || task.CreatedAt == "" || seen[task.ID] {
			return nil, fmt.Errorf("%s does not contain a valid task list", path)
		}
		seen[task.ID] = true
	}
	return tasks, nil
}

func saveTasks(path string, tasks []Task) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	file, err := os.CreateTemp(directory, ".tasks-*.tmp")
	if err != nil {
		return err
	}
	tempPath := file.Name()
	defer os.Remove(tempPath)

	if _, err := file.Write(data); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := os.Rename(tempPath, path); err != nil {
		return err
	}
	return nil
}

func run(args []string, out, errOut io.Writer) error {
	flags := flag.NewFlagSet("task-tracker", flag.ContinueOnError)
	flags.SetOutput(errOut)
	path := flags.String("file", "tasks.json", "task file to use")
	flags.Usage = func() {
		fmt.Fprintln(errOut, "Usage: task-tracker [--file PATH] <add|list|complete|delete> [arguments]")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		return err
	}

	arguments := flags.Args()
	if len(arguments) == 0 {
		flags.Usage()
		return errors.New("a command is required")
	}

	tasks, err := loadTasks(*path)
	if err != nil {
		return err
	}

	switch arguments[0] {
	case "add":
		if len(arguments) < 2 {
			return errors.New("usage: task-tracker add <description>")
		}
		nextID := 1
		for _, task := range tasks {
			if task.ID >= nextID {
				nextID = task.ID + 1
			}
		}
		task := Task{
			ID:          nextID,
			Description: arguments[1],
			Completed:   false,
			CreatedAt:   time.Now().UTC().Format(time.RFC3339Nano),
		}
		tasks = append(tasks, task)
		if err := saveTasks(*path, tasks); err != nil {
			return err
		}
		fmt.Fprintf(out, "Added task %d: %s\n", task.ID, task.Description)
	case "list":
		if len(tasks) == 0 {
			fmt.Fprintln(out, "No tasks.")
			return nil
		}
		for _, task := range tasks {
			status := " "
			if task.Completed {
				status = "x"
			}
			fmt.Fprintf(out, "[%s] %d: %s\n", status, task.ID, task.Description)
		}
	case "complete", "delete":
		if len(arguments) != 2 {
			return fmt.Errorf("usage: task-tracker %s <id>", arguments[0])
		}
		id, err := strconv.Atoi(arguments[1])
		if err != nil || id <= 0 {
			return fmt.Errorf("invalid task ID %q", arguments[1])
		}
		index := -1
		for i := range tasks {
			if tasks[i].ID == id {
				index = i
				break
			}
		}
		if index == -1 {
			return fmt.Errorf("task %d not found", id)
		}
		task := tasks[index]
		if arguments[0] == "complete" {
			if task.Completed {
				fmt.Fprintf(out, "Task %d is already complete.\n", id)
				return nil
			}
			tasks[index].Completed = true
			if err := saveTasks(*path, tasks); err != nil {
				return err
			}
			fmt.Fprintf(out, "Completed task %d: %s\n", id, task.Description)
		} else {
			tasks = append(tasks[:index], tasks[index+1:]...)
			if err := saveTasks(*path, tasks); err != nil {
				return err
			}
			fmt.Fprintf(out, "Deleted task %d: %s\n", id, task.Description)
		}
	default:
		return fmt.Errorf("unknown command %q", arguments[0])
	}
	return nil
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
