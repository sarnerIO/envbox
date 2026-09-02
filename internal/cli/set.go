package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"sarner/envbox/internal/config"
	"sarner/envbox/internal/core"
)

var setCmd = &cobra.Command{
	Use:   "set <KEY> <VALUE>",
	Short: "Set or update an environment variable",
	Args:  cobra.ExactArgs(2),
	RunE:  runSet,
}

func init() {
	rootCmd.AddCommand(setCmd)
}

func runSet(cmd *cobra.Command, args []string) error {
	key, value := args[0], args[1]

	cfg, err := config.Load("envbox.toml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	envPath := cfg.Env.EnvFile

	envFile, err := core.ParseFile(envPath)
	if err != nil {
		if os.IsNotExist(err) {
			envFile = &core.EnvFile{
				Path:               envPath,
				Lines:              []core.EnvLine{},
				HadTrailingNewline: false,
			}
		} else {
			return fmt.Errorf("failed to parse %s: %w", envPath, err)
		}
	}

	envFile.Set(key, value)

	content := envFile.Render()
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", envPath, err)
	}

	return nil
}
