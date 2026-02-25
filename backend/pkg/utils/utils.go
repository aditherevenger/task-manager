package utils

import (
	"fmt"
	"strconv"
	"strings"
	"task-manager/internal/model"
	"time"
)

// ParseDueDate parses a date string in the format yyyy-mm-dd
func ParseDueDate(dateStr string) (time.Time, error) {
	//check if the date string is empty
	if dateStr == "" {
		return time.Time{}, nil // Return zero value for time if empty
	}

	//Parse the date string
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format, please use YYYY-MM-DD %w", err)
	}
	return t, nil
}

// ParsePriority parses a priority string to an integer (1-5)
func ParsePriority(priorityStr string) (int, error) {
	// Check if the priority string is empty
	if priorityStr == "" {
		return 0, nil // Return 0 for empty priority
	}

	// Try to parse as text first
	priorityStr = strings.ToLower(strings.TrimSpace(priorityStr))
	switch priorityStr {
	case "highest":
		return 1, nil
	case "high":
		return 2, nil
	case "medium":
		return 3, nil
	case "low":
		return 4, nil
	case "lowest":
		return 5, nil
	}

	// Try to parse as integer
	priority, err := strconv.Atoi(priorityStr)
	if err != nil {
		return 0, fmt.Errorf("invalid priority format, please use a number between 1 and 5 or one of the following: highest, high, medium, low, lowest")
	}

	// Validate the priority range
	if priority < 1 || priority > 5 {
		return 0, fmt.Errorf("invalid priority, it should be between 1 and 5")
	}

	return priority, nil
}

// FormatTaskList formats a list of tasks for display
func FormatTaskList(tasks []*model.Task, showDetails bool) string {
	if len(tasks) == 0 {
		return "No tasks available."
	}

	var sb strings.Builder
	for i, task := range tasks {
		if i > 0 {
			sb.WriteString("\n")
			if showDetails {
				sb.WriteString("\n")
			}
		}
		if showDetails {
			sb.WriteString(task.DetailString())
		} else {
			sb.WriteString(task.String())
		}
	}
	return sb.String()
}
