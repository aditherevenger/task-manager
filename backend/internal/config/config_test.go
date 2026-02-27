package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Test with default values
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	if config == nil {
		t.Fatal("LoadConfig() returned nil config")
	}

	// Check default values
	if config.DBHost != "localhost" {
		t.Errorf("Expected DBHost to be 'localhost', got '%s'", config.DBHost)
	}
	if config.DBPort != "5432" {
		t.Errorf("Expected DBPort to be '5432', got '%s'", config.DBPort)
	}
	if config.DBUser != "postgres" {
		t.Errorf("Expected DBUser to be 'postgres', got '%s'", config.DBUser)
	}
	if config.DBName != "taskmanager" {
		t.Errorf("Expected DBName to be 'taskmanager', got '%s'", config.DBName)
	}
	if config.DBSSLMode != "disable" {
		t.Errorf("Expected DBSSLMode to be 'disable', got '%s'", config.DBSSLMode)
	}
	if config.APIAddr != ":8080" {
		t.Errorf("Expected APIAddr to be ':8080', got '%s'", config.APIAddr)
	}
}

func TestLoadConfigWithEnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("DB_HOST", "testhost")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_SSLMODE", "require")
	os.Setenv("API_ADDR", ":9090")

	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_SSLMODE")
		os.Unsetenv("API_ADDR")
	}()

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	// Check environment values
	if config.DBHost != "testhost" {
		t.Errorf("Expected DBHost to be 'testhost', got '%s'", config.DBHost)
	}
	if config.DBPort != "5433" {
		t.Errorf("Expected DBPort to be '5433', got '%s'", config.DBPort)
	}
	if config.DBUser != "testuser" {
		t.Errorf("Expected DBUser to be 'testuser', got '%s'", config.DBUser)
	}
	if config.DBPassword != "testpass" {
		t.Errorf("Expected DBPassword to be 'testpass', got '%s'", config.DBPassword)
	}
	if config.DBName != "testdb" {
		t.Errorf("Expected DBName to be 'testdb', got '%s'", config.DBName)
	}
	if config.DBSSLMode != "require" {
		t.Errorf("Expected DBSSLMode to be 'require', got '%s'", config.DBSSLMode)
	}
	if config.APIAddr != ":9090" {
		t.Errorf("Expected APIAddr to be ':9090', got '%s'", config.APIAddr)
	}
}

func TestGetEnv(t *testing.T) {
	// Test with no environment variable set
	result := getEnv("NONEXISTENT_VAR", "default")
	if result != "default" {
		t.Errorf("Expected 'default', got '%s'", result)
	}

	// Test with environment variable set
	os.Setenv("TEST_VAR", "value")
	defer os.Unsetenv("TEST_VAR")

	result = getEnv("TEST_VAR", "default")
	if result != "value" {
		t.Errorf("Expected 'value', got '%s'", result)
	}

	// Test with empty environment variable
	os.Setenv("EMPTY_VAR", "")
	defer os.Unsetenv("EMPTY_VAR")

	result = getEnv("EMPTY_VAR", "default")
	if result != "default" {
		t.Errorf("Expected 'default' for empty var, got '%s'", result)
	}
}
