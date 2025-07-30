package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pankona/knowledges/pkg/config"
)

func TestLoad_MinimalConfig(t *testing.T) {
	// Arrange
	configYAML := `
github:
  repositories:
    - owner/repo1
database:
  path: ./test.db
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatal(err)
	}

	// Act
	cfg, err := config.Load(configPath)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.GitHub.Repositories) != 1 {
		t.Errorf("expected 1 repository, got %d", len(cfg.GitHub.Repositories))
	}
	if cfg.GitHub.Repositories[0] != "owner/repo1" {
		t.Errorf("expected 'owner/repo1', got %q", cfg.GitHub.Repositories[0])
	}
	if cfg.Database.Path != "./test.db" {
		t.Errorf("expected './test.db', got %q", cfg.Database.Path)
	}
}

func TestLoad_FullConfig(t *testing.T) {
	// Arrange
	configYAML := `
github:
  repositories:
    - owner/repo1
    - owner/repo2

llm:
  primary: claude
  parallel: 5
  retry:
    max_attempts: 3
    initial_delay: 1s
    max_delay: 10s
  drivers:
    claude:
      command: claude
      args: [-p]

database:
  path: ./knowledge.db

collection:
  batch_size: 10
  max_prs_per_run: 100

server:
  port: 8080
  read_timeout: 30
  write_timeout: 30
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatal(err)
	}

	// Act
	cfg, err := config.Load(configPath)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// GitHub
	if len(cfg.GitHub.Repositories) != 2 {
		t.Errorf("expected 2 repositories, got %d", len(cfg.GitHub.Repositories))
	}

	// LLM
	if cfg.LLM.Primary != "claude" {
		t.Errorf("expected primary LLM 'claude', got %q", cfg.LLM.Primary)
	}
	if cfg.LLM.Parallel != 5 {
		t.Errorf("expected parallel 5, got %d", cfg.LLM.Parallel)
	}
	if cfg.LLM.Retry.MaxAttempts != 3 {
		t.Errorf("expected max attempts 3, got %d", cfg.LLM.Retry.MaxAttempts)
	}

	// Collection
	if cfg.Collection.BatchSize != 10 {
		t.Errorf("expected batch size 10, got %d", cfg.Collection.BatchSize)
	}
	if cfg.Collection.MaxPRsPerRun != 100 {
		t.Errorf("expected max PRs 100, got %d", cfg.Collection.MaxPRsPerRun)
	}

	// Server
	if cfg.Server.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Server.Port)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	// Act
	_, err := config.Load("non-existent-file.yaml")

	// Assert
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	// Arrange
	invalidYAML := `
github:
  repositories
    - invalid yaml syntax
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatal(err)
	}

	// Act
	_, err := config.Load(configPath)

	// Assert
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoad_DefaultValues(t *testing.T) {
	// Arrange
	configYAML := `
github:
  repositories:
    - owner/repo
database:
  path: ./test.db
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatal(err)
	}

	// Act
	cfg, err := config.Load(configPath)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check default values
	if cfg.LLM.Primary != "claude" {
		t.Errorf("expected default primary LLM 'claude', got %q", cfg.LLM.Primary)
	}
	if cfg.LLM.Parallel != 3 {
		t.Errorf("expected default parallel 3, got %d", cfg.LLM.Parallel)
	}
	if cfg.Collection.BatchSize != 5 {
		t.Errorf("expected default batch size 5, got %d", cfg.Collection.BatchSize)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Server.Port)
	}
}