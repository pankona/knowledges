package github

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// PullRequest はPRの情報を表現します
type PullRequest struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"createdAt"`
	Author    Author    `json:"author"`
	Labels    []Label   `json:"labels,omitempty"`
}

// Label はPRのラベル情報を表現します
type Label struct {
	Name string `json:"name"`
}

// Author はユーザー情報を表現します
type Author struct {
	Login string `json:"login"`
}

// Comment はPRのレビューコメントを表現します
type Comment struct {
	Author     Author    `json:"author"`
	Body       string    `json:"body"`
	CreatedAt  time.Time `json:"createdAt"`
	URL        string    `json:"url"`
	FilePath   string    `json:"filePath"`   // GraphQLレスポンスから抽出
	LineNumber int       `json:"lineNumber"` // GraphQLレスポンスから抽出
}

// GraphQLレスポンス用の構造体
type graphQLResponse struct {
	Data struct {
		Repository struct {
			PullRequest *struct {
				ReviewThreads struct {
					Nodes []struct {
						Path     string `json:"path"`
						Line     int    `json:"line"`
						Comments struct {
							Nodes []struct {
								Author    Author `json:"author"`
								Body      string `json:"body"`
								CreatedAt string `json:"createdAt"`
								URL       string `json:"url"`
							} `json:"nodes"`
						} `json:"comments"`
					} `json:"nodes"`
				} `json:"reviewThreads"`
			} `json:"pullRequest"`
		} `json:"repository"`
	} `json:"data"`
}

// CommandExecutor は外部コマンドを実行するインターフェース
type CommandExecutor interface {
	Execute(ctx context.Context, cmd string, args ...string) ([]byte, error)
}

// DefaultCommandExecutor は実際のコマンドを実行します
type DefaultCommandExecutor struct{}

func (e *DefaultCommandExecutor) Execute(ctx context.Context, cmd string, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, cmd, args...)
	return command.Output()
}

// GHWrapper はghコマンドのラッパーです
type GHWrapper struct {
	repo     string
	executor CommandExecutor
}

// NewGHWrapper は新しいGHWrapperを作成します
func NewGHWrapper(repo string) *GHWrapper {
	return &GHWrapper{
		repo:     repo,
		executor: &DefaultCommandExecutor{},
	}
}

// SetExecutor はコマンド実行器を設定します（テスト用）
func (g *GHWrapper) SetExecutor(executor CommandExecutor) {
	g.executor = executor
}

// GetMergedPRs は最新のマージ済みPRを取得します
func (g *GHWrapper) GetMergedPRs(ctx context.Context, limit int) ([]PullRequest, error) {
	args := []string{
		"pr", "list",
		"--repo", g.repo,
		"--state", "merged",
		"--limit", fmt.Sprintf("%d", limit),
		"--json", "number,title,url,createdAt,author",
	}

	output, err := g.executor.Execute(ctx, "gh", args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute gh command: %w", err)
	}

	var prs []PullRequest
	if err := json.Unmarshal(output, &prs); err != nil {
		return nil, fmt.Errorf("failed to parse gh output: %w", err)
	}

	return prs, nil
}

// GetMergedPRsWithLabel は指定されたラベルを持つマージ済みPRを取得します
func (g *GHWrapper) GetMergedPRsWithLabel(ctx context.Context, limit int, label string) ([]PullRequest, error) {
	args := []string{
		"pr", "list",
		"--repo", g.repo,
		"--state", "merged",
		"--limit", fmt.Sprintf("%d", limit),
		"--search", fmt.Sprintf("label:%s", label),
		"--json", "number,title,url,createdAt,author,labels",
	}

	output, err := g.executor.Execute(ctx, "gh", args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute gh command: %w", err)
	}

	var prs []PullRequest
	if err := json.Unmarshal(output, &prs); err != nil {
		return nil, fmt.Errorf("failed to parse gh output: %w", err)
	}

	return prs, nil
}

// GetMergedPRsExcludingBots は bot を除外してマージ済みPRを取得します
func (g *GHWrapper) GetMergedPRsExcludingBots(ctx context.Context, limit int, label string) ([]PullRequest, error) {
	// 除外するbot作成者のリスト
	botAuthors := []string{
		"dependabot[bot]",
		"github-actions[bot]",
		"renovate[bot]",
		"codecov[bot]",
	}

	args := []string{
		"pr", "list",
		"--repo", g.repo,
		"--state", "merged",
		"--limit", fmt.Sprintf("%d", limit),
	}

	// 検索条件を一つにまとめる
	var searchTerms []string
	
	// ラベルフィルタを追加
	if label != "" {
		searchTerms = append(searchTerms, fmt.Sprintf("label:%s", label))
	}

	// bot作成者を除外
	for _, bot := range botAuthors {
		searchTerms = append(searchTerms, fmt.Sprintf("-author:%s", bot))
	}

	// 検索条件がある場合は一つの--searchオプションにまとめる
	if len(searchTerms) > 0 {
		args = append(args, "--search", strings.Join(searchTerms, " "))
	}

	args = append(args, "--json", "number,title,url,createdAt,author,labels")

	output, err := g.executor.Execute(ctx, "gh", args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute gh command: %w", err)
	}

	var prs []PullRequest
	if err := json.Unmarshal(output, &prs); err != nil {
		return nil, fmt.Errorf("failed to parse gh output: %w", err)
	}

	return prs, nil
}

// GetPRComments は指定PRのレビューコメントを取得します
func (g *GHWrapper) GetPRComments(ctx context.Context, prNumber int) ([]Comment, error) {
	owner, name := parseRepo(g.repo)
	
	query := `
	query($owner: String!, $repo: String!, $number: Int!) {
		repository(owner: $owner, name: $repo) {
			pullRequest(number: $number) {
				reviewThreads(first: 100) {
					nodes {
						path
						line
						comments(first: 50) {
							nodes {
								author { login }
								body
								createdAt
								url
							}
						}
					}
				}
			}
		}
	}`

	args := []string{
		"api", "graphql",
		"-f", fmt.Sprintf("query=%s", query),
		"-f", fmt.Sprintf("owner=%s", owner),
		"-f", fmt.Sprintf("repo=%s", name),
		"-F", fmt.Sprintf("number=%d", prNumber),
	}

	output, err := g.executor.Execute(ctx, "gh", args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute gh GraphQL query: %w", err)
	}

	var response graphQLResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("failed to parse GraphQL response: %w", err)
	}

	if response.Data.Repository.PullRequest == nil {
		return nil, fmt.Errorf("PR #%d not found", prNumber)
	}

	var comments []Comment
	for _, thread := range response.Data.Repository.PullRequest.ReviewThreads.Nodes {
		for _, comment := range thread.Comments.Nodes {
			createdAt, err := time.Parse(time.RFC3339, comment.CreatedAt)
			if err != nil {
				// Skip invalid timestamps but continue processing
				continue
			}

			comments = append(comments, Comment{
				Author:     comment.Author,
				Body:       comment.Body,
				CreatedAt:  createdAt,
				URL:        comment.URL,
				FilePath:   thread.Path,
				LineNumber: thread.Line,
			})
		}
	}

	return comments, nil
}

// GetPR は指定されたPR番号の詳細情報を取得します
func (g *GHWrapper) GetPR(ctx context.Context, prNumber int) (*PullRequest, error) {
	owner, name := parseRepo(g.repo)
	if owner == "" || name == "" {
		return nil, fmt.Errorf("invalid repository format: %s", g.repo)
	}

	args := []string{
		"pr", "view", strconv.Itoa(prNumber),
		"--repo", g.repo,
		"--json", "number,title,url,createdAt,author",
	}

	output, err := g.executor.Execute(ctx, "gh", args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute gh command: %w", err)
	}

	var pr PullRequest
	if err := json.Unmarshal(output, &pr); err != nil {
		return nil, fmt.Errorf("failed to parse PR JSON: %w", err)
	}

	return &pr, nil
}

// parseRepo はrepo文字列を owner/name に分割します
func parseRepo(repo string) (owner, name string) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}