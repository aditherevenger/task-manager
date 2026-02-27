package utils

import (
	"task-manager/internal/model"
	"testing"
	"time"
)

func TestParseDueDate(t *testing.T) {
	tests := []struct {
		name    string
		dateStr string
		want    time.Time
		wantErr bool
	}{
		{
			name:    "Valid date",
			dateStr: "2025-07-18",
			want:    time.Date(2025, 7, 18, 0, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "Empty date",
			dateStr: "",
			want:    time.Time{},
			wantErr: false,
		},
		{
			name:    "Invalid format",
			dateStr: "2025/07/18",
			want:    time.Time{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDueDate(tt.dateStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDueDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !got.Equal(tt.want) {
				t.Errorf("ParseDueDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParsePriority(t *testing.T) {
	tests := []struct {
		name        string
		priorityStr string
		want        int
		wantErr     bool
	}{
		{
			name:        "Valid numeric priority",
			priorityStr: "3",
			want:        3,
			wantErr:     false,
		},
		{
			name:        "Empty priority",
			priorityStr: "",
			want:        0,
			wantErr:     false,
		},
		{
			name:        "Invalid number",
			priorityStr: "6",
			want:        0,
			wantErr:     true,
		},
		{
			name:        "Text highest",
			priorityStr: "highest",
			want:        1,
			wantErr:     false,
		},
		{
			name:        "Text medium",
			priorityStr: "medium",
			want:        3,
			wantErr:     false,
		},
		{
			name:        "Text lowest",
			priorityStr: "lowest",
			want:        5,
			wantErr:     false,
		},
		{
			name:        "Invalid text",
			priorityStr: "invalid",
			want:        0,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePriority(tt.priorityStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePriority() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParsePriority() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatTaskList(t *testing.T) {
	task1 := &model.Task{
		ID:          1,
		Title:       "Test Task 1",
		Description: "Description 1",
		DueDate:     time.Date(2025, 7, 18, 0, 0, 0, 0, time.UTC),
		Priority:    1,
	}
	task2 := &model.Task{
		ID:          2,
		Title:       "Test Task 2",
		Description: "Description 2",
		DueDate:     time.Date(2025, 7, 19, 0, 0, 0, 0, time.UTC),
		Priority:    2,
	}

	tests := []struct {
		name        string
		tasks       []*model.Task
		showDetails bool
		want        string
	}{
		{
			name:        "Empty task list",
			tasks:       []*model.Task{},
			showDetails: false,
			want:        "No tasks available.",
		},
		{
			name:        "Single task without details",
			tasks:       []*model.Task{task1},
			showDetails: false,
			want:        task1.String(),
		},
		{
			name:        "Single task with details",
			tasks:       []*model.Task{task1},
			showDetails: true,
			want:        task1.DetailString(),
		},
		{
			name:        "Multiple tasks without details",
			tasks:       []*model.Task{task1, task2},
			showDetails: false,
			want:        task1.String() + "\n" + task2.String(),
		},
		{
			name:        "Multiple tasks with details",
			tasks:       []*model.Task{task1, task2},
			showDetails: true,
			want:        task1.DetailString() + "\n\n" + task2.DetailString(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTaskList(tt.tasks, tt.showDetails)
			if got != tt.want {
				t.Errorf("FormatTaskList() = %v, want %v", got, tt.want)
			}
		})
	}
}
