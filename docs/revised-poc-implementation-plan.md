# 改訂版PoC実装計画

## 主な変更点

1. **GitHub API** → **ghコマンド**: 認証管理が不要、実装が簡素化
2. **OpenAI API** → **claude -p**: 複数LLM対応、並列処理で高速化

## 改訂版実装ステップ

### 0. 前提条件の確認（実装前）

```bash
# ghコマンドの認証確認
gh auth status

# claudeコマンドの確認
claude --version

# 必要に応じてセットアップ
gh auth login
```

### 1. 基盤構築（1日目）

```bash
# プロジェクト初期化
go mod init github.com/pankona/knowledges
```

**実装内容:**
- プロジェクト構造の作成
- 設定ファイル読み込み（APIキー不要の簡素な設定）
- データベース接続とマイグレーション

**設定ファイル例:**
```yaml
# config.yaml
github:
  repositories:
    - pankona/knowledges

llm:
  primary: claude
  parallel: 5
  retry:
    max_attempts: 3
    initial_delay: 1s
  drivers:
    claude:
      command: claude
      args: [-p]

database:
  path: ./knowledge.db

collection:
  batch_size: 10
  max_prs_per_run: 100
```

### 2. GitHub CLI統合（2日目）

**実装内容:**
- ghコマンドのラッパー実装
- JSON出力のパース
- 並列でのPR情報取得

**実装例:**
```go
// internal/github/gh_wrapper.go
func (g *GHWrapper) GetMergedPRs(ctx context.Context, limit int) ([]PullRequest, error) {
    cmd := exec.CommandContext(ctx, "gh", "pr", "list",
        "--repo", g.repo,
        "--state", "merged",
        "--limit", fmt.Sprintf("%d", limit),
        "--json", "number,title,url,createdAt,author,files")
    
    output, err := cmd.Output()
    // JSONパース処理
}
```

### 3. LLM CLI統合（3日目）

**実装内容:**
- claude -p のラッパー実装
- プロンプトテンプレート管理
- 並列実行による高速化

**動作確認:**
```bash
# 単一コメントの分析テスト
echo "テストコメント" | claude -p

# 並列実行のテスト
for i in {1..5}; do
    echo "コメント$i" | claude -p &
done
wait
```

### 4. データ永続化とコレクター実装（4日目）

**実装内容:**
- SQLiteのCRUD操作
- PR収集→分析→保存のパイプライン
- エラーハンドリングとリトライ

### 5. API実装（5日目）

**実装内容:**
- 検索エンドポイント
- コンテキスト生成エンドポイント
- CLIツール

### 6. 統合テストと最適化（6日目）

**実装内容:**
- エンドツーエンドテスト
- パフォーマンス測定
- ドキュメント整備

## シンプルになった実装例

### メインの収集フロー

```go
// cmd/collector/main.go
func main() {
    // 1. ghでPR一覧取得
    prs := ghWrapper.GetMergedPRs(ctx, 100)
    
    // 2. 並列でコメント取得
    comments := ghWrapper.FetchPRsWithCommentsBatch(ctx, prNumbers)
    
    // 3. LLMで並列分析
    results := llmDriver.AnalyzeBatch(ctx, comments)
    
    // 4. DBに保存
    store.SaveDocuments(ctx, results)
}
```

## 動作確認シナリオ

### 1. 最小限の動作確認

```bash
# 1件のPRで動作確認
go run cmd/collector/main.go --repo pankona/knowledges --limit 1

# ログ確認
[INFO] Fetching PRs with gh command...
[INFO] Found 1 merged PR
[INFO] Fetching comments for PR #123...
[INFO] Analyzing 3 comments with claude...
[INFO] Saved 3 documents to database
```

### 2. 並列処理の確認

```bash
# 10件のPRで並列処理確認
go run cmd/collector/main.go --repo pankona/knowledges --limit 10 --parallel 5

# CPU使用率とメモリ使用量を監視
```

## リスクと対策

### リスク1: CLIコマンドの可用性
**対策:**
- 起動時にコマンドの存在確認
- 明確なエラーメッセージ

### リスク2: 並列実行時のレート制限
**対策:**
- セマフォによる並列数制限
- エラー時の自動リトライ

## 期待される成果

1. **開発速度**: APIライブラリ不要で実装が高速化
2. **保守性**: CLIツールのアップデートに自動追従
3. **柔軟性**: 新しいLLMの追加が容易
4. **デバッグ**: コマンドを直接実行して動作確認可能