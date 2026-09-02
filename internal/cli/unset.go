package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"sarner/envbox/internal/config"
	"sarner/envbox/internal/core"
)

var unsetCmd = &cobra.Command{
	Use:   "unset <KEY>",
	Short: "Remove an environment variable",
	Args:  cobra.ExactArgs(1),
	RunE:  runUnset,
}

func init() {
	rootCmd.AddCommand(unsetCmd)
}

func runUnset(cmd *cobra.Command, args []string) error {
	key := args[0]

	cfg, err := config.Load("envbox.toml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	envPath := cfg.Env.EnvFile

	envFile, err := core.ParseFile(envPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file %s does not exist", envPath)
		}
		return fmt.Errorf("failed to parse %s: %w", envPath, err)
	}

	if !envFile.Unset(key) {
		return fmt.Errorf("key %q not found", key)
	}

	content := envFile.Render()
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", envPath, err)
	}

	return nil
}
