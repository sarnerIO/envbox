package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"sarner/envbox/internal/config"
	"sarner/envbox/internal/core"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Interactive wizard to create .env from .env.example",
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func readLine(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func runInit(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load("envbox.toml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	examplePath := cfg.Env.ExampleFile
	envPath := cfg.Env.EnvFile

	if _, err := os.Stat(envPath); err == nil {
		fmt.Printf(".env already exists at %s\n", envPath)
		response := readLine("Overwrite? [y/N]: ")
		if strings.ToLower(response) != "y" {
			fmt.Println("Aborted")
			return nil
		}
	}

	exampleFile, err := core.ParseFile(examplePath)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", examplePath, err)
	}

	envValues := make(map[string]string)

	for _, line := range exampleFile.Lines {
		if line.Type != core.LineEnvVar {
			continue
		}

		key := line.Key
		currentValue := line.Value

		var input string
		if core.IsSensitiveKey(key) {
			input = readLine(fmt.Sprintf("%s [%s]: ", key, maskString(currentValue)))
			if input == "" {
				input = currentValue
			}
		} else {
			input = readLine(fmt.Sprintf("%s [%s]: ", key, currentValue))
			if input == "" {
				input = currentValue
			}
		}

		envValues[key] = input
	}

	file, err := os.Create(envPath)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", envPath, err)
	}
	defer file.Close()

	for i, line := range exampleFile.Lines {
		if i > 0 {
			file.WriteString("\n")
		}
		switch line.Type {
		case core.LineEmpty, core.LineComment:
			file.WriteString(line.Raw)
		case core.LineEnvVar:
			value := envValues[line.Key]
			if line.Quote != 0 {
				file.WriteString(fmt.Sprintf("%s=%s%s%s", line.Key, string(line.Quote), value, string(line.Quote)))
			} else {
				file.WriteString(fmt.Sprintf("%s=%s", line.Key, value))
			}
		}
	}

	if exampleFile.HadTrailingNewline {
		file.WriteString("\n")
	}

	fmt.Printf("Created %s\n", envPath)
	return nil
}

func maskString(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return s[:2] + strings.Repeat("*", len(s)-4)
}
