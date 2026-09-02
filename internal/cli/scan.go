package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"sarner/envbox/internal/config"
	"sarner/envbox/internal/core"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan for hardcoded secrets in project files",
	RunE:  runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)
}

type secretMatch struct {
	filePath string
	lineNum  int
	line     string
	pattern  string
	match    string
}

var telegramBotTokenRe = regexp.MustCompile(`[0-9]{9,10}:[a-zA-Z0-9_-]{35}`)
var awsAccessKeyRe = regexp.MustCompile(`AKIA[0-9A-Z]{16}`)
var privateKeyRe = regexp.MustCompile(`-----BEGIN PRIVATE KEY-----`)
var jwtRe = regexp.MustCompile(`ey[A-Za-z0-9-_]+\.ey`)

var secretPatterns = []struct {
	name string
	re   *regexp.Regexp
}{
	{"Telegram Bot Token", telegramBotTokenRe},
	{"AWS Access Key", awsAccessKeyRe},
	{"Private Key", privateKeyRe},
	{"JWT Token", jwtRe},
}

var ignoredDirs = map[string]bool{
	".git":         true,
	"vendor":       true,
	"node_modules": true,
	"dist":         true,
	".svn":         true,
	"__pycache__":  true,
	".idea":        true,
	".vscode":      true,
	"internal":     true,
}

func runScan(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load("envbox.toml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	for _, path := range cfg.Scan.IgnorePaths {
		ignoredDirs[path] = true
	}

	if err := loadGitignore(); err == nil {
		for _, pattern := range gitignorePatterns {
			ignoredDirs[pattern] = true
		}
	}

	envPath := cfg.Env.EnvFile
	envFile, err := core.ParseFile(envPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file %s does not exist", envPath)
		}
		return fmt.Errorf("failed to parse %s: %w", envPath, err)
	}

	envSecrets := extractEnvSecrets(envFile)

	matches, err := scanDirectory(".", envSecrets)
	if err != nil {
		return err
	}

	if len(matches) == 0 {
		fmt.Println("No secrets found")
		return nil
	}

	for _, m := range matches {
		highlighted := highlightMatch(m.line, m.match)
		fmt.Printf("%s:%d: %s\n  %s\n  Pattern: %s\n\n", m.filePath, m.lineNum, m.match, highlighted, m.pattern)
	}

	return fmt.Errorf("found %d secret(s)", len(matches))
}

func extractEnvSecrets(f *core.EnvFile) map[string]string {
	secrets := make(map[string]string)
	for _, line := range f.Lines {
		if line.Type == core.LineEnvVar && len(line.Value) > 6 {
			if !isSensitiveKey(line.Key) {
				continue
			}
			secrets[line.Value] = line.Key
		}
	}
	return secrets
}

func isSensitiveKey(key string) bool {
	upper := strings.ToUpper(key)
	sensitive := []string{"KEY", "SECRET", "TOKEN", "PASSWORD", "PASS", "AUTH", "PRIVATE", "CREDENTIAL"}
	for _, s := range sensitive {
		if strings.Contains(upper, s) {
			return true
		}
	}
	return false
}

func loadGitignore() error {
	data, err := os.ReadFile(".gitignore")
	if err != nil {
		return err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		gitignorePatterns = append(gitignorePatterns, line)
	}
	return nil
}

var gitignorePatterns []string

func scanDirectory(root string, envSecrets map[string]string) ([]secretMatch, error) {
	var matches []secretMatch

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			dirName := info.Name()
			if ignoredDirs[dirName] {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(path)
		if ext == ".env" || ext == ".env.example" {
			return nil
		}

		if !isScannableFile(path) {
			return nil
		}

		fileMatches, err := scanFile(path, envSecrets)
		if err != nil {
			return nil
		}
		matches = append(matches, fileMatches...)

		return nil
	})

	return matches, err
}

func isScannableFile(path string) bool {
	ext := filepath.Ext(path)
	scannableExts := map[string]bool{
		".go": true, ".js": true, ".ts": true, ".jsx": true, ".tsx": true,
		".json": true, ".yml": true, ".yaml": true, ".py": true,
		".rb": true, ".java": true, ".cs": true, ".cpp": true, ".c": true,
		".php": true, ".sh": true, ".bash": true, ".zsh": true,
		".tf": true, ".xml": true, ".toml": true, ".ini": true, ".conf": true,
	}
	if scannableExts[ext] {
		return true
	}
	name := filepath.Base(path)
	if name == "Dockerfile" || name == "Makefile" || name == "docker-compose.yml" {
		return true
	}
	return false
}

func scanFile(path string, envSecrets map[string]string) ([]secretMatch, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil
	}
	content := string(data)
	lines := strings.Split(content, "\n")

	var matches []secretMatch

	for i, line := range lines {
		for _, pattern := range secretPatterns {
			found := pattern.re.FindAllString(line, -1)
			for _, match := range found {
				matches = append(matches, secretMatch{
					filePath: path,
					lineNum:  i + 1,
					line:     line,
					pattern:  pattern.name,
					match:    match,
				})
			}
		}

		for value := range envSecrets {
			if strings.Contains(line, value) {
				matches = append(matches, secretMatch{
					filePath: path,
					lineNum:  i + 1,
					line:     line,
					pattern:  "Env value match",
					match:    value,
				})
			}
		}
	}

	return matches, nil
}

func highlightMatch(line, match string) string {
	idx := strings.Index(line, match)
	if idx < 0 {
		return line
	}
	start := idx - 20
	if start < 0 {
		start = 0
	}
	end := idx + len(match) + 20
	if end > len(line) {
		end = len(line)
	}

	prefix := ""
	suffix := ""
	if start > 0 {
		prefix = "..."
	}
	if end < len(line) {
		suffix = "..."
	}

	return prefix + line[start:end] + suffix
}
