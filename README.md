# Knowledge Base

AI のための knowledge base - Pull Requestレビューコメントから学習するナレッジベース

## 概要

過去のPull Requestレビューコメントを収集・分析し、AIがコード生成やレビューを行う際のコンテキストとして活用できるツールです。

## ✨ 特徴

- **🚀 シンプルな実装**: `gh`コマンドとClaude CLIを活用した軽量設計
- **📊 効率的な収集**: GitHub APIの代わりに`gh`コマンドを使用
- **🤖 LLM統合**: Claude、Gemini等のCLIツールで並列分析
- **🗄️ ローカル完結**: SQLiteでのオフライン動作
- **🔧 拡張可能**: 新しいLLMや機能の追加が容易

## 🛠️ 必要な環境

- Go 1.21+
- `gh` コマンド（GitHub CLI）
- `claude` コマンド（オプション、LLM分析用）

## 📦 インストール

```bash
# リポジトリをクローン
git clone https://github.com/pankona/knowledges.git
cd knowledges

# 依存関係をインストール
go mod tidy

# ツールをビルド
go build -o bin/collector ./cmd/collector
go build -o bin/query ./cmd/query
go build -o bin/test-gh ./cmd/test-gh
```

## ⚙️ セットアップ

### 1. GitHub CLI認証

```bash
gh auth login
```

### 2. 設定ファイル

```bash
cp config.yaml.example config.yaml
# 必要に応じて設定を編集
```

## 🚀 使用方法

### 1. GitHub統合テスト

```bash
# ghコマンドが正常に動作するかテスト
./bin/test-gh -repo golang/go -limit 3
```

### 2. データ収集（PoC）

```bash
# 最小限のPoC収集
./bin/collector -repo golang/go -limit 1

# 設定ファイルを使用
./bin/collector -config config.yaml -limit 5
```

### 3. 収集データの確認

```bash
# データベースの内容を表示
./bin/query

# 特定のデータベースファイルを指定
./bin/query -db /path/to/knowledge.db
```

## 📊 現在の機能（PoC版）

### ✅ 実装済み
- **GitHub PR・コメント取得**（`gh`コマンド連携）
- **インテリジェントなコメントフィルタリング**（LGTM等の除外）
- **ファイル情報の自動抽出**（言語・ディレクトリ検出）
- **LLM分析**（Claude CLI対応、フォールバック機能付き）
- **SQLiteデータベース**（マイグレーション・重複排除対応）
- **エンドツーエンドパイプライン**（実PR→分析→保存）
- **基本的なクエリ機能**

### 🚧 今後の実装予定
- バッチ処理・並列処理の最適化
- REST API
- 複数LLMの並列実行
- Web UI
- 検索精度の向上
- パフォーマンス監視

## 🏗️ アーキテクチャ

```
┌─────────────────────────────────────────┐
│           Knowledge Base                │
├─────────────────┬───────────────────────┤
│ Data Collector  │   Knowledge Manager   │
│                 │                       │
│ gh wrapper      │   Document Store      │
│ LLM driver      │   Query Engine        │
│ PR parser       │   Context Builder     │
└─────────────────┴───────────────────────┘
          │                │
          ▼                ▼
    ┌──────────┐    ┌─────────────┐
    │    gh    │    │  SQLite DB  │
    │ claude   │    │             │
    └──────────┘    └─────────────┘
```

## 📈 期待する効果

- **ドメイン固有知識の獲得**: プロジェクト特有のコーディング規約やパターンを学習
- **レビュー品質向上**: 過去のレビューコメントから一般的でない知見を抽出
- **AIコンテキスト強化**: コード生成・レビュー時に関連する過去の知識を活用
- **継続的改善**: 新しいレビューコメントが自動的にナレッジベースを拡充

## 🧪 テスト

```bash
# 全テストを実行
go test ./... -v

# 特定のパッケージをテスト
go test ./internal/github -v
go test ./internal/llm -v
go test ./internal/database -v
```

## 📄 ライセンス

MIT License

## 🤝 コントリビューション

PRやIssueを歓迎します！詳細は[設計ドキュメント](./docs/)を参照してください。
