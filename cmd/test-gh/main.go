package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pankona/knowledges/internal/github"
)

func main() {
	var (
		repo  = flag.String("repo", "", "Repository in format owner/repo")
		limit = flag.Int("limit", 5, "Number of PRs to fetch")
	)
	flag.Parse()

	if *repo == "" {
		fmt.Println("Usage: test-gh -repo owner/repo [-limit 5]")
		fmt.Println()
		fmt.Println("This tool tests the gh command integration by fetching merged PRs.")
		fmt.Println("Make sure you have 'gh' command installed and authenticated.")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  test-gh -repo pankona/knowledges -limit 3")
		os.Exit(1)
	}

	fmt.Printf("Testing gh command integration with repo: %s\n", *repo)
	fmt.Printf("Fetching last %d merged PRs...\n\n", *limit)

	wrapper := github.NewGHWrapper(*repo)
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	prs, err := wrapper.GetMergedPRs(ctx, *limit)
	if err != nil {
		log.Fatalf("Failed to fetch PRs: %v", err)
	}

	if len(prs) == 0 {
		fmt.Println("No merged PRs found.")
		return
	}

	fmt.Printf("Found %d merged PRs:\n\n", len(prs))
	
	for i, pr := range prs {
		fmt.Printf("%d. PR #%d: %s\n", i+1, pr.Number, pr.Title)
		fmt.Printf("   Author: %s\n", pr.Author.Login)
		fmt.Printf("   Created: %s\n", pr.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("   URL: %s\n", pr.URL)
		fmt.Println()
	}

	fmt.Println("✅ gh command integration test successful!")
}