# Knowledge Base

PRレビューコメントを収集・分析するツール

## セットアップ

```bash
# ビルド
go build -o bin/collector ./cmd/collector
go build -o bin/query ./cmd/query

# 認証設定
gh auth login
claude auth

# 設定ファイル
cp config.yaml.example config.yaml
```

## 使用方法

### collector - PRコメント収集

```bash
# 基本的な使用方法
./bin/collector -repo owner/repo -limit 10

# オプション
-repo string       # 対象リポジトリ (owner/repo 形式)
-limit int         # 処理するPR数 (default: 1)
-label string      # ラベルでフィルタ
-exclude-bots      # ボットPRを除外 (default: true)
-skip-processed    # 処理済みPRをスキップ (default: true)
-pr-url string     # 特定PRを再処理
-config string     # 設定ファイル (default: config.yaml)

# 使用例
./bin/collector -repo golang/go -limit 5
./bin/collector -repo owner/repo -label "bug" -limit 20
./bin/collector -repo owner/repo -skip-processed=false -limit 5
./bin/collector -pr-url https://github.com/owner/repo/pull/123
```

### query - データ検索

```bash
# 基本的な使用方法
./bin/query

# オプション
-db string         # データベースファイル (default: knowledge.db)
-dir string        # ディレクトリで絞り込み
-file string       # ファイルパターンで絞り込み
-author string     # 作成者で絞り込み
-type string       # コメント種類で絞り込み
-keyword string    # キーワード検索
-v                 # 詳細表示

# 使用例
./bin/query -dir "src/components"
./bin/query -file "*.ts"
./bin/query -type security
./bin/query -keyword "authentication"
./bin/query -dir "src/" -v
```

## コメント分類

コメントは以下の9種類に分類されます：

- `implementation` - コード改善提案、パフォーマンス、リファクタリング
- `security` - セキュリティ関連の指摘
- `testing` - テスト関連のコメント
- `business` - ビジネスロジック、仕様に関する指摘
- `design` - アーキテクチャ、設計パターンに関する提案
- `maintenance` - 保守性、可読性に関する改善提案
- `explanation` - 説明、質問、情報共有
- `bug` - バグ報告、問題の指摘
- `noise` - 価値の低いコメント (LGTM等)
