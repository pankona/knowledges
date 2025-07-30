package llm

import (
	"context"
	"errors"
	"testing"
)

// Mock executor for testing
type MockCommandExecutor struct {
	output []byte
	err    error
}

func (m *MockCommandExecutor) Execute(ctx context.Context, cmd string, args []string, input string) ([]byte, error) {
	return m.output, m.err
}

func TestExtractJSON_WithCodeBlock(t *testing.T) {
	input := []byte("```json\n{\n  \"summary\": \"test\",\n  \"type\": \"bug\"\n}\n```")
	expected := `{
  "summary": "test",
  "type": "bug"
}`
	
	result := extractJSON(input)
	if string(result) != expected {
		t.Errorf("Expected %s, got %s", expected, string(result))
	}
}

func TestExtractJSON_WithoutCodeBlock(t *testing.T) {
	input := []byte(`{"summary": "test", "type": "bug"}`)
	expected := `{"summary": "test", "type": "bug"}`
	
	result := extractJSON(input)
	if string(result) != expected {
		t.Errorf("Expected %s, got %s", expected, string(result))
	}
}

func TestExtractJSON_WithMarkdownJsonBlock(t *testing.T) {
	input := []byte("```json\n{\"summary\": \"Feature flag check\", \"type\": \"suggestion\"}\n```")
	expected := `{"summary": "Feature flag check", "type": "suggestion"}`
	
	result := extractJSON(input)
	if string(result) != expected {
		t.Errorf("Expected %s, got %s", expected, string(result))
	}
}

func TestAnalyzeComment_Success(t *testing.T) {
	// Arrange
	mockExecutor := &MockCommandExecutor{
		output: []byte(`{"summary": "Test summary", "type": "bug", "tags": ["test"], "relevance_score": 0.8}`),
		err:    nil,
	}
	
	driver := NewDriver("claude", []string{"-p"})
	driver.SetExecutor(mockExecutor)
	
	// Act
	result, err := driver.AnalyzeComment(context.Background(), "test prompt")
	
	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if result.Summary != "Test summary" {
		t.Errorf("Expected summary 'Test summary', got '%s'", result.Summary)
	}
	
	if result.Type != "bug" {
		t.Errorf("Expected type 'bug', got '%s'", result.Type)
	}
	
	if result.RelevanceScore != 0.8 {
		t.Errorf("Expected relevance score 0.8, got %f", result.RelevanceScore)
	}
}

func TestAnalyzeComment_WithCodeBlock_Success(t *testing.T) {
	// Arrange
	mockExecutor := &MockCommandExecutor{
		output: []byte("```json\n{\"summary\": \"Feature flag analysis\", \"type\": \"domain\", \"tags\": [\"feature-flag\"], \"relevance_score\": 0.9}\n```"),
		err:    nil,
	}
	
	driver := NewDriver("claude", []string{"-p"})
	driver.SetExecutor(mockExecutor)
	
	// Act
	result, err := driver.AnalyzeComment(context.Background(), "test prompt")
	
	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if result.Summary != "Feature flag analysis" {
		t.Errorf("Expected summary 'Feature flag analysis', got '%s'", result.Summary)
	}
	
	if result.Type != "domain" {
		t.Errorf("Expected type 'domain', got '%s'", result.Type)
	}
}

func TestAnalyzeComment_CommandFailed(t *testing.T) {
	// Arrange
	mockExecutor := &MockCommandExecutor{
		output: nil,
		err:    errors.New("command failed"),
	}
	
	driver := NewDriver("claude", []string{"-p"})
	driver.SetExecutor(mockExecutor)
	
	// Act
	_, err := driver.AnalyzeComment(context.Background(), "test prompt")
	
	// Assert
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestAnalyzeComment_EmptyPrompt(t *testing.T) {
	// Arrange
	driver := NewDriver("claude", []string{"-p"})
	
	// Act
	_, err := driver.AnalyzeComment(context.Background(), "")
	
	// Assert
	if err == nil {
		t.Fatal("Expected error for empty prompt, got nil")
	}
}