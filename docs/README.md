# Knowledge Base システム設計ドキュメント

## 概要

このディレクトリには、Knowledge Baseシステムの設計ドキュメントが含まれています。

## ドキュメント構成

1. **[architecture.md](./architecture.md)** - システムアーキテクチャ設計
   - システム全体の構成
   - 主要コンポーネントの説明
   - データフロー
   - 技術スタック

2. **[database-design.md](./database-design.md)** - データベース設計
   - SQLiteスキーマ定義
   - テーブル構成とインデックス
   - クエリ例
   - データ管理戦略

3. **[component-design.md](./component-design.md)** - コンポーネント設計
   - パッケージ構成
   - インターフェース定義
   - データ構造
   - エラー処理

4. **[data-collection-strategy.md](./data-collection-strategy.md)** - データ収集戦略
   - 収集対象の優先順位付け
   - API制限対策
   - 効率的なデータ取得方法
   - LLM分析の最適化

5. **[api-design.md](./api-design.md)** - API設計
   - RESTエンドポイント定義
   - リクエスト/レスポンス形式
   - CLI統合
   - エラーハンドリング

## 実装優先順位（PoC向け）

### フェーズ1: 基本機能実装
1. データベースセットアップ
2. GitHub APIクライアント実装
3. 基本的なPR収集機能
4. シンプルなLLM分析

### フェーズ2: API実装
1. ドキュメント検索API
2. コンテキスト生成API
3. CLIツール

### フェーズ3: 拡張機能
1. 増分収集
2. 高度なフィルタリング
3. パフォーマンス最適化

## 開発開始手順

1. **環境設定**
   ```bash
   # Go modules初期化
   go mod init github.com/pankona/knowledges
   
   # 必要な依存関係をインストール
   go get github.com/google/go-github/v58
   go get github.com/mattn/go-sqlite3
   go get github.com/sashabaranov/go-openai
   ```

2. **設定ファイル作成**
   ```yaml
   # config.yaml
   github:
     token: ${GITHUB_TOKEN}
     repositories:
       - owner/repo1
   
   llm:
     provider: openai
     api_key: ${OPENAI_API_KEY}
     model: gpt-3.5-turbo
   
   database:
     path: ./knowledge.db
   ```

3. **データベース初期化**
   ```bash
   # マイグレーション実行
   go run cmd/migrate/main.go up
   ```

## 注意事項

- このシステムはPoCとして設計されており、本番環境での使用には追加の考慮が必要です
- API制限に注意し、適切なレート制限を実装してください
- センシティブな情報（トークン、APIキー）は環境変数で管理してください
- 大量のデータ収集を行う前に、小規模なテストを実施してください

## 今後の拡張可能性

1. **複数のVCSプロバイダー対応**
   - GitLab
   - Bitbucket
   
2. **より高度な分析**
   - コードの変更内容との相関分析
   - チーム固有のパターン学習
   
3. **インテグレーション拡張**
   - IDE プラグイン
   - CI/CD パイプライン統合
   - Slack/Teams通知

## 参考リンク

- [GitHub REST API](https://docs.github.com/en/rest)
- [GitHub GraphQL API](https://docs.github.com/en/graphql)
- [OpenAI API](https://platform.openai.com/docs/api-reference)
- [SQLite Documentation](https://www.sqlite.org/docs.html)