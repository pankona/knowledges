# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Knowledge Base system that collects PR review comments from GitHub repositories and uses LLM analysis to create a searchable knowledge base for AI-assisted code generation and review. The system uses GitHub CLI (`gh` command) instead of GitHub API for authentication simplicity, and integrates with Claude CLI for comment analysis.

## Essential Commands

### Building
```bash
go build -o bin/collector ./cmd/collector
go build -o bin/query ./cmd/query  
go build -o bin/test-gh ./cmd/test-gh
```

### Testing
```bash
# Run all tests
go test ./... -v

# Test specific packages
go test ./internal/github -v
go test ./internal/llm -v
go test ./internal/database -v
go test ./cmd/collector -v

# Run single test
go test ./internal/github -run TestGHWrapper_GetPR_Success -v
```

### Running the Tools
```bash
# Test GitHub CLI integration
./bin/test-gh -repo golang/go -limit 3

# Collect PR comments (skips processed PRs by default)
./bin/collector -repo owner/repo -limit 5 -label "bug"

# Reprocess a specific PR (deletes existing data first)
./bin/collector -pr-url https://github.com/owner/repo/pull/123

# Query the knowledge base
./bin/query -dir "src/" -type "security" -v
```

## Architecture Overview

The system has two main components:

### Data Collector (`cmd/collector`)
- **GH Wrapper** (`internal/github/gh_wrapper.go`): Wraps `gh` CLI commands for PR and comment fetching
- **LLM Driver** (`internal/llm/driver.go`): Integrates with Claude CLI for comment analysis with JSON extraction
- **Comment Filter** (`internal/collector/filter.go`): Filters out noise (LGTM, bots, empty comments)
- **File Info Extractor** (`internal/collector/file_info.go`): Extracts language and directory information

### Knowledge Manager (`cmd/query`)
- **Database Layer** (`internal/database/db.go`): SQLite with migration support and multi-repository schema
- **Query Engine**: Supports filtering by directory, file patterns, author, comment type, and keywords

## Key Design Decisions

### CLI Integration Over APIs
- Uses `gh` command instead of GitHub API (no token management needed)
- Uses `claude -p` command instead of OpenAI API (enables parallel processing)
- LLM responses are wrapped in markdown code blocks requiring JSON extraction

### Processing Efficiency
- **Skip processed PRs by default**: `--skip-processed=true` (default behavior)
- **PR-level duplicate detection**: Queries existing `documents` table for processed PR numbers
- **Single PR reprocessing**: `--pr-url` option deletes existing PR data before reprocessing

### Comment Classification System
9-type classification: `implementation`, `security`, `testing`, `business`, `design`, `maintenance`, `explanation`, `bug`, `noise`

### Database Schema
- **Multi-repository support**: `repository` column with appropriate indexes
- **Unique constraint**: `(repository, pr_number, comment_url)` prevents duplicates
- **Rich metadata**: Includes file paths, directories, languages, relevance scores

## Test-Driven Development (TDD)
The codebase follows TDD methodology with:
- Comprehensive unit tests for all components
- Mock executors for external command testing
- Database integration tests with temporary SQLite files
- Test coverage for both success and error scenarios

## Configuration Structure
- `config.yaml`: Main configuration (see `config.yaml.example`)
- Supports multiple repositories in `github.repositories` array (collector uses first one)
- `server` section is for future REST API (not yet implemented)
- LLM configuration supports different CLI commands and arguments