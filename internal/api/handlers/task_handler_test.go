package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"reflect"
	"task-manager/internal/manager"
	"task-manager/internal/model"
	"task-manager/internal/storage"
	"testing"
	"time"
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

func TestListTasks(t *testing.T) {
	type TaskResponse struct {
		ID          int       `json:"id"`
		Title       string    `json:"title"`
		Description string    `json:"description"`
		Completed   bool      `json:"completed"`
		CreatedAt   time.Time `json:"created_at"`
		CompletedAt time.Time `json:"completed_at,omitempty"`
		DueDate     time.Time `json:"due_date,omitempty"`
		Priority    int       `json:"priority"`
		IsOverdue   bool      `json:"is_overdue"`
	}

	now := time.Now()
	dueDate := now.Add(24 * time.Hour)
	overdueDueDate := now.Add(-24 * time.Hour)

	tests := []struct {
		name           string
		query          string
		setupMock      func(storage.TaskStorage, *manager.TaskManager)
		expectedStatus int
		expectedLen    int
		validate       func([]TaskResponse) bool
	}{
		{
			name:  "list all tasks",
			query: "",
			setupMock: func(s storage.TaskStorage, tm *manager.TaskManager) {
				mock := s.(*MockStorage)
				mock.tasks = []*model.Task{
					model.NewTask(1, "Task 1", "Description 1"),
					model.NewTask(2, "Task 2", "Description 2"),
				}
				mock.tasks[0].SetDueDate(dueDate)
				if err := tm.LoadTasks(); err != nil {
					t.Fatalf("Failed to load tasks: %v", err)
				}
			},
			expectedStatus: http.StatusOK,
			expectedLen:    2,
			validate: func(tasks []TaskResponse) bool {
				return len(tasks) == 2 && !tasks[0].IsOverdue
			},
		},
		{
			name:  "list completed tasks",
			query: "?completed=true",
			setupMock: func(s storage.TaskStorage, tm *manager.TaskManager) {
				mock := s.(*MockStorage)
				task := model.NewTask(1, "Task 1", "Description 1")
				task.MarkComplete()
				mock.tasks = []*model.Task{task}
				if err := tm.LoadTasks(); err != nil {
					t.Fatalf("Failed to load tasks: %v", err)
				}
			},
			expectedStatus: http.StatusOK,
			expectedLen:    1,
			validate: func(tasks []TaskResponse) bool {
				return len(tasks) == 1 && tasks[0].Completed && !tasks[0].CompletedAt.IsZero()
			},
		},
		{
			name:  "list overdue tasks",
			query: "?overdue=true",
			setupMock: func(s storage.TaskStorage, tm *manager.TaskManager) {
				mock := s.(*MockStorage)
				task := model.NewTask(1, "Task 1", "Description 1")
				task.SetDueDate(overdueDueDate)
				mock.tasks = []*model.Task{task}
				if err := tm.LoadTasks(); err != nil {
					t.Fatalf("Failed to load tasks: %v", err)
				}
			},
			expectedStatus: http.StatusOK,
			expectedLen:    1,
			validate: func(tasks []TaskResponse) bool {
				return len(tasks) == 1 && tasks[0].IsOverdue
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, taskMgr, mockStore := setupTest()
			tt.setupMock(mockStore, taskMgr)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/tasks"+tt.query, nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response []TaskResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if len(response) != tt.expectedLen {
				t.Errorf("Expected %d tasks, got %d", tt.expectedLen, len(response))
			}

			if tt.validate != nil && !tt.validate(response) {
				t.Error("Response validation failed")
			}
		})
	}
}

func TestCreateTask(t *testing.T) {
	tests := []struct {
		name           string
		request        CreateTaskRequest
		setupMock      func(storage.TaskStorage)
		expectedStatus int
		expectedError  string
		validate       func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "create task successfully",
			request: CreateTaskRequest{
				Title:       "Test Task",
				Description: "Test Description",
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
				if title, ok := response["title"].(string); !ok || title != "Test Task" {
					t.Errorf("Expected title 'Test Task', got %v", title)
				}
				if desc, ok := response["description"].(string); !ok || desc != "Test Description" {
					t.Errorf("Expected description 'Test Description', got %v", desc)
				}
			},
		},
		{
			name: "create task with missing title",
			request: CreateTaskRequest{
				Description: "Test Description",
			},
			setupMock: func(m storage.TaskStorage) {
				m.(*MockStorage).err = nil
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "create task with error",
			request: CreateTaskRequest{
				Title:       "Test Task",
				Description: "Test Description",
			},
			setupMock: func(m storage.TaskStorage) {
				m.(*MockStorage).err = errors.New("internal error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to save tasks: internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, _, mockStore := setupTest()
			tt.setupMock(mockStore)

			jsonData, _ := json.Marshal(tt.request)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
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
				if response["error"] != tt.expectedError {
					t.Errorf("Expected error %s, got %s", tt.expectedError, response["error"])
				}
			}
		})
	}
}

func TestGetTask(t *testing.T) {
	tests := []struct {
		name           string
		taskID         string
		setupMock      func(storage.TaskStorage)
		expectedStatus int
		expectedError  string
		validate       func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "get existing task",
			taskID: "1",
			setupMock: func(s storage.TaskStorage) {
				mock := s.(*MockStorage)
				mock.tasks = []*model.Task{model.NewTask(1, "Task 1", "Description 1")}
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if id, ok := response["id"].(float64); !ok || int(id) != 1 {
					t.Errorf("Expected id 1, got %v", id)
				}
				if title, ok := response["title"].(string); !ok || title != "Task 1" {
					t.Errorf("Expected title 'Task 1', got %v", title)
				}
				if desc, ok := response["description"].(string); !ok || desc != "Description 1" {
					t.Errorf("Expected description 'Description 1', got %v", desc)
				}
			},
		},
		{
			name:   "get non-existent task",
			taskID: "999",
			setupMock: func(s storage.TaskStorage) {
				mock := s.(*MockStorage)
				mock.tasks = []*model.Task{}
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "task with ID 999 not found",
		},
		{
			name:   "invalid task ID",
			taskID: "invalid",
			setupMock: func(s storage.TaskStorage) {
				mock := s.(*MockStorage)
				mock.tasks = []*model.Task{}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid task ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, taskMgr, mockStore := setupTest()
			tt.setupMock(mockStore)
			if err := taskMgr.LoadTasks(); err != nil && tt.expectedError == "" {
				t.Fatalf("Failed to load tasks: %v", err)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/tasks/"+tt.taskID, nil)
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
				if response["error"] != tt.expectedError {
					t.Errorf("Expected error %s, got %s", tt.expectedError, response["error"])
				}
			}

			if tt.validate != nil {
				tt.validate(t, w)
			}
		})
	}
}

func TestUpdateTask(t *testing.T) {
	tests := []struct {
		name           string
		taskID         string
		request        UpdateTaskRequest
		setupMock      func(storage.TaskStorage)
		expectedStatus int
		expectedError  string
		validate       func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "update existing task",
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
				if title, ok := response["title"].(string); !ok || title != "Updated Title" {
					t.Errorf("Expected title to be 'Updated Title', got %v", title)
				}
				if desc, ok := response["description"].(string); !ok || desc != "Updated Description" {
					t.Errorf("Expected description to be 'Updated Description', got %v", desc)
				}
			},
		},
		{
			name:   "update non-existent task",
			taskID: "999",
			request: UpdateTaskRequest{
				Title:       "Updated Title",
				Description: "Updated Description",
			},
			setupMock: func(m storage.TaskStorage) {
				m.(*MockStorage).tasks = []*model.Task{}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "task with ID 999 not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, taskMgr, mockStore := setupTest()
			tt.setupMock(mockStore)
			if err := taskMgr.LoadTasks(); err != nil && tt.expectedError == "" {
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

			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response["error"] != tt.expectedError {
					t.Errorf("Expected error %s, got %s", tt.expectedError, response["error"])
				}
			}

			if tt.validate != nil {
				tt.validate(t, w)
			}
		})
	}
}

func TestCompleteTask(t *testing.T) {
	tests := []struct {
		name           string
		taskID         string
		setupMock      func(storage.TaskStorage)
		expectedStatus int
		expectedError  string
		validate       func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "complete existing task",
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
				completed, ok := response["completed"].(bool)
				if !ok || !completed {
					t.Error("Expected task to be marked as completed")
				}
			},
		},
		{
			name:   "complete non-existent task",
			taskID: "999",
			setupMock: func(m storage.TaskStorage) {
				m.(*MockStorage).tasks = []*model.Task{}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "task with ID 999 not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, taskMgr, mockStore := setupTest()
			tt.setupMock(mockStore)
			if err := taskMgr.LoadTasks(); err != nil && tt.expectedError == "" {
				t.Fatalf("Failed to load tasks: %v", err)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("PATCH", "/api/v1/tasks/"+tt.taskID+"/complete", nil)
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
				if response["error"] != tt.expectedError {
					t.Errorf("Expected error %s, got %s", tt.expectedError, response["error"])
				}
			}

			if tt.validate != nil {
				tt.validate(t, w)
			}
		})
	}
}

func TestSetTaskDueDate(t *testing.T) {
	tests := []struct {
		name           string
		taskID         string
		request        SetDueDateRequest
		setupMock      func(storage.TaskStorage)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "set due date for existing task",
			taskID: "1",
			request: SetDueDateRequest{
				DueDate: "2025-12-31",
			},
			setupMock: func(m storage.TaskStorage) {
				m.(*MockStorage).tasks = []*model.Task{model.NewTask(1, "Task 1", "Description 1")}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "set invalid due date",
			taskID: "1",
			request: SetDueDateRequest{
				DueDate: "invalid-date",
			},
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
			if err := taskMgr.LoadTasks(); err != nil {
				t.Fatalf("Failed to load tasks: %v", err)
			}

			jsonData, _ := json.Marshal(tt.request)
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

func TestGetStats(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(storage.TaskStorage, *manager.TaskManager)
		expectedStatus int
		expectedStats  map[string]int
	}{
		{
			name: "get stats with mixed tasks",
			setupMock: func(s storage.TaskStorage, tm *manager.TaskManager) {
				mock := s.(*MockStorage)
				task1 := model.NewTask(1, "Task 1", "Description 1")
				task1.MarkComplete()
				task2 := model.NewTask(2, "Task 2", "Description 2")
				task3 := model.NewTask(3, "Task 3", "Description 3")
				task3.SetDueDate(time.Now().Add(-24 * time.Hour)) // overdue task
				mock.tasks = []*model.Task{task1, task2, task3}
				tm.LoadTasks()
			},
			expectedStatus: http.StatusOK,
			expectedStats: map[string]int{
				"Total tasks":     3,
				"Completed tasks": 1,
				"Pending":         2,
				"Overdue":         1,
			},
		},
		{
			name: "get stats with no tasks",
			setupMock: func(s storage.TaskStorage, tm *manager.TaskManager) {
				mock := s.(*MockStorage)
				mock.tasks = []*model.Task{}
				tm.LoadTasks()
			},
			expectedStatus: http.StatusOK,
			expectedStats: map[string]int{
				"Total tasks":     0,
				"Completed tasks": 0,
				"Pending":         0,
				"Overdue":         0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, taskMgr, mockStore := setupTest()
			tt.setupMock(mockStore, taskMgr)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/stats", nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response map[string]int
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if !reflect.DeepEqual(response, tt.expectedStats) {
				t.Errorf("Expected stats %v, got %v", tt.expectedStats, response)
			}
		})
	}
}

func TestSetPriority(t *testing.T) {
	tests := []struct {
		name           string
		taskID         string
		request        SetPriorityRequest
		setupMock      func(storage.TaskStorage, *manager.TaskManager)
		expectedStatus int
		expectedError  string
		validate       func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "set valid priority",
			taskID: "1",
			request: SetPriorityRequest{
				Priority: "high",
			},
			setupMock: func(s storage.TaskStorage, tm *manager.TaskManager) {
				mock := s.(*MockStorage)
				task := model.NewTask(1, "Task 1", "Description 1")
				mock.tasks = []*model.Task{task}
				if err := tm.LoadTasks(); err != nil {
					t.Fatalf("Failed to load tasks: %v", err)
				}
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response TaskResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Priority != 2 { // high priority is actually 2 in the implementation
					t.Errorf("Expected priority 2, got %d", response.Priority)
				}
			},
		},
		{
			name:   "set invalid priority",
			taskID: "1",
			request: SetPriorityRequest{
				Priority: "invalid",
			},
			setupMock: func(s storage.TaskStorage, tm *manager.TaskManager) {
				mock := s.(*MockStorage)
				mock.tasks = []*model.Task{model.NewTask(1, "Task 1", "Description 1")}
				if err := tm.LoadTasks(); err != nil {
					t.Fatalf("Failed to load tasks: %v", err)
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid priority format, please use a number between 1 and 5 or one of the following: highest, high, medium, low, lowest",
		},
		{
			name:   "set priority for non-existent task",
			taskID: "999",
			request: SetPriorityRequest{
				Priority: "high",
			},
			setupMock: func(s storage.TaskStorage, tm *manager.TaskManager) {
				mock := s.(*MockStorage)
				mock.tasks = []*model.Task{}
				if err := tm.LoadTasks(); err != nil {
					t.Fatalf("Failed to load tasks: %v", err)
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "task with ID 999 not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, taskMgr, mockStore := setupTest()
			tt.setupMock(mockStore, taskMgr)

			jsonData, _ := json.Marshal(tt.request)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("PATCH", "/api/v1/tasks/"+tt.taskID+"/priority", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
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
				if response["error"] != tt.expectedError {
					t.Errorf("Expected error %s, got %s", tt.expectedError, response["error"])
				}
			}

			if tt.validate != nil {
				tt.validate(t, w)
			}
		})
	}
}
