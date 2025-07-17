package storage

import (
	"fmt"
	"task-manager/internal/model"
	"testing"
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

func (m *MockStorage) save(tasks []*model.Task) error {
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

	err := storage.save(tasks)
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

	err := storage.save([]*model.Task{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}
