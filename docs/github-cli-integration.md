# GitHub CLI (gh) 統合設計

## 概要

GitHub APIの代わりに`gh`コマンドを使用することで、認証管理を簡素化し、実装を軽量化します。

## ghコマンドの活用

### 1. PR一覧の取得

```bash
# 最新100件のマージ済みPRを取得
gh pr list --repo owner/repo --state merged --limit 100 --json number,title,url,createdAt,author,files

# 出力例（JSON）
[
  {
    "number": 123,
    "title": "Add user authentication",
    "url": "https://github.com/owner/repo/pull/123",
    "createdAt": "2024-01-15T10:00:00Z",
    "author": {
      "login": "user1"
    },
    "files": [
      {"path": "src/auth.go"},
      {"path": "src/auth_test.go"}
    ]
  }
]
```

### 2. PR詳細とレビューコメントの取得

```bash
# PR詳細とレビューコメントを取得
gh pr view 123 --repo owner/repo --json reviews,comments

# レビュースレッドの取得（GraphQL使用）
gh api graphql -f query='
  query($owner: String!, $repo: String!, $number: Int!) {
    repository(owner: $owner, name: $repo) {
      pullRequest(number: $number) {
        reviewThreads(first: 100) {
          nodes {
            path
            line
            isResolved
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
  }' -f owner=owner -f repo=repo -F number=123
```

### 3. 効率的なバッチ取得

```bash
# 複数PRの情報を一度に取得するスクリプト
#!/bin/bash
pr_numbers=(123 124 125)
for pr in "${pr_numbers[@]}"; do
    gh pr view "$pr" --repo owner/repo --json number,reviews,comments &
done
wait
```

## Go実装での統合

### GH Wrapperの実装

```go
// internal/github/gh_wrapper.go
package github

import (
    "encoding/json"
    "fmt"
    "os/exec"
    "context"
)

type GHWrapper struct {
    repo string
}

func NewGHWrapper(repo string) *GHWrapper {
    return &GHWrapper{repo: repo}
}

// GetMergedPRs は最新のマージ済みPRを取得
func (g *GHWrapper) GetMergedPRs(ctx context.Context, limit int) ([]PullRequest, error) {
    cmd := exec.CommandContext(ctx, "gh", "pr", "list",
        "--repo", g.repo,
        "--state", "merged",
        "--limit", fmt.Sprintf("%d", limit),
        "--json", "number,title,url,createdAt,author,files")
    
    output, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("failed to execute gh command: %w", err)
    }
    
    var prs []PullRequest
    if err := json.Unmarshal(output, &prs); err != nil {
        return nil, fmt.Errorf("failed to parse gh output: %w", err)
    }
    
    return prs, nil
}

// GetPRComments はPRのレビューコメントを取得
func (g *GHWrapper) GetPRComments(ctx context.Context, prNumber int) (*PRDetail, error) {
    // GraphQLクエリを使用してレビュースレッドを取得
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
    
    owner, name := parseRepo(g.repo)
    cmd := exec.CommandContext(ctx, "gh", "api", "graphql",
        "-f", fmt.Sprintf("query=%s", query),
        "-f", fmt.Sprintf("owner=%s", owner),
        "-f", fmt.Sprintf("repo=%s", name),
        "-F", fmt.Sprintf("number=%d", prNumber))
    
    output, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("failed to get PR comments: %w", err)
    }
    
    // 結果をパース
    var result GraphQLResponse
    if err := json.Unmarshal(output, &result); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }
    
    return convertToPRDetail(result), nil
}
```

### 並列処理での高速化

```go
// internal/collector/parallel_fetcher.go
package collector

import (
    "context"
    "sync"
)

type ParallelFetcher struct {
    gh        *github.GHWrapper
    workers   int
}

func (f *ParallelFetcher) FetchPRsWithComments(ctx context.Context, prNumbers []int) ([]*PRWithComments, error) {
    results := make([]*PRWithComments, len(prNumbers))
    errors := make([]error, len(prNumbers))
    
    // ワーカープールで並列実行
    sem := make(chan struct{}, f.workers)
    var wg sync.WaitGroup
    
    for i, prNum := range prNumbers {
        wg.Add(1)
        go func(idx int, pr int) {
            defer wg.Done()
            
            sem <- struct{}{}
            defer func() { <-sem }()
            
            detail, err := f.gh.GetPRComments(ctx, pr)
            if err != nil {
                errors[idx] = err
                return
            }
            
            results[idx] = &PRWithComments{
                Number:   pr,
                Comments: detail.Comments,
            }
        }(i, prNum)
    }
    
    wg.Wait()
    
    // エラーチェック
    for _, err := range errors {
        if err != nil {
            return results, fmt.Errorf("some PRs failed to fetch: %w", err)
        }
    }
    
    return results, nil
}
```

## エラーハンドリング

### ghコマンドのエラー処理

```go
func handleGHError(err error) error {
    if exitErr, ok := err.(*exec.ExitError); ok {
        stderr := string(exitErr.Stderr)
        
        // 一般的なエラーパターンの処理
        switch {
        case strings.Contains(stderr, "authentication"):
            return fmt.Errorf("gh authentication required: run 'gh auth login'")
        case strings.Contains(stderr, "rate limit"):
            return fmt.Errorf("GitHub rate limit exceeded")
        case strings.Contains(stderr, "not found"):
            return fmt.Errorf("repository or PR not found")
        default:
            return fmt.Errorf("gh command failed: %s", stderr)
        }
    }
    return err
}
```

## 利点とトレードオフ

### 利点
1. **認証の簡素化**: ghの認証を利用（`gh auth login`で設定済み）
2. **保守性**: GitHub APIの変更に対してghコマンドが吸収
3. **デバッグ容易性**: コマンドを直接実行して確認可能
4. **セキュリティ**: トークンをコード内で管理不要

### トレードオフ
1. **依存性**: ghコマンドのインストールが必要
2. **パフォーマンス**: プロセス起動のオーバーヘッド
3. **エラー処理**: 標準エラー出力の解析が必要

### パフォーマンス対策
- 並列実行で複数PRを同時取得
- 必要な情報のみをJSONで取得
- GraphQLで1回のリクエストで多くの情報を取得