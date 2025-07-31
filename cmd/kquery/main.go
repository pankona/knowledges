package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/pankona/knowledges/internal/database"
)

func main() {
	var (
		dbPath    = flag.String("db", "knowledge.db", "Path to database file")
		directory = flag.String("dir", "", "Filter by directory (e.g., 'payment-service')")
		filePath  = flag.String("file", "", "Filter by file path pattern (e.g., '*.rb', 'Orders.ts')")
		author    = flag.String("author", "", "Filter by comment author")
		commentType = flag.String("type", "", "Filter by comment type (e.g., 'security', 'performance')")
		keyword   = flag.String("keyword", "", "Search in summary and original comment text")
		verbose   = flag.Bool("v", false, "Show detailed output including original comment")
	)
	flag.Parse()

	fmt.Println("📊 Knowledge Base Query Tool")
	fmt.Println("============================")

	db, err := database.New(*dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Build query with filters
	baseQuery := `
	SELECT id, summary, original_comment, file_path, directory_path, repository, 
	       pr_number, pr_title, author, comment_type, relevance_score, commented_at
	FROM documents WHERE 1=1`
	
	var conditions []string
	var args []interface{}
	argIndex := 1

	if *directory != "" {
		conditions = append(conditions, fmt.Sprintf(" AND (directory_path LIKE $%d OR file_path LIKE $%d)", argIndex, argIndex+1))
		args = append(args, "%"+*directory+"%", "%"+*directory+"/%")
		argIndex += 2
	}

	if *filePath != "" {
		conditions = append(conditions, fmt.Sprintf(" AND file_path LIKE $%d", argIndex))
		args = append(args, "%"+*filePath+"%")
		argIndex++
	}

	if *author != "" {
		conditions = append(conditions, fmt.Sprintf(" AND author LIKE $%d", argIndex))
		args = append(args, "%"+*author+"%")
		argIndex++
	}

	if *commentType != "" {
		conditions = append(conditions, fmt.Sprintf(" AND comment_type = $%d", argIndex))
		args = append(args, *commentType)
		argIndex++
	}

	if *keyword != "" {
		conditions = append(conditions, fmt.Sprintf(" AND (summary LIKE $%d OR original_comment LIKE $%d)", argIndex, argIndex+1))
		args = append(args, "%"+*keyword+"%", "%"+*keyword+"%")
		argIndex += 2
	}

	for _, condition := range conditions {
		baseQuery += condition
	}

	baseQuery += " ORDER BY relevance_score DESC, commented_at DESC"

	// Execute query
	rows, err := db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		log.Fatalf("Failed to query documents: %v", err)
	}
	defer rows.Close()

	// Count results
	var results []map[string]interface{}
	for rows.Next() {
		var id int64
		var summary, originalComment, filePath, directoryPath, repository, prTitle, author, commentType string
		var prNumber int
		var relevanceScore float64
		var commentedAt string

		err := rows.Scan(&id, &summary, &originalComment, &filePath, &directoryPath, 
			&repository, &prNumber, &prTitle, &author, &commentType, &relevanceScore, &commentedAt)
		if err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}

		results = append(results, map[string]interface{}{
			"id": id, "summary": summary, "originalComment": originalComment,
			"filePath": filePath, "directoryPath": directoryPath, "repository": repository,
			"prNumber": prNumber, "prTitle": prTitle, "author": author,
			"commentType": commentType, "relevanceScore": relevanceScore, "commentedAt": commentedAt,
		})
	}

	if err = rows.Err(); err != nil {
		log.Fatalf("Error iterating rows: %v", err)
	}

	// Show results
	fmt.Printf("📈 Found %d documents", len(results))
	if *directory != "" {
		fmt.Printf(" in directory: %s", *directory)
	}
	if *filePath != "" {
		fmt.Printf(" matching file: %s", *filePath)
	}
	fmt.Println()

	if len(results) == 0 {
		fmt.Println("No documents found matching the criteria.")
		fmt.Println("\nTip: Try broader search terms or use different filters:")
		fmt.Println("  -dir payment-service          # Search by directory")
		fmt.Println("  -file *.rb                    # Search by file pattern")
		fmt.Println("  -author username              # Search by reviewer")
		fmt.Println("  -type implementation          # Search by comment type")
		fmt.Println("  -keyword security             # Search by keyword")
		fmt.Println("  -v                            # Show full comment text")
		fmt.Println("\nAvailable types:")
		fmt.Println("  implementation, security, testing, business, design,")
		fmt.Println("  maintenance, explanation, bug, noise")
		return
	}

	fmt.Println("📝 Documents:")
	fmt.Println("-------------")

	for _, result := range results {
		fmt.Printf("ID: %d\n", result["id"])
		fmt.Printf("📁 File: %s\n", result["filePath"])
		fmt.Printf("📦 Repository: %s\n", result["repository"])
		fmt.Printf("🔗 PR: #%d - %s\n", result["prNumber"], result["prTitle"])
		fmt.Printf("👤 Author: %s\n", result["author"])
		fmt.Printf("🏷️  Type: %s (Score: %.2f)\n", result["commentType"], result["relevanceScore"])
		fmt.Printf("📅 Date: %s\n", result["commentedAt"])
		fmt.Printf("💭 Summary: %s\n", result["summary"])
		
		if *verbose {
			fmt.Printf("📝 Original Comment:\n%s\n", result["originalComment"])
		}
		fmt.Println("---")
	}

	if !*verbose && len(results) > 0 {
		fmt.Println("\nTip: Use -v flag to see full comment text")
	}
}