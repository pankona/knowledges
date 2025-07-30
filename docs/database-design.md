# データベース設計

## 概要

Knowledge BaseシステムのデータベースとしてSQLite3を採用します。ローカルで動作し、セットアップが簡単で、PoCには十分な性能を持っています。

## スキーマ設計

### テーブル構成

#### 1. documents テーブル
レビューコメントの要約とメタデータを格納するメインテーブル

```sql
CREATE TABLE documents (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    
    -- コメント情報
    summary TEXT NOT NULL,                -- LLMによる要約
    original_comment TEXT NOT NULL,       -- 元のコメント全文
    thread_context TEXT,                  -- スレッド全体のコンテキスト
    
    -- ファイル情報
    file_path TEXT NOT NULL,              -- 対象ファイルパス
    directory_path TEXT NOT NULL,         -- 所属ディレクトリパス
    language TEXT NOT NULL,               -- プログラミング言語
    line_number INTEGER,                  -- 行番号（該当する場合）
    
    -- PR情報
    repository TEXT NOT NULL,             -- リポジトリ名（owner/repo形式）
    pr_number INTEGER NOT NULL,           -- PR番号
    pr_title TEXT NOT NULL,               -- PRタイトル
    pr_url TEXT NOT NULL,                 -- PR URL
    comment_url TEXT NOT NULL,            -- コメントURL
    
    -- メタデータ
    author TEXT NOT NULL,                 -- コメント作成者
    comment_type TEXT NOT NULL,           -- コメントタイプ（下記参照）
    tags TEXT,                           -- カンマ区切りのタグ
    relevance_score REAL DEFAULT 1.0,     -- 関連度スコア
    
    -- タイムスタンプ
    commented_at DATETIME NOT NULL,       -- コメント作成日時
    collected_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    -- インデックス用
    UNIQUE(repository, pr_number, comment_url)
);

-- インデックス
CREATE INDEX idx_documents_file_path ON documents(file_path);
CREATE INDEX idx_documents_directory_path ON documents(directory_path);
CREATE INDEX idx_documents_language ON documents(language);
CREATE INDEX idx_documents_comment_type ON documents(comment_type);
CREATE INDEX idx_documents_repository ON documents(repository);
CREATE INDEX idx_documents_commented_at ON documents(commented_at);
```

#### 2. collection_progress テーブル
PR収集の進捗を管理するテーブル

```sql
CREATE TABLE collection_progress (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    repository TEXT NOT NULL UNIQUE,
    last_pr_number INTEGER NOT NULL,      -- 最後に処理したPR番号
    last_collected_at DATETIME NOT NULL,  -- 最後の収集日時
    total_prs_processed INTEGER DEFAULT 0,
    total_comments_collected INTEGER DEFAULT 0,
    status TEXT DEFAULT 'active',         -- active, paused, completed
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

#### 3. api_rate_limits テーブル
API制限管理用テーブル

```sql
CREATE TABLE api_rate_limits (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    service TEXT NOT NULL UNIQUE,         -- 'github', 'openai'
    remaining_calls INTEGER,
    reset_at DATETIME,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### データ型定義

#### comment_type の値
- `performance`: パフォーマンス関連
- `security`: セキュリティ関連
- `readability`: 可読性関連
- `domain`: ドメイン知識関連
- `testing`: テスト関連
- `architecture`: アーキテクチャ関連
- `style`: コーディングスタイル関連
- `bug`: バグ指摘
- `suggestion`: 改善提案
- `question`: 質問
- `other`: その他

### クエリ例

#### 1. 特定ファイルに関連するドキュメント取得
```sql
SELECT * FROM documents 
WHERE file_path = ? OR directory_path = ?
ORDER BY relevance_score DESC, commented_at DESC
LIMIT ?;
```

#### 2. 言語別ドキュメント取得
```sql
SELECT * FROM documents 
WHERE language = ? 
  AND comment_type IN (?, ?, ?)
ORDER BY relevance_score DESC
LIMIT ?;
```

#### 3. 最近のドメイン知識取得
```sql
SELECT * FROM documents 
WHERE comment_type = 'domain'
  AND commented_at > datetime('now', '-6 months')
ORDER BY commented_at DESC;
```

## データ管理戦略

### 1. 重複管理
- `repository`, `pr_number`, `comment_url` の組み合わせでユニーク制約
- 更新時は既存レコードを上書き

### 2. データ量管理
- 古いデータの定期的なアーカイブ（将来的な実装）
- relevance_score による優先度管理

### 3. パフォーマンス最適化
- 適切なインデックスの設定
- 定期的なVACUUM実行
- クエリ結果のキャッシング（アプリケーション層）

## マイグレーション戦略

### 初期セットアップ
```sql
-- schema_version テーブル
CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER PRIMARY KEY,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### バージョン管理
- マイグレーションファイルは `migrations/` ディレクトリで管理
- ファイル名: `001_initial_schema.sql`, `002_add_tags.sql` など
- 各マイグレーションは冪等性を保つ

## 想定されるデータ量とパフォーマンス

### PoCフェーズの想定
- リポジトリ数: 1-5
- PR数: 各リポジトリ100-500件
- コメント数: 各PR平均5-10件
- 総ドキュメント数: 2,500-25,000件

### パフォーマンス目標
- 単一ファイルのドキュメント検索: < 100ms
- ディレクトリ単位の検索: < 200ms
- 言語別検索: < 300ms

SQLiteはこの規模であれば十分な性能を発揮します。