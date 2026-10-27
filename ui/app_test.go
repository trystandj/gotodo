package ui

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"gotodo/functions"
)

// TestMain re-runs the tests with GOTODO_FILE pointing at a temp file. The
// functions package picks its file at init, before any test code runs, so
// this is the only way to keep these tests away from the real task file.
func TestMain(m *testing.M) {
	if os.Getenv("GOTODO_FILE") != "" {
		os.Exit(m.Run())
	}
	dir, err := os.MkdirTemp("", "gotodo-ui-test")
	if err != nil {
		panic(err)
	}
	cmd := exec.Command(os.Args[0], os.Args[1:]...)
	cmd.Env = append(os.Environ(), "GOTODO_FILE="+filepath.Join(dir, "tasks.json"))
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	err = cmd.Run()
	os.RemoveAll(dir)
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

var ansi = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// press sends keys to the model one at a time; plain strings are typed.
func press(t *testing.T, m model, keys ...any) model {
	t.Helper()
	for _, k := range keys {
		var msg tea.KeyMsg
		switch k := k.(type) {
		case tea.KeyType:
			msg = tea.KeyMsg{Type: k}
		case string:
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
		}
		next, _ := m.Update(msg)
		m = next.(model)
	}
	return m
}

// snapshot returns the rendered view without colours and logs it.
func snapshot(t *testing.T, label string, m model) string {
	t.Helper()
	v := ansi.ReplaceAllString(m.View(), "")
	t.Logf("---- %s ----\n%s", label, v)
	return v
}

func TestMenuDeleteTask(t *testing.T) {
	for _, d := range []string{"Buy milk", "Write report", "Call mom"} {
		if err := functions.AddTask(d); err != nil {
			t.Fatalf("AddTask: %v", err)
		}
	}
	t.Cleanup(func() {
		tasks, _ := functions.GetTasks()
		for _, task := range tasks {
			functions.DeleteTaskById(task.Id)
		}
	})

	m := newModel()
	if v := snapshot(t, "menu", m); !strings.Contains(v, "6. Delete a task") {
		t.Fatal("menu has no delete entry")
	}

	// Choose "Delete a task" and type an ID that doesn't exist.
	m = press(t, m, "6", "9", tea.KeyEnter)
	if v := snapshot(t, "unknown id", m); m.screen != inputScreen || !strings.Contains(v, "Task #9 not found") {
		t.Fatal("unknown ID was not rejected")
	}

	// Correct it to task 2 and confirm.
	m = press(t, m, tea.KeyBackspace, "2", tea.KeyEnter)
	if v := snapshot(t, "confirm", m); m.screen != confirmScreen || !strings.Contains(v, "#2  Write report") {
		t.Fatal("confirmation for task 2 not shown")
	}
	m = press(t, m, "y")
	v := snapshot(t, "after delete", m)
	if m.screen != listScreen || !strings.Contains(v, "Deleted task #2") || strings.Contains(v, "Write report") {
		t.Fatal("task 2 not deleted in the UI")
	}

	tasks, err := functions.GetTasks()
	if err != nil {
		t.Fatalf("GetTasks: %v", err)
	}
	if len(tasks) != 2 || tasks[0].Id != "1" || tasks[1].Id != "3" {
		t.Fatalf("file holds %+v, want tasks 1 and 3", tasks)
	}
}

func TestMenuDeleteCancel(t *testing.T) {
	if err := functions.AddTask("Keep me"); err != nil {
		t.Fatalf("AddTask: %v", err)
	}
	t.Cleanup(func() { functions.DeleteTaskById("1") })

	m := press(t, newModel(), "6", "1", tea.KeyEnter)
	if m.screen != confirmScreen {
		t.Fatal("confirmation not shown")
	}
	m = press(t, m, "n")
	snapshot(t, "after cancel", m)
	if m.screen != listScreen {
		t.Fatal("cancel did not return to the list")
	}
	if n, _ := functions.Count(); n != 1 {
		t.Fatalf("Count = %d after cancel, want 1", n)
	}
}
