# LLM CLI統合設計

## 概要

OpenAI APIの代わりに`claude -p`や`gemini -p`などのCLIツールを使用することで、複数のLLMを柔軟に活用し、並列処理による高速化を実現します。

## LLM CLIの活用

### 1. 基本的な使用方法

```bash
# Claude CLIでの分析
echo "以下のコードレビューコメントを分析してください..." | claude -p

# Gemini CLIでの分析（仮想的な例）
echo "以下のコードレビューコメントを分析してください..." | gemini -p

# ファイルからプロンプトを読み込む
claude -p < prompt.txt
```

### 2. 並列実行による高速化

```bash
#!/bin/bash
# 複数のコメントを並列で分析
for comment_file in comments/*.txt; do
    claude -p < "$comment_file" > "results/$(basename $comment_file .txt).json" &
done
wait
```

## Go実装での統合

### LLM Driverの実装

```go
// internal/llm/driver.go
package llm

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "os/exec"
    "sync"
)

type Driver struct {
    command     string   // "claude", "gemini" など
    args        []string // ["-p"] など
    maxParallel int      // 並列実行数
}

func NewDriver(command string, maxParallel int) *Driver {
    return &Driver{
        command:     command,
        args:        []string{"-p"},
        maxParallel: maxParallel,
    }
}

// AnalyzeComment は単一のコメントを分析
func (d *Driver) AnalyzeComment(ctx context.Context, prompt string) (*AnalysisResult, error) {
    cmd := exec.CommandContext(ctx, d.command, d.args...)
    cmd.Stdin = bytes.NewBufferString(prompt)
    
    output, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("LLM command failed: %w", err)
    }
    
    // 結果をパース
    var result AnalysisResult
    if err := json.Unmarshal(output, &result); err != nil {
        return nil, fmt.Errorf("failed to parse LLM output: %w", err)
    }
    
    return &result, nil
}

// AnalyzeBatch は複数のコメントを並列で分析
func (d *Driver) AnalyzeBatch(ctx context.Context, prompts []CommentPrompt) ([]*AnalysisResult, error) {
    results := make([]*AnalysisResult, len(prompts))
    errors := make([]error, len(prompts))
    
    // セマフォで並列数を制限
    sem := make(chan struct{}, d.maxParallel)
    var wg sync.WaitGroup
    
    for i, prompt := range prompts {
        wg.Add(1)
        go func(idx int, p CommentPrompt) {
            defer wg.Done()
            
            // 並列数制限
            sem <- struct{}{}
            defer func() { <-sem }()
            
            result, err := d.AnalyzeComment(ctx, p.Content)
            if err != nil {
                errors[idx] = fmt.Errorf("failed to analyze comment %d: %w", p.ID, err)
                return
            }
            
            results[idx] = result
        }(i, prompt)
    }
    
    wg.Wait()
    
    // エラー集約
    var errs []error
    for _, err := range errors {
        if err != nil {
            errs = append(errs, err)
        }
    }
    
    if len(errs) > 0 {
        return results, fmt.Errorf("some analyses failed: %v", errs)
    }
    
    return results, nil
}
```

### プロンプトテンプレート管理

```go
// internal/llm/prompt.go
package llm

import (
    "bytes"
    "encoding/json"
    "text/template"
)

type PromptBuilder struct {
    templates map[string]*template.Template
}

func NewPromptBuilder() *PromptBuilder {
    pb := &PromptBuilder{
        templates: make(map[string]*template.Template),
    }
    
    // デフォルトテンプレートの登録
    pb.RegisterTemplate("analyze_comment", `
以下のコードレビューコメントを分析し、JSON形式で結果を返してください。

コメント情報:
- ファイル: {{.FilePath}}
- 行番号: {{.LineNumber}}
- 作成者: {{.Author}}
- 内容: {{.Body}}

求める出力形式:
{
  "summary": "1-2文での要約",
  "type": "performance|security|readability|domain|testing|architecture|style|bug|suggestion|question|other",
  "tags": ["タグ1", "タグ2", "タグ3"],
  "relevance_score": 0.0-1.0の数値
}

JSONのみを出力し、それ以外の説明は不要です。
`)
    
    return pb
}

func (pb *PromptBuilder) Build(templateName string, data interface{}) (string, error) {
    tmpl, ok := pb.templates[templateName]
    if !ok {
        return "", fmt.Errorf("template not found: %s", templateName)
    }
    
    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, data); err != nil {
        return "", fmt.Errorf("failed to execute template: %w", err)
    }
    
    return buf.String(), nil
}
```

### 複数LLMの切り替え

```go
// internal/llm/multi_driver.go
package llm

type MultiDriver struct {
    drivers map[string]*Driver
    primary string
}

func NewMultiDriver() *MultiDriver {
    return &MultiDriver{
        drivers: map[string]*Driver{
            "claude": NewDriver("claude", 5),
            "gemini": NewDriver("gemini", 5),
            // 他のLLMも追加可能
        },
        primary: "claude", // デフォルトはClaude
    }
}

// SetPrimary はメインで使用するLLMを設定
func (m *MultiDriver) SetPrimary(name string) error {
    if _, ok := m.drivers[name]; !ok {
        return fmt.Errorf("unknown LLM: %s", name)
    }
    m.primary = name
    return nil
}

// AnalyzeWithFallback はプライマリLLMで失敗した場合に他のLLMにフォールバック
func (m *MultiDriver) AnalyzeWithFallback(ctx context.Context, prompt string) (*AnalysisResult, error) {
    // まずプライマリLLMで試行
    result, err := m.drivers[m.primary].AnalyzeComment(ctx, prompt)
    if err == nil {
        return result, nil
    }
    
    // 失敗した場合、他のLLMで試行
    for name, driver := range m.drivers {
        if name == m.primary {
            continue
        }
        
        result, err := driver.AnalyzeComment(ctx, prompt)
        if err == nil {
            return result, nil
        }
    }
    
    return nil, fmt.Errorf("all LLMs failed")
}
```

### 効率的なバッチ処理

```go
// internal/collector/analyzer.go
package collector

type CommentAnalyzer struct {
    llm           *llm.MultiDriver
    promptBuilder *llm.PromptBuilder
}

func (a *CommentAnalyzer) AnalyzeComments(ctx context.Context, comments []Comment) ([]*Document, error) {
    // プロンプトを準備
    prompts := make([]llm.CommentPrompt, len(comments))
    for i, comment := range comments {
        prompt, err := a.promptBuilder.Build("analyze_comment", comment)
        if err != nil {
            return nil, err
        }
        prompts[i] = llm.CommentPrompt{
            ID:      comment.ID,
            Content: prompt,
        }
    }
    
    // バッチで分析（並列実行）
    results, err := a.llm.AnalyzeBatch(ctx, prompts)
    if err != nil {
        // 部分的な成功も処理
        return a.convertPartialResults(comments, results), err
    }
    
    // 結果をDocumentに変換
    return a.convertToDocuments(comments, results), nil
}
```

## エラーハンドリングとリトライ

```go
// internal/llm/retry.go
package llm

import (
    "context"
    "time"
)

type RetryConfig struct {
    MaxAttempts int
    InitialDelay time.Duration
    MaxDelay     time.Duration
}

func (d *Driver) AnalyzeWithRetry(ctx context.Context, prompt string, config RetryConfig) (*AnalysisResult, error) {
    var lastErr error
    delay := config.InitialDelay
    
    for attempt := 0; attempt < config.MaxAttempts; attempt++ {
        if attempt > 0 {
            select {
            case <-time.After(delay):
            case <-ctx.Done():
                return nil, ctx.Err()
            }
            
            // 指数バックオフ
            delay *= 2
            if delay > config.MaxDelay {
                delay = config.MaxDelay
            }
        }
        
        result, err := d.AnalyzeComment(ctx, prompt)
        if err == nil {
            return result, nil
        }
        
        lastErr = err
    }
    
    return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
```

## 利点とトレードオフ

### 利点
1. **柔軟性**: 複数のLLMを簡単に切り替え・併用可能
2. **並列処理**: 高速な分析が可能
3. **コスト管理**: 各CLIツールの課金体系を活用
4. **デバッグ**: CLIを直接実行して動作確認可能
5. **拡張性**: 新しいLLMの追加が容易

### トレードオフ
1. **依存性**: 各CLIツールのインストールが必要
2. **プロセス起動**: オーバーヘッドが存在
3. **出力形式**: 各CLIの出力形式に依存

### 設定例

```yaml
# config.yaml
llm:
  primary: claude
  parallel: 5
  retry:
    max_attempts: 3
    initial_delay: 1s
    max_delay: 10s
  drivers:
    claude:
      command: claude
      args: [-p]
    gemini:
      command: gemini
      args: [-p]
```