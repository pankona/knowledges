package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config はアプリケーション全体の設定
type Config struct {
	GitHub     GitHubConfig     `yaml:"github"`
	LLM        LLMConfig        `yaml:"llm"`
	Database   DatabaseConfig   `yaml:"database"`
	Collection CollectionConfig `yaml:"collection"`
	Server     ServerConfig     `yaml:"server"`
}

// GitHubConfig はGitHub関連の設定
type GitHubConfig struct {
	Repositories []string `yaml:"repositories"`
}

// LLMConfig はLLM関連の設定
type LLMConfig struct {
	Primary  string                 `yaml:"primary"`
	Parallel int                    `yaml:"parallel"`
	Retry    RetryConfig            `yaml:"retry"`
	Drivers  map[string]DriverConfig `yaml:"drivers"`
}

// RetryConfig はリトライ設定
type RetryConfig struct {
	MaxAttempts  int           `yaml:"max_attempts"`
	InitialDelay time.Duration `yaml:"initial_delay"`
	MaxDelay     time.Duration `yaml:"max_delay"`
}

// DriverConfig はLLMドライバー設定
type DriverConfig struct {
	Command string   `yaml:"command"`
	Args    []string `yaml:"args"`
}

// DatabaseConfig はデータベース設定
type DatabaseConfig struct {
	Path string `yaml:"path"`
}

// CollectionConfig はデータ収集設定
type CollectionConfig struct {
	BatchSize     int `yaml:"batch_size"`
	MaxPRsPerRun  int `yaml:"max_prs_per_run"`
}

// ServerConfig はサーバー設定
type ServerConfig struct {
	Port         int `yaml:"port"`
	ReadTimeout  int `yaml:"read_timeout"`
	WriteTimeout int `yaml:"write_timeout"`
}

// Load は指定されたパスから設定ファイルを読み込みます
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// デフォルト値を設定
	if cfg.LLM.Primary == "" {
		cfg.LLM.Primary = "claude"
	}
	if cfg.LLM.Parallel == 0 {
		cfg.LLM.Parallel = 3
	}
	if cfg.Collection.BatchSize == 0 {
		cfg.Collection.BatchSize = 5
	}
	if cfg.Collection.MaxPRsPerRun == 0 {
		cfg.Collection.MaxPRsPerRun = 100
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 30
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 30
	}

	// Retry設定のデフォルト値
	if cfg.LLM.Retry.MaxAttempts == 0 {
		cfg.LLM.Retry.MaxAttempts = 3
	}
	if cfg.LLM.Retry.InitialDelay == 0 {
		cfg.LLM.Retry.InitialDelay = time.Second
	}
	if cfg.LLM.Retry.MaxDelay == 0 {
		cfg.LLM.Retry.MaxDelay = 10 * time.Second
	}

	return cfg, nil
}