package manager

import (
	"fmt"
	"sort"
	"task-manager/internal/model"
	"task-manager/internal/storage"
	"time"
)

// TaskManager handler handles the business logic for managing tasks
type TaskManager struct {
	storage storage.TaskStorage
	tasks   []*model.Task
	nextId  int
}

// NewTaskManager creates a new TaskManager with the given storage
func NewTaskManager(storage storage.TaskStorage) (*TaskManager, error) {
	manager := &TaskManager{
		storage: storage,
		tasks:   []*model.Task{},
		nextId:  1,
	}

	// Load tasks from storage
	err := manager.LoadTasks()
	if err != nil {
		return nil, fmt.Errorf("error loading tasks: %w", err)
	}
	return manager, nil
}

// LoadTasks loads task from storage
func (m *TaskManager) LoadTasks() error {
	tasks, err := m.storage.Load()
	if err != nil {
		return err
	}
	m.tasks = tasks

	//Find the highest ID to set nextID
	if len(tasks) > 0 {
		highestID := 0
		for _, task := range tasks {
			if task.ID > highestID {
				highestID = task.ID
			}
		}
		m.nextId = highestID + 1
	}
	return nil
}

// SaveTasks save tasks to storage
func (m *TaskManager) SaveTasks() error {
	return m.storage.Save(m.tasks)
}

// AddTask AddTasks add a new task
func (m *TaskManager) AddTask(title, description string) (*model.Task, error) {
	if title == "" {
		return nil, fmt.Errorf("title cannot be empty")
	}
	task := model.NewTask(m.nextId, title, description)
	m.tasks = append(m.tasks, task)
	m.nextId++

	err := m.SaveTasks()
	if err != nil {
		return nil, fmt.Errorf("failed to save tasks: %w", err)
	}
	return task, nil
}

// GetTask GetTasks gets a task by ID
func (m *TaskManager) GetTask(id int) (*model.Task, error) {
	for _, task := range m.tasks {
		if task.ID == id {
			return task, nil
		}
	}
	return nil, fmt.Errorf("task with ID %d not found", id)
}

// UpdateTask updates a task
func (m *TaskManager) UpdateTask(id int, title, description string) error {
	task, err := m.GetTask(id)
	if err != nil {
		return err
	}

	if title != "" {
		task.Title = title
	}

	if description != "" {
		task.Description = description
	}

	return m.SaveTasks()
}

// DeleteTask deletes a task by ID
func (m *TaskManager) DeleteTask(id int) error {
	for i, task := range m.tasks {
		if task.ID == id {
			//Remove task from slice
			m.tasks = append(m.tasks[:i], m.tasks[i+1:]...)
			return m.SaveTasks()
		}
	}
	return fmt.Errorf("task with ID %d not found", id)
}

// MarkTaskComplete MarkTaskAsComplete marks a task as complete
func (m *TaskManager) MarkTaskComplete(id int) error {
	task, err := m.GetTask(id)
	if err != nil {
		return err
	}
	task.MarkComplete()
	return m.SaveTasks()
}

// MarkTaskIncomplete SetTaskIncomplete marks a task as incomplete
func (m *TaskManager) MarkTaskIncomplete(id int) error {
	task, err := m.GetTask(id)
	if err != nil {
		return err
	}
	task.MarkIncomplete()
	return m.SaveTasks()
}

// SetTaskDueDate sets the due date for a task
func (m *TaskManager) SetTaskDueDate(id int, dueDate time.Time) error {
	task, err := m.GetTask(id)
	if err != nil {
		return err
	}
	task.SetDueDate(dueDate)
	return m.SaveTasks()
}

// SetTaskPriority sets the priority for a task
func (m *TaskManager) SetTaskPriority(id int, priority int) error {
	task, err := m.GetTask(id)
	if err != nil {
		return err
	}

	err = task.SetPriority(priority)
	if err != nil {
		return err
	}
	return m.SaveTasks()
}

// GetAllTasks return all tasks
func (m *TaskManager) GetAllTasks() []*model.Task {
	return m.tasks
}

// GetCompletedTasks return all tasks
func (m *TaskManager) GetCompletedTasks() []*model.Task {
	var completedTasks []*model.Task
	for _, task := range m.tasks {
		if task.Completed {
			completedTasks = append(completedTasks, task)
		}
	}
	return completedTasks
}

// GetPendingTasks returns all pending tasks
func (m *TaskManager) GetPendingTasks() []*model.Task {
	var pendingTasks []*model.Task
	for _, task := range m.tasks {
		if !task.Completed {
			pendingTasks = append(pendingTasks, task)
		}
	}
	return pendingTasks
}

// GetOverdueTasks returns all overdue tasks
func (m *TaskManager) GetOverdueTasks() []*model.Task {
	var overdueTasks []*model.Task
	for _, task := range m.tasks {
		if task.IsOverdue() {
			overdueTasks = append(overdueTasks, task)
		}
	}
	return overdueTasks
}

// SortTasksByPriority sorts task by priority (highest first)
func (m *TaskManager) SortTasksByPriority(tasks []*model.Task) []*model.Task {
	sortedTasks := make([]*model.Task, len(tasks))
	copy(sortedTasks, tasks)

	sort.SliceStable(sortedTasks, func(i, j int) bool {
		return sortedTasks[i].Priority < sortedTasks[j].Priority
	})
	return sortedTasks
}

// SortTasksByDueDate sorts task by due date (earliest first)
func (m *TaskManager) SortTasksByDueDate(tasks []*model.Task) []*model.Task {
	sortedTasks := make([]*model.Task, len(tasks))
	copy(sortedTasks, tasks)

	sort.SliceStable(sortedTasks, func(i, j int) bool {
		//Tasks without due date go last
		if sortedTasks[i].DueDate.IsZero() {
			return false
		}
		if sortedTasks[j].DueDate.IsZero() {
			return true
		}
		return sortedTasks[i].DueDate.Before(sortedTasks[j].DueDate)
	})
	return sortedTasks
}

// GetTaskStats returns statistics about tasks
func (m *TaskManager) GetTaskStats() map[string]int {
	total := len(m.tasks)
	completed := len(m.GetCompletedTasks())
	pending := len(m.GetPendingTasks())
	overdue := len(m.GetOverdueTasks())

	return map[string]int{
		"Total tasks":     total,
		"Completed tasks": completed,
		"Pending":         pending,
		"Overdue":         overdue,
	}
}
