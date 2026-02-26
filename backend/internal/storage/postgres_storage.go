package storage

import (
	"database/sql"
	"fmt"
	"task-manager/internal/model"
	"time"

	_ "github.com/lib/pq"
)

// TaskStorage interface defines methods for task persistence
type TaskStorage interface {
	// Save saves all tasks to storage
	Save(tasks []*model.Task) error

	// Load loads all tasks from storage
	Load() ([]*model.Task, error)
}

// PostgresStorage implements TaskStorage interface using PostgreSQL
type PostgresStorage struct {
	db *sql.DB
}

// NewPostgresStorage creates a new PostgresStorage instance
func NewPostgresStorage(host, port, user, password, dbname, sslmode string) (*PostgresStorage, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storage := &PostgresStorage{db: db}

	// Initialize the schema
	if err := storage.InitSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return storage, nil
}

// InitSchema creates the tasks table if it doesn't exist
func (p *PostgresStorage) InitSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS tasks (
		id SERIAL PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		description TEXT,
		completed BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		completed_at TIMESTAMP,
		due_date TIMESTAMP,
		priority INTEGER DEFAULT 3 CHECK (priority >= 1 AND priority <= 5)
	);
	`

	_, err := p.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

// Save saves all tasks to the database
func (p *PostgresStorage) Save(tasks []*model.Task) error {
	tx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Clear existing tasks
	_, err = tx.Exec("DELETE FROM tasks")
	if err != nil {
		return fmt.Errorf("failed to clear tasks: %w", err)
	}

	// Insert all tasks
	for _, task := range tasks {
		var completedAt interface{}
		if !task.CompletedAt.IsZero() {
			completedAt = task.CompletedAt
		}

		var dueDate interface{}
		if !task.DueDate.IsZero() {
			dueDate = task.DueDate
		}

		_, err := tx.Exec(`
			INSERT INTO tasks (id, title, description, completed, created_at, completed_at, due_date, priority)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, task.ID, task.Title, task.Description, task.Completed, task.CreatedAt, completedAt, dueDate, task.Priority)

		if err != nil {
			return fmt.Errorf("failed to insert task: %w", err)
		}
	}

	// Update the sequence to match the highest ID
	if len(tasks) > 0 {
		maxID := 0
		for _, task := range tasks {
			if task.ID > maxID {
				maxID = task.ID
			}
		}
		_, err = tx.Exec("SELECT setval('tasks_id_seq', $1, true)", maxID)
		if err != nil {
			return fmt.Errorf("failed to update sequence: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Load loads all tasks from the database
func (p *PostgresStorage) Load() ([]*model.Task, error) {
	rows, err := p.db.Query(`
		SELECT id, title, description, completed, created_at, completed_at, due_date, priority
		FROM tasks
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*model.Task
	for rows.Next() {
		task := &model.Task{}
		var completedAt, dueDate sql.NullTime

		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Completed,
			&task.CreatedAt,
			&completedAt,
			&dueDate,
			&task.Priority,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		if completedAt.Valid {
			task.CompletedAt = completedAt.Time
		} else {
			task.CompletedAt = time.Time{}
		}

		if dueDate.Valid {
			task.DueDate = dueDate.Time
		} else {
			task.DueDate = time.Time{}
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return tasks, nil
}

// Close closes the database connection
func (p *PostgresStorage) Close() error {
	return p.db.Close()
}
