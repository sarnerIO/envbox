package cli

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"sarner/envbox/internal/config"
	"sarner/envbox/internal/core"
)

var dryRun bool

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync .env.example with new keys from .env",
	RunE:  runSync,
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be added without modifying files")
}

func runSync(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load("envbox.toml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	envPath := cfg.Env.EnvFile
	examplePath := cfg.Env.ExampleFile

	envFile, err := core.ParseFile(envPath)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", envPath, err)
	}

	exampleFile, err := core.ParseFile(examplePath)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", examplePath, err)
	}

	envKeys := keysFromEnvFile(envFile)
	exampleKeys := keysFromEnvFile(exampleFile)

	var missing []string
	for key := range envKeys {
		if _, ok := exampleKeys[key]; !ok {
			missing = append(missing, key)
		}
	}

	if len(missing) == 0 {
		fmt.Println("No new keys to sync")
		return nil
	}

	sort.Strings(missing)

	if dryRun {
		fmt.Printf("Would add %d key(s) to %s:\n", len(missing), examplePath)
		for _, key := range missing {
			fmt.Printf("  + %s=\n", key)
		}
		return nil
	}

	f, err := os.OpenFile(examplePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", examplePath, err)
	}
	defer f.Close()

	for _, key := range missing {
		line := fmt.Sprintf("%s=\n", key)
		if _, err := f.WriteString(line); err != nil {
			return fmt.Errorf("failed to write: %w", err)
		}
	}

	fmt.Printf("Added %d key(s) to %s:\n", len(missing), examplePath)
	for _, key := range missing {
		fmt.Printf("  + %s=\n", key)
	}

	return nil
}

func keysFromEnvFile(f *core.EnvFile) map[string]bool {
	m := make(map[string]bool)
	for _, line := range f.Lines {
		if line.Type == core.LineEnvVar {
			m[line.Key] = true
		}
	}
	return m
}
