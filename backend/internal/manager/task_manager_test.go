package manager

import (
	"errors"
	"reflect"
	"sort"
	"task-manager/internal/model"
	"testing"
	"time"
)

// MockTaskStorage is a mock implementation of the TaskStorage interface for testing
type MockTaskStorage struct {
	tasks     []*model.Task
	failLoad  bool
	failSave  bool
	saveCalls int
	loadCalls int
}

func (m *MockTaskStorage) Save(tasks []*model.Task) error {
	m.saveCalls++
	if m.failSave {
		return errors.New("save error")
	}
	m.tasks = tasks
	return nil
}

func (m *MockTaskStorage) Load() ([]*model.Task, error) {
	m.loadCalls++
	if m.failLoad {
		return nil, errors.New("load error")
	}
	return m.tasks, nil
}

func TestNewTaskManager(t *testing.T) {
	tests := []struct {
		name        string
		storageData []*model.Task
		failLoad    bool
		wantErr     bool
		wantNextID  int
	}{
		{
			name:        "success with empty storage",
			storageData: []*model.Task{},
			failLoad:    false,
			wantErr:     false,
			wantNextID:  1,
		},
		{
			name: "success with existing tasks",
			storageData: []*model.Task{
				model.NewTask(1, "Task 1", "Description 1"),
				model.NewTask(3, "Task 3", "Description 3"),
			},
			failLoad:   false,
			wantErr:    false,
			wantNextID: 4, // Next ID should be highest ID + 1
		},
		{
			name:        "error loading tasks",
			storageData: nil,
			failLoad:    true,
			wantErr:     true,
			wantNextID:  0, // Not relevant as we expect an error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &MockTaskStorage{
				tasks:    tt.storageData,
				failLoad: tt.failLoad,
			}

			manager, err := NewTaskManager(mockStorage)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTaskManager() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If we wanted an error, we don't need to check anything else
			if tt.wantErr {
				return
			}

			// Check that Load was called
			if mockStorage.loadCalls != 1 {
				t.Errorf("Expected Load to be called once, got %d", mockStorage.loadCalls)
			}

			// Check that the tasks were loaded correctly
			if len(manager.tasks) != len(tt.storageData) {
				t.Errorf("Expected %d tasks, got %d", len(tt.storageData), len(manager.tasks))
			}

			// Check that the next ID was set correctly
			if manager.nextId != tt.wantNextID {
				t.Errorf("Expected nextId to be %d, got %d", tt.wantNextID, manager.nextId)
			}
		})
	}
}

func TestTaskManager_LoadTasks(t *testing.T) {
	tests := []struct {
		name        string
		storageData []*model.Task
		failLoad    bool
		wantErr     bool
		wantNextID  int
	}{
		{
			name:        "success with empty storage",
			storageData: []*model.Task{},
			failLoad:    false,
			wantErr:     false,
			wantNextID:  1,
		},
		{
			name: "success with existing tasks",
			storageData: []*model.Task{
				model.NewTask(1, "Task 1", "Description 1"),
				model.NewTask(5, "Task 5", "Description 5"),
			},
			failLoad:   false,
			wantErr:    false,
			wantNextID: 6, // Next ID should be highest ID + 1
		},
		{
			name:        "error loading tasks",
			storageData: nil,
			failLoad:    true,
			wantErr:     true,
			wantNextID:  0, // Not relevant as we expect an error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &MockTaskStorage{
				tasks:    tt.storageData,
				failLoad: tt.failLoad,
			}

			manager := &TaskManager{
				storage: mockStorage,
				tasks:   []*model.Task{},
				nextId:  1,
			}

			err := manager.LoadTasks()

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadTasks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If we wanted an error, we don't need to check anything else
			if tt.wantErr {
				return
			}

			// Check that the tasks were loaded correctly
			if len(manager.tasks) != len(tt.storageData) {
				t.Errorf("Expected %d tasks, got %d", len(tt.storageData), len(manager.tasks))
			}

			// Check that the next ID was set correctly
			if manager.nextId != tt.wantNextID {
				t.Errorf("Expected nextId to be %d, got %d", tt.wantNextID, manager.nextId)
			}
		})
	}
}

func TestTaskManager_SaveTasks(t *testing.T) {
	tests := []struct {
		name     string
		tasks    []*model.Task
		failSave bool
		wantErr  bool
	}{
		{
			name:     "success with empty tasks",
			tasks:    []*model.Task{},
			failSave: false,
			wantErr:  false,
		},
		{
			name: "success with tasks",
			tasks: []*model.Task{
				model.NewTask(1, "Task 1", "Description 1"),
				model.NewTask(2, "Task 2", "Description 2"),
			},
			failSave: false,
			wantErr:  false,
		},
		{
			name:     "error saving tasks",
			tasks:    []*model.Task{model.NewTask(1, "Task 1", "Description 1")},
			failSave: true,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &MockTaskStorage{
				failSave: tt.failSave,
			}

			manager := &TaskManager{
				storage: mockStorage,
				tasks:   tt.tasks,
				nextId:  1,
			}

			err := manager.SaveTasks()

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveTasks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check that Save was called
			if mockStorage.saveCalls != 1 {
				t.Errorf("Expected Save to be called once, got %d", mockStorage.saveCalls)
			}

			// If there was no error, check that the tasks were saved correctly
			if !tt.wantErr && !reflect.DeepEqual(mockStorage.tasks, tt.tasks) {
				t.Errorf("Tasks were not saved correctly")
			}
		})
	}
}

func TestTaskManager_AddTask(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		description string
		nextId      int
		failSave    bool
		wantErr     bool
	}{
		{
			name:        "success",
			title:       "New Task",
			description: "Description",
			nextId:      1,
			failSave:    false,
			wantErr:     false,
		},
		{
			name:        "empty title",
			title:       "",
			description: "Description",
			nextId:      1,
			failSave:    false,
			wantErr:     true,
		},
		{
			name:        "save error",
			title:       "New Task",
			description: "Description",
			nextId:      1,
			failSave:    true,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &MockTaskStorage{
				failSave: tt.failSave,
			}

			manager := &TaskManager{
				storage: mockStorage,
				tasks:   []*model.Task{},
				nextId:  tt.nextId,
			}

			task, err := manager.AddTask(tt.title, tt.description)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("AddTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If we expected an error, we don't need to check anything else
			if tt.wantErr {
				return
			}

			// Check that the task was added correctly
			if task.ID != tt.nextId {
				t.Errorf("Expected task ID %d, got %d", tt.nextId, task.ID)
			}
			if task.Title != tt.title {
				t.Errorf("Expected task title %s, got %s", tt.title, task.Title)
			}
			if task.Description != tt.description {
				t.Errorf("Expected task description %s, got %s", tt.description, task.Description)
			}

			// Check that the task was added to the manager's tasks
			if len(manager.tasks) != 1 {
				t.Errorf("Expected 1 task, got %d", len(manager.tasks))
			}

			// Check that nextId was incremented
			if manager.nextId != tt.nextId+1 {
				t.Errorf("Expected nextId to be %d, got %d", tt.nextId+1, manager.nextId)
			}

			// Check that Save was called
			if mockStorage.saveCalls != 1 {
				t.Errorf("Expected Save to be called once, got %d", mockStorage.saveCalls)
			}
		})
	}
}

func TestTaskManager_GetTask(t *testing.T) {
	tasks := []*model.Task{
		model.NewTask(1, "Task 1", "Description 1"),
		model.NewTask(2, "Task 2", "Description 2"),
	}

	tests := []struct {
		name    string
		id      int
		wantErr bool
	}{
		{
			name:    "existing task",
			id:      1,
			wantErr: false,
		},
		{
			name:    "non-existent task",
			id:      3,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := &TaskManager{
				tasks:  tasks,
				nextId: 3,
			}

			task, err := manager.GetTask(tt.id)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If we expected an error, we don't need to check anything else
			if tt.wantErr {
				return
			}

			// Check that the correct task was returned
			if task.ID != tt.id {
				t.Errorf("Expected task ID %d, got %d", tt.id, task.ID)
			}
		})
	}
}

func TestTaskManager_UpdateTask(t *testing.T) {
	tasks := []*model.Task{
		model.NewTask(1, "Task 1", "Description 1"),
		model.NewTask(2, "Task 2", "Description 2"),
	}

	tests := []struct {
		name           string
		id             int
		newTitle       string
		newDescription string
		failSave       bool
		wantErr        bool
	}{
		{
			name:           "update title and description",
			id:             1,
			newTitle:       "Updated Title",
			newDescription: "Updated Description",
			failSave:       false,
			wantErr:        false,
		},
		{
			name:           "update title only",
			id:             1,
			newTitle:       "Updated Title",
			newDescription: "",
			failSave:       false,
			wantErr:        false,
		},
		{
			name:           "update description only",
			id:             1,
			newTitle:       "",
			newDescription: "Updated Description",
			failSave:       false,
			wantErr:        false,
		},
		{
			name:           "non-existent task",
			id:             3,
			newTitle:       "Updated Title",
			newDescription: "Updated Description",
			failSave:       false,
			wantErr:        true,
		},
		{
			name:           "save error",
			id:             1,
			newTitle:       "Updated Title",
			newDescription: "Updated Description",
			failSave:       true,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh copy of tasks for each test
			tasksCopy := make([]*model.Task, len(tasks))
			for i, task := range tasks {
				taskCopy := *task // Create a copy of the task
				tasksCopy[i] = &taskCopy
			}

			mockStorage := &MockTaskStorage{
				failSave: tt.failSave,
			}

			manager := &TaskManager{
				storage: mockStorage,
				tasks:   tasksCopy,
				nextId:  3,
			}

			err := manager.UpdateTask(tt.id, tt.newTitle, tt.newDescription)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If we expected an error, we don't need to check anything else
			if tt.wantErr {
				return
			}

			// Get the updated task
			task, err := manager.GetTask(tt.id)
			if err != nil {
				t.Errorf("Unexpected error getting task: %v", err)
				return
			}

			// Check that the task was updated correctly
			if tt.newTitle != "" && task.Title != tt.newTitle {
				t.Errorf("Expected task title %s, got %s", tt.newTitle, task.Title)
			}
			if tt.newDescription != "" && task.Description != tt.newDescription {
				t.Errorf("Expected task description %s, got %s", tt.newDescription, task.Description)
			}

			// Check that Save was called
			if mockStorage.saveCalls != 1 {
				t.Errorf("Expected Save to be called once, got %d", mockStorage.saveCalls)
			}
		})
	}
}

func TestTaskManager_DeleteTask(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		failSave bool
		wantErr  bool
	}{
		{
			name:     "existing task",
			id:       1,
			failSave: false,
			wantErr:  false,
		},
		{
			name:     "non-existent task",
			id:       3,
			failSave: false,
			wantErr:  true,
		},
		{
			name:     "save error",
			id:       1,
			failSave: true,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh set of tasks for each test
			tasks := []*model.Task{
				model.NewTask(1, "Task 1", "Description 1"),
				model.NewTask(2, "Task 2", "Description 2"),
			}

			mockStorage := &MockTaskStorage{
				failSave: tt.failSave,
			}

			manager := &TaskManager{
				storage: mockStorage,
				tasks:   tasks,
				nextId:  3,
			}

			originalLen := len(manager.tasks)
			err := manager.DeleteTask(tt.id)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If we expected an error, we don't need to check anything else
			if tt.wantErr {
				return
			}

			// Check that the task was deleted
			if len(manager.tasks) != originalLen-1 {
				t.Errorf("Expected %d tasks, got %d", originalLen-1, len(manager.tasks))
			}

			// Check that the deleted task is no longer present
			_, err = manager.GetTask(tt.id)
			if err == nil {
				t.Errorf("Expected error getting deleted task, got nil")
			}

			// Check that Save was called
			if mockStorage.saveCalls != 1 {
				t.Errorf("Expected Save to be called once, got %d", mockStorage.saveCalls)
			}
		})
	}
}

func TestTaskManager_MarkTaskComplete(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		failSave bool
		wantErr  bool
	}{
		{
			name:     "existing task",
			id:       1,
			failSave: false,
			wantErr:  false,
		},
		{
			name:     "non-existent task",
			id:       3,
			failSave: false,
			wantErr:  true,
		},
		{
			name:     "save error",
			id:       1,
			failSave: true,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh set of tasks for each test
			tasks := []*model.Task{
				model.NewTask(1, "Task 1", "Description 1"),
				model.NewTask(2, "Task 2", "Description 2"),
			}

			mockStorage := &MockTaskStorage{
				failSave: tt.failSave,
			}

			manager := &TaskManager{
				storage: mockStorage,
				tasks:   tasks,
				nextId:  3,
			}

			err := manager.MarkTaskComplete(tt.id)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("MarkTaskComplete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If we expected an error, we don't need to check anything else
			if tt.wantErr {
				return
			}

			// Check that the task was marked as complete
			task, err := manager.GetTask(tt.id)
			if err != nil {
				t.Errorf("Unexpected error getting task: %v", err)
				return
			}
			if !task.Completed {
				t.Error("Task should be marked as completed")
			}

			// Check that Save was called
			if mockStorage.saveCalls != 1 {
				t.Errorf("Expected Save to be called once, got %d", mockStorage.saveCalls)
			}
		})
	}
}

func TestTaskManager_MarkTaskIncomplete(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		failSave bool
		wantErr  bool
	}{
		{
			name:     "existing task",
			id:       1,
			failSave: false,
			wantErr:  false,
		},
		{
			name:     "non-existent task",
			id:       3,
			failSave: false,
			wantErr:  true,
		},
		{
			name:     "save error",
			id:       1,
			failSave: true,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh set of tasks for each test
			tasks := []*model.Task{
				model.NewTask(1, "Task 1", "Description 1"),
				model.NewTask(2, "Task 2", "Description 2"),
			}
			// Mark the first task as complete
			tasks[0].MarkComplete()

			mockStorage := &MockTaskStorage{
				failSave: tt.failSave,
			}

			manager := &TaskManager{
				storage: mockStorage,
				tasks:   tasks,
				nextId:  3,
			}

			err := manager.MarkTaskIncomplete(tt.id)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("MarkTaskIncomplete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If we expected an error, we don't need to check anything else
			if tt.wantErr {
				return
			}

			// Check that the task was marked as incomplete
			task, err := manager.GetTask(tt.id)
			if err != nil {
				t.Errorf("Unexpected error getting task: %v", err)
				return
			}
			if task.Completed {
				t.Error("Task should be marked as incomplete")
			}

			// Check that Save was called
			if mockStorage.saveCalls != 1 {
				t.Errorf("Expected Save to be called once, got %d", mockStorage.saveCalls)
			}
		})
	}
}

func TestTaskManager_SetTaskDueDate(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		dueDate  time.Time
		failSave bool
		wantErr  bool
	}{
		{
			name:     "existing task with future date",
			id:       1,
			dueDate:  time.Now().Add(24 * time.Hour),
			failSave: false,
			wantErr:  false,
		},
		{
			name:     "existing task with zero date",
			id:       1,
			dueDate:  time.Time{},
			failSave: false,
			wantErr:  false,
		},
		{
			name:     "non-existent task",
			id:       3,
			dueDate:  time.Now().Add(24 * time.Hour),
			failSave: false,
			wantErr:  true,
		},
		{
			name:     "save error",
			id:       1,
			dueDate:  time.Now().Add(24 * time.Hour),
			failSave: true,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh set of tasks for each test
			tasks := []*model.Task{
				model.NewTask(1, "Task 1", "Description 1"),
				model.NewTask(2, "Task 2", "Description 2"),
			}

			mockStorage := &MockTaskStorage{
				failSave: tt.failSave,
			}

			manager := &TaskManager{
				storage: mockStorage,
				tasks:   tasks,
				nextId:  3,
			}

			err := manager.SetTaskDueDate(tt.id, tt.dueDate)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("SetTaskDueDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If we expected an error, we don't need to check anything else
			if tt.wantErr {
				return
			}

			// Check that the due date was set correctly
			task, err := manager.GetTask(tt.id)
			if err != nil {
				t.Errorf("Unexpected error getting task: %v", err)
				return
			}
			if !task.DueDate.Equal(tt.dueDate) {
				t.Errorf("Expected due date %v, got %v", tt.dueDate, task.DueDate)
			}

			// Check that Save was called
			if mockStorage.saveCalls != 1 {
				t.Errorf("Expected Save to be called once, got %d", mockStorage.saveCalls)
			}
		})
	}
}

func TestTaskManager_SetTaskPriority(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		priority int
		failSave bool
		wantErr  bool
	}{
		{
			name:     "valid priority",
			id:       1,
			priority: 2,
			failSave: false,
			wantErr:  false,
		},
		{
			name:     "invalid priority",
			id:       1,
			priority: 6,
			failSave: false,
			wantErr:  true,
		},
		{
			name:     "non-existent task",
			id:       3,
			priority: 2,
			failSave: false,
			wantErr:  true,
		},
		{
			name:     "save error",
			id:       1,
			priority: 2,
			failSave: true,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh set of tasks for each test
			tasks := []*model.Task{
				model.NewTask(1, "Task 1", "Description 1"),
				model.NewTask(2, "Task 2", "Description 2"),
			}

			mockStorage := &MockTaskStorage{
				failSave: tt.failSave,
			}

			manager := &TaskManager{
				storage: mockStorage,
				tasks:   tasks,
				nextId:  3,
			}

			err := manager.SetTaskPriority(tt.id, tt.priority)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("SetTaskPriority() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If we expected an error, we don't need to check anything else
			if tt.wantErr {
				return
			}

			// Check that the priority was set correctly
			task, err := manager.GetTask(tt.id)
			if err != nil {
				t.Errorf("Unexpected error getting task: %v", err)
				return
			}
			if task.Priority != tt.priority {
				t.Errorf("Expected priority %d, got %d", tt.priority, task.Priority)
			}

			// Check that Save was called
			if mockStorage.saveCalls != 1 {
				t.Errorf("Expected Save to be called once, got %d", mockStorage.saveCalls)
			}
		})
	}
}

func TestTaskManager_GetAllTasks(t *testing.T) {
	tasks := []*model.Task{
		model.NewTask(1, "Task 1", "Description 1"),
		model.NewTask(2, "Task 2", "Description 2"),
	}

	manager := &TaskManager{
		tasks:  tasks,
		nextId: 3,
	}

	returnedTasks := manager.GetAllTasks()

	// Check that the correct tasks were returned
	if len(returnedTasks) != len(tasks) {
		t.Errorf("Expected %d tasks, got %d", len(tasks), len(returnedTasks))
	}

	// Check that the returned tasks are the same as the original tasks
	for i, task := range tasks {
		if task != returnedTasks[i] {
			t.Errorf("Task at index %d does not match", i)
		}
	}
}

func TestTaskManager_GetCompletedTasks(t *testing.T) {
	tasks := []*model.Task{
		model.NewTask(1, "Task 1", "Description 1"),
		model.NewTask(2, "Task 2", "Description 2"),
		model.NewTask(3, "Task 3", "Description 3"),
	}

	// Mark some tasks as completed
	tasks[0].MarkComplete()
	tasks[2].MarkComplete()

	manager := &TaskManager{
		tasks:  tasks,
		nextId: 4,
	}

	completedTasks := manager.GetCompletedTasks()

	// Check that the correct number of completed tasks were returned
	if len(completedTasks) != 2 {
		t.Errorf("Expected 2 completed tasks, got %d", len(completedTasks))
	}

	// Check that all returned tasks are completed
	for _, task := range completedTasks {
		if !task.Completed {
			t.Errorf("Task %d should be completed", task.ID)
		}
	}

	// Check that the specific completed tasks were returned
	completedIDs := []int{}
	for _, task := range completedTasks {
		completedIDs = append(completedIDs, task.ID)
	}
	sort.Ints(completedIDs)
	expectedIDs := []int{1, 3}

	if !reflect.DeepEqual(completedIDs, expectedIDs) {
		t.Errorf("Expected completed task IDs %v, got %v", expectedIDs, completedIDs)
	}
}

func TestTaskManager_GetPendingTasks(t *testing.T) {
	tasks := []*model.Task{
		model.NewTask(1, "Task 1", "Description 1"),
		model.NewTask(2, "Task 2", "Description 2"),
		model.NewTask(3, "Task 3", "Description 3"),
	}

	// Mark some tasks as completed
	tasks[0].MarkComplete()
	tasks[2].MarkComplete()

	manager := &TaskManager{
		tasks:  tasks,
		nextId: 4,
	}

	pendingTasks := manager.GetPendingTasks()

	// Check that the correct number of pending tasks were returned
	if len(pendingTasks) != 1 {
		t.Errorf("Expected 1 pending task, got %d", len(pendingTasks))
	}

	// Check that all returned tasks are pending
	for _, task := range pendingTasks {
		if task.Completed {
			t.Errorf("Task %d should be pending", task.ID)
		}
	}

	// Check that the specific pending task was returned
	if pendingTasks[0].ID != 2 {
		t.Errorf("Expected pending task ID 2, got %d", pendingTasks[0].ID)
	}
}

func TestTaskManager_GetOverdueTasks(t *testing.T) {
	tasks := []*model.Task{
		model.NewTask(1, "Task 1", "Description 1"),
		model.NewTask(2, "Task 2", "Description 2"),
		model.NewTask(3, "Task 3", "Description 3"),
		model.NewTask(4, "Task 4", "Description 4"),
	}

	// Set up tasks with various conditions
	// Task 1: No due date - not overdue
	// Task 2: Due date in the future - not overdue
	tasks[1].SetDueDate(time.Now().Add(24 * time.Hour))
	// Task 3: Due date in the past - overdue
	tasks[2].SetDueDate(time.Now().Add(-24 * time.Hour))
	// Task 4: Due date in the past but completed - not overdue
	tasks[3].SetDueDate(time.Now().Add(-24 * time.Hour))
	tasks[3].MarkComplete()

	manager := &TaskManager{
		tasks:  tasks,
		nextId: 5,
	}

	overdueTasks := manager.GetOverdueTasks()

	// According to the IsOverdue() implementation, a task is overdue if:
	// it has a non-zero due date in the past and is not completed
	expectedCount := 1 // Only task 3 should be overdue
	if len(overdueTasks) != expectedCount {
		t.Errorf("Expected %d overdue tasks, got %d", expectedCount, len(overdueTasks))
	}

	// Check that the specific overdue task was returned
	if len(overdueTasks) > 0 && overdueTasks[0].ID != 3 {
		t.Errorf("Expected overdue task ID 3, got %d", overdueTasks[0].ID)
	}
}

func TestTaskManager_SortTasksByPriority(t *testing.T) {
	tasks := []*model.Task{
		model.NewTask(1, "Medium Priority", "Description 1"), // Default priority 3
		model.NewTask(2, "Highest Priority", "Description 2"),
		model.NewTask(3, "Lowest Priority", "Description 3"),
	}

	_ = tasks[1].SetPriority(1) // Highest priority
	_ = tasks[2].SetPriority(5) // Lowest priority

	manager := &TaskManager{
		tasks:  tasks,
		nextId: 4,
	}

	sortedTasks := manager.SortTasksByPriority(tasks)

	// The implementation sorts by priority from lowest to highest numeric value
	// Lower priority value = higher importance (1 is highest priority)
	expectedOrder := []int{2, 1, 3} // IDs in expected sorted order

	if len(sortedTasks) != len(tasks) {
		t.Errorf("Expected %d tasks, got %d", len(tasks), len(sortedTasks))
	}

	for i, expectedID := range expectedOrder {
		if sortedTasks[i].ID != expectedID {
			t.Errorf("Expected task with ID %d at position %d, got task with ID %d",
				expectedID, i, sortedTasks[i].ID)
		}
	}
}

func TestTaskManager_SortTasksByDueDate(t *testing.T) {
	tasks := []*model.Task{
		model.NewTask(1, "No Due Date", "Description 1"),
		model.NewTask(2, "Future Due Date", "Description 2"),
		model.NewTask(3, "Past Due Date", "Description 3"),
	}

	// Set due dates
	tasks[1].SetDueDate(time.Now().Add(24 * time.Hour))
	tasks[2].SetDueDate(time.Now().Add(-24 * time.Hour))

	manager := &TaskManager{
		tasks:  tasks,
		nextId: 4,
	}

	sortedTasks := manager.SortTasksByDueDate(tasks)

	// Expected order: past due, future due, no due (tasks without due dates go last)
	expectedOrder := []int{3, 2, 1} // IDs in expected sorted order

	if len(sortedTasks) != len(tasks) {
		t.Errorf("Expected %d tasks, got %d", len(tasks), len(sortedTasks))
	}

	for i, expectedID := range expectedOrder {
		if sortedTasks[i].ID != expectedID {
			t.Errorf("Expected task with ID %d at position %d, got task with ID %d",
				expectedID, i, sortedTasks[i].ID)
		}
	}
}

func TestTaskManager_GetTaskStats(t *testing.T) {
	tasks := []*model.Task{
		model.NewTask(1, "Task 1", "Description 1"),
		model.NewTask(2, "Task 2", "Description 2"),
		model.NewTask(3, "Task 3", "Description 3"),
		model.NewTask(4, "Task 4", "Description 4"),
	}

	// Mark some tasks as completed
	tasks[0].MarkComplete()
	tasks[2].MarkComplete()

	// Set due dates for tasks
	// Task 1: Completed with past due date - not overdue
	tasks[0].SetDueDate(time.Now().Add(-24 * time.Hour))
	// Task 2: Past due date and not completed - overdue
	tasks[1].SetDueDate(time.Now().Add(-24 * time.Hour))
	// Task 3: Completed with past due date - not overdue
	tasks[2].SetDueDate(time.Now().Add(-24 * time.Hour))
	// Task 4: Past due date and not completed - overdue
	tasks[3].SetDueDate(time.Now().Add(-24 * time.Hour))

	manager := &TaskManager{
		tasks:  tasks,
		nextId: 5,
	}

	stats := manager.GetTaskStats()

	// Check that all expected stats are present
	expectedStats := map[string]int{
		"Total tasks":     4,
		"Completed tasks": 2,
		"Pending":         2,
		"Overdue":         2, // Tasks 2 and 4 are overdue (past due date and not completed)
	}

	for key, expectedValue := range expectedStats {
		actualValue, ok := stats[key]
		if !ok {
			t.Errorf("Expected stat '%s' is missing", key)
		} else if actualValue != expectedValue {
			t.Errorf("Expected '%s' to be %d, got %d", key, expectedValue, actualValue)
		}
	}

	// Check for unexpected stats
	for key := range stats {
		_, ok := expectedStats[key]
		if !ok {
			t.Errorf("Unexpected stat '%s'", key)
		}
	}
}
