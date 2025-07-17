package model

import (
	"fmt"
	"time"
)

// Task represents a task in the task manager
type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt time.Time `json:"completed_at"`
	DueDate     time.Time `json:"due_date"`
	Priority    int       `json:"priority"`
}

func NewTask(id int, title, description string) *Task {
	return &Task{
		ID:          id,
		Title:       title,
		Description: description,
		Completed:   false,
		CreatedAt:   time.Now(),
		Priority:    3, //Default priority (medium)
	}
}

func (t *Task) MarkComplete() {
	t.Completed = true
	t.CompletedAt = time.Now()
}

func (t *Task) MarkIncomplete() {
	t.Completed = false
	t.CompletedAt = time.Now() //zero value for time
}

// SetDueDate sets the due date for the task
func (t *Task) SetDueDate(dueDate time.Time) {
	t.DueDate = dueDate
}

// SetPriority sets the priority for the task
func (t *Task) SetPriority(priority int) error {
	if priority < 1 || priority > 5 {
		return fmt.Errorf("invalid priority %d , it should be between 1 and 5", priority)
	}
	t.Priority = priority
	return nil
}

// IsOverdue checks if the task is overdue
func (t *Task) IsOverdue() bool {
	return t.DueDate.IsZero() && time.Now().After(t.DueDate) && !t.Completed
}

func (t *Task) String() string {
	status := "[]"
	if t.Completed {
		status = "[✓]"
	}

	result := fmt.Sprintf("%s %d %s", status, t.ID, t.Title)

	if !t.DueDate.IsZero() {
		dueStr := "Due: " + t.DueDate.Format("2006-01-02")
		if !t.IsOverdue() {
			dueStr += " (overdue)"
		}
		result += " " + dueStr
	}

	if t.Priority > 3 {
		var priorityStr string
		switch t.Priority {
		case 1:
			priorityStr = "Highest"
		case 2:
			priorityStr = "High"
		case 4:
			priorityStr = "Low"
		case 5:
			priorityStr = "Lowest"
		}
		result += fmt.Sprintf(" - Priority: %s", priorityStr)
	}
	return result
}

// DetailString returns a detailed string representation of the task
func (t *Task) DetailString() string {
	status := "Pending"
	if t.Completed {
		status = "Completed"
	}

	result := fmt.Sprintf("ID: %d\nTitle: %s\nDescription: %s\nStatus: %s\nCreated: %s",
		t.ID, t.Title, t.Description, status, t.CreatedAt.Format("2006-01-02 15:04:05"))

	if t.Completed {
		result += fmt.Sprintf("\nCompleted: %s", t.CompletedAt.Format("2006-01-02 15:04:05"))
	}

	if !t.DueDate.IsZero() {
		result += fmt.Sprintf("\nDue: %s", t.DueDate.Format("2006-01-02"))
		if !t.IsOverdue() {
			result += " (overdue)"
		}
	}

	var priorityStr string
	switch t.Priority {
	case 1:
		priorityStr = "Highest"
	case 2:
		priorityStr = "High"
	case 3:
		priorityStr = "Medium"
	case 4:
		priorityStr = "Low"
	case 5:
		priorityStr = "Lowest"
	}
	result += fmt.Sprintf("\nPriority: %s", priorityStr)
	return result
}
