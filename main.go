package main

import (
	"fmt"
	"os"
	"strings"

	"gotodo/functions"
	"gotodo/ui"
)

const usage = `Usage:
  todo                              open the interactive app
  todo add <description>            add a task
  todo update <id> <description>    change a task's description
  todo delete <id>                  delete a task
  todo mark-in-progress <id>        mark a task as in progress
  todo mark-done <id>               mark a task as done
  todo list [todo|in-progress|done] list tasks, optionally by status`

// fail prints msg to stderr and exits with a non-zero status.
func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func main() {

	// With no arguments, open the interactive app
	if len(os.Args) < 2 {
		if err := ui.Run(); err != nil {
			fail("Error running app: %v", err)
		}
		return
	}

	command := os.Args[1]

	switch command {
	case "add":
		if len(os.Args) < 3 {
			fail("Usage: todo add \"task description\"")
		}
		// Join the rest so unquoted multi-word descriptions are kept whole.
		task := strings.Join(os.Args[2:], " ")
		if err := functions.AddTask(task); err != nil {
			fail("Error adding task: %v", err)
		}
		fmt.Printf("Added task: %s\n", strings.TrimSpace(task))
	case "update":
		if len(os.Args) < 4 {
			fail("Usage: todo update <task_id> \"new task description\"")
		}
		taskID := os.Args[2]
		newDescription := strings.Join(os.Args[3:], " ")
		if strings.TrimSpace(newDescription) == "" {
			fail("Error updating task: description cannot be empty")
		}
		if err := functions.UpdateTask(taskID, newDescription, ""); err != nil {
			fail("Error updating task: %v", err)
		}
		fmt.Printf("Updated task %s to: %s\n", taskID, strings.TrimSpace(newDescription))
	case "delete":
		if len(os.Args) < 3 {
			fail("Usage: todo delete <task_id>")
		}
		taskID := os.Args[2]
		if err := functions.DeleteTaskById(taskID); err != nil {
			fail("Error deleting task: %v", err)
		}
		fmt.Printf("Deleted task with ID: %s\n", taskID)
	case "list":
		var option string
		if len(os.Args) > 2 {
			option = os.Args[2]
		}
		var tasks []functions.Task
		var err error
		switch status := functions.Status(option); status {
		case "":
			tasks, err = functions.GetTasks()
		case functions.StatusTodo, functions.StatusInProgress, functions.StatusDone:
			tasks, err = functions.GetTasksByStatus(status)
		default:
			fail("Unknown status %q. Use todo, in-progress or done.", option)
		}
		if err != nil {
			fail("Error retrieving tasks: %v", err)
		}
		if len(tasks) == 0 {
			fmt.Println("No tasks found.")
			return
		}
		fmt.Println("Tasks:")
		for _, task := range tasks {
			fmt.Printf("ID: %s, Description: %s, Status: %s, CreatedAt: %s, UpdatedAt: %s\n",
				task.Id, task.Description, task.Status, task.CreatedAt, task.UpdatedAt)
		}
	case "mark-done":
		if len(os.Args) < 3 {
			fail("Usage: todo mark-done <task_id>")
		}
		taskID := os.Args[2]
		if err := functions.UpdateTask(taskID, "", functions.StatusDone); err != nil {
			fail("Error marking task as done: %v", err)
		}
		fmt.Printf("Marked task %s as done\n", taskID)
	case "mark-in-progress":
		if len(os.Args) < 3 {
			fail("Usage: todo mark-in-progress <task_id>")
		}
		taskID := os.Args[2]
		if err := functions.UpdateTask(taskID, "", functions.StatusInProgress); err != nil {
			fail("Error marking task as in-progress: %v", err)
		}
		fmt.Printf("Marked task %s as in-progress\n", taskID)
	case "help", "-h", "--help":
		fmt.Println(usage)
	default:
		fail("Unknown command: %s\n\n%s", command, usage)
	}
}
