package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// New はSQLiteデータベースへの接続を作成します
func New(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// SQLiteの設定
	pragmas := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to set pragma: %w", err)
		}
	}

	return db, nil
}

// Migrate はデータベースのマイグレーションを実行します
func Migrate(db *sql.DB) error {
	// documentsテーブルの作成
	createDocumentsTable := `
	CREATE TABLE IF NOT EXISTS documents (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		
		-- コメント情報
		summary TEXT NOT NULL,
		original_comment TEXT NOT NULL,
		thread_context TEXT,
		
		-- ファイル情報
		file_path TEXT NOT NULL,
		directory_path TEXT NOT NULL,
		language TEXT NOT NULL,
		line_number INTEGER,
		
		-- PR情報
		repository TEXT NOT NULL,
		pr_number INTEGER NOT NULL,
		pr_title TEXT NOT NULL,
		pr_url TEXT NOT NULL,
		comment_url TEXT NOT NULL,
		
		-- メタデータ
		author TEXT NOT NULL,
		comment_type TEXT NOT NULL,
		tags TEXT,
		relevance_score REAL DEFAULT 1.0,
		
		-- タイムスタンプ
		commented_at DATETIME NOT NULL,
		collected_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		
		-- ユニーク制約
		UNIQUE(repository, pr_number, comment_url)
	)`

	if _, err := db.Exec(createDocumentsTable); err != nil {
		return fmt.Errorf("failed to create documents table: %w", err)
	}

	// インデックスの作成
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_documents_file_path ON documents(file_path)",
		"CREATE INDEX IF NOT EXISTS idx_documents_directory_path ON documents(directory_path)",
		"CREATE INDEX IF NOT EXISTS idx_documents_language ON documents(language)",
		"CREATE INDEX IF NOT EXISTS idx_documents_comment_type ON documents(comment_type)",
		"CREATE INDEX IF NOT EXISTS idx_documents_repository ON documents(repository)",
		"CREATE INDEX IF NOT EXISTS idx_documents_commented_at ON documents(commented_at)",
	}

	for _, index := range indexes {
		if _, err := db.Exec(index); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	// collection_progressテーブルの作成
	createProgressTable := `
	CREATE TABLE IF NOT EXISTS collection_progress (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		repository TEXT NOT NULL UNIQUE,
		last_pr_number INTEGER NOT NULL,
		last_collected_at DATETIME NOT NULL,
		total_prs_processed INTEGER DEFAULT 0,
		total_comments_collected INTEGER DEFAULT 0,
		status TEXT DEFAULT 'active',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`

	if _, err := db.Exec(createProgressTable); err != nil {
		return fmt.Errorf("failed to create collection_progress table: %w", err)
	}

	return nil
}