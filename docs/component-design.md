# コンポーネント設計（改訂版）

## パッケージ構成

```
knowledges/
├── cmd/
│   ├── collector/      # PR収集CLI
│   │   └── main.go
│   └── server/         # API サーバー
│       └── main.go
├── internal/
│   ├── collector/      # データ収集関連
│   │   ├── fetcher.go
│   │   ├── parser.go
│   │   └── analyzer.go
│   ├── database/       # データベース関連
│   │   ├── db.go
│   │   ├── document.go
│   │   ├── progress.go
│   │   └── migrations/
│   ├── knowledge/      # ナレッジ管理
│   │   ├── store.go
│   │   ├── query.go
│   │   └── context.go
│   ├── github/         # GitHub CLI (gh) ラッパー
│   │   ├── gh_wrapper.go
│   │   └── types.go
│   └── llm/           # LLM CLI統合
│       ├── driver.go
│       ├── prompt.go
│       └── multi_driver.go
├── pkg/
│   ├── models/        # 共通データモデル
│   │   └── document.go
│   └── config/        # 設定管理
│       └── config.go
└── migrations/        # DBマイグレーションファイル
    └── 001_initial.sql
```

## 主要インターフェース定義

### 1. Collector インターフェース

```go
// internal/collector/interfaces.go

package collector

import (
    "context"
    "time"
)

// Fetcher はPRデータを取得するインターフェース
type Fetcher interface {
    // FetchMergedPRs は最新のマージ済みPRを取得
    FetchMergedPRs(ctx context.Context, limit int) ([]*PullRequest, error)
    
    // FetchPRComments は指定PRのレビューコメントを取得
    FetchPRComments(ctx context.Context, prNumber int) (*PRDetail, error)
    
    // FetchPRsWithCommentsBatch は複数PRのコメントを並列で取得
    FetchPRsWithCommentsBatch(ctx context.Context, prNumbers []int) ([]*PRWithComments, error)
}

// Parser はPRデータを解析するインターフェース
type Parser interface {
    // ParseComment はコメントから必要な情報を抽出
    ParseComment(comment *Comment) (*ParsedComment, error)
    
    // ExtractFileInfo はファイル情報を抽出
    ExtractFileInfo(comment *Comment) (*FileInfo, error)
}

// Analyzer はコメントを分析するインターフェース
type Analyzer interface {
    // AnalyzeComment は単一のコメントを分析
    AnalyzeComment(ctx context.Context, comment Comment) (*AnalysisResult, error)
    
    // AnalyzeBatch は複数のコメントを並列で分析
    AnalyzeBatch(ctx context.Context, comments []Comment) ([]*AnalysisResult, error)
    
    // SetLLM は使用するLLMを設定
    SetLLM(llmName string) error
}
```

### 2. Knowledge インターフェース

```go
// internal/knowledge/interfaces.go

package knowledge

import (
    "context"
    "knowledges/pkg/models"
)

// Store はドキュメントを永続化するインターフェース
type Store interface {
    // SaveDocument はドキュメントを保存
    SaveDocument(ctx context.Context, doc *models.Document) error
    
    // GetDocument はIDでドキュメントを取得
    GetDocument(ctx context.Context, id int64) (*models.Document, error)
    
    // DeleteDocument はドキュメントを削除
    DeleteDocument(ctx context.Context, id int64) error
}

// QueryEngine は検索を処理するインターフェース
type QueryEngine interface {
    // QueryByFile はファイルパスで検索
    QueryByFile(ctx context.Context, filePath string, limit int) ([]*models.Document, error)
    
    // QueryByDirectory はディレクトリで検索
    QueryByDirectory(ctx context.Context, dirPath string, limit int) ([]*models.Document, error)
    
    // QueryByLanguage は言語で検索
    QueryByLanguage(ctx context.Context, language string, types []string, limit int) ([]*models.Document, error)
}

// ContextBuilder はAI用コンテキストを構築するインターフェース
type ContextBuilder interface {
    // BuildContext は関連ドキュメントからコンテキストを構築
    BuildContext(docs []*models.Document, params ContextParams) (string, error)
}
```

## 主要データ構造

### 1. Document モデル

```go
// pkg/models/document.go

package models

import "time"

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
```

### 2. 設定構造体

```go
// pkg/config/config.go

package config

type Config struct {
    // GitHub設定
    GitHub GitHubConfig `yaml:"github"`
    
    // LLM設定
    LLM LLMConfig `yaml:"llm"`
    
    // データベース設定
    Database DatabaseConfig `yaml:"database"`
    
    // 収集設定
    Collection CollectionConfig `yaml:"collection"`
    
    // サーバー設定
    Server ServerConfig `yaml:"server"`
}

type GitHubConfig struct {
    Repositories   []string `yaml:"repositories"`
    // ghコマンドの認証は gh auth login で管理
}

type LLMConfig struct {
    Primary        string            `yaml:"primary"` // "claude", "gemini"など
    Parallel       int               `yaml:"parallel"`
    Retry          RetryConfig       `yaml:"retry"`
    Drivers        map[string]Driver `yaml:"drivers"`
}

type Driver struct {
    Command        string   `yaml:"command"`
    Args           []string `yaml:"args"`
}

type DatabaseConfig struct {
    Path           string   `yaml:"path"`
}

type CollectionConfig struct {
    BatchSize      int      `yaml:"batch_size"`
    MaxPRsPerRun   int      `yaml:"max_prs_per_run"`
    LookbackDays   int      `yaml:"lookback_days"`
    RateLimit      int      `yaml:"rate_limit"`
}

type ServerConfig struct {
    Port           int      `yaml:"port"`
    ReadTimeout    int      `yaml:"read_timeout"`
    WriteTimeout   int      `yaml:"write_timeout"`
}
```

## エラー処理

### カスタムエラー型

```go
// internal/errors/errors.go

package errors

import "fmt"

// AppError はアプリケーション全体で使用するエラー型
type AppError struct {
    Code    string
    Message string
    Err     error
}

func (e *AppError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
    }
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// エラーコード定義
const (
    ErrCodeGitHubAPI     = "GITHUB_API_ERROR"
    ErrCodeLLMAPI        = "LLM_API_ERROR"
    ErrCodeDatabase      = "DATABASE_ERROR"
    ErrCodeNotFound      = "NOT_FOUND"
    ErrCodeInvalidInput  = "INVALID_INPUT"
    ErrCodeRateLimit     = "RATE_LIMIT"
)
```

## 依存性注入

### DIコンテナ

```go
// internal/container/container.go

package container

type Container struct {
    config         *config.Config
    db             *sql.DB
    githubClient   github.Client
    llmClient      llm.Client
    documentStore  knowledge.Store
    queryEngine    knowledge.QueryEngine
    contextBuilder knowledge.ContextBuilder
}

// New は新しいコンテナを作成
func New(cfg *config.Config) (*Container, error) {
    // 各種クライアントとサービスの初期化
}
```

## テスト戦略

### モックインターフェース

各インターフェースに対してモック実装を用意：

```go
// internal/collector/mocks/fetcher.go

type MockFetcher struct {
    FetchPullRequestsFunc func(ctx context.Context, repo string, since time.Time) ([]*PullRequest, error)
    FetchCommentsFunc     func(ctx context.Context, repo string, prNumber int) ([]*Comment, error)
}
```

### テストヘルパー

```go
// internal/testutil/database.go

// SetupTestDB はテスト用のDBを作成
func SetupTestDB(t *testing.T) *sql.DB {
    // インメモリSQLiteを使用
}
```