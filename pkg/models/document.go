package models

import "time"

// Document はレビューコメントから生成されたナレッジドキュメント
type Document struct {
	ID               int64     `json:"id"`
	Summary          string    `json:"summary"`
	OriginalComment  string    `json:"original_comment"`
	ThreadContext    string    `json:"thread_context,omitempty"`
	
	// ファイル情報
	FilePath        string    `json:"file_path"`
	DirectoryPath   string    `json:"directory_path"`
	Language        string    `json:"language"`
	LineNumber      *int      `json:"line_number,omitempty"`
	
	// PR情報
	Repository      string    `json:"repository"`
	PRNumber        int       `json:"pr_number"`
	PRTitle         string    `json:"pr_title"`
	PRURL           string    `json:"pr_url"`
	CommentURL      string    `json:"comment_url"`
	
	// メタデータ
	Author          string    `json:"author"`
	CommentType     string    `json:"comment_type"`
	Tags            []string  `json:"tags"`
	RelevanceScore  float64   `json:"relevance_score"`
	
	// タイムスタンプ
	CommentedAt     time.Time `json:"commented_at"`
	CollectedAt     time.Time `json:"collected_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CommentType の定義
type CommentType string

const (
	CommentTypePerformance   CommentType = "performance"
	CommentTypeSecurity      CommentType = "security"
	CommentTypeReadability   CommentType = "readability"
	CommentTypeDomain        CommentType = "domain"
	CommentTypeTesting       CommentType = "testing"
	CommentTypeArchitecture  CommentType = "architecture"
	CommentTypeStyle         CommentType = "style"
	CommentTypeBug           CommentType = "bug"
	CommentTypeSuggestion    CommentType = "suggestion"
	CommentTypeQuestion      CommentType = "question"
	CommentTypeOther         CommentType = "other"
)