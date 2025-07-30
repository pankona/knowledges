package collector

import (
	"strings"

	"github.com/pankona/knowledges/internal/github"
)

// CommentFilter はレビューコメントをフィルタリングします
type CommentFilter struct {
	minLength       int
	excludePatterns []string
	excludeAuthors  []string
}

// NewCommentFilter は新しいCommentFilterを作成します
func NewCommentFilter() *CommentFilter {
	return &CommentFilter{
		minLength: 10, // 最小10文字
		excludePatterns: []string{
			// 短い承認コメント
			"lgtm",
			"looks good to me",
			"approved",
			"👍",
			"✅",
			"+1",
			
			// 短い応答コメント  
			"thanks",
			"thank you",
			"done",
			"fixed",
			"ok",
			"sure",
			"yes",
			"no",
			"nope",
			"agree",
			"agreed",
			
			// 自動生成っぽいパターン
			"automatically generated",
			"bumps version",
			"dependency update",
		},
		excludeAuthors: []string{
			"github-actions[bot]",
			"dependabot[bot]",
			"renovate[bot]",
			"codecov[bot]",
		},
	}
}

// IsUseful はコメントが有用かどうかを判定します
func (f *CommentFilter) IsUseful(comment github.Comment) bool {
	// 自動化されたアカウントからのコメントを除外
	for _, excludeAuthor := range f.excludeAuthors {
		if strings.EqualFold(comment.Author.Login, excludeAuthor) {
			return false
		}
	}

	// 最小文字数チェック
	if !f.HasMinimumLength(comment.Body) {
		return false
	}

	// 除外パターンチェック
	bodyLower := strings.ToLower(strings.TrimSpace(comment.Body))
	
	for _, pattern := range f.excludePatterns {
		if strings.Contains(bodyLower, pattern) {
			// 完全一致または単語として一致する場合のみ除外
			if bodyLower == pattern || f.isWordMatch(bodyLower, pattern) {
				return false
			}
		}
	}

	return true
}

// HasMinimumLength はコメントが最小文字数を満たしているかチェックします
func (f *CommentFilter) HasMinimumLength(body string) bool {
	// 空白を除去してから文字数をチェック
	trimmed := strings.TrimSpace(body)
	return len(trimmed) >= f.minLength
}

// FilterComments は有用なコメントのみを抽出します
func (f *CommentFilter) FilterComments(comments []github.Comment) []github.Comment {
	var filtered []github.Comment
	
	for _, comment := range comments {
		if f.IsUseful(comment) {
			filtered = append(filtered, comment)
		}
	}
	
	return filtered
}

// isWordMatch はパターンが単語として一致するかチェックします
func (f *CommentFilter) isWordMatch(text, pattern string) bool {
	words := strings.Fields(text)
	for _, word := range words {
		// 句読点を除去
		cleanWord := strings.Trim(word, ".,!?;:")
		if cleanWord == pattern {
			return true
		}
	}
	return false
}