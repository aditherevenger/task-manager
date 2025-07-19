# Task Manager

![Go Version](https://img.shields.io/badge/Go-1.24-blue )
![License](https://img.shields.io/badge/License-MIT-green )
![Test Coverage](https://img.shields.io/badge/Coverage-80%2B%25-brightgreen )
![Build Status](https://img.shields.io/github/actions/workflow/status/yourusername/task-manager/go.yml?branch=master )

A robust task management application built with Go, featuring both CLI and REST API interfaces. Designed with clean architecture principles and best practices for Go development.

## 📋 Features

- **Task Management**: Create, read, update, and delete tasks
- **Status Tracking**: Mark tasks as complete or incomplete
- **Due Dates**: Set and track task deadlines
- **Priority Levels**: Assign importance to tasks (highest to lowest)
- **Filtering**: View tasks by status (all, pending, completed, overdue)
- **Statistics**: Get insights into task completion and status
- **Dual Interface**: Use as CLI tool or REST API server
- **Persistent Storage**: Tasks saved to JSON file
- **Comprehensive Testing**: 80%+ code coverage with automated enforcement
- **Security Scanning**: CodeQL integration for vulnerability detection
- **Dependency Monitoring**: Automated dependency review for security issues

## 🚀 Installation

### Prerequisites

- Go 1.24 or higher

### From Source

```bash
# Clone the repository
git clone https://github.com/yourusername/task-manager.git
cd task-manager

# Build the application
make build

# Run tests with coverage
make test
```

## 🖥️ Usage
CLI Mode

```bash
task_manager [options] [command] [arguments]
```

## Options
- **file string:** Storage file path (default "tasks.json" )
- **api:** Run in API mode
- **addr string:** API server address (default ":8080")
- **help:** Show help

## ⌨️ Commands

| Command          | Alias         | Description                    | Example                                           |
|------------------|---------------|--------------------------------|---------------------------------------------------|
| **add ➕**        | `a`           | Add a new task                 | `task-manager add "Buy groceries" "Milk, eggs, bread"` |
| **list 📜**       | `ls`, `l`     | List pending tasks             | `task-manager list`                               |
| **list-all 📋**   | `la`          | List all tasks                 | `task-manager list-all`                           |
| **list-completed ✅** | `lc`      | List completed tasks           | `task-manager list-completed`                     |
| **list-pending ⏳**   | `lp`      | List pending tasks             | `task-manager list-pending`                       |
| **list-overdue ⚠️**   | `lo`      | List overdue tasks             | `task-manager list-overdue`                       |
| **detail 🔍**     | `d`           | Show task details              | `task-manager detail 1`                           |
| **complete ✔️**   | `c`           | Mark task as complete          | `task-manager complete 1`                         |
| **uncomplete ↩️** | `uc`          | Mark task as incomplete        | `task-manager uncomplete 1`                       |
| **update ✏️**     | `u`           | Update task title/description  | `task-manager update 1 "New title" "New description"` |
| **delete 🗑️**     | `del`, `rm`   | Delete a task                  | `task-manager delete 1`                           |
| **due 📅**        | `due-date`    | Set task due date              | `task-manager due 1 2023-12-31`                   |
| **priority 🔝**   | `p`           | Set task priority              | `task-manager priority 1 high`                    |
| **stats 📊**      | `st`          | Show task statistics           | `task-manager stats`                              |
| **help ❓**       | `h`           | Show help                      | `task-manager help`                               |


```bash
  # Add a new task
task_manager add "Complete project documentation" "Add usage examples and API documentation"

# List all pending tasks
task_manager list

# Mark a task as complete
task_manager complete 1

# Set a due date
task_manager due 2 2023-12-31

# Set priority
task_manager priority 3 high

# View task statistics
task_manager stats
```

## API Mode
Start the API server:

```bash
task_manager -api -addr :8080
```

## 📡 API Endpoints

| Method  | Endpoint                         | Description                |
|---------|---------------------------------|----------------------------|
| **GET**    | `/api/v1/tasks`                | List all tasks             |
| **GET**    | `/api/v1/tasks/:id`            | Get a specific task        |
| **POST**   | `/api/v1/tasks`                | Create a new task          |
| **PUT**    | `/api/v1/tasks/:id`            | Update a task              |
| **DELETE** | `/api/v1/tasks/:id`            | Delete a task              |
| **PATCH**  | `/api/v1/tasks/:id/complete`   | Mark a task as complete    |
| **PATCH**  | `/api/v1/tasks/:id/uncomplete` | Mark a task as incomplete  |
| **PATCH**  | `/api/v1/tasks/:id/due-date`   | Set a task's due date      |
| **PATCH**  | `/api/v1/tasks/:id/priority`   | Set a task's priority      |
| **GET**    | `/api/v1/stats`                | Get task statistics        |

## 📐 Architecture
The project follows clean architecture principles with clear separation of concerns:

```bash
task-manager/
├── task_manager/
│   ├── main.go                      # Main application entry point (CLI & API startup logic)
│   └── main_test.go                 # Tests for CLI entry point and flags
├── internal/
│   ├── api/                         # API-related code
│   │   ├── handlers/                # HTTP request handlers (routes implementation)
│   │   │   ├── task_handler.go      # Handlers for task CRUD operations
│   │   │   └── task_handler_test.go # Unit tests for task handlers
│   │   ├── middleware/              # HTTP middleware logic
│   │   │   ├── logger.go            # Logging middleware for requests/responses
│   │   │   └── logger_test.go       # Tests for logging middleware
│   │   ├── server.go                # API server setup and route registration
│   │   └── server_test.go           # Tests for API server setup and routes
│   ├── manager/                     # Business logic (core task management)
│   │   ├── task_manager.go          # Core logic for managing tasks (CRUD, filters, stats)
│   │   └── task_manager_test.go     # Tests for business logic
│   ├── model/                       # Domain models (data structures)
│   │   ├── task.go                  # Task model (fields: ID, Title, DueDate, etc.)
│   │   └── task_test.go             # Tests for task model (validation, struct checks)
│   └── storage/                     # Data persistence layer
│       ├── storage.go               # Logic to save/load tasks (e.g., JSON, file operations)
│       └── storage_test.go          # Tests for storage persistence logic
└── pkg/
    └── utils/                       # Reusable utility functions
        ├── utils.go                 # Helper functions (date parsing, string utilities)
        └── utils_test.go            # Tests for utility functions
```

## 🧪 Testing
The project maintains a high standard of code quality with comprehensive test coverage:

```bash
# Run all tests with coverage report
make test

# View detailed coverage report
go tool cover -html=coverage.out
```
Our CI pipeline enforces a minimum of 80% test coverage across the codebase.

## 🔒 Security
This project uses multiple security tools to ensure code quality and security:

- **CodeQL Analysis:** Advanced static analysis to detect security vulnerabilities
- **Dependency Review:** Automated scanning of dependencies for known vulnerabilities
- **Supply Chain Security:** Monitoring and enforcement of secure dependencies

## 🛠️ Development
Makefile Commands

The project includes a comprehensive Makefile to simplify development tasks:
```bash
make build      # Build the application
make run        # Run the application
make test       # Run tests with coverage
make clean      # Clean build artifacts
make fmt        # Format code
make deps       # Download dependencies
make tidy       # Tidy go.mod file
```

CI/CD Pipeline
The project uses GitHub Actions for continuous integration and deployment:
- **Build & Test:** Automatically builds and tests the code on each push
- **Coverage Enforcement:** Ensures test coverage remains above 80%
- **CodeQL Analysis:** Performs security scanning
- **Dependency Review:** Checks for vulnerable dependencies

## 🤝 Contributing
Contributions are welcome! Please feel free to submit a Pull Request.

- Fork the repository
- Create your feature branch (git checkout -b feature/amazing-feature)
- Commit your changes (git commit -m 'Add some amazing feature')
- Push to the branch (git push origin feature/amazing-feature)
- Open a Pull Request

Please make sure your code passes all tests and follows the project's coding style.

## 📄 License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details

## 📊 Project Status
This project is actively maintained and in development. New features and improvements are being added regularly.
