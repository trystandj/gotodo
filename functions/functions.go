package functions

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Status string

const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in-progress"
	StatusDone       Status = "done"
)

type Task struct {
	Id          string `json:"id"`
	Description string `json:"description"`
	Status      Status `json:"status"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// legacyTimeLayout is the format older versions wrote: local time with a
// literal "Z" suffix, which wrongly claims to be UTC.
const legacyTimeLayout = "2006-01-02T15:04:05Z"

// legacyFilename is where older versions stored tasks, relative to the
// current directory.
const legacyFilename = "./functions/task.json"

var filename = dataFile()

// dataFile picks the task file: $GOTODO_FILE if set, otherwise
// <user config dir>/gotodo/tasks.json so the app works from any directory.
func dataFile() string {
	if f := os.Getenv("GOTODO_FILE"); f != "" {
		return f
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return legacyFilename
	}
	return filepath.Join(dir, "gotodo", "tasks.json")
}

// Now returns the current time in the format tasks are stored in.
func Now() string {
	return time.Now().Format(time.RFC3339)
}

// ParseTime parses a stored timestamp, accepting the legacy local-time format.
func ParseTime(s string) (time.Time, error) {
	if strings.HasSuffix(s, "Z") {
		return time.ParseInLocation(legacyTimeLayout, s, time.Local)
	}
	return time.Parse(time.RFC3339, s)
}

func MarshalJSON(tasks []Task) ([]byte, error) {
	jsonData, err := json.MarshalIndent(tasks, "", "  ") // Indented for readability
	if err != nil {
		return nil, fmt.Errorf("error marshaling JSON: %w", err)
	}
	return jsonData, nil
}
func ReadJSONFromFile() ([]Task, error) {
	var tasks []Task
	data, err := os.ReadFile(filename)
	if errors.Is(err, os.ErrNotExist) && filename != legacyFilename {
		// Fall back to the file older versions kept in the project folder;
		// it is copied to the new location on the next save.
		data, err = os.ReadFile(legacyFilename)
	}
	if errors.Is(err, os.ErrNotExist) {
		// No tasks saved yet
		return tasks, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error reading from file %s: %w", filename, err)
	}
	if len(data) > 0 {
		// Unmarshal existing data into the slice
		if err := json.Unmarshal(data, &tasks); err != nil {
			return nil, fmt.Errorf("error parsing %s: %w", filename, err)
		}
	}
	return tasks, nil
}

// WriteJSONToFile writes to a temporary file and renames it into place, so a
// crash mid-write cannot leave a half-written task file behind.
func WriteJSONToFile(jsonData []byte) error {
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return fmt.Errorf("error creating directory for %s: %w", filename, err)
	}
	tmp := filename + ".tmp"
	if err := os.WriteFile(tmp, jsonData, 0644); err != nil {
		return fmt.Errorf("error writing to file %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, filename); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("error replacing file %s: %w", filename, err)
	}
	return nil
}

func save(tasks []Task) error {
	jsonData, err := MarshalJSON(tasks)
	if err != nil {
		return err
	}
	return WriteJSONToFile(jsonData)
}

func Count() (int, error) {
	tasks, err := ReadJSONFromFile()
	if err != nil {
		return 0, err
	}
	return len(tasks), nil
}
func NextId(tasks []Task) string {
	highest := 0
	for _, task := range tasks {
		if n, err := strconv.Atoi(task.Id); err == nil && n > highest {
			highest = n
		}
	}
	return strconv.Itoa(highest + 1)
}
func AddTask(title string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return fmt.Errorf("description cannot be empty")
	}
	tasks, err := ReadJSONFromFile()
	if err != nil {
		return err
	}

	// Create a new task
	now := Now()
	newTask := Task{
		Id:          NextId(tasks),
		Description: title,
		Status:      StatusTodo,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return save(append(tasks, newTask))
}

// UpdateTask changes a task's description and/or status; pass "" to leave
// either one unchanged.
func UpdateTask(id string, description string, status Status) error {
	description = strings.TrimSpace(description)
	if id == "" {
		return fmt.Errorf("task ID cannot be empty")
	}
	if description == "" && status == "" {
		return fmt.Errorf("at least one of description or status must be provided for update")
	}
	tasks, err := ReadJSONFromFile()
	if err != nil {
		return err
	}

	found := false
	for i, task := range tasks {
		if task.Id == id {
			found = true
			if description != "" {
				tasks[i].Description = description
			}
			if status != "" {
				tasks[i].Status = status
			}
			tasks[i].UpdatedAt = Now()
			break
		}
	}
	if !found {
		return fmt.Errorf("task with id %s not found", id)
	}

	return save(tasks)
}
func GetTasks() ([]Task, error) {
	return ReadJSONFromFile()
}
func GetTaskById(id string) (*Task, error) {
	tasks, err := ReadJSONFromFile()
	if err != nil {
		return nil, err
	}

	for _, task := range tasks {
		if task.Id == id {
			return &task, nil
		}
	}
	return nil, fmt.Errorf("task with id %s not found", id)
}
func GetTasksByStatus(status Status) ([]Task, error) {
	tasks, err := ReadJSONFromFile()
	if err != nil {
		return nil, err
	}

	var filteredTasks []Task
	for _, task := range tasks {
		if task.Status == status {
			filteredTasks = append(filteredTasks, task)
		}
	}
	return filteredTasks, nil
}
func DeleteTaskById(id string) error {
	tasks, err := ReadJSONFromFile()
	if err != nil {
		return err
	}

	updatedTasks := []Task{}
	for _, task := range tasks {
		if task.Id != id {
			updatedTasks = append(updatedTasks, task)
		}
	}
	if len(updatedTasks) == len(tasks) {
		return fmt.Errorf("task with id %s not found", id)
	}

	return save(updatedTasks)
}
