package collector_test

import (
	"testing"
	"time"

	"github.com/pankona/knowledges/internal/collector"
	"github.com/pankona/knowledges/internal/github"
)

func TestCommentFilter_IsUseful_ValidComments(t *testing.T) {
	filter := collector.NewCommentFilter()

	tests := []struct {
		name    string
		comment github.Comment
		want    bool
	}{
		{
			name: "meaningful review comment",
			comment: github.Comment{
				Body: "Consider using a more descriptive variable name here. This would improve readability.",
				Author: github.Author{Login: "reviewer1"},
			},
			want: true,
		},
		{
			name: "security suggestion",
			comment: github.Comment{
				Body: "This endpoint should validate user input to prevent SQL injection.",
				Author: github.Author{Login: "security-team"},
			},
			want: true,
		},
		{
			name: "architecture feedback",
			comment: github.Comment{
				Body: "This function is doing too many things. Consider breaking it into smaller, focused functions.",
				Author: github.Author{Login: "architect"},
			},
			want: true,
		},
		{
			name: "performance suggestion",
			comment: github.Comment{
				Body: "This loop could be optimized by using a map lookup instead of linear search.",
				Author: github.Author{Login: "reviewer2"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filter.IsUseful(tt.comment)
			if got != tt.want {
				t.Errorf("IsUseful() = %v, want %v for comment: %q", got, tt.want, tt.comment.Body)
			}
		})
	}
}

func TestCommentFilter_IsUseful_FilterOutUseless(t *testing.T) {
	filter := collector.NewCommentFilter()

	tests := []struct {
		name    string
		comment github.Comment
		want    bool
	}{
		{
			name: "LGTM only",
			comment: github.Comment{
				Body: "LGTM",
				Author: github.Author{Login: "reviewer1"},
			},
			want: false,
		},
		{
			name: "emoji only",
			comment: github.Comment{
				Body: "👍",
				Author: github.Author{Login: "reviewer1"},
			},
			want: false,
		},
		{
			name: "single word",
			comment: github.Comment{
				Body: "approved",
				Author: github.Author{Login: "reviewer1"},
			},
			want: false,
		},
		{
			name: "very short comment",
			comment: github.Comment{
				Body: "ok",
				Author: github.Author{Login: "reviewer1"},
			},
			want: false,
		},
		{
			name: "thanks only",
			comment: github.Comment{
				Body: "thanks!",
				Author: github.Author{Login: "author1"},
			},
			want: false,
		},
		{
			name: "done comment",
			comment: github.Comment{
				Body: "done",
				Author: github.Author{Login: "author1"},
			},
			want: false,
		},
		{
			name: "fixed comment",
			comment: github.Comment{
				Body: "fixed",
				Author: github.Author{Login: "author1"},
			},
			want: false,
		},
		{
			name: "automated comment",
			comment: github.Comment{
				Body: "Automatically generated comment from CI",
				Author: github.Author{Login: "github-actions[bot]"},
			},
			want: false,
		},
		{
			name: "dependabot comment",
			comment: github.Comment{
				Body: "Bumps version from 1.0.0 to 1.0.1",
				Author: github.Author{Login: "dependabot[bot]"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filter.IsUseful(tt.comment)
			if got != tt.want {
				t.Errorf("IsUseful() = %v, want %v for comment: %q", got, tt.want, tt.comment.Body)
			}
		})
	}
}

func TestCommentFilter_FilterComments(t *testing.T) {
	filter := collector.NewCommentFilter()

	comments := []github.Comment{
		{
			Body: "This function could benefit from better error handling.",
			Author: github.Author{Login: "reviewer1"},
			CreatedAt: time.Now(),
		},
		{
			Body: "LGTM",
			Author: github.Author{Login: "reviewer2"},
			CreatedAt: time.Now(),
		},
		{
			Body: "Consider adding unit tests for this function.",
			Author: github.Author{Login: "reviewer3"},
			CreatedAt: time.Now(),
		},
		{
			Body: "👍",
			Author: github.Author{Login: "reviewer4"},
			CreatedAt: time.Now(),
		},
		{
			Body: "This variable name is confusing. Could you rename it to something more descriptive?",
			Author: github.Author{Login: "reviewer5"},
			CreatedAt: time.Now(),
		},
	}

	filtered := filter.FilterComments(comments)

	expectedCount := 3 // Only meaningful comments should remain
	if len(filtered) != expectedCount {
		t.Errorf("FilterComments() returned %d comments, want %d", len(filtered), expectedCount)
	}

	// Check that filtered comments are the meaningful ones
	meaningfulBodies := []string{
		"This function could benefit from better error handling.",
		"Consider adding unit tests for this function.",
		"This variable name is confusing. Could you rename it to something more descriptive?",
	}

	for i, comment := range filtered {
		if comment.Body != meaningfulBodies[i] {
			t.Errorf("FilterComments() filtered comment %d: got %q, want %q", i, comment.Body, meaningfulBodies[i])
		}
	}
}

func TestCommentFilter_HasMinimumLength(t *testing.T) {
	filter := collector.NewCommentFilter()

	tests := []struct {
		body string
		want bool
	}{
		{"This is a meaningful comment with sufficient length", true},
		{"Short but meaningful suggestion", true},
		{"ok", false},
		{"", false},
		{"a", false},
		{"This is exactly twenty characters", true},
	}

	for _, tt := range tests {
		got := filter.HasMinimumLength(tt.body)
		if got != tt.want {
			t.Errorf("HasMinimumLength(%q) = %v, want %v", tt.body, got, tt.want)
		}
	}
}