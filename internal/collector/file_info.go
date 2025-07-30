package collector

import (
	"path/filepath"
	"strings"
)

// FileInfoExtractor はファイルパスから情報を抽出します
type FileInfoExtractor struct {
	languageMap map[string]string
	testPatterns []string
	configPatterns []string
}

// NewFileInfoExtractor は新しいFileInfoExtractorを作成します
func NewFileInfoExtractor() *FileInfoExtractor {
	return &FileInfoExtractor{
		languageMap: map[string]string{
			".go":   "go",
			".js":   "javascript",
			".ts":   "typescript",
			".tsx":  "typescript",
			".jsx":  "javascript",
			".py":   "python",
			".java": "java",
			".c":    "c",
			".h":    "c",
			".cpp":  "cpp",
			".hpp":  "cpp",
			".cc":   "cpp",
			".cxx":  "cpp",
			".cs":   "csharp",
			".php":  "php",
			".rb":   "ruby",
			".rs":   "rust",
			".vue":  "vue",
			".css":  "css",
			".scss": "scss",
			".sass": "sass",
			".less": "less",
			".html": "html",
			".xml":  "xml",
			".json": "json",
			".yaml": "yaml",
			".yml":  "yaml",
			".toml": "toml",
			".md":   "markdown",
			".sh":   "shell",
			".sql":  "sql",
		},
		testPatterns: []string{
			"_test.",
			".test.",
			".spec.",
			"test_",
			"/test/",
			"/tests/",
			"/__tests__/",
			"/spec/",
			"spec/", // spec/ で始まるパスも追加
		},
		configPatterns: []string{
			"config.",
			".env",
			"docker-compose.",
			"Dockerfile",
			"package.json",
			"go.mod",
			"go.sum",
			"Cargo.toml",
			"requirements.txt",
			"Makefile",
			".gitignore",
			".dockerignore",
		},
	}
}

// ExtractLanguage はファイルパスから言語を推定します
func (e *FileInfoExtractor) ExtractLanguage(filePath string) string {
	if filePath == "" {
		return "unknown"
	}

	// 特殊ケース
	fileName := filepath.Base(filePath)
	switch fileName {
	case "Dockerfile":
		return "dockerfile"
	case "Makefile":
		return "makefile"
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	if language, ok := e.languageMap[ext]; ok {
		return language
	}

	return "unknown"
}

// ExtractDirectory はファイルパスからディレクトリを抽出します
func (e *FileInfoExtractor) ExtractDirectory(filePath string) string {
	if filePath == "" {
		return "."
	}

	dir := filepath.Dir(filePath)
	
	// 現在ディレクトリの場合は "." を返す
	if dir == "." || dir == "/" {
		return dir
	}

	// 相対パスの場合は "./" を除去
	if strings.HasPrefix(dir, "./") {
		dir = strings.TrimPrefix(dir, "./")
	}

	return dir
}

// IsTestFile はテストファイルかどうかを判定します
func (e *FileInfoExtractor) IsTestFile(filePath string) bool {
	if filePath == "" {
		return false
	}

	lowerPath := strings.ToLower(filePath)
	
	for _, pattern := range e.testPatterns {
		if strings.Contains(lowerPath, pattern) {
			return true
		}
	}

	return false
}

// IsConfigFile は設定ファイルかどうかを判定します
func (e *FileInfoExtractor) IsConfigFile(filePath string) bool {
	if filePath == "" {
		return false
	}

	fileName := strings.ToLower(filepath.Base(filePath))

	for _, pattern := range e.configPatterns {
		if strings.Contains(fileName, pattern) || fileName == strings.ToLower(pattern) {
			return true
		}
	}

	return false
}