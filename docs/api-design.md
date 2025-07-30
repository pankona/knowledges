# API設計

## 概要

Knowledge BaseシステムのREST APIインターフェースを定義します。主にAIツールやCLIからナレッジを取得するために使用されます。

## エンドポイント一覧

### 1. ドキュメント検索

#### GET /api/v1/documents/search

指定された条件でドキュメントを検索します。

**クエリパラメータ:**
```
file_path    (string, optional): ファイルパス
directory    (string, optional): ディレクトリパス
language     (string, optional): プログラミング言語
types        (string, optional): カンマ区切りのコメントタイプ
limit        (int, optional):    結果の最大数（デフォルト: 20, 最大: 100）
offset       (int, optional):    オフセット（ページネーション用）
```

**レスポンス例:**
```json
{
  "documents": [
    {
      "id": 123,
      "summary": "非同期処理では適切なエラーハンドリングが必要です",
      "original_comment": "This async function should handle errors...",
      "file_path": "src/handlers/user.go",
      "directory_path": "src/handlers",
      "language": "go",
      "line_number": 42,
      "repository": "owner/repo",
      "pr_number": 456,
      "pr_title": "Add user authentication",
      "author": "reviewer1",
      "comment_type": "bug",
      "tags": ["error-handling", "async", "golang"],
      "relevance_score": 0.95,
      "commented_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 45,
  "limit": 20,
  "offset": 0
}
```

### 2. コンテキスト生成

#### POST /api/v1/context/generate

AIツール用のコンテキストを生成します。

**リクエストボディ:**
```json
{
  "file_path": "src/handlers/user.go",
  "directory": "src/handlers",
  "language": "go",
  "context_type": "code_review",
  "max_tokens": 4000,
  "filters": {
    "types": ["security", "performance", "domain"],
    "min_relevance_score": 0.7,
    "max_age_days": 180
  }
}
```

**レスポンス例:**
```json
{
  "context": "Based on previous code reviews in this repository:\n\n1. **Security**: Always validate user input...\n2. **Performance**: Use connection pooling...\n3. **Domain**: Follow the repository's error handling pattern...",
  "document_count": 15,
  "tokens_used": 3500,
  "metadata": {
    "most_relevant_files": [
      "src/handlers/auth.go",
      "src/handlers/middleware.go"
    ],
    "common_issues": [
      "error-handling",
      "input-validation"
    ]
  }
}
```

### 3. 統計情報

#### GET /api/v1/stats

システムの統計情報を取得します。

**レスポンス例:**
```json
{
  "total_documents": 5234,
  "repositories": [
    {
      "name": "owner/repo1",
      "document_count": 2156,
      "last_updated": "2024-01-20T15:00:00Z"
    }
  ],
  "comment_types": {
    "security": 523,
    "performance": 412,
    "readability": 1523,
    "domain": 234
  },
  "languages": {
    "go": 3421,
    "javascript": 1234,
    "python": 579
  },
  "collection_status": {
    "last_run": "2024-01-20T10:00:00Z",
    "next_run": "2024-01-21T10:00:00Z",
    "status": "idle"
  }
}
```

### 4. 収集タスク管理

#### POST /api/v1/collection/trigger

手動でデータ収集をトリガーします。

**リクエストボディ:**
```json
{
  "repository": "owner/repo",
  "mode": "incremental",
  "options": {
    "since": "2024-01-01T00:00:00Z",
    "pr_limit": 100
  }
}
```

**レスポンス例:**
```json
{
  "task_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "queued",
  "estimated_duration": "15m"
}
```

#### GET /api/v1/collection/status/:task_id

収集タスクのステータスを確認します。

**レスポンス例:**
```json
{
  "task_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "in_progress",
  "progress": {
    "current": 45,
    "total": 100,
    "phase": "analyzing_comments"
  },
  "started_at": "2024-01-20T15:30:00Z",
  "metrics": {
    "prs_processed": 45,
    "comments_collected": 234,
    "documents_created": 156
  }
}
```

## CLI統合

### CLIコマンド例

```bash
# ファイルに関連するナレッジを取得
knowledges query --file src/handlers/user.go --limit 10

# AIコンテキストを生成
knowledges context --file src/handlers/user.go --type code_review

# 統計情報を表示
knowledges stats

# データ収集を実行
knowledges collect --repo owner/repo --mode incremental
```

### CLI出力フォーマット

```bash
$ knowledges query --file src/handlers/user.go --limit 5

Found 5 relevant documents:

1. [Security] Input validation required
   File: src/handlers/user.go:42
   PR: #456 "Add user authentication"
   Score: 0.95

2. [Performance] Use connection pooling
   File: src/handlers/user.go:156
   PR: #234 "Optimize database queries"
   Score: 0.87

...
```

## エラーレスポンス

### 標準エラー形式

```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "Invalid language parameter",
    "details": {
      "field": "language",
      "value": "unknown",
      "allowed_values": ["go", "javascript", "python", "java"]
    }
  }
}
```

### HTTPステータスコード

- `200 OK`: 正常終了
- `400 Bad Request`: 不正なリクエスト
- `404 Not Found`: リソースが見つからない
- `429 Too Many Requests`: レート制限
- `500 Internal Server Error`: サーバーエラー

## 認証とセキュリティ

### 認証方式（将来実装）

```
Authorization: Bearer <api_token>
```

### レート制限

- デフォルト: 100リクエスト/分
- バースト: 10リクエスト/秒

### CORS設定

```go
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

## パフォーマンス最適化

### キャッシング戦略

```go
type CacheConfig struct {
    TTL             time.Duration
    MaxSize         int
    EvictionPolicy  string // "lru", "lfu"
}

// レスポンスヘッダー
Cache-Control: public, max-age=300
ETag: "33a64df551425fcc55e4d42a148795d9f25f89d4"
```

### レスポンス圧縮

```go
func gzipMiddleware(next http.Handler) http.Handler {
    return gziphandler.GzipHandler(next)
}
```

## 使用例

### curl

```bash
# ドキュメント検索
curl -X GET "http://localhost:8080/api/v1/documents/search?file_path=src/main.go&limit=10"

# コンテキスト生成
curl -X POST "http://localhost:8080/api/v1/context/generate" \
  -H "Content-Type: application/json" \
  -d '{
    "file_path": "src/main.go",
    "language": "go",
    "context_type": "code_review"
  }'
```

### Go クライアント

```go
type KnowledgeClient struct {
    baseURL string
    client  *http.Client
}

func (c *KnowledgeClient) SearchDocuments(query DocumentQuery) (*DocumentResponse, error) {
    // 実装
}

func (c *KnowledgeClient) GenerateContext(req ContextRequest) (*ContextResponse, error) {
    // 実装
}
```