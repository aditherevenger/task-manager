package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"task-manager/internal/model"
	"testing"
	"time"
)

// MockStorage implements TaskStorage interface for testing
type MockStorage struct {
	tasks []*model.Task
	err   error
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		tasks: make([]*model.Task, 0),
	}
}

func (m *MockStorage) Save(tasks []*model.Task) error {
	if m.err != nil {
		return m.err
	}
	m.tasks = tasks
	return nil
}

func (m *MockStorage) Load() ([]*model.Task, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.tasks, nil
}

func TestMockStorage_Save(t *testing.T) {
	storage := NewMockStorage()
	tasks := []*model.Task{
		model.NewTask(1, "Task 1", "Description 1"),
		model.NewTask(2, "Task 2", "Description 2"),
	}

	err := storage.Save(tasks)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	savedTasks, err := storage.Load()
	if err != nil {
		t.Errorf("Unexpected error on load: %v", err)
	}

	if len(savedTasks) != len(tasks) {
		t.Errorf("Expected %d tasks, got %d", len(tasks), len(savedTasks))
	}
}

func TestMockStorage_LoadError(t *testing.T) {
	storage := NewMockStorage()
	storage.err = fmt.Errorf("mock error")

	_, err := storage.Load()
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestMockStorage_SaveError(t *testing.T) {
	storage := NewMockStorage()
	storage.err = fmt.Errorf("mock error")

	err := storage.Save([]*model.Task{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// JSONStorage Tests

func TestNewJSONStorage(t *testing.T) {
	filePath := "test.json"
	storage := NewJSONStorage(filePath)

	if storage == nil {
		t.Error("Expected non-nil JSONStorage")
	}

	if storage.filePath != filePath {
		t.Errorf("Expected filePath %s, got %s", filePath, storage.filePath)
	}
}

func TestJSONStorage_SaveAndLoad_EmptyTasks(t *testing.T) {
	// Create temporary file
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test_empty.json")
	storage := NewJSONStorage(filePath)

	// Test saving empty tasks
	err := storage.Save([]*model.Task{})
	if err != nil {
		t.Errorf("Unexpected error saving empty tasks: %v", err)
	}

	// Test loading empty tasks
	tasks, err := storage.Load()
	if err != nil {
		t.Errorf("Unexpected error loading empty tasks: %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks, got %d", len(tasks))
	}
}

func TestJSONStorage_SaveAndLoad_WithTasks(t *testing.T) {
	// Create temporary file
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test_tasks.json")
	storage := NewJSONStorage(filePath)

	// Create test tasks
	now := time.Now()
	originalTasks := []*model.Task{
		{
			ID:          1,
			Title:       "Task 1",
			Description: "Description 1",
			Completed:   false,
			CreatedAt:   now,
			Priority:    1,
		},
		{
			ID:          2,
			Title:       "Task 2",
			Description: "Description 2",
			Completed:   true,
			CreatedAt:   now.Add(time.Hour),
			CompletedAt: now.Add(2 * time.Hour),
			Priority:    3,
		},
	}

	// Test saving tasks
	err := storage.Save(originalTasks)
	if err != nil {
		t.Errorf("Unexpected error saving tasks: %v", err)
	}

	// Test loading tasks
	loadedTasks, err := storage.Load()
	if err != nil {
		t.Errorf("Unexpected error loading tasks: %v", err)
	}

	// Verify task count
	if len(loadedTasks) != len(originalTasks) {
		t.Errorf("Expected %d tasks, got %d", len(originalTasks), len(loadedTasks))
	}

	// Verify task details
	for i, task := range loadedTasks {
		original := originalTasks[i]
		if task.ID != original.ID {
			t.Errorf("Task %d: Expected ID %d, got %d", i, original.ID, task.ID)
		}
		if task.Title != original.Title {
			t.Errorf("Task %d: Expected title %s, got %s", i, original.Title, task.Title)
		}
		if task.Description != original.Description {
			t.Errorf("Task %d: Expected description %s, got %s", i, original.Description, task.Description)
		}
		if task.Completed != original.Completed {
			t.Errorf("Task %d: Expected completed %t, got %t", i, original.Completed, task.Completed)
		}
		if task.Priority != original.Priority {
			t.Errorf("Task %d: Expected priority %d, got %d", i, original.Priority, task.Priority)
		}
	}
}

func TestJSONStorage_Load_NonExistentFile(t *testing.T) {
	// Use a non-existent file path
	filePath := filepath.Join(t.TempDir(), "nonexistent.json")
	storage := NewJSONStorage(filePath)

	// Test loading from non-existent file
	tasks, err := storage.Load()
	if err != nil {
		t.Errorf("Expected no error for non-existent file, got: %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks for non-existent file, got %d", len(tasks))
	}
}

func TestJSONStorage_Load_EmptyFile(t *testing.T) {
	// Create temporary empty file
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "empty.json")
	
	// Create empty file
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}
	file.Close()

	storage := NewJSONStorage(filePath)

	// Test loading from empty file
	tasks, err := storage.Load()
	if err != nil {
		t.Errorf("Expected no error for empty file, got: %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks for empty file, got %d", len(tasks))
	}
}

func TestJSONStorage_Load_InvalidJSON(t *testing.T) {
	// Create temporary file with invalid JSON
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "invalid.json")
	
	// Write invalid JSON
	err := os.WriteFile(filePath, []byte("invalid json content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid JSON file: %v", err)
	}

	storage := NewJSONStorage(filePath)

	// Test loading from invalid JSON file
	_, err = storage.Load()
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestJSONStorage_Save_DirectoryCreation(t *testing.T) {
	// Create path with nested directories
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "nested", "dir", "test.json")
	storage := NewJSONStorage(filePath)

	// Test saving (should create directories)
	tasks := []*model.Task{
		model.NewTask(1, "Test Task", "Test Description"),
	}

	err := storage.Save(tasks)
	if err != nil {
		t.Errorf("Unexpected error saving to nested path: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("Expected file to be created, but it doesn't exist")
	}

	// Verify content can be loaded
	loadedTasks, err := storage.Load()
	if err != nil {
		t.Errorf("Unexpected error loading from created file: %v", err)
	}

	if len(loadedTasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(loadedTasks))
	}
}

func TestJSONStorage_Save_ReadOnlyDirectory(t *testing.T) {
	// Skip this test on Windows as it behaves differently with permissions
	if runtime.GOOS == "windows" {
		t.Skip("Skipping read-only directory test on Windows")
	}

	// Create read-only directory
	tempDir := t.TempDir()
	readOnlyDir := filepath.Join(tempDir, "readonly")
	err := os.Mkdir(readOnlyDir, 0444) // Read-only permissions
	if err != nil {
		t.Fatalf("Failed to create read-only directory: %v", err)
	}

	filePath := filepath.Join(readOnlyDir, "test.json")
	storage := NewJSONStorage(filePath)

	// Test saving to read-only directory (should fail)
	tasks := []*model.Task{
		model.NewTask(1, "Test Task", "Test Description"),
	}

	err = storage.Save(tasks)
	if err == nil {
		t.Error("Expected error saving to read-only directory, got nil")
	}
}

func TestJSONStorage_InterfaceCompliance(t *testing.T) {
	// Test that JSONStorage implements TaskStorage interface
	var _ TaskStorage = (*JSONStorage)(nil)
	var _ TaskStorage = NewJSONStorage("test.json")
}
