package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"task-manager/internal/manager"
	"task-manager/internal/model"
	"task-manager/pkg/utils"
	"time"
)

// TaskHandlers handles HTTP requests for tasks
type TaskHandlers struct {
	taskManager *manager.TaskManager
}

// NewTaskHandlers creates a new instance of TaskHandlers
func NewTaskHandlers(taskManager *manager.TaskManager) *TaskHandlers {
	return &TaskHandlers{
		taskManager: taskManager,
	}
}

// TaskResponse represents a response for a task
type TaskResponse struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt time.Time `json:"completed_at",omitempty`
	DueDate     time.Time `json:"due_date",omitempty`
	Priority    int       `json:"priority"`
	IsOverdue   bool      `json:"is_overdue"`
}

// CreateTaskRequest represents a request to create a task
type CreateTaskRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	DueDate     string `json:"due_date"`
	Priority    string `json:"priority"`
}

// UpdateTaskRequest represents a request to update a task
type UpdateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// SetDueDateRequest represents a request to set a due date for a task
type SetDueDateRequest struct {
	DueDate string `json:"due_date" binding:"required"`
}

// SetPriorityRequest represents a request to set the priority of a task
type SetPriorityRequest struct {
	Priority string `json:"priority" binding:"required"`
}

// taskResponse converts a task to a TaskResponse
func taskToResponse(task *model.Task) TaskResponse {
	return TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Completed:   task.Completed,
		CreatedAt:   task.CreatedAt,
		CompletedAt: task.CompletedAt,
		DueDate:     task.DueDate,
		Priority:    task.Priority,
		IsOverdue:   task.IsOverdue(),
	}
}

// ListTasks handles GET /api/v1/tasks
func (h *TaskHandlers) ListTasks(c *gin.Context) {
	//Get query parameters
	completed := c.Query("completed")
	overdue := c.Query("overdue")

	var tasks []*model.Task

	//Filter tasks based on query parameters
	if completed == "true" {
		tasks = h.taskManager.GetCompletedTasks()
	} else if completed == "false" {
		tasks = h.taskManager.GetPendingTasks()
	} else if overdue == "true" {
		tasks = h.taskManager.GetOverdueTasks()
	} else {
		tasks = h.taskManager.GetAllTasks()
	}

	//convert tasks to response format
	response := make([]TaskResponse, len(tasks))
	for i, task := range tasks {
		response[i] = taskToResponse(task)
	}
	c.JSON(http.StatusOK, response)
}

// GetTask handles GET /api/v1/tasks/:id
func (h *TaskHandlers) GetTask(c *gin.Context) {
	//Parse task ID from URL
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}
	//Get task
	task, err := h.taskManager.GetTask(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, taskToResponse(task))
}

// CreateTask handles POST /api/v1/tasks
func (h *TaskHandlers) CreateTask(c *gin.Context) {
	var request CreateTaskRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create a new task
	task, err := h.taskManager.AddTask(request.Title, request.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set due date if provided
	if request.DueDate != "" {
		dueDate, err := utils.ParseDueDate(request.DueDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		h.taskManager.SetTaskDueDate(task.ID, dueDate)
	}

	// Set priority if provided
	if request.Priority != "" {
		priority, err := utils.ParsePriority(request.Priority)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		h.taskManager.SetTaskPriority(task.ID, priority)
	}

	//Reload the task to get updated fields
	task, _ = h.taskManager.GetTask(task.ID)
	c.JSON(http.StatusCreated, taskToResponse(task))
}

// UpdateTask handles PUT /api/v1/tasks/:id
func (h *TaskHandlers) UpdateTask(c *gin.Context) {
	// Parse task ID
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}
	var request UpdateTaskRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update the task
	err = h.taskManager.UpdateTask(id, request.Title, request.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//Get the updated task
	task, _ := h.taskManager.GetTask(id)
	c.JSON(http.StatusOK, taskToResponse(task))
}

// DeleteTask handles DELETE /api/v1/tasks/:id
func (h *TaskHandlers) DeleteTask(c *gin.Context) {
	// Parse task ID
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	// Delete the task
	err = h.taskManager.DeleteTask(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// CompleteTask handles PATCH /api/v1/tasks/:id/complete
func (h *TaskHandlers) CompleteTask(c *gin.Context) {
	// Parse task ID
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	// Marks Task as Completed
	err = h.taskManager.MarkTaskComplete(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//Get the updated task
	task, _ := h.taskManager.GetTask(id)
	c.JSON(http.StatusOK, taskToResponse(task))
}

// UncompleteTask handles PATCH /api/v1/tasks/:id/uncomplete
func (h *TaskHandlers) UncompleteTask(c *gin.Context) {
	// Parse task ID
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	// Marks Task as Incompleted
	err = h.taskManager.MarkTaskIncomplete(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//Get the updated task
	task, _ := h.taskManager.GetTask(id)
	c.JSON(http.StatusOK, taskToResponse(task))
}

// SetDueDate handles PATCH /api/v1/tasks/:id/due-date
func (h *TaskHandlers) SetDueDate(c *gin.Context) {
	// Parse task ID
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	var request SetDueDateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse due date
	dueDate, err := utils.ParseDueDate(request.DueDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set the due date for the task
	err = h.taskManager.SetTaskDueDate(id, dueDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//Get the updated task
	task, _ := h.taskManager.GetTask(id)
	c.JSON(http.StatusOK, taskToResponse(task))
}

// SetPriority handles PATCH /api/v1/tasks/:id/priority
func (h *TaskHandlers) SetPriority(c *gin.Context) {
	// Parse task ID
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	var request SetPriorityRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse priority
	priority, err := utils.ParsePriority(request.Priority)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set the priority for the task
	err = h.taskManager.SetTaskPriority(id, priority)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//Get the updated task
	task, _ := h.taskManager.GetTask(id)
	c.JSON(http.StatusOK, taskToResponse(task))
}

// GetStats handles GET /api/v1/stats
func (h *TaskHandlers) GetStats(c *gin.Context) {
	stats := h.taskManager.GetTaskStats()
	c.JSON(http.StatusOK, stats)
}
