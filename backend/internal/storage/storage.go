package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"task-manager/internal/model"
)

type TaskStorage interface {

	// Save save saves all task to storage
	Save(tasks []*model.Task) error

	//Load loads all task from storage
	Load() ([]*model.Task, error)
}

// JSONStorage implements TaskStorage interface using JSON file storage
type JSONStorage struct {
	filePath string
}

// NewJSONStorage creates a new JSONStorage instance
func NewJSONStorage(filePath string) *JSONStorage {
	return &JSONStorage{
		filePath: filePath,
	}
}

// Save saves all tasks to the JSON file
func (j *JSONStorage) Save(tasks []*model.Task) error {
	// Ensure the directory exists
	dir := filepath.Dir(j.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Marshal tasks to JSON
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(j.filePath, data, 0644)
}

// Load loads all tasks from the JSON file
func (j *JSONStorage) Load() ([]*model.Task, error) {
	// Check if file exists
	if _, err := os.Stat(j.filePath); os.IsNotExist(err) {
		// Return empty slice if file doesn't exist
		return []*model.Task{}, nil
	}

	// Read file
	data, err := os.ReadFile(j.filePath)
	if err != nil {
		return nil, err
	}

	// Handle empty file
	if len(data) == 0 {
		return []*model.Task{}, nil
	}

	// Unmarshal JSON
	var tasks []*model.Task
	err = json.Unmarshal(data, &tasks)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}
