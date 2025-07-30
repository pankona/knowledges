package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pankona/knowledges/internal/collector"
	"github.com/pankona/knowledges/internal/database"
	"github.com/pankona/knowledges/internal/github"
	"github.com/pankona/knowledges/internal/llm"
	"github.com/pankona/knowledges/pkg/config"
	"github.com/pankona/knowledges/pkg/models"
)

func main() {
	var (
		configPath     = flag.String("config", "config.yaml", "Path to config file")
		repo           = flag.String("repo", "", "Repository to collect from (overrides config)")
		limit          = flag.Int("limit", 1, "Number of PRs to process")
		label          = flag.String("label", "", "Filter PRs by label (e.g., 'payment-service')")
		excludeBots    = flag.Bool("exclude-bots", true, "Exclude PRs created by bots")
		skipProcessed  = flag.Bool("skip-processed", true, "Skip already processed PRs (default: true)")
		prURL          = flag.String("pr-url", "", "Process specific PR by URL (forces reprocessing)")
	)
	flag.Parse()

	fmt.Println("🚀 Knowledge Base Collector (Minimal PoC)")
	fmt.Println("==========================================")

	// Load config
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Override repo if specified
	targetRepo := *repo
	if targetRepo == "" && len(cfg.GitHub.Repositories) > 0 {
		targetRepo = cfg.GitHub.Repositories[0]
	}
	if targetRepo == "" && *prURL == "" {
		fmt.Println("Usage: collector -repo owner/repo [-limit 1] [-label label-name] [-exclude-bots] [-skip-processed] [-config config.yaml]")
		fmt.Println("   or: collector -pr-url https://github.com/owner/repo/pull/123")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  collector -repo owner/repo -limit 2")
		fmt.Println("  collector -repo owner/repo -label bug -limit 10")
		fmt.Println("  collector -repo owner/repo -exclude-bots=false")
		fmt.Println("  collector -repo owner/repo -skip-processed=false  # Reprocess all PRs")
		fmt.Println("  collector -pr-url https://github.com/owner/repo/pull/123  # Reprocess specific PR")
		os.Exit(1)
	}

	// Handle PR URL mode
	if *prURL != "" {
		fmt.Printf("🔄 Single PR reprocessing mode\n")
		fmt.Printf("🔗 PR URL: %s\n", *prURL)
		// Extract repo and PR number from URL
		// e.g., https://github.com/owner/repo/pull/123
		parts := strings.Split(*prURL, "/")
		if len(parts) < 6 || parts[2] != "github.com" || parts[5] != "pull" {
			log.Fatalf("Invalid PR URL format. Expected: https://github.com/owner/repo/pull/123")
		}
		targetRepo = parts[3] + "/" + parts[4]
		prNumber, err := strconv.Atoi(parts[6])
		if err != nil {
			log.Fatalf("Invalid PR number in URL: %v", err)
		}
		fmt.Printf("📦 Extracted repository: %s\n", targetRepo)
		fmt.Printf("🔢 PR number: %d\n", prNumber)
	} else {
		fmt.Printf("📦 Target repository: %s\n", targetRepo)
		fmt.Printf("🔢 Processing limit: %d PRs\n", *limit)
		if *label != "" {
			fmt.Printf("🏷️  Label filter: %s\n", *label)
		}
		if *excludeBots {
			fmt.Printf("🤖 Excluding bot PRs: enabled\n")
		}
		if *skipProcessed {
			fmt.Printf("⏭️  Skip processed PRs: enabled\n")
		}
	}
	fmt.Println()

	// Initialize database
	dbPath := cfg.Database.Path
	if !filepath.IsAbs(dbPath) {
		dbPath = filepath.Join(".", dbPath)
	}

	fmt.Printf("🗄️  Initializing database: %s\n", dbPath)
	db, err := database.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	fmt.Println("✅ Database ready")

	// Initialize components
	ghWrapper := github.NewGHWrapper(targetRepo)
	llmDriver := llm.NewDriver("claude", []string{"-p"})
	commentFilter := collector.NewCommentFilter()
	fileInfoExtractor := collector.NewFileInfoExtractor()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Step 1: Fetch PRs
	var prs []github.PullRequest

	if *prURL != "" {
		// Single PR mode: extract PR number and fetch specific PR
		parts := strings.Split(*prURL, "/")
		prNumber, _ := strconv.Atoi(parts[6])
		
		fmt.Printf("\n🔄 Deleting existing data for PR #%d...\n", prNumber)
		err = deletePRData(ctx, db, targetRepo, prNumber)
		if err != nil {
			log.Printf("⚠️  Failed to delete existing PR data: %v", err)
		} else {
			fmt.Println("✅ Existing PR data deleted")
		}
		
		fmt.Printf("📥 Fetching PR #%d from %s...\n", prNumber, targetRepo)
		// Fetch actual PR information from GitHub
		pr, err := ghWrapper.GetPR(ctx, prNumber)
		if err != nil {
			log.Fatalf("Failed to fetch PR #%d: %v", prNumber, err)
		}
		prs = []github.PullRequest{*pr}
	} else {
		// Regular mode: fetch multiple PRs
		var fetchMessage string
		if *label != "" {
			fetchMessage = fmt.Sprintf("📥 Fetching %d merged PRs with label '%s' from %s", *limit, *label, targetRepo)
		} else {
			fetchMessage = fmt.Sprintf("📥 Fetching %d merged PRs from %s", *limit, targetRepo)
		}
		if *excludeBots {
			fetchMessage += " (excluding bots)"
		}
		fmt.Printf("\n%s...\n", fetchMessage)

		if *excludeBots {
			prs, err = ghWrapper.GetMergedPRsExcludingBots(ctx, *limit, *label)
		} else if *label != "" {
			prs, err = ghWrapper.GetMergedPRsWithLabel(ctx, *limit, *label)
		} else {
			prs, err = ghWrapper.GetMergedPRs(ctx, *limit)
		}

		if err != nil {
			log.Fatalf("Failed to fetch PRs: %v", err)
		}

		if len(prs) == 0 {
			fmt.Println("⚠️  No merged PRs found")
			return
		}

		// Filter out already processed PRs if skip-processed is enabled
		if *skipProcessed {
			fmt.Printf("🔍 Filtering out processed PRs...\n")
			processedPRs, err := getProcessedPRNumbers(ctx, db, targetRepo)
			if err != nil {
				log.Printf("⚠️  Failed to get processed PRs: %v", err)
			} else {
				originalCount := len(prs)
				prs = filterUnprocessedPRs(prs, processedPRs)
				skippedCount := originalCount - len(prs)
				if skippedCount > 0 {
					fmt.Printf("⏭️  Skipped %d already processed PRs\n", skippedCount)
				}
			}
		}

		fmt.Printf("✅ Found %d PRs to process\n", len(prs))
		
		if len(prs) == 0 {
			fmt.Println("ℹ️  All PRs have already been processed")
			fmt.Println("💡 Use -skip-processed=false to reprocess all PRs")
			return
		}
	}

	// Step 2: Process each PR and its comments
	var totalDocuments int

	for i, pr := range prs {
		fmt.Printf("\n🔍 Processing PR #%d (%d/%d): %s\n", pr.Number, i+1, len(prs), pr.Title)
		fmt.Printf("👤 Author: %s\n", pr.Author.Login)
		fmt.Printf("📅 Created: %s\n", pr.CreatedAt.Format("2006-01-02 15:04:05"))

		// Fetch actual PR comments
		fmt.Printf("📥 Fetching PR comments...\n")
		comments, err := ghWrapper.GetPRComments(ctx, pr.Number)
		if err != nil {
			fmt.Printf("⚠️  Failed to fetch comments for PR #%d: %v\n", pr.Number, err)
			continue
		}

		if len(comments) == 0 {
			fmt.Printf("ℹ️  No comments found for PR #%d\n", pr.Number)
			continue
		}

		fmt.Printf("✅ Found %d comments\n", len(comments))

		// Filter useful comments
		fmt.Printf("🔍 Filtering useful comments...\n")
		filteredComments := commentFilter.FilterComments(comments)

		if len(filteredComments) == 0 {
			fmt.Printf("ℹ️  No useful comments found after filtering\n")
			continue
		}

		fmt.Printf("✅ %d useful comments after filtering\n", len(filteredComments))

		// Process each filtered comment
		for j, comment := range filteredComments {
			fmt.Printf("\n🤖 Analyzing comment %d/%d...\n", j+1, len(filteredComments))
			fmt.Printf("💬 Author: %s\n", comment.Author.Login)
			fmt.Printf("📂 File: %s:%d\n", comment.FilePath, comment.LineNumber)
			fmt.Printf("📝 Content: %.100s...\n", comment.Body)

			// Extract file information
			language := fileInfoExtractor.ExtractLanguage(comment.FilePath)
			directory := fileInfoExtractor.ExtractDirectory(comment.FilePath)

			// Create prompt for LLM analysis
			prompt := fmt.Sprintf(`
Analyze this code review comment and provide structured output in JSON format:

Context:
- Repository: %s  
- PR #%d: %s
- File: %s (line %d)
- Language: %s
- Author: %s

Comment:
%s

Please provide:
{
  "summary": "Detailed actionable review guidance (3-8 sentences) that includes: 1) What to check/ensure, 2) Why it matters (context/reasoning), 3) Specific implementation details or patterns, 4) Code examples if relevant (before/after snippets)",
  "type": "implementation|security|testing|business|design|maintenance|explanation|bug|noise",
  "tags": ["relevant", "keywords", "max-5-tags"],
  "relevance_score": 0.0-1.0
}

Type definitions:
- implementation: Code improvement suggestions (performance, refactoring, code quality)
- security: Security-related concerns or suggestions
- testing: Test-related comments (test methods, coverage, test cases)
- business: Business logic, domain knowledge, specifications
- design: Architecture, design patterns, structure
- maintenance: Maintainability, readability, naming, code style
- explanation: Explanations, questions, information sharing
- bug: Bug reports or issue identification
- noise: Low-value comments (use relevance_score 0.1-0.3)

Summary guidelines:
- Start with actionable language: "When reviewing X, ensure...", "Check that...", "Verify..."
- Explain the reasoning: why this matters, what problems it prevents
- Include specific technical details: patterns, methods, configurations
- Add code examples when helpful (use backticks for inline code, triple backticks for blocks)
- Reference specific files, functions, or patterns mentioned in the comment
- Extract generalizable principles that apply to similar situations
- Make it comprehensive enough that a reviewer can apply the knowledge without reading the original comment

Return only the JSON, no other text.
`, targetRepo, pr.Number, pr.Title, comment.FilePath, comment.LineNumber, language, comment.Author.Login, comment.Body)

			// Analyze with LLM
			result, err := llmDriver.AnalyzeComment(ctx, prompt)
			if err != nil {
				fmt.Printf("⚠️  LLM analysis failed: %v\n", err)
				fmt.Println("📝 Creating fallback analysis...")

				// Create fallback result
				result = &llm.AnalysisResult{
					Summary:        fmt.Sprintf("Review comment about %s", comment.FilePath),
					Type:           "suggestion",
					Tags:           []string{"review", "feedback"},
					RelevanceScore: 0.7,
				}
			} else {
				fmt.Println("✅ LLM analysis completed")
			}

			// Create document
			document := &models.Document{
				Summary:         result.Summary,
				OriginalComment: comment.Body,
				FilePath:        comment.FilePath,
				DirectoryPath:   directory,
				Language:        language,
				Repository:      targetRepo,
				PRNumber:        pr.Number,
				PRTitle:         pr.Title,
				PRURL:           pr.URL,
				CommentURL:      comment.URL,
				Author:          comment.Author.Login,
				CommentType:     result.Type,
				Tags:            result.Tags,
				RelevanceScore:  result.RelevanceScore,
				CommentedAt:     comment.CreatedAt,
				CollectedAt:     time.Now(),
				UpdatedAt:       time.Now(),
			}

			// Save to database
			err = saveDocument(ctx, db, document)
			if err != nil {
				fmt.Printf("⚠️  Failed to save document: %v\n", err)
				continue
			}

			totalDocuments++
			fmt.Printf("✅ Document %d saved\n", totalDocuments)
		}
	}

	// Step 3: Final verification
	fmt.Printf("\n🔍 Verifying saved data...\n")

	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM documents WHERE repository = ?", targetRepo).Scan(&count)
	if err != nil {
		log.Fatalf("Failed to query documents: %v", err)
	}

	fmt.Printf("📊 Total documents for %s: %d\n", targetRepo, count)

	// Success!
	fmt.Printf("\n🎉 PoC Collection completed successfully!\n")
	fmt.Println("====================================")
	fmt.Printf("✅ Processed %d PRs\n", len(prs))
	fmt.Printf("✅ Created %d documents\n", totalDocuments)
	fmt.Printf("✅ Saved to database: %s\n", dbPath)
	fmt.Println("\nNext steps:")
	fmt.Println("- Add parallel processing")
	fmt.Println("- Implement REST API")
	fmt.Println("- Add batch processing for large repositories")
	fmt.Println("- Enhance LLM prompts for better analysis")
}

// saveDocument はドキュメントをデータベースに保存します
func saveDocument(ctx context.Context, db *sql.DB, document *models.Document) error {
	query := `
	INSERT INTO documents (
		summary, original_comment, file_path, directory_path, language,
		repository, pr_number, pr_title, pr_url, comment_url,
		author, comment_type, tags, relevance_score,
		commented_at, collected_at, updated_at
	) VALUES (
		?, ?, ?, ?, ?,
		?, ?, ?, ?, ?,
		?, ?, ?, ?,
		?, ?, ?
	) ON CONFLICT(repository, pr_number, comment_url) DO UPDATE SET
		summary = excluded.summary,
		original_comment = excluded.original_comment,
		file_path = excluded.file_path,
		directory_path = excluded.directory_path,
		language = excluded.language,
		pr_title = excluded.pr_title,
		author = excluded.author,
		comment_type = excluded.comment_type,
		tags = excluded.tags,
		relevance_score = excluded.relevance_score,
		updated_at = excluded.updated_at
	`

	tagsStr := ""
	if len(document.Tags) > 0 {
		tagsStr = fmt.Sprintf("%v", document.Tags) // Simple serialization
	}

	_, err := db.ExecContext(ctx, query,
		document.Summary, document.OriginalComment, document.FilePath,
		document.DirectoryPath, document.Language,
		document.Repository, document.PRNumber, document.PRTitle,
		document.PRURL, document.CommentURL,
		document.Author, document.CommentType, tagsStr, document.RelevanceScore,
		document.CommentedAt, document.CollectedAt, document.UpdatedAt,
	)

	return err
}

// getProcessedPRNumbers は指定されたリポジトリで既に処理済みのPR番号リストを取得します
func getProcessedPRNumbers(ctx context.Context, db *sql.DB, repository string) (map[int]bool, error) {
	query := `SELECT DISTINCT pr_number FROM documents WHERE repository = ?`
	rows, err := db.QueryContext(ctx, query, repository)
	if err != nil {
		return nil, fmt.Errorf("failed to query processed PRs: %w", err)
	}
	defer rows.Close()

	processedPRs := make(map[int]bool)
	for rows.Next() {
		var prNumber int
		if err := rows.Scan(&prNumber); err != nil {
			return nil, fmt.Errorf("failed to scan PR number: %w", err)
		}
		processedPRs[prNumber] = true
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating processed PRs: %w", err)
	}

	return processedPRs, nil
}

// filterUnprocessedPRs は未処理のPRのみを返します
func filterUnprocessedPRs(prs []github.PullRequest, processedPRs map[int]bool) []github.PullRequest {
	var unprocessedPRs []github.PullRequest
	for _, pr := range prs {
		if !processedPRs[pr.Number] {
			unprocessedPRs = append(unprocessedPRs, pr)
		}
	}
	return unprocessedPRs
}

// deletePRData は指定されたPRに関連するすべてのドキュメントを削除します
func deletePRData(ctx context.Context, db *sql.DB, repository string, prNumber int) error {
	query := `DELETE FROM documents WHERE repository = ? AND pr_number = ?`
	result, err := db.ExecContext(ctx, query, repository, prNumber)
	if err != nil {
		return fmt.Errorf("failed to delete PR data: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected > 0 {
		fmt.Printf("🗑️  Deleted %d existing documents for PR #%d\n", rowsAffected, prNumber)
	} else {
		fmt.Printf("ℹ️  No existing documents found for PR #%d\n", prNumber)
	}

	return nil
}
