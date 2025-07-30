package database_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/pankona/knowledges/internal/database"
)

func TestNew_CreateDatabase(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Act
	db, err := database.New(dbPath)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer db.Close()

	// Verify database connection
	if err := db.Ping(); err != nil {
		t.Errorf("failed to ping database: %v", err)
	}
}

func TestNew_ExistingDatabase(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create database first
	db1, err := database.New(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	db1.Close()

	// Act - open existing database
	db2, err := database.New(dbPath)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer db2.Close()

	if err := db2.Ping(); err != nil {
		t.Errorf("failed to ping existing database: %v", err)
	}
}

func TestMigrate_InitialMigration(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Act
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	// Assert - verify documents table exists
	ctx := context.Background()
	var tableName string
	query := `SELECT name FROM sqlite_master WHERE type='table' AND name='documents'`
	err = db.QueryRowContext(ctx, query).Scan(&tableName)
	if err != nil {
		t.Fatalf("documents table not found: %v", err)
	}
	if tableName != "documents" {
		t.Errorf("expected table name 'documents', got %q", tableName)
	}
}

func TestMigrate_TableStructure(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Act
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	// Assert - verify key columns exist
	ctx := context.Background()
	columns := []string{"id", "summary", "original_comment", "file_path", "repository", "pr_number"}
	
	for _, column := range columns {
		query := `SELECT COUNT(*) FROM pragma_table_info('documents') WHERE name = ?`
		var count int
		err := db.QueryRowContext(ctx, query, column).Scan(&count)
		if err != nil {
			t.Fatalf("failed to check column %s: %v", column, err)
		}
		if count != 1 {
			t.Errorf("column %s not found", column)
		}
	}
}

func TestMigrate_Idempotent(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Act - run migration twice
	if err := database.Migrate(db); err != nil {
		t.Fatalf("first migration failed: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("second migration failed: %v", err)
	}

	// Assert - no error should occur
	ctx := context.Background()
	var count int
	query := `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='documents'`
	err = db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		t.Fatalf("failed to count tables: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 documents table, got %d", count)
	}
}