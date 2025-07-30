# PoC実装計画

## 目的

Knowledge Baseシステムの有用性を検証するための最小限の実装を行います。

## PoC スコープ

### 含むもの
- 単一リポジトリからのPR収集（最新100件）
- 基本的なコメント分析とDB保存
- シンプルな検索API
- CLIツールでの動作確認

### 含まないもの
- 認証機能
- 複雑なフィルタリング
- パフォーマンス最適化
- Web UI

## 実装ステップ

### 1. 基盤構築（1日目）

```bash
# プロジェクト初期化
mkdir -p cmd/collector cmd/server
mkdir -p internal/{collector,database,github,llm,knowledge}
mkdir -p pkg/{models,config}
```

**実装内容:**
- 基本的なプロジェクト構造
- 設定ファイル読み込み
- データベース接続とマイグレーション

### 2. GitHub連携（2日目）

**実装内容:**
- GitHub APIクライアントのラッパー
- PR一覧取得機能
- レビューコメント取得機能
- 基本的なレート制限対応

**サンプルコード:**
```go
// internal/github/client.go
type Client struct {
    client *github.Client
    limit  *RateLimiter
}

func (c *Client) GetPullRequests(ctx context.Context, owner, repo string, limit int) ([]*github.PullRequest, error) {
    // 実装
}
```

### 3. LLM統合（3日目）

**実装内容:**
- OpenAI APIクライアント
- コメント分析プロンプト
- 結果のパース

**重要な決定事項:**
- 最初はシンプルな分類のみ（タイプ判定）
- バッチ処理は後回し

### 4. データ永続化（4日目）

**実装内容:**
- SQLiteでのCRUD操作
- 基本的なクエリ実装
- トランザクション管理

### 5. API実装（5日目）

**実装内容:**
- `/api/v1/documents/search` エンドポイント
- `/api/v1/context/generate` エンドポイント
- 基本的なエラーハンドリング

### 6. CLI実装（6日目）

**実装内容:**
- `collect` コマンド（データ収集）
- `query` コマンド（検索）
- `context` コマンド（コンテキスト生成）

## 動作確認シナリオ

### 1. データ収集

```bash
# 設定
export GITHUB_TOKEN="your-token"
export OPENAI_API_KEY="your-key"

# 収集実行
go run cmd/collector/main.go --repo pankona/knowledges --limit 50
```

期待される出力:
```
[INFO] Starting collection for pankona/knowledges
[INFO] Fetched 50 PRs
[INFO] Processing PR #123: "Add feature X"
[INFO] Found 5 review comments
[INFO] Analyzing comments with LLM...
[INFO] Saved 3 documents to database
...
[INFO] Collection completed: 150 documents created
```

### 2. データ検索

```bash
# ファイルで検索
go run cmd/cli/main.go query --file "src/main.go"

# 結果
Found 5 documents:
1. [Security] Input validation needed (PR #123)
2. [Performance] Consider using goroutines (PR #456)
...
```

### 3. コンテキスト生成

```bash
# AIコンテキスト生成
go run cmd/cli/main.go context --file "src/main.go" --type code_review

# 結果
Context for AI code review:
Based on previous reviews in this repository:
- Always validate user input before processing
- Use connection pooling for database operations
- Follow the error wrapping pattern: fmt.Errorf("failed to X: %w", err)
```

## 成功基準

1. **技術的成功**
   - 50以上のPRからデータ収集できる
   - 収集したデータを検索できる
   - AIが使えるコンテキストを生成できる

2. **ビジネス価値**
   - 過去のレビューコメントが実際に有用である
   - AIのコード生成/レビューの質が向上する
   - 開発者が価値を感じる

## リスクと対策

### リスク1: API制限
**対策:** 
- 小規模なリポジトリから開始
- キャッシュ機能の早期実装

### リスク2: LLMコストの増大
**対策:**
- GPT-3.5を使用
- 必要最小限の分析に絞る

### リスク3: データ品質
**対策:**
- 明確なフィルタリング基準
- 手動での品質確認

## 次のステップ

PoCが成功した場合:
1. 複数リポジトリ対応
2. より高度な分析機能
3. IDE統合
4. チーム固有のカスタマイズ