package main

import (
	"context"
	"os"
	"testing"

	"github.com/pankona/knowledges/internal/database"
	"github.com/pankona/knowledges/internal/github"
	"github.com/pankona/knowledges/pkg/models"
)

func TestGetProcessedPRNumbers_Success(t *testing.T) {
	// Arrange
	dbPath := "test_processed_prs.db"
	defer os.Remove(dbPath)

	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// Insert test documents for different PRs
	testDocs := []*models.Document{
		{
			Summary:         "Test comment 1",
			OriginalComment: "Test comment",
			FilePath:        "test.go",
			DirectoryPath:   ".",
			Language:        "go",
			Repository:      "owner/repo",
			PRNumber:        123,
			PRTitle:         "Test PR 123",
			PRURL:           "https://github.com/owner/repo/pull/123",
			CommentURL:      "https://github.com/owner/repo/pull/123#comment1",
			Author:          "reviewer1",
			CommentType:     "implementation",
			RelevanceScore:  0.8,
		},
		{
			Summary:         "Test comment 2",
			OriginalComment: "Another test comment",
			FilePath:        "test2.go",
			DirectoryPath:   ".",
			Language:        "go",
			Repository:      "owner/repo",
			PRNumber:        124,
			PRTitle:         "Test PR 124",
			PRURL:           "https://github.com/owner/repo/pull/124",
			CommentURL:      "https://github.com/owner/repo/pull/124#comment1",
			Author:          "reviewer2",
			CommentType:     "security",
			RelevanceScore:  0.9,
		},
		{
			Summary:         "Test comment 3",
			OriginalComment: "Third test comment",
			FilePath:        "test3.go",
			DirectoryPath:   ".",
			Language:        "go",
			Repository:      "owner/repo",
			PRNumber:        123, // Same PR as first document
			PRTitle:         "Test PR 123",
			PRURL:           "https://github.com/owner/repo/pull/123",
			CommentURL:      "https://github.com/owner/repo/pull/123#comment2",
			Author:          "reviewer3",
			CommentType:     "testing",
			RelevanceScore:  0.7,
		},
	}

	ctx := context.Background()
	for _, doc := range testDocs {
		err := saveDocument(ctx, db, doc)
		if err != nil {
			t.Fatalf("Failed to save test document: %v", err)
		}
	}

	// Act
	processedPRs, err := getProcessedPRNumbers(ctx, db, "owner/repo")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedPRs := map[int]bool{123: true, 124: true}
	if len(processedPRs) != len(expectedPRs) {
		t.Errorf("expected %d processed PRs, got %d", len(expectedPRs), len(processedPRs))
	}

	for prNumber := range expectedPRs {
		if !processedPRs[prNumber] {
			t.Errorf("expected PR #%d to be in processed list", prNumber)
		}
	}
}

func TestFilterUnprocessedPRs_Success(t *testing.T) {
	// Arrange
	allPRs := []github.PullRequest{
		{Number: 100, Title: "PR 100"},
		{Number: 101, Title: "PR 101"},
		{Number: 102, Title: "PR 102"},
		{Number: 103, Title: "PR 103"},
	}

	processedPRs := map[int]bool{
		101: true,
		103: true,
	}

	// Act
	unprocessedPRs := filterUnprocessedPRs(allPRs, processedPRs)

	// Assert
	expectedCount := 2 // PRs 100 and 102 should remain
	if len(unprocessedPRs) != expectedCount {
		t.Errorf("expected %d unprocessed PRs, got %d", expectedCount, len(unprocessedPRs))
	}

	expectedNumbers := []int{100, 102}
	for i, pr := range unprocessedPRs {
		if pr.Number != expectedNumbers[i] {
			t.Errorf("expected PR #%d at index %d, got #%d", expectedNumbers[i], i, pr.Number)
		}
	}
}

func TestDeletePRData_Success(t *testing.T) {
	// Arrange
	dbPath := "test_delete_pr.db"
	defer os.Remove(dbPath)

	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// Insert test documents
	testDocs := []*models.Document{
		{
			Summary:         "Comment for PR 123",
			OriginalComment: "Test comment",
			FilePath:        "test.go",
			DirectoryPath:   ".",
			Language:        "go",
			Repository:      "owner/repo",
			PRNumber:        123,
			PRTitle:         "Test PR 123",
			PRURL:           "https://github.com/owner/repo/pull/123",
			CommentURL:      "https://github.com/owner/repo/pull/123#comment1",
			Author:          "reviewer1",
			CommentType:     "implementation",
			RelevanceScore:  0.8,
		},
		{
			Summary:         "Comment for PR 124",
			OriginalComment: "Another test comment",
			FilePath:        "test2.go",
			DirectoryPath:   ".",
			Language:        "go",
			Repository:      "owner/repo",
			PRNumber:        124,
			PRTitle:         "Test PR 124",
			PRURL:           "https://github.com/owner/repo/pull/124",
			CommentURL:      "https://github.com/owner/repo/pull/124#comment1",
			Author:          "reviewer2",
			CommentType:     "security",
			RelevanceScore:  0.9,
		},
	}

	ctx := context.Background()
	for _, doc := range testDocs {
		err := saveDocument(ctx, db, doc)
		if err != nil {
			t.Fatalf("Failed to save test document: %v", err)
		}
	}

	// Act - Delete PR 123 data
	err = deletePRData(ctx, db, "owner/repo", 123)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify PR 123 data is deleted
	var count123 int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM documents WHERE repository = ? AND pr_number = ?", 
		"owner/repo", 123).Scan(&count123)
	if err != nil {
		t.Fatalf("Failed to query documents: %v", err)
	}
	if count123 != 0 {
		t.Errorf("expected 0 documents for PR 123 after deletion, got %d", count123)
	}

	// Verify PR 124 data still exists
	var count124 int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM documents WHERE repository = ? AND pr_number = ?", 
		"owner/repo", 124).Scan(&count124)
	if err != nil {
		t.Fatalf("Failed to query documents: %v", err)
	}
	if count124 != 1 {
		t.Errorf("expected 1 document for PR 124 after deletion, got %d", count124)
	}
}