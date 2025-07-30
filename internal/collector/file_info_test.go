package collector_test

import (
	"testing"

	"github.com/pankona/knowledges/internal/collector"
)

func TestFileInfoExtractor_ExtractLanguage(t *testing.T) {
	extractor := collector.NewFileInfoExtractor()

	tests := []struct {
		filePath string
		want     string
	}{
		{"main.go", "go"},
		{"src/utils.go", "go"},
		{"app.js", "javascript"},
		{"src/components/Button.tsx", "typescript"},
		{"styles.css", "css"},
		{"README.md", "markdown"},
		{"Dockerfile", "dockerfile"},
		{"config.yaml", "yaml"},
		{"config.yml", "yaml"},
		{"data.json", "json"},
		{"script.py", "python"},
		{"main.java", "java"},
		{"Component.vue", "vue"},
		{"style.scss", "scss"},
		{"test.rb", "ruby"},
		{"main.rs", "rust"},
		{"app.php", "php"},
		{"main.c", "c"},
		{"main.cpp", "cpp"},
		{"main.h", "c"},
		{"main.hpp", "cpp"},
		{"Program.cs", "csharp"},
		{"no_extension", "unknown"},
		{"", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			got := extractor.ExtractLanguage(tt.filePath)
			if got != tt.want {
				t.Errorf("ExtractLanguage(%q) = %q, want %q", tt.filePath, got, tt.want)
			}
		})
	}
}

func TestFileInfoExtractor_ExtractDirectory(t *testing.T) {
	extractor := collector.NewFileInfoExtractor()

	tests := []struct {
		filePath string
		want     string
	}{
		{"main.go", "."},
		{"src/main.go", "src"},
		{"src/utils/helper.go", "src/utils"},
		{"a/b/c/d/file.js", "a/b/c/d"},
		{"./relative/path.py", "relative"},
		{"/absolute/path/file.java", "/absolute/path"},
		{"", "."},
		{"no_slash", "."},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			got := extractor.ExtractDirectory(tt.filePath)
			if got != tt.want {
				t.Errorf("ExtractDirectory(%q) = %q, want %q", tt.filePath, got, tt.want)
			}
		})
	}
}

func TestFileInfoExtractor_IsTestFile(t *testing.T) {
	extractor := collector.NewFileInfoExtractor()

	tests := []struct {
		filePath string
		want     bool
	}{
		{"main_test.go", true},
		{"utils_test.go", true},
		{"test_utils.go", true},
		{"src/components/Button.test.js", true},
		{"src/components/Button.spec.js", true},
		{"tests/integration_test.py", true},
		{"test/unit/helper_test.js", true},
		{"__tests__/Component.test.jsx", true},
		{"spec/user_spec.rb", true},
		{"main.go", false},
		{"utils.js", false},
		{"README.md", false},
		{"testing_utils.go", false}, // testing in middle doesn't count
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			got := extractor.IsTestFile(tt.filePath)
			if got != tt.want {
				t.Errorf("IsTestFile(%q) = %v, want %v", tt.filePath, got, tt.want)
			}
		})
	}
}

func TestFileInfoExtractor_IsConfigFile(t *testing.T) {
	extractor := collector.NewFileInfoExtractor()

	tests := []struct {
		filePath string
		want     bool
	}{
		{"config.yaml", true},
		{"config.yml", true},
		{"config.json", true},
		{".env", true},
		{".env.local", true},
		{"docker-compose.yml", true},
		{"Dockerfile", true},
		{"package.json", true},
		{"go.mod", true},
		{"go.sum", true},
		{"Cargo.toml", true},
		{"requirements.txt", true},
		{"main.go", false},
		{"README.md", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			got := extractor.IsConfigFile(tt.filePath)
			if got != tt.want {
				t.Errorf("IsConfigFile(%q) = %v, want %v", tt.filePath, got, tt.want)
			}
		})
	}
}