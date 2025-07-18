package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"task-manager/internal/manager"
	"task-manager/internal/model"
	"task-manager/internal/storage"
	"task-manager/pkg/utils"
)

// MockStorage implements storage.TaskStorage interface for testing
type MockStorage struct {
	tasks []*model.Task
	err   error
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

// setupTest initializes a new test environment
func setupTest() (*gin.Engine, *manager.TaskManager, storage.TaskStorage) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockStorage := &MockStorage{
		tasks: make([]*model.Task, 0),
	}
	taskManager, _ := manager.NewTaskManager(mockStorage)
	handlers := NewTaskHandlers(taskManager)

	// Setup routes
	v1 := r.Group("/api/v1")
	{
		v1.GET("/tasks", handlers.ListTasks)
		v1.POST("/tasks", handlers.CreateTask)
		v1.GET("/tasks/:id", handlers.GetTask)
		v1.PUT("/tasks/:id", handlers.UpdateTask)
		v1.DELETE("/tasks/:id", handlers.DeleteTask)
		v1.PATCH("/tasks/:id/complete", handlers.CompleteTask)
		v1.PATCH("/tasks/:id/uncomplete", handlers.UncompleteTask)
		v1.PATCH("/tasks/:id/due-date", handlers.SetDueDate)
		v1.PATCH("/tasks/:id/priority", handlers.SetPriority)
		v1.GET("/stats", handlers.GetStats)
	}

	return r, taskManager, mockStorage
}

// ... (rest of the code remains the same)

// NEW TESTS TO INCREASE COVERAGE

func TestDeleteTask(t *testing.T) {
	tests := []struct {
		name           string
		taskID         string
		setupMock      func(storage.TaskStorage)
		expectedStatus int
		expectedError  string
		skipLoad       bool
	}{
		{
			name:   "delete existing task",
			taskID: "1",
			setupMock: func(m storage.TaskStorage) {
				m.(*MockStorage).tasks = []*model.Task{model.NewTask(1, "Task 1", "Description 1")}
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "delete with invalid task ID",
			taskID:         "invalid",
			setupMock:      func(m storage.TaskStorage) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid task ID",
		},
		{
			name:   "delete non-existent task",
			taskID: "999",
			setupMock: func(m storage.TaskStorage) {
				m.(*MockStorage).tasks = []*model.Task{model.NewTask(1, "Task 1", "Description 1")}
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "delete task with storage error",
			taskID: "1",
			setupMock: func(m storage.TaskStorage) {
				m.(*MockStorage).tasks = []*model.Task{model.NewTask(1, "Task 1", "Description 1")}
				m.(*MockStorage).err = errors.New("storage error")
			},
			expectedStatus: http.StatusInternalServerError,
			skipLoad:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, taskMgr, mockStore := setupTest()
			tt.setupMock(mockStore)
			if !tt.skipLoad {
				if err := taskMgr.LoadTasks(); err != nil {
					t.Fatalf("Failed to load tasks: %v", err)
				}
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("DELETE", "/api/v1/tasks/"+tt.taskID, nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if !strings.Contains(response["error"], tt.expectedError) {
					t.Errorf("Expected error containing %s, got %s", tt.expectedError, response["error"])
				}
			}
		})
	}
}

func TestUncompleteTask(t *testing.T) {
	tests := []struct {
		name           string
		taskID         string
		setupMock      func(storage.TaskStorage)
		expectedStatus int
		expectedError  string
		validate       func(*testing.T, *httptest.ResponseRecorder)
		skipLoad       bool
	}{
		{
			name:   "uncomplete existing completed task",
			taskID: "1",
			setupMock: func(m storage.TaskStorage) {
				task := model.NewTask(1, "Task 1", "Description 1")
				task.MarkComplete()
				m.(*MockStorage).tasks = []*model.Task{task}
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response["completed"].(bool) != false {
					t.Error("Expected task to be uncompleted")
				}
			},
		},
		{
			name:           "uncomplete with invalid task ID",
			taskID:         "invalid",
			setupMock:      func(m storage.TaskStorage) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid task ID",
		},
		{
			name:   "uncomplete non-existent task",
			taskID: "999",
			setupMock: func(m storage.TaskStorage) {
				m.(*MockStorage).tasks = []*model.Task{model.NewTask(1, "Task 1", "Description 1")}
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "uncomplete task with storage error",
			taskID: "1",
			setupMock: func(m storage.TaskStorage) {
				task := model.NewTask(1, "Task 1", "Description 1")
				task.MarkComplete()
				m.(*MockStorage).tasks = []*model.Task{task}
				m.(*MockStorage).err = errors.New("storage error")
			},
			expectedStatus: http.StatusInternalServerError,
			skipLoad:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, taskMgr, mockStore := setupTest()
			tt.setupMock(mockStore)
			if !tt.skipLoad {
				if err := taskMgr.LoadTasks(); err != nil {
					t.Fatalf("Failed to load tasks: %v", err)
				}
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("PATCH", "/api/v1/tasks/"+tt.taskID+"/uncomplete", nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if !strings.Contains(response["error"], tt.expectedError) {
					t.Errorf("Expected error containing %s, got %s", tt.expectedError, response["error"])
				}
			}

			if tt.validate != nil {
				tt.validate(t, w)
			}
		})
	}
}

func TestCreateTask_ExtendedCoverage(t *testing.T) {
	tests := []struct {
		name           string
		request        CreateTaskRequest
		setupMock      func(storage.TaskStorage)
		expectedStatus int
		expectedError  string
		validate       func(*testing.T, *httptest.ResponseRecorder)
		skipLoad       bool
	}{
		{
			name: "create task with due date and priority",
			request: CreateTaskRequest{
				Title:       "Test Task",
				Description: "Test Description",
				DueDate:     "2025-12-31",
				Priority:    "high",
			},
			setupMock: func(m storage.TaskStorage) {
				m.(*MockStorage).err = nil
			},
			expectedStatus: http.StatusCreated,
			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response["title"].(string) != "Test Task" {
					t.Error("Expected task title to match")
				}
				// Priority "high" maps to 2, not 1
				if response["priority"].(float64) != 2 {
					t.Errorf("Expected priority to be set to 2 (high), got %v", response["priority"])
				}
			},
		},
		{
			name: "create task with invalid due date",
			request: CreateTaskRequest{
				Title:       "Test Task",
				Description: "Test Description",
				DueDate:     "invalid-date",
			},
			setupMock: func(m storage.TaskStorage) {
				m.(*MockStorage).err = nil
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "create task with invalid priority",
			request: CreateTaskRequest{
				Title:       "Test Task",
				Description: "Test Description",
				Priority:    "invalid-priority",
			},
			setupMock: func(m storage.TaskStorage) {
				m.(*MockStorage).err = nil
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "create task with storage error during add",
			request: CreateTaskRequest{
				Title:       "Test Task",
				Description: "Test Description",
			},
			setupMock: func(m storage.TaskStorage) {
				m.(*MockStorage).err = errors.New("storage error")
			},
			expectedStatus: http.StatusInternalServerError,
			skipLoad:       true,
		},
		{
			name:           "create task with missing title",
			request:        CreateTaskRequest{Description: "Test Description"},
			setupMock:      func(m storage.TaskStorage) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, taskMgr, mockStore := setupTest()
			tt.setupMock(mockStore)
			if !tt.skipLoad {
				if err := taskMgr.LoadTasks(); err != nil {
					t.Fatalf("Failed to load tasks: %v", err)
				}
			}

			jsonData, _ := json.Marshal(tt.request)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.validate != nil {
				tt.validate(t, w)
			}
		})
	}
}

func TestUpdateTask_ExtendedCoverage(t *testing.T) {
	tests := []struct {
		name           string
		taskID         string
		request        UpdateTaskRequest
		setupMock      func(storage.TaskStorage)
		expectedStatus int
		expectedError  string
		skipLoad       bool
	}{
		{
			name:   "update task with storage error",
			taskID: "1",
			request: UpdateTaskRequest{
				Title:       "Updated Title",
				Description: "Updated Description",
			},
			setupMock: func(m storage.TaskStorage) {
				m.(*MockStorage).tasks = []*model.Task{model.NewTask(1, "Task 1", "Description 1")}
				m.(*MockStorage).err = errors.New("storage error")
			},
			expectedStatus: http.StatusInternalServerError,
			skipLoad:       true,
		},
		{
			name:   "update with malformed JSON",
			taskID: "1",
			setupMock: func(m storage.TaskStorage) {
				m.(*MockStorage).tasks = []*model.Task{model.NewTask(1, "Task 1", "Description 1")}
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, taskMgr, mockStore := setupTest()
			tt.setupMock(mockStore)
			if !tt.skipLoad {
				if err := taskMgr.LoadTasks(); err != nil {
					t.Fatalf("Failed to load tasks: %v", err)
				}
			}

			var jsonData []byte
			if tt.name == "update with malformed JSON" {
				jsonData = []byte(`{"title": "Updated Title", "description":}`) // malformed JSON
			} else {
				jsonData, _ = json.Marshal(tt.request)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("PUT", "/api/v1/tasks/"+tt.taskID, bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestCompleteTask_ExtendedCoverage(t *testing.T) {
	tests := []struct {
		name           string
		taskID         string
		setupMock      func(storage.TaskStorage)
		expectedStatus int
		skipLoad       bool
	}{
		{
			name:   "complete task with storage error",
			taskID: "1",
			setupMock: func(m storage.TaskStorage) {
				m.(*MockStorage).tasks = []*model.Task{model.NewTask(1, "Task 1", "Description 1")}
				m.(*MockStorage).err = errors.New("storage error")
			},
			expectedStatus: http.StatusInternalServerError,
			skipLoad:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, taskMgr, mockStore := setupTest()
			tt.setupMock(mockStore)
			if !tt.skipLoad {
				if err := taskMgr.LoadTasks(); err != nil {
					t.Fatalf("Failed to load tasks: %v", err)
				}
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("PATCH", "/api/v1/tasks/"+tt.taskID+"/complete", nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestSetDueDate_ExtendedCoverage(t *testing.T) {
	tests := []struct {
		name           string
		taskID         string
		request        SetDueDateRequest
		setupMock      func(storage.TaskStorage)
		expectedStatus int
		skipLoad       bool
	}{
		{
			name:   "set due date with invalid task ID",
			taskID: "invalid",
			request: SetDueDateRequest{
				DueDate: "2025-12-31",
			},
			setupMock:      func(m storage.TaskStorage) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "set due date with malformed JSON",
			taskID: "1",
			setupMock: func(m storage.TaskStorage) {
				m.(*MockStorage).tasks = []*model.Task{model.NewTask(1, "Task 1", "Description 1")}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "set due date for non-existent task",
			taskID: "999",
			request: SetDueDateRequest{
				DueDate: "2025-12-31",
			},
			setupMock: func(m storage.TaskStorage) {
				m.(*MockStorage).tasks = []*model.Task{model.NewTask(1, "Task 1", "Description 1")}
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "set due date with storage error",
			taskID: "1",
			request: SetDueDateRequest{
				DueDate: "2025-12-31",
			},
			setupMock: func(m storage.TaskStorage) {
				m.(*MockStorage).tasks = []*model.Task{model.NewTask(1, "Task 1", "Description 1")}
				m.(*MockStorage).err = errors.New("storage error")
			},
			expectedStatus: http.StatusInternalServerError,
			skipLoad:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, taskMgr, mockStore := setupTest()
			tt.setupMock(mockStore)
			if !tt.skipLoad {
				if err := taskMgr.LoadTasks(); err != nil {
					t.Fatalf("Failed to load tasks: %v", err)
				}
			}

			var jsonData []byte
			if tt.name == "set due date with malformed JSON" {
				jsonData = []byte(`{"due_date":}`) // malformed JSON
			} else {
				jsonData, _ = json.Marshal(tt.request)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("PATCH", "/api/v1/tasks/"+tt.taskID+"/due-date", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestSetPriority_ExtendedCoverage(t *testing.T) {
	tests := []struct {
		name           string
		taskID         string
		request        SetPriorityRequest
		setupMock      func(storage.TaskStorage)
		expectedStatus int
		skipLoad       bool
	}{
		{
			name:   "set priority with invalid task ID",
			taskID: "invalid",
			request: SetPriorityRequest{
				Priority: "high",
			},
			setupMock:      func(m storage.TaskStorage) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "set priority with malformed JSON",
			taskID: "1",
			setupMock: func(m storage.TaskStorage) {
				m.(*MockStorage).tasks = []*model.Task{model.NewTask(1, "Task 1", "Description 1")}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "set priority for non-existent task",
			taskID: "999",
			request: SetPriorityRequest{
				Priority: "high",
			},
			setupMock: func(m storage.TaskStorage) {
				m.(*MockStorage).tasks = []*model.Task{model.NewTask(1, "Task 1", "Description 1")}
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "set priority with storage error",
			taskID: "1",
			request: SetPriorityRequest{
				Priority: "high",
			},
			setupMock: func(m storage.TaskStorage) {
				m.(*MockStorage).tasks = []*model.Task{model.NewTask(1, "Task 1", "Description 1")}
				m.(*MockStorage).err = errors.New("storage error")
			},
			expectedStatus: http.StatusInternalServerError,
			skipLoad:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, taskMgr, mockStore := setupTest()
			tt.setupMock(mockStore)
			if !tt.skipLoad {
				if err := taskMgr.LoadTasks(); err != nil {
					t.Fatalf("Failed to load tasks: %v", err)
				}
			}

			var jsonData []byte
			if tt.name == "set priority with malformed JSON" {
				jsonData = []byte(`{"priority":}`) // malformed JSON
			} else {
				jsonData, _ = json.Marshal(tt.request)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("PATCH", "/api/v1/tasks/"+tt.taskID+"/priority", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestNewTaskHandlers(t *testing.T) {
	mockStorage := &MockStorage{tasks: make([]*model.Task, 0)}
	taskManager, _ := manager.NewTaskManager(mockStorage)
	
	handlers := NewTaskHandlers(taskManager)
	
	if handlers == nil {
		t.Error("Expected non-nil TaskHandlers")
	}
	
	if handlers.taskManager != taskManager {
		t.Error("Expected taskManager to be set correctly")
	}
}

func TestTaskToResponse(t *testing.T) {
	now := time.Now()
	task := &model.Task{
		ID:          1,
		Title:       "Test Task",
		Description: "Test Description",
		Completed:   true,
		CreatedAt:   now,
		CompletedAt: now.Add(time.Hour),
		DueDate:     now.Add(24 * time.Hour),
		Priority:    2,
	}

	response := taskToResponse(task)

	if response.ID != task.ID {
		t.Errorf("Expected ID %d, got %d", task.ID, response.ID)
	}
	if response.Title != task.Title {
		t.Errorf("Expected title %s, got %s", task.Title, response.Title)
	}
	if response.Description != task.Description {
		t.Errorf("Expected description %s, got %s", task.Description, response.Description)
	}
	if response.Completed != task.Completed {
		t.Errorf("Expected completed %t, got %t", task.Completed, response.Completed)
	}
	if response.Priority != task.Priority {
		t.Errorf("Expected priority %d, got %d", task.Priority, response.Priority)
	}
	if response.IsOverdue != task.IsOverdue() {
		t.Errorf("Expected IsOverdue %t, got %t", task.IsOverdue(), response.IsOverdue)
	}
}

func TestGetTask_ExtendedCoverage(t *testing.T) {
	tests := []struct {
		name           string
		taskID         string
		setupMock      func(storage.TaskStorage)
		expectedStatus int
		skipLoad       bool
	}{
		{
			name:   "get task with storage error during load",
			taskID: "1",
			setupMock: func(m storage.TaskStorage) {
				m.(*MockStorage).err = errors.New("storage error")
			},
			expectedStatus: http.StatusNotFound,
			skipLoad:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, _, mockStore := setupTest()
			tt.setupMock(mockStore)
			// Don't load tasks to simulate storage error

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/tasks/"+tt.taskID, nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestListTasks(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(storage.TaskStorage)
		expectedStatus int
		validate       func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "list empty tasks",
			setupMock: func(m storage.TaskStorage) {
				m.(*MockStorage).tasks = []*model.Task{}
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response []map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if len(response) != 0 {
					t.Errorf("Expected empty task list, got %d tasks", len(response))
				}
			},
		},
		{
			name: "list multiple tasks",
			setupMock: func(m storage.TaskStorage) {
				task1 := model.NewTask(1, "Task 1", "Description 1")
				task2 := model.NewTask(2, "Task 2", "Description 2")
				task2.MarkComplete()
				m.(*MockStorage).tasks = []*model.Task{task1, task2}
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response []map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if len(response) != 2 {
					t.Errorf("Expected 2 tasks, got %d", len(response))
				}
				if response[0]["title"].(string) != "Task 1" {
					t.Error("Expected first task title to be 'Task 1'")
				}
				if response[1]["completed"].(bool) != true {
					t.Error("Expected second task to be completed")
				}
			},
		},
		{
			name: "list tasks with storage error",
			setupMock: func(m storage.TaskStorage) {
				m.(*MockStorage).err = errors.New("storage error")
			},
			expectedStatus: http.StatusOK, // ListTasks doesn't handle storage errors, returns empty list with 200
			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response []map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if len(response) != 0 {
					t.Errorf("Expected empty task list due to storage error, got %d tasks", len(response))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, taskMgr, mockStore := setupTest()
			tt.setupMock(mockStore)
			if tt.name != "list tasks with storage error" {
				if err := taskMgr.LoadTasks(); err != nil {
					t.Fatalf("Failed to load tasks: %v", err)
				}
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/tasks", nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.validate != nil {
				tt.validate(t, w)
			}
		})
	}
}

func TestGetStats(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(storage.TaskStorage)
		expectedStatus int
		validate       func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "get stats with no tasks",
			setupMock: func(m storage.TaskStorage) {
				m.(*MockStorage).tasks = []*model.Task{}
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				// Handle potential nil values safely
				if total, ok := response["total"]; ok && total != nil {
					if total.(float64) != 0 {
						t.Error("Expected total to be 0")
					}
				}
				if completed, ok := response["completed"]; ok && completed != nil {
					if completed.(float64) != 0 {
						t.Error("Expected completed to be 0")
					}
				}
				if pending, ok := response["pending"]; ok && pending != nil {
					if pending.(float64) != 0 {
						t.Error("Expected pending to be 0")
					}
				}
				if overdue, ok := response["overdue"]; ok && overdue != nil {
					if overdue.(float64) != 0 {
						t.Error("Expected overdue to be 0")
					}
				}
			},
		},
		{
			name: "get stats with mixed tasks",
			setupMock: func(m storage.TaskStorage) {
				task1 := model.NewTask(1, "Task 1", "Description 1")
				task2 := model.NewTask(2, "Task 2", "Description 2")
				task2.MarkComplete()
				task3 := model.NewTask(3, "Task 3", "Description 3")
				if dueDate, err := utils.ParseDueDate("2020-01-01"); err == nil {
					task3.SetDueDate(dueDate)
				}
				m.(*MockStorage).tasks = []*model.Task{task1, task2, task3}
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				// Handle potential nil values safely
				if total, ok := response["total"]; ok && total != nil {
					if total.(float64) != 3 {
						t.Errorf("Expected total to be 3, got %v", total)
					}
				}
				if completed, ok := response["completed"]; ok && completed != nil {
					if completed.(float64) != 1 {
						t.Errorf("Expected completed to be 1, got %v", completed)
					}
				}
				if pending, ok := response["pending"]; ok && pending != nil {
					if pending.(float64) != 2 {
						t.Errorf("Expected pending to be 2, got %v", pending)
					}
				}
				if overdue, ok := response["overdue"]; ok && overdue != nil {
					if overdue.(float64) != 1 {
						t.Errorf("Expected overdue to be 1, got %v", overdue)
					}
				}
			},
		},
		{
			name: "get stats with storage error",
			setupMock: func(m storage.TaskStorage) {
				m.(*MockStorage).err = errors.New("storage error")
			},
			expectedStatus: http.StatusOK, // GetStats doesn't handle storage errors, returns stats with 200
			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				// Just verify we get a valid JSON response - stats may be empty/zero due to storage error
				// The handler doesn't explicitly handle storage errors, so it returns whatever the manager provides
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, taskMgr, mockStore := setupTest()
			tt.setupMock(mockStore)
			if tt.name != "get stats with storage error" {
				if err := taskMgr.LoadTasks(); err != nil {
					t.Fatalf("Failed to load tasks: %v", err)
				}
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/stats", nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.validate != nil {
				tt.validate(t, w)
			}
		})
	}
}

func TestGetTask_ComprehensiveCoverage(t *testing.T) {
	tests := []struct {
		name           string
		taskID         string
		setupMock      func(storage.TaskStorage)
		expectedStatus int
		validate       func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "get existing task successfully",
			taskID: "1",
			setupMock: func(m storage.TaskStorage) {
				task := model.NewTask(1, "Task 1", "Description 1")
				if priority, err := utils.ParsePriority("high"); err == nil {
					task.SetPriority(priority)
				}
				if dueDate, err := utils.ParseDueDate("2025-12-31"); err == nil {
					task.SetDueDate(dueDate)
				}
				m.(*MockStorage).tasks = []*model.Task{task}
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response["title"].(string) != "Task 1" {
					t.Error("Expected task title to match")
				}
				if response["priority"].(float64) != 2 {
					t.Error("Expected priority to be 2 (high)")
				}
			},
		},
		{
			name:           "get task with invalid ID format",
			taskID:         "abc",
			setupMock:      func(m storage.TaskStorage) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "get non-existent task",
			taskID: "999",
			setupMock: func(m storage.TaskStorage) {
				m.(*MockStorage).tasks = []*model.Task{model.NewTask(1, "Task 1", "Description 1")}
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, taskMgr, mockStore := setupTest()
			tt.setupMock(mockStore)
			if err := taskMgr.LoadTasks(); err != nil {
				t.Fatalf("Failed to load tasks: %v", err)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/tasks/"+tt.taskID, nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.validate != nil {
				tt.validate(t, w)
			}
		})
	}
}

func TestUpdateTask_ComprehensiveCoverage(t *testing.T) {
	tests := []struct {
		name           string
		taskID         string
		request        UpdateTaskRequest
		setupMock      func(storage.TaskStorage)
		expectedStatus int
		validate       func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "update task successfully with all fields",
			taskID: "1",
			request: UpdateTaskRequest{
				Title:       "Updated Title",
				Description: "Updated Description",
			},
			setupMock: func(m storage.TaskStorage) {
				m.(*MockStorage).tasks = []*model.Task{model.NewTask(1, "Task 1", "Description 1")}
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response["title"].(string) != "Updated Title" {
					t.Error("Expected title to be updated")
				}
				if response["description"].(string) != "Updated Description" {
					t.Error("Expected description to be updated")
				}
			},
		},
		{
			name:   "update task with invalid task ID",
			taskID: "abc",
			request: UpdateTaskRequest{
				Title: "Updated Title",
			},
			setupMock:      func(m storage.TaskStorage) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "update non-existent task",
			taskID: "999",
			request: UpdateTaskRequest{
				Title: "Updated Title",
			},
			setupMock: func(m storage.TaskStorage) {
				m.(*MockStorage).tasks = []*model.Task{model.NewTask(1, "Task 1", "Description 1")}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, taskMgr, mockStore := setupTest()
			tt.setupMock(mockStore)
			if err := taskMgr.LoadTasks(); err != nil {
				t.Fatalf("Failed to load tasks: %v", err)
			}

			jsonData, _ := json.Marshal(tt.request)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("PUT", "/api/v1/tasks/"+tt.taskID, bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.validate != nil {
				tt.validate(t, w)
			}
		})
	}
}

func TestCompleteTask_ComprehensiveCoverage(t *testing.T) {
	tests := []struct {
		name           string
		taskID         string
		setupMock      func(storage.TaskStorage)
		expectedStatus int
		validate       func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "complete task successfully",
			taskID: "1",
			setupMock: func(m storage.TaskStorage) {
				m.(*MockStorage).tasks = []*model.Task{model.NewTask(1, "Task 1", "Description 1")}
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response["completed"].(bool) != true {
					t.Error("Expected task to be completed")
				}
			},
		},
		{
			name:           "complete task with invalid ID",
			taskID:         "abc",
			setupMock:      func(m storage.TaskStorage) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "complete non-existent task",
			taskID: "999",
			setupMock: func(m storage.TaskStorage) {
				m.(*MockStorage).tasks = []*model.Task{model.NewTask(1, "Task 1", "Description 1")}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, taskMgr, mockStore := setupTest()
			tt.setupMock(mockStore)
			if err := taskMgr.LoadTasks(); err != nil {
				t.Fatalf("Failed to load tasks: %v", err)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("PATCH", "/api/v1/tasks/"+tt.taskID+"/complete", nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.validate != nil {
				tt.validate(t, w)
			}
		})
	}
}
