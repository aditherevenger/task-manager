package api

import (
	"github.com/gin-gonic/gin"
	"task-manager/internal/api/handlers"
	"task-manager/internal/api/middleware"
	"task-manager/internal/manager"
)

// Server represents the API server.
type Server struct {
	route       *gin.Engine
	taskManager *manager.TaskManager
}

// NewServer creates a new API server with the given task manager.
func NewServer(taskManager *manager.TaskManager) *Server {
	server := &Server{
		route:       gin.Default(),
		taskManager: taskManager,
	}

	// Add middlewares
	server.route.Use(middleware.Logger())

	//Register routes
	server.registerRoutes()

	return server
}

// registerRoutes registers the API routes.
func (s *Server) registerRoutes() {
	// Create task handlers
	taskHandlers := handlers.NewTaskHandlers(s.taskManager)

	//API v1 group
	v1 := s.route.Group("/api/v1")
	{
		// Tasks endpoints
		tasks := v1.Group("/tasks")
		{
			tasks.GET("", taskHandlers.ListTasks)
			tasks.GET("/:id", taskHandlers.GetTask)
			tasks.POST("", taskHandlers.CreateTask)
			tasks.PUT("/:id", taskHandlers.UpdateTask)
			tasks.DELETE("/:id", taskHandlers.DeleteTask)
			tasks.PATCH("/:id/complete", taskHandlers.CompleteTask)
			tasks.PATCH("/:id/uncomplete", taskHandlers.UncompleteTask)
			tasks.PATCH("/:id/due-date", taskHandlers.SetDueDate)
			tasks.PATCH("/:id/priority", taskHandlers.SetPriority)
		}

		// Stats endpoints
		v1.GET("/stats", taskHandlers.GetStats)
	}
}

// Run starts the API server on the specified address.
func (s *Server) Run(addr string) error {
	return s.route.Run(addr)
}
