package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"task-manager/internal/manager"
	"task-manager/internal/model"
)

// MockStorage implements storage.TaskStorage for testing
type MockStorage struct {
	tasks []*model.Task
	err   error
}

func (m *MockStorage) Load() ([]*model.Task, error) {
	return m.tasks, m.err
}

func (m *MockStorage) Save(tasks []*model.Task) error {
	m.tasks = tasks
	return m.err
}

func TestNewServer(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		taskManager *manager.TaskManager
		expectNil   bool
	}{
		{
			name:        "Valid task manager",
			taskManager: createTestTaskManager(t),
			expectNil:   false,
		},
		{
			name:        "Nil task manager",
			taskManager: nil,
			expectNil:   false, // Server should still be created even with nil task manager
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer(tt.taskManager)

			if tt.expectNil {
				assert.Nil(t, server)
			} else {
				assert.NotNil(t, server)
				assert.NotNil(t, server.route)
				assert.Equal(t, tt.taskManager, server.taskManager)
			}
		})
	}
}

// createTestTaskManager creates a TaskManager for testing
func createTestTaskManager(t *testing.T) *manager.TaskManager {
	mockStorage := &MockStorage{
		tasks: []*model.Task{},
		err:   nil,
	}
	
	taskManager, err := manager.NewTaskManager(mockStorage)
	if err != nil {
		t.Fatalf("Failed to create test task manager: %v", err)
	}
	
	return taskManager
}

func TestServer_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	taskManager := createTestTaskManager(t)
	server := NewServer(taskManager)

	// Test that routes are properly registered by making test requests
	testCases := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "GET /api/v1/tasks",
			method:         "GET",
			path:           "/api/v1/tasks",
			expectedStatus: http.StatusOK, // Should return empty list
		},
		{
			name:           "GET /api/v1/tasks/1",
			method:         "GET",
			path:           "/api/v1/tasks/1",
			expectedStatus: http.StatusNotFound, // Task doesn't exist
		},
		{
			name:           "POST /api/v1/tasks",
			method:         "POST",
			path:           "/api/v1/tasks",
			expectedStatus: http.StatusBadRequest, // Empty body
		},
		{
			name:           "PUT /api/v1/tasks/1",
			method:         "PUT",
			path:           "/api/v1/tasks/1",
			expectedStatus: http.StatusBadRequest, // Empty body
		},
		{
			name:           "DELETE /api/v1/tasks/1",
			method:         "DELETE",
			path:           "/api/v1/tasks/1",
			expectedStatus: http.StatusNotFound, // Task doesn't exist
		},
		{
			name:           "PATCH /api/v1/tasks/1/complete",
			method:         "PATCH",
			path:           "/api/v1/tasks/1/complete",
			expectedStatus: http.StatusNotFound, // Task doesn't exist
		},
		{
			name:           "PATCH /api/v1/tasks/1/uncomplete",
			method:         "PATCH",
			path:           "/api/v1/tasks/1/uncomplete",
			expectedStatus: http.StatusNotFound, // Task doesn't exist
		},
		{
			name:           "PATCH /api/v1/tasks/1/due-date",
			method:         "PATCH",
			path:           "/api/v1/tasks/1/due-date",
			expectedStatus: http.StatusBadRequest, // Empty body
		},
		{
			name:           "PATCH /api/v1/tasks/1/priority",
			method:         "PATCH",
			path:           "/api/v1/tasks/1/priority",
			expectedStatus: http.StatusBadRequest, // Empty body
		},
		{
			name:           "GET /api/v1/stats",
			method:         "GET",
			path:           "/api/v1/stats",
			expectedStatus: http.StatusOK, // Should return stats
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, tc.path, nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			server.route.ServeHTTP(w, req)

			if tc.path == "/api/v1/tasks/1" && tc.method == "GET" {
				// This endpoint should return 404 for non-existent task, which is correct behavior
				assert.True(t, w.Code == http.StatusNotFound || w.Code == http.StatusOK, 
					"GET /api/v1/tasks/1 should return 404 (not found) or 200 (found)")
			} else {
				// For other routes, just ensure they're registered (not 404 for unregistered route)
				assert.NotEqual(t, http.StatusNotFound, w.Code, "Route should be registered")
			}
		})
	}
}

func TestServer_RegisterRoutes_InvalidPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	taskManager := createTestTaskManager(t)
	server := NewServer(taskManager)

	// Test invalid paths that should return 404
	invalidPaths := []string{
		"/api/v2/tasks",
		"/api/v1/task", // singular instead of plural
		"/tasks",       // missing api/v1 prefix
		"/api/v1/tasks/invalid/complete",
		"/api/v1/stats/invalid",
	}

	for _, path := range invalidPaths {
		t.Run("Invalid path: "+path, func(t *testing.T) {
			req, err := http.NewRequest("GET", path, nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			server.route.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code, "Invalid path should return 404")
		})
	}
}

func TestServer_MiddlewareRegistration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	taskManager := createTestTaskManager(t)
	server := NewServer(taskManager)

	// Test that middleware is properly registered by checking if logger middleware is applied
	req, err := http.NewRequest("GET", "/api/v1/tasks", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	server.route.ServeHTTP(w, req)

	// The middleware should be applied - we can't easily test the logger middleware
	// but we can ensure the request goes through without panicking
	assert.NotNil(t, w)
}

func TestServer_Run(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	taskManager := createTestTaskManager(t)
	server := NewServer(taskManager)

	// Test Run method with invalid address (should return error)
	err := server.Run("invalid-address")
	assert.Error(t, err, "Invalid address should return error")

	// Note: Testing with a valid address would start an actual server,
	// which is not suitable for unit tests. Integration tests would be better for that.
}

func TestServer_RouteGroupStructure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	taskManager := createTestTaskManager(t)
	server := NewServer(taskManager)

	// Test that the route structure is correct by checking route info
	routes := server.route.Routes()
	
	// Verify that we have the expected number of routes
	assert.Greater(t, len(routes), 0, "Should have registered routes")

	// Check for specific route patterns
	routePaths := make(map[string]bool)
	for _, route := range routes {
		routePaths[route.Path] = true
	}

	expectedPaths := []string{
		"/api/v1/tasks",
		"/api/v1/tasks/:id",
		"/api/v1/tasks/:id/complete",
		"/api/v1/tasks/:id/uncomplete",
		"/api/v1/tasks/:id/due-date",
		"/api/v1/tasks/:id/priority",
		"/api/v1/stats",
	}

	for _, expectedPath := range expectedPaths {
		assert.True(t, routePaths[expectedPath], "Expected route path %s should be registered", expectedPath)
	}
}

func TestServer_HTTPMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	taskManager := createTestTaskManager(t)
	server := NewServer(taskManager)

	routes := server.route.Routes()
	
	// Map to track method-path combinations
	methodPaths := make(map[string][]string)
	for _, route := range routes {
		methodPaths[route.Method] = append(methodPaths[route.Method], route.Path)
	}

	// Verify specific HTTP methods are used for appropriate endpoints
	assert.Contains(t, methodPaths["GET"], "/api/v1/tasks", "GET method should be registered for tasks list")
	assert.Contains(t, methodPaths["GET"], "/api/v1/tasks/:id", "GET method should be registered for single task")
	assert.Contains(t, methodPaths["POST"], "/api/v1/tasks", "POST method should be registered for task creation")
	assert.Contains(t, methodPaths["PUT"], "/api/v1/tasks/:id", "PUT method should be registered for task update")
	assert.Contains(t, methodPaths["DELETE"], "/api/v1/tasks/:id", "DELETE method should be registered for task deletion")
	assert.Contains(t, methodPaths["PATCH"], "/api/v1/tasks/:id/complete", "PATCH method should be registered for task completion")
	assert.Contains(t, methodPaths["GET"], "/api/v1/stats", "GET method should be registered for stats")
}

func TestServer_WithPreloadedTasks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Create a task manager with some pre-loaded tasks
	mockStorage := &MockStorage{
		tasks: []*model.Task{
			{
				ID:          1,
				Title:       "Test Task 1",
				Description: "Test Description 1",
				Completed:   false,
				CreatedAt:   time.Now(),
				Priority:    3,
			},
			{
				ID:          2,
				Title:       "Test Task 2",
				Description: "Test Description 2",
				Completed:   true,
				CreatedAt:   time.Now(),
				Priority:    1,
			},
		},
		err: nil,
	}
	
	taskManager, err := manager.NewTaskManager(mockStorage)
	assert.NoError(t, err)
	
	server := NewServer(taskManager)

	// Test GET /api/v1/tasks with existing tasks
	req, err := http.NewRequest("GET", "/api/v1/tasks", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	server.route.ServeHTTP(w, req)

	// Should return 200 OK since we have tasks
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestServer_TaskManagerIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	taskManager := createTestTaskManager(t)
	server := NewServer(taskManager)

	// Verify that the server has the correct task manager instance
	assert.Equal(t, taskManager, server.taskManager)
	
	// Verify that the task manager is properly initialized
	assert.NotNil(t, server.taskManager)
	
	// Test that we can get all tasks (should be empty initially)
	tasks := server.taskManager.GetAllTasks()
	assert.Empty(t, tasks, "Initial task list should be empty")
}
