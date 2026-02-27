package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"task-manager/internal/api"
	"task-manager/internal/manager"
	"task-manager/internal/storage"
	"task-manager/pkg/utils"
)

const (
	defaultStorageFile = "tasks.json"
	defaultPort        = "8080"
)

func main() {

	//Define command line flags
	var (
		storageFile string
		apiMode     bool
		apiAddr     string
		showHelp    bool
	)

	flag.StringVar(&storageFile, "file", defaultStorageFile, "Path to the storage file")
	flag.BoolVar(&apiMode, "api", false, "Run the API server")
	flag.StringVar(&apiAddr, "addr", ":"+defaultPort, "API server address")
	flag.BoolVar(&showHelp, "help", false, "Show help message")
	flag.Parse()

	if showHelp {
		printHelp()
		return
	}

	//Ensure storage file path is absolute
	if !filepath.IsAbs(storageFile) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
			os.Exit(1)
		}
		storageFile = filepath.Join(homeDir, ".task_manager", storageFile)
	}

	//Create storage and task manager
	jsonStorage := storage.NewJSONStorage(storageFile)
	taskManager, err := manager.NewTaskManager(jsonStorage)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating task manager: %v\n", err)
		os.Exit(1)
	}

	//Rin in API mode if specified
	if apiMode {
		fmt.Printf("Starting API server on %s\n", apiAddr)
		server := api.NewServer(taskManager)
		if err := server.Run(apiAddr); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting API server: %v\n", err)
			os.Exit(1)
		}
		return
	}

	//Otherwise, run the CLI mode
	runCLI(taskManager)
}

func runCLI(taskManager *manager.TaskManager) {
	//Process commands
	args := flag.Args()
	if len(args) == 0 {
		//Default action: List all tasks
		listTasks(taskManager, false)
		return
	}

	command := args[0]
	switch command {
	case "add", "a":
		addTask(taskManager, args[1:])
	case "list", "ls", "l":
		listTasks(taskManager, false)
	case "list-all", "la":
		ListAllTasks(taskManager, false)
	case "list-completed", "lc":
		listCompletedTasks(taskManager, false)
	case "list-pending", "lp":
		listPendingTasks(taskManager, false)
	case "list-overdue", "lo":
		listOverdueTasks(taskManager, false)
	case "detail", "d":
		showTaskDetail(taskManager, args[1:])
	case "complete", "comp", "c":
		completeTask(taskManager, args[1:])
	case "uncomplete", "uc":
		uncompleteTask(taskManager, args[1:])
	case "update", "u":
		updateTask(taskManager, args[1:])
	case "delete", "del", "rm":
		deleteTask(taskManager, args[1:])
	case "due", "due-date":
		setDueDate(taskManager, args[1:])
	case "priority", "p":
		setPriority(taskManager, args[1:])
	case "stats", "st":
		showStats(taskManager)
	case "help", "h":
		printHelp()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printHelp()
		os.Exit(1)
	}
}

func addTask(tm *manager.TaskManager, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: Title is required for adding a task.")
		return
	}

	title := args[0]
	description := ""
	if len(args) > 1 {
		description = strings.Join(args[1:], " ")
	}

	task, err := tm.AddTask(title, description)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error adding task: %v\n", err)
		return
	}

	fmt.Printf("Task added: %s\n", task.String())
}

func listTasks(tm *manager.TaskManager, showDetails bool) {
	tasks := tm.GetPendingTasks()
	fmt.Println("Pending Tasks:")
	fmt.Println(utils.FormatTaskList(tasks, showDetails))
}

func ListAllTasks(tm *manager.TaskManager, showDetails bool) {
	tasks := tm.GetAllTasks()
	if len(tasks) == 0 {
		fmt.Println("No tasks found.")
		return
	}
	fmt.Println("All Tasks:")
	fmt.Println(utils.FormatTaskList(tasks, showDetails))
}

func listCompletedTasks(tm *manager.TaskManager, showDetails bool) {
	tasks := tm.GetCompletedTasks()
	fmt.Println("Completed Tasks:")
	fmt.Println(utils.FormatTaskList(tasks, showDetails))
}

func listPendingTasks(tm *manager.TaskManager, showDetails bool) {
	tasks := tm.GetPendingTasks()
	fmt.Println("Pending Tasks:")
	fmt.Println(utils.FormatTaskList(tasks, showDetails))
}

func listOverdueTasks(tm *manager.TaskManager, showDetails bool) {
	tasks := tm.GetOverdueTasks()
	fmt.Println("Overdue Tasks:")
	fmt.Println(utils.FormatTaskList(tasks, showDetails))
}

func showTaskDetail(tm *manager.TaskManager, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: Task ID is required for showing details.")
		return
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting task: %v\n", err)
		return
	}
	task, err := tm.GetTask(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting task: %v\n", err)
		return
	}

	fmt.Println(task.DetailString())
}

func completeTask(tm *manager.TaskManager, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: Task ID is required")
		return
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing task ID: %v\n", err)
		return
	}

	err = tm.MarkTaskComplete(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error completing task: %v\n", err)
		return
	}

	fmt.Printf("Task %d marked as completed.\n", id)
}

func uncompleteTask(tm *manager.TaskManager, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: Task ID is required")
		return
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing task ID: %v\n", err)
		return
	}

	err = tm.MarkTaskIncomplete(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error uncompleting task: %v\n", err)
		return
	}

	fmt.Printf("Task %d marked as uncompleted.\n", id)
}

func updateTask(tm *manager.TaskManager, args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Error: Task ID and new title are required for updating a task.")
		return
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing task ID: %v\n", err)
		return
	}

	title := args[1]
	description := ""
	if len(args) > 2 {
		description = strings.Join(args[2:], " ")
	}

	err = tm.UpdateTask(id, title, description)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error updating task: %v\n", err)
		return
	}

	fmt.Printf("Task updated: %d\n", id)
}

func deleteTask(tm *manager.TaskManager, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: Task ID is required for deleting a task.")
		return
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing task ID: %v\n", err)
		return
	}

	err = tm.DeleteTask(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting task: %v\n", err)
		return
	}

	fmt.Printf("Task %d deleted successfully.\n", id)
}

func setDueDate(tm *manager.TaskManager, args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Error: Task ID and due date are required for setting a due date.")
		return
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing task ID: %v\n", err)
		return
	}

	dueDate, err := utils.ParseDueDate(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing due date: %v\n", err)
		return
	}

	err = tm.SetTaskDueDate(id, dueDate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting due date: %v\n", err)
		return
	}

	fmt.Printf("Due date set for task %d.\n", id)
}

func setPriority(tm *manager.TaskManager, args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Error: Task ID and priority are required for setting a priority.")
		return
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing task ID: %v\n", err)
		return
	}

	priority, err := utils.ParsePriority(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing priority: %v\n", err)
		return
	}

	err = tm.SetTaskPriority(id, priority)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting priority: %v\n", err)
		return
	}

	fmt.Printf("Priority set for task %d.\n", id)
}

func showStats(tm *manager.TaskManager) {
	stats := tm.GetTaskStats()
	if stats == nil {
		fmt.Println("No tasks available to show statistics.")
		return
	}
	fmt.Println("Task Statistics:")
	fmt.Printf("Total Tasks: %d\n", stats["total"])
	fmt.Printf("Completed Tasks: %d\n", stats["completed"])
	fmt.Printf("Pending Tasks: %d\n", stats["pending"])
	fmt.Printf("Overdue Tasks: %d\n", stats["overdue"])
}

func printHelp() {
	fmt.Println("Task Manager - A simple CLI task management application")
	fmt.Println("\nUsage:")
	fmt.Println("  task_manager [options] [command] [arguments]")
	fmt.Println("\nOptions:")
	fmt.Println("  -file string    Storage file path (default \"tasks.json\")")
	fmt.Println("  -api            Run in API mode")
	fmt.Println("  -addr string    API server address (default \":8080\")")
	fmt.Println("  -help           Show help")
	fmt.Println("\nCommands (CLI mode only):")
	fmt.Println("  add, a <title> [description]       Add a new task")
	fmt.Println("  list, ls, l                        List pending tasks")
	fmt.Println("  list-all, la                       List all tasks")
	fmt.Println("  list-completed, lc                 List completed tasks")
	fmt.Println("  list-pending, lp                   List pending tasks")
	fmt.Println("  list-overdue, lo                   List overdue tasks")
	fmt.Println("  detail, d <id>                     Show task details")
	fmt.Println("  complete, c <id>                   Mark task as complete")
	fmt.Println("  uncomplete, uc <id>                Mark task as incomplete")
	fmt.Println("  update, u <id> <title> [desc]      Update task title and description")
	fmt.Println("  delete, del, rm <id>               Delete task")
	fmt.Println("  due, due-date <id> <YYYY-MM-DD>    Set task due date")
	fmt.Println("  priority, p <id> <1-5>             Set task priority (1=highest, 5=lowest)")
	fmt.Println("  stats, st                          Show task statistics")
	fmt.Println("  help, h                            Show help")
	fmt.Println("\nExamples:")
	fmt.Println("  task_manager add \"Buy groceries\" \"Milk, eggs, bread\"")
	fmt.Println("  task_manager list")
	fmt.Println("  task_manager complete 1")
	fmt.Println("  task_manager due 2 2023-12-31")
	fmt.Println("  task_manager priority 3 high")
	fmt.Println("  task_manager -api -addr :8080")
}
