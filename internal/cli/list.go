package cli

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"sarner/envbox/internal/config"
	"sarner/envbox/internal/core"
)

var reveal bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all environment variables",
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVarP(&reveal, "reveal", "r", false, "Show sensitive values unmasked")
}

var sensitivePatterns = []string{
	"KEY",
	"SECRET",
	"TOKEN",
	"PASSWORD",
	"PASS",
	"AUTH",
	"PRIVATE",
}

func isSensitive(key string) bool {
	upper := strings.ToUpper(key)
	for _, pattern := range sensitivePatterns {
		if strings.Contains(upper, pattern) {
			return true
		}
	}
	return false
}

func mask(value string) string {
	return "********"
}

func runList(cmd *cobra.Command, args []string) error {
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

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "KEY\tVALUE")

	for _, line := range envFile.Lines {
		if line.Type != core.LineEnvVar {
			continue
		}
		value := line.Value
		if !reveal && isSensitive(line.Key) {
			value = mask(value)
		}
		fmt.Fprintf(tw, "%s\t%s\n", line.Key, value)
	}

	return tw.Flush()
}
