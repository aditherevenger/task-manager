package storage

import (
	"database/sql"
	"fmt"
	"task-manager/internal/model"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNewPostgresStorage(t *testing.T) {
	// Test with invalid connection - should fail since DB doesn't exist
	_, err := NewPostgresStorage("invalid", "5432", "user", "pass", "db", "disable")
	if err == nil {
		t.Error("Expected error for invalid connection, got nil")
	}
}

func TestInitSchema(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	storage := &PostgresStorage{db: db}

	// Mock the schema creation
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS tasks").WillReturnResult(sqlmock.NewResult(0, 0))

	err = storage.InitSchema()
	if err != nil {
		t.Errorf("InitSchema() failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestInitSchemaError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	storage := &PostgresStorage{db: db}

	// Mock schema creation error
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS tasks").WillReturnError(fmt.Errorf("schema error"))

	err = storage.InitSchema()
	if err == nil {
		t.Error("Expected error from InitSchema, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestSave(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	storage := &PostgresStorage{db: db}

	// Create test tasks
	now := time.Now()
	tasks := []*model.Task{
		{
			ID:          1,
			Title:       "Test Task 1",
			Description: "Description 1",
			Completed:   false,
			CreatedAt:   now,
			Priority:    3,
		},
		{
			ID:          2,
			Title:       "Test Task 2",
			Description: "Description 2",
			Completed:   true,
			CreatedAt:   now,
			CompletedAt: now,
			DueDate:     now.Add(24 * time.Hour),
			Priority:    5,
		},
	}

	// Mock transaction
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM tasks").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO tasks").WithArgs(
		1, "Test Task 1", "Description 1", false, now, nil, nil, 3,
	).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO tasks").WithArgs(
		2, "Test Task 2", "Description 2", true, now, now, tasks[1].DueDate, 5,
	).WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectExec("SELECT setval").WithArgs(2).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = storage.Save(tasks)
	if err != nil {
		t.Errorf("Save() failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestSaveEmptyTasks(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	storage := &PostgresStorage{db: db}

	// Mock transaction for empty tasks
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM tasks").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err = storage.Save([]*model.Task{})
	if err != nil {
		t.Errorf("Save() failed for empty tasks: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestSaveBeginError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	storage := &PostgresStorage{db: db}

	// Mock begin error
	mock.ExpectBegin().WillReturnError(fmt.Errorf("begin error"))

	err = storage.Save([]*model.Task{})
	if err == nil {
		t.Error("Expected error from Save, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestSaveDeleteError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	storage := &PostgresStorage{db: db}

	// Mock delete error
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM tasks").WillReturnError(fmt.Errorf("delete error"))
	mock.ExpectRollback()

	err = storage.Save([]*model.Task{})
	if err == nil {
		t.Error("Expected error from Save, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestSaveInsertError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	storage := &PostgresStorage{db: db}

	task := &model.Task{
		ID:          1,
		Title:       "Test",
		Description: "Test",
		CreatedAt:   time.Now(),
		Priority:    3,
	}

	// Mock insert error
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM tasks").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO tasks").WillReturnError(fmt.Errorf("insert error"))
	mock.ExpectRollback()

	err = storage.Save([]*model.Task{task})
	if err == nil {
		t.Error("Expected error from Save, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestSaveSetvalError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	storage := &PostgresStorage{db: db}

	task := &model.Task{
		ID:          1,
		Title:       "Test",
		Description: "Test",
		CreatedAt:   time.Now(),
		Priority:    3,
	}

	// Mock setval error
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM tasks").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO tasks").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("SELECT setval").WillReturnError(fmt.Errorf("setval error"))
	mock.ExpectRollback()

	err = storage.Save([]*model.Task{task})
	if err == nil {
		t.Error("Expected error from Save, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestSaveCommitError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	storage := &PostgresStorage{db: db}

	task := &model.Task{
		ID:          1,
		Title:       "Test",
		Description: "Test",
		CreatedAt:   time.Now(),
		Priority:    3,
	}

	// Mock commit error
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM tasks").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO tasks").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("SELECT setval").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit().WillReturnError(fmt.Errorf("commit error"))

	err = storage.Save([]*model.Task{task})
	if err == nil {
		t.Error("Expected error from Save, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestLoad(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	storage := &PostgresStorage{db: db}

	// Create expected rows
	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "title", "description", "completed", "created_at", "completed_at", "due_date", "priority",
	}).AddRow(
		1, "Task 1", "Desc 1", false, now, sql.NullTime{}, sql.NullTime{}, 3,
	).AddRow(
		2, "Task 2", "Desc 2", true, now, sql.NullTime{Valid: true, Time: now}, sql.NullTime{Valid: true, Time: now}, 5,
	)

	mock.ExpectQuery("SELECT (.+) FROM tasks").WillReturnRows(rows)

	tasks, err := storage.Load()
	if err != nil {
		t.Errorf("Load() failed: %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(tasks))
	}

	// Check first task
	if tasks[0].ID != 1 || tasks[0].Title != "Task 1" {
		t.Error("First task data incorrect")
	}

	// Check second task
	if tasks[1].ID != 2 || tasks[1].Title != "Task 2" {
		t.Error("Second task data incorrect")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestLoadQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	storage := &PostgresStorage{db: db}

	// Mock query error
	mock.ExpectQuery("SELECT (.+) FROM tasks").WillReturnError(fmt.Errorf("query error"))

	_, err = storage.Load()
	if err == nil {
		t.Error("Expected error from Load, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestLoadScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	storage := &PostgresStorage{db: db}

	// Create rows with incorrect data types to cause scan error
	rows := sqlmock.NewRows([]string{
		"id", "title", "description", "completed", "created_at", "completed_at", "due_date", "priority",
	}).AddRow(
		"invalid", "Task 1", "Desc 1", false, time.Now(), sql.NullTime{}, sql.NullTime{}, 3,
	)

	mock.ExpectQuery("SELECT (.+) FROM tasks").WillReturnRows(rows)

	_, err = storage.Load()
	if err == nil {
		t.Error("Expected error from Load with invalid data, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestLoadRowsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	storage := &PostgresStorage{db: db}

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "title", "description", "completed", "created_at", "completed_at", "due_date", "priority",
	}).AddRow(
		1, "Task 1", "Desc 1", false, now, sql.NullTime{}, sql.NullTime{}, 3,
	).RowError(0, fmt.Errorf("row error"))

	mock.ExpectQuery("SELECT (.+) FROM tasks").WillReturnRows(rows)

	_, err = storage.Load()
	if err == nil {
		t.Error("Expected error from Load with row error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestClose(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}

	storage := &PostgresStorage{db: db}

	mock.ExpectClose()

	err = storage.Close()
	if err != nil {
		t.Errorf("Close() failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}
