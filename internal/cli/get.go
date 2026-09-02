package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"sarner/envbox/internal/config"
	"sarner/envbox/internal/core"
)

var getCmd = &cobra.Command{
	Use:   "get <KEY>",
	Short: "Print value of an environment variable",
	Args:  cobra.ExactArgs(1),
	RunE:  runGet,
}

func init() {
	rootCmd.AddCommand(getCmd)
}

func runGet(cmd *cobra.Command, args []string) error {
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

	value, ok := envFile.Get(key)
	if !ok {
		return fmt.Errorf("key %q not found", key)
	}

	fmt.Print(value)
	return nil
}
