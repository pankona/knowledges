package github_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pankona/knowledges/internal/github"
)

func TestGHWrapper_GetMergedPRs_Success(t *testing.T) {
	// Arrange
	mockJSON := `[
		{
			"number": 123,
			"title": "Add user authentication",
			"url": "https://github.com/owner/repo/pull/123",
			"createdAt": "2024-01-15T10:00:00Z",
			"author": {
				"login": "user1"
			}
		},
		{
			"number": 124,
			"title": "Fix database connection",
			"url": "https://github.com/owner/repo/pull/124",
			"createdAt": "2024-01-14T15:30:00Z",
			"author": {
				"login": "user2"
			}
		}
	]`

	mockExecutor := &MockCommandExecutor{
		output: mockJSON,
		err:    nil,
	}
	
	wrapper := github.NewGHWrapper("owner/repo")
	wrapper.SetExecutor(mockExecutor)

	// Act
	prs, err := wrapper.GetMergedPRs(context.Background(), 2)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(prs) != 2 {
		t.Errorf("expected 2 PRs, got %d", len(prs))
	}

	if prs[0].Number != 123 {
		t.Errorf("expected PR number 123, got %d", prs[0].Number)
	}
	if prs[0].Title != "Add user authentication" {
		t.Errorf("expected title 'Add user authentication', got %q", prs[0].Title)
	}
	if prs[0].Author.Login != "user1" {
		t.Errorf("expected author 'user1', got %q", prs[0].Author.Login)
	}

	// Verify command was called correctly
	expectedArgs := []string{
		"pr", "list",
		"--repo", "owner/repo",
		"--state", "merged",
		"--limit", "2",
		"--json", "number,title,url,createdAt,author",
	}
	if !equalStringSlices(mockExecutor.lastArgs, expectedArgs) {
		t.Errorf("expected args %v, got %v", expectedArgs, mockExecutor.lastArgs)
	}
}

func TestGHWrapper_GetMergedPRs_CommandError(t *testing.T) {
	// Arrange
	mockExecutor := &MockCommandExecutor{
		output: "",
		err:    fmt.Errorf("gh: command not found"),
	}
	
	wrapper := github.NewGHWrapper("owner/repo")
	wrapper.SetExecutor(mockExecutor)

	// Act
	_, err := wrapper.GetMergedPRs(context.Background(), 10)

	// Assert
	if err == nil {
		t.Error("expected error when gh command fails")
	}
}

func TestGHWrapper_GetMergedPRs_InvalidJSON(t *testing.T) {
	// Arrange
	invalidJSON := `[{invalid json`
	
	mockExecutor := &MockCommandExecutor{
		output: invalidJSON,
		err:    nil,
	}
	
	wrapper := github.NewGHWrapper("owner/repo")
	wrapper.SetExecutor(mockExecutor)

	// Act
	_, err := wrapper.GetMergedPRs(context.Background(), 10)

	// Assert
	if err == nil {
		t.Error("expected error when JSON is invalid")
	}
}

func TestGHWrapper_GetMergedPRs_EmptyResult(t *testing.T) {
	// Arrange
	emptyJSON := `[]`
	
	mockExecutor := &MockCommandExecutor{
		output: emptyJSON,
		err:    nil,
	}
	
	wrapper := github.NewGHWrapper("owner/repo")
	wrapper.SetExecutor(mockExecutor)

	// Act
	prs, err := wrapper.GetMergedPRs(context.Background(), 10)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prs) != 0 {
		t.Errorf("expected 0 PRs, got %d", len(prs))
	}
}

func TestGHWrapper_GetMergedPRsWithLabel_Success(t *testing.T) {
	// Arrange
	mockJSON := `[
		{
			"number": 123,
			"title": "Add user authentication",
			"url": "https://github.com/owner/repo/pull/123",
			"createdAt": "2024-01-15T10:00:00Z",
			"author": {
				"login": "user1"
			},
			"labels": [
				{"name": "payment-service"},
				{"name": "backend"}
			]
		}
	]`

	mockExecutor := &MockCommandExecutor{
		output: mockJSON,
		err:    nil,
	}
	
	wrapper := github.NewGHWrapper("owner/repo")
	wrapper.SetExecutor(mockExecutor)

	// Act
	prs, err := wrapper.GetMergedPRsWithLabel(context.Background(), 10, "payment-service")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(prs) != 1 {
		t.Errorf("expected 1 PR, got %d", len(prs))
	}

	// Verify command was called with label parameter
	if mockExecutor.lastCmd != "gh" {
		t.Errorf("expected command 'gh', got %q", mockExecutor.lastCmd)
	}
	
	// Check that label filter was included in args
	foundLabel := false
	for _, arg := range mockExecutor.lastArgs {
		if arg == "label:payment-service" {
			foundLabel = true
			break
		}
	}
	if !foundLabel {
		t.Errorf("expected label filter in args, got %v", mockExecutor.lastArgs)
	}
}

func TestGHWrapper_GetMergedPRsExcludingBots_Success(t *testing.T) {
	// Arrange
	mockJSON := `[
		{
			"number": 123,
			"title": "Add user authentication",
			"url": "https://github.com/owner/repo/pull/123",
			"createdAt": "2024-01-15T10:00:00Z",
			"author": {
				"login": "human-developer"
			}
		}
	]`

	mockExecutor := &MockCommandExecutor{
		output: mockJSON,
		err:    nil,
	}
	
	wrapper := github.NewGHWrapper("owner/repo")
	wrapper.SetExecutor(mockExecutor)

	// Act
	prs, err := wrapper.GetMergedPRsExcludingBots(context.Background(), 10, "payment-service")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(prs) != 1 {
		t.Errorf("expected 1 PR, got %d", len(prs))
	}

	// Verify command includes search argument with bot exclusion
	foundSearchArg := false
	for _, arg := range mockExecutor.lastArgs {
		if arg == "--search" {
			foundSearchArg = true
			break
		}
	}
	if !foundSearchArg {
		t.Errorf("expected --search argument, got %v", mockExecutor.lastArgs)
	}
	
	// Verify search contains bot exclusion
	foundExclude := false
	for _, arg := range mockExecutor.lastArgs {
		if arg == "label:payment-service -author:dependabot[bot] -author:github-actions[bot] -author:renovate[bot] -author:codecov[bot]" {
			foundExclude = true
			break
		}
	}
	if !foundExclude {
		t.Errorf("expected combined search string with bot exclusion, got %v", mockExecutor.lastArgs)
	}
}

func TestGHWrapper_GetPRComments_Success(t *testing.T) {
	// Arrange
	mockGraphQLResponse := `{
		"data": {
			"repository": {
				"pullRequest": {
					"reviewThreads": {
						"nodes": [
							{
								"path": "src/main.go",
								"line": 42,
								"comments": {
									"nodes": [
										{
											"author": { "login": "reviewer1" },
											"body": "Consider using a more descriptive variable name here.",
											"createdAt": "2024-01-15T10:00:00Z",
											"url": "https://github.com/owner/repo/pull/123#discussion_r1"
										},
										{
											"author": { "login": "author1" },
											"body": "Good point, I'll rename it to 'userRepository'.",
											"createdAt": "2024-01-15T10:05:00Z",
											"url": "https://github.com/owner/repo/pull/123#discussion_r2"
										}
									]
								}
							},
							{
								"path": "src/utils.go",
								"line": 15,
								"comments": {
									"nodes": [
										{
											"author": { "login": "reviewer2" },
											"body": "This function could benefit from error handling.",
											"createdAt": "2024-01-15T11:00:00Z",
											"url": "https://github.com/owner/repo/pull/123#discussion_r3"
										}
									]
								}
							}
						]
					}
				}
			}
		}
	}`

	mockExecutor := &MockCommandExecutor{
		output: mockGraphQLResponse,
		err:    nil,
	}
	
	wrapper := github.NewGHWrapper("owner/repo")
	wrapper.SetExecutor(mockExecutor)

	// Act
	comments, err := wrapper.GetPRComments(context.Background(), 123)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(comments) != 3 {
		t.Errorf("expected 3 comments, got %d", len(comments))
	}

	// Check first comment
	comment1 := comments[0]
	if comment1.Author.Login != "reviewer1" {
		t.Errorf("expected author 'reviewer1', got %q", comment1.Author.Login)
	}
	if comment1.Body != "Consider using a more descriptive variable name here." {
		t.Errorf("unexpected comment body: %q", comment1.Body)
	}
	if comment1.FilePath != "src/main.go" {
		t.Errorf("expected file path 'src/main.go', got %q", comment1.FilePath)
	}
	if comment1.LineNumber != 42 {
		t.Errorf("expected line number 42, got %d", comment1.LineNumber)
	}

	// Verify GraphQL command was called correctly
	if mockExecutor.lastCmd != "gh" {
		t.Errorf("expected command 'gh', got %q", mockExecutor.lastCmd)
	}
	expectedArgs := []string{"api", "graphql", "-f"}
	if len(mockExecutor.lastArgs) < 3 || !equalStringSlices(mockExecutor.lastArgs[:3], expectedArgs) {
		t.Errorf("expected args to start with %v, got %v", expectedArgs, mockExecutor.lastArgs[:3])
	}
}

func TestGHWrapper_GetPRComments_NoPRFound(t *testing.T) {
	// Arrange
	mockResponse := `{
		"data": {
			"repository": {
				"pullRequest": null
			}
		}
	}`

	mockExecutor := &MockCommandExecutor{
		output: mockResponse,
		err:    nil,
	}
	
	wrapper := github.NewGHWrapper("owner/repo")
	wrapper.SetExecutor(mockExecutor)

	// Act
	_, err := wrapper.GetPRComments(context.Background(), 999)

	// Assert
	if err == nil {
		t.Error("expected error when PR not found")
	}
}

func TestGHWrapper_GetPRComments_NoComments(t *testing.T) {
	// Arrange
	mockResponse := `{
		"data": {
			"repository": {
				"pullRequest": {
					"reviewThreads": {
						"nodes": []
					}
				}
			}
		}
	}`

	mockExecutor := &MockCommandExecutor{
		output: mockResponse,
		err:    nil,
	}
	
	wrapper := github.NewGHWrapper("owner/repo")
	wrapper.SetExecutor(mockExecutor)

	// Act
	comments, err := wrapper.GetPRComments(context.Background(), 123)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(comments) != 0 {
		t.Errorf("expected 0 comments, got %d", len(comments))
	}
}

func TestGHWrapper_GetPR_Success(t *testing.T) {
	// Arrange
	mockJSON := `{
		"number": 123,
		"title": "Add user authentication",
		"url": "https://github.com/owner/repo/pull/123",
		"createdAt": "2024-01-15T10:00:00Z",
		"author": {
			"login": "user1"
		}
	}`

	mockExecutor := &MockCommandExecutor{
		output: mockJSON,
		err:    nil,
	}
	
	wrapper := github.NewGHWrapper("owner/repo")
	wrapper.SetExecutor(mockExecutor)

	// Act
	pr, err := wrapper.GetPR(context.Background(), 123)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pr.Number != 123 {
		t.Errorf("expected PR number 123, got %d", pr.Number)
	}
	if pr.Title != "Add user authentication" {
		t.Errorf("expected title 'Add user authentication', got %q", pr.Title)
	}
	if pr.Author.Login != "user1" {
		t.Errorf("expected author 'user1', got %q", pr.Author.Login)
	}

	// Verify command was called correctly
	expectedArgs := []string{
		"pr", "view", "123",
		"--repo", "owner/repo",
		"--json", "number,title,url,createdAt,author",
	}
	if !equalStringSlices(mockExecutor.lastArgs, expectedArgs) {
		t.Errorf("expected args %v, got %v", expectedArgs, mockExecutor.lastArgs)
	}
}

func TestGHWrapper_GetPR_CommandError(t *testing.T) {
	// Arrange
	mockExecutor := &MockCommandExecutor{
		output: "",
		err:    fmt.Errorf("PR not found"),
	}
	
	wrapper := github.NewGHWrapper("owner/repo")
	wrapper.SetExecutor(mockExecutor)

	// Act
	_, err := wrapper.GetPR(context.Background(), 999)

	// Assert
	if err == nil {
		t.Error("expected error when PR not found")
	}
}

func TestParsePRCreatedAt(t *testing.T) {
	// Test time parsing
	timeStr := "2024-01-15T10:00:00Z"
	expected := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	
	parsed, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		t.Fatalf("failed to parse time: %v", err)
	}
	
	if !parsed.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, parsed)
	}
}

// MockCommandExecutor は外部コマンド実行をモックします
type MockCommandExecutor struct {
	output   string
	err      error
	lastCmd  string
	lastArgs []string
}

func (m *MockCommandExecutor) Execute(ctx context.Context, cmd string, args ...string) ([]byte, error) {
	m.lastCmd = cmd
	m.lastArgs = args
	return []byte(m.output), m.err
}

// Helper function to compare string slices
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}