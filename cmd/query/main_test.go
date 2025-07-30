package main

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/pankona/knowledges/internal/database"
	"github.com/pankona/knowledges/pkg/models"
)

func TestQueryWithDirectoryFilter_Success(t *testing.T) {
	// Arrange
	dbPath := "test_query.db"
	defer os.Remove(dbPath)

	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// Insert test data
	now := time.Now()
	testDocs := []*models.Document{
		{
			Summary:         "Payment validation issue",
			OriginalComment: "This payment logic needs validation",
			FilePath:        "payment-service/app/models/payment.rb",
			DirectoryPath:   "payment-service/app/models",
			Language:        "ruby",
			Repository:      "example-org/payment-system",
			PRNumber:        123,
			PRTitle:         "Add payment validation",
			PRURL:           "https://github.com/example-org/payment-system/pull/123",
			CommentURL:      "https://github.com/example-org/payment-system/pull/123#discussion_r1",
			Author:          "reviewer1",
			CommentType:     "security",
			RelevanceScore:  0.9,
			CommentedAt:     now,
			CollectedAt:     now,
			UpdatedAt:       now,
		},
		{
			Summary:         "React component optimization",
			OriginalComment: "This component could be optimized",
			FilePath:        "frontend/src/components/Quiz.tsx",
			DirectoryPath:   "frontend/src/components",
			Language:        "typescript",
			Repository:      "example-org/frontend-app",
			PRNumber:        124,
			PRTitle:         "Optimize Quiz component",
			PRURL:           "https://github.com/example-org/frontend-app/pull/124",
			CommentURL:      "https://github.com/example-org/frontend-app/pull/124#discussion_r2",
			Author:          "reviewer2",
			CommentType:     "performance",
			RelevanceScore:  0.8,
			CommentedAt:     now,
			CollectedAt:     now,
			UpdatedAt:       now,
		},
	}

	for _, doc := range testDocs {
		err := insertTestDocument(db, doc)
		if err != nil {
			t.Fatalf("Failed to insert test document: %v", err)
		}
	}

	// Act & Assert - Test that we can query documents by directory
	// This would ideally be done by creating a separate query function
	// For now, we verify data exists (will implement actual query logic next)
	
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM documents WHERE directory_path LIKE ?", "%payment-service%").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query by directory: %v", err)
	}
	
	if count != 1 {
		t.Errorf("Expected 1 document in payment-service directory, got %d", count)
	}
}

func TestQueryWithFileFilter_Success(t *testing.T) {
	// Test file pattern filtering
	t.Skip("Test implementation pending - TDD Red phase")
}

func TestQueryWithAuthorFilter_Success(t *testing.T) {
	// Test author filtering
	t.Skip("Test implementation pending - TDD Red phase")
}

func TestQueryWithKeywordSearch_Success(t *testing.T) {
	// Test keyword search in summary and comment text
	t.Skip("Test implementation pending - TDD Red phase")
}

func TestQueryWithMultipleFilters_Success(t *testing.T) {
	// Test combining multiple filters
	t.Skip("Test implementation pending - TDD Red phase")
}

func TestQueryWithNoResults_ShowsHelpMessage(t *testing.T) {
	// Test showing helpful message when no results found
	t.Skip("Test implementation pending - TDD Red phase")
}

// Helper function to insert test documents
func insertTestDocument(db *sql.DB, doc *models.Document) error {
	query := `
	INSERT INTO documents (
		summary, original_comment, file_path, directory_path, language,
		repository, pr_number, pr_title, pr_url, comment_url,
		author, comment_type, relevance_score, commented_at, collected_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := db.Exec(query,
		doc.Summary, doc.OriginalComment, doc.FilePath, doc.DirectoryPath, doc.Language,
		doc.Repository, doc.PRNumber, doc.PRTitle, doc.PRURL, doc.CommentURL,
		doc.Author, doc.CommentType, doc.RelevanceScore, doc.CommentedAt, doc.CollectedAt, doc.UpdatedAt)
	
	return err
}