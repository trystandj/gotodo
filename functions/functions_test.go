package functions

import (
	"path/filepath"
	"testing"
)

// useTempFile points the package at an empty task file for one test.
func useTempFile(t *testing.T) {
	t.Helper()
	old := filename
	filename = filepath.Join(t.TempDir(), "tasks.json")
	t.Cleanup(func() { filename = old })
}

func TestDeleteTaskById(t *testing.T) {
	useTempFile(t)
	for _, d := range []string{"first", "second", "third"} {
		if err := AddTask(d); err != nil {
			t.Fatalf("AddTask(%q): %v", d, err)
		}
	}

	if err := DeleteTaskById("2"); err != nil {
		t.Fatalf("DeleteTaskById(2): %v", err)
	}

	tasks, err := GetTasks()
	if err != nil {
		t.Fatalf("GetTasks: %v", err)
	}
	if len(tasks) != 2 || tasks[0].Id != "1" || tasks[1].Id != "3" {
		t.Fatalf("after delete got %+v, want tasks 1 and 3", tasks)
	}
	if _, err := GetTaskById("2"); err == nil {
		t.Fatal("task 2 still found after delete")
	}
}

func TestDeleteTaskByIdNotFound(t *testing.T) {
	useTempFile(t)
	if err := AddTask("only"); err != nil {
		t.Fatalf("AddTask: %v", err)
	}
	if err := DeleteTaskById("42"); err == nil {
		t.Fatal("deleting a missing ID returned no error")
	}
	if n, _ := Count(); n != 1 {
		t.Fatalf("Count = %d after failed delete, want 1", n)
	}
}
