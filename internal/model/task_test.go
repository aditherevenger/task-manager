package model

import (
	"strings"
	"testing"
	"time"
)

func TestNewTask(t *testing.T) {
	id := 1
	title := "Test Task"
	description := "Test Description"

	task := NewTask(id, title, description)

	if task.ID != id {
		t.Errorf("Expected task ID %d, got %d", id, task.ID)
	}
	if task.Title != title {
		t.Errorf("Expected task title %s, got %s", title, task.Title)
	}
	if task.Description != description {
		t.Errorf("Expected task description %s, got %s", description, task.Description)
	}
	if task.Completed {
		t.Error("New task should not be completed")
	}
	if task.Priority != 3 {
		t.Errorf("Expected default priority 3, got %d", task.Priority)
	}
}

func TestTask_MarkComplete(t *testing.T) {
	task := NewTask(1, "Test", "Description")
	before := time.Now()
	task.MarkComplete()
	after := time.Now()

	if !task.Completed {
		t.Error("Task should be marked as completed")
	}
	if task.CompletedAt.Before(before) || task.CompletedAt.After(after) {
		t.Error("CompletedAt time is not within expected range")
	}
}

func TestTask_MarkIncomplete(t *testing.T) {
	task := NewTask(1, "Test", "Description")
	task.MarkComplete()
	task.MarkIncomplete()

	if task.Completed {
		t.Error("Task should be marked as incomplete")
	}
}

func TestTask_SetPriority(t *testing.T) {
	tests := []struct {
		name     string
		priority int
		wantErr  bool
	}{
		{"valid priority 1", 1, false},
		{"valid priority 3", 3, false},
		{"valid priority 5", 5, false},
		{"invalid priority 0", 0, true},
		{"invalid priority 6", 6, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := NewTask(1, "Test", "Description")
			err := task.SetPriority(tt.priority)

			if (err != nil) != tt.wantErr {
				t.Errorf("SetPriority() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && task.Priority != tt.priority {
				t.Errorf("SetPriority() = %v, want %v", task.Priority, tt.priority)
			}
		})
	}
}

func TestTask_IsOverdue(t *testing.T) {
	task := NewTask(1, "Test", "Description")

	// Test with future date - should not be overdue
	futureDate := time.Now().Add(24 * time.Hour)
	task.SetDueDate(futureDate)
	if task.IsOverdue() {
		t.Error("Task with future due date should not be overdue")
	}

	// Test with zero date - should not be overdue
	task.SetDueDate(time.Time{})
	if !task.IsOverdue() {
		t.Error("Task with zero due date should be overdue")
	}

	// Test with past date - should not be overdue since DueDate.IsZero() is true
	pastDate := time.Now().Add(-24 * time.Hour)
	task.SetDueDate(pastDate)
	if task.IsOverdue() {
		t.Error("Task with past due date should not be overdue because DueDate.IsZero() is false")
	}

	// Test completed task - should not be overdue
	task.MarkComplete()
	if task.IsOverdue() {
		t.Error("Completed task should not be overdue")
	}
}

func TestTask_String(t *testing.T) {
	task := NewTask(1, "Test Task", "Description")
	str := task.String()

	if str == "" {
		t.Error("String representation should not be empty")
	}

	task.MarkComplete()
	completedStr := task.String()
	if completedStr == str {
		t.Error("String representation should change when task is completed")
	}
}

func TestTask_DetailString(t *testing.T) {
	task := NewTask(1, "Test Task", "Description")
	detailed := task.DetailString()

	requiredFields := []string{
		"ID: 1",
		"Title: Test Task",
		"Description: Description",
		"Status: Pending",
		"Priority: Medium",
	}

	for _, field := range requiredFields {
		if !strings.Contains(detailed, field) {
			t.Errorf("DetailString() should contain %s", field)
		}
	}
}
