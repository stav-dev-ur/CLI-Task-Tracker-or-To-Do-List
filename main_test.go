package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func invoke(t *testing.T, path string, args ...string) (string, error) {
	t.Helper()
	args = append([]string{"--file", path}, args...)
	var output bytes.Buffer
	err := run(args, &output, &output)
	return output.String(), err
}

func TestAddCompleteListAndDelete(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")

	output, err := invoke(t, path, "add", "Write tests")
	if err != nil || !strings.Contains(output, "Added task 1") {
		t.Fatalf("add output = %q, err = %v", output, err)
	}

	output, err = invoke(t, path, "complete", "1")
	if err != nil || !strings.Contains(output, "Completed task 1") {
		t.Fatalf("complete output = %q, err = %v", output, err)
	}

	output, err = invoke(t, path, "list")
	if err != nil || !strings.Contains(output, "[x] 1: Write tests") {
		t.Fatalf("list output = %q, err = %v", output, err)
	}

	output, err = invoke(t, path, "delete", "1")
	if err != nil || !strings.Contains(output, "Deleted task 1") {
		t.Fatalf("delete output = %q, err = %v", output, err)
	}
	tasks, err := loadTasks(path)
	if err != nil || len(tasks) != 0 {
		t.Fatalf("tasks after delete = %#v, err = %v", tasks, err)
	}
}

func TestMissingTaskIsReported(t *testing.T) {
	_, err := invoke(t, filepath.Join(t.TempDir(), "tasks.json"), "complete", "99")
	if err == nil || !strings.Contains(err.Error(), "task 99 not found") {
		t.Fatalf("expected missing task error, got %v", err)
	}
}

func TestInvalidJSONIsReported(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	if err := os.WriteFile(path, []byte("{"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := invoke(t, path, "list")
	if err == nil || !strings.Contains(err.Error(), "contains invalid JSON") {
		t.Fatalf("expected invalid JSON error, got %v", err)
	}
}
