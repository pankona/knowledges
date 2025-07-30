package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// AnalysisResult はLLMによる分析結果を表現します
type AnalysisResult struct {
	Summary         string   `json:"summary"`
	Type            string   `json:"type"`
	Tags            []string `json:"tags"`
	RelevanceScore  float64  `json:"relevance_score"`
}

// CommandExecutor はLLMコマンドを実行するインターフェース
type CommandExecutor interface {
	Execute(ctx context.Context, cmd string, args []string, input string) ([]byte, error)
}

// DefaultCommandExecutor は実際のコマンドを実行します
type DefaultCommandExecutor struct{}

func (e *DefaultCommandExecutor) Execute(ctx context.Context, cmd string, args []string, input string) ([]byte, error) {
	command := exec.CommandContext(ctx, cmd, args...)
	command.Stdin = bytes.NewBufferString(input)
	return command.Output()
}

// Driver はLLMコマンドのドライバーです
type Driver struct {
	command  string
	args     []string  
	executor CommandExecutor
}

// NewDriver は新しいDriverを作成します
func NewDriver(command string, args []string) *Driver {
	return &Driver{
		command:  command,
		args:     args,
		executor: &DefaultCommandExecutor{},
	}
}

// SetExecutor はコマンド実行器を設定します（テスト用）
func (d *Driver) SetExecutor(executor CommandExecutor) {
	d.executor = executor
}

// AnalyzeComment は単一のコメントを分析します
func (d *Driver) AnalyzeComment(ctx context.Context, prompt string) (*AnalysisResult, error) {
	if prompt == "" {
		return nil, fmt.Errorf("prompt cannot be empty")
	}

	output, err := d.executor.Execute(ctx, d.command, d.args, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM command failed: %w", err)
	}

	if len(output) == 0 {
		return nil, fmt.Errorf("LLM returned empty response")
	}

	// LLMの出力からJSONを抽出（コードブロック対応）
	jsonOutput := extractJSON(output)

	var result AnalysisResult
	if err := json.Unmarshal(jsonOutput, &result); err != nil {
		return nil, fmt.Errorf("failed to parse LLM output: %w", err)
	}

	return &result, nil
}

// extractJSON はLLMの出力からJSON部分を抽出します
func extractJSON(output []byte) []byte {
	text := string(output)
	
	// ```json ... ``` のコードブロックを探す
	codeBlockPattern := regexp.MustCompile("```(?:json)?\n?([^`]+)```")
	matches := codeBlockPattern.FindStringSubmatch(text)
	if len(matches) > 1 {
		return []byte(strings.TrimSpace(matches[1]))
	}
	
	// { } で囲まれたJSONを探す
	jsonPattern := regexp.MustCompile(`(\{[^{}]*(?:\{[^{}]*\}[^{}]*)*\})`)
	jsonMatches := jsonPattern.FindStringSubmatch(text)
	if len(jsonMatches) > 1 {
		return []byte(strings.TrimSpace(jsonMatches[1]))
	}
	
	// そのまま返す（既にJSONの場合）
	return bytes.TrimSpace(output)
}