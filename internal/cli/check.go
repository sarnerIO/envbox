package cli

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"sarner/envbox/internal/config"
	"sarner/envbox/internal/core"
)

var strict bool
var ci bool

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check .env and .env.example consistency",
	RunE:  runCheck,
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().BoolVar(&strict, "strict", false, "Treat warnings as errors")
	checkCmd.Flags().BoolVar(&ci, "ci", false, "Disable ANSI colors for CI logs")
}

var (
	red    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	green  = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
	yellow = lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	dim    = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

func checkMarker(ok bool, ciMode bool) string {
	if ciMode {
		if ok {
			return "[OK]"
		}
		return "[FAIL]"
	}
	if ok {
		return green.Render("✓")
	}
	return red.Render("✗")
}

func warnMarker(ciMode bool) string {
	if ciMode {
		return "[WARN]"
	}
	return yellow.Render("⚠")
}

func runCheck(cmd *cobra.Command, args []string) error {
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

	envKeys := keysFromFile(envFile)
	exampleKeys := keysFromFile(exampleFile)
	requiredKeys := cfg.Required

	var errors, warnings []string

	for key, field := range requiredKeys {
		val, ok := envKeys[key]
		if !ok {
			errors = append(errors, fmt.Sprintf("%s Required key '%s' is missing in %s", checkMarker(false, ci), key, envPath))
			continue
		}
		result := config.Validate(key, val, field.Type)
		if !result.Valid {
			errors = append(errors, fmt.Sprintf("%s Key '%s' has invalid value for type '%s': %s", checkMarker(false, ci), key, field.Type, result.Reason))
		} else {
			fmt.Printf("%s %s (%s)\n", checkMarker(true, ci), key, field.Type)
		}
	}

	for key := range exampleKeys {
		if _, ok := envKeys[key]; !ok {
			if _, required := requiredKeys[key]; required {
				errors = append(errors, fmt.Sprintf("%s Key '%s' is required but missing in %s", checkMarker(false, ci), key, envPath))
			} else {
				if strict {
					errors = append(errors, fmt.Sprintf("%s Key '%s' is in .env.example but missing in .env", checkMarker(false, ci), key))
				} else {
					warnings = append(warnings, fmt.Sprintf("%s Key '%s' is in .env.example but missing in .env", warnMarker(ci), key))
				}
			}
		}
	}

	for key := range envKeys {
		if _, ok := exampleKeys[key]; !ok {
			warnings = append(warnings, fmt.Sprintf("%s Key '%s' is in %s but not in %s", warnMarker(ci), key, envPath, examplePath))
		}
	}

	for _, w := range warnings {
		fmt.Println(w)
	}
	for _, e := range errors {
		fmt.Println(e)
	}

	if len(errors) > 0 {
		return fmt.Errorf("check failed: %d error(s)", len(errors))
	}

	if len(warnings) > 0 && strict {
		return fmt.Errorf("check failed: %d warning(s) in strict mode", len(warnings))
	}

	return nil
}

func keysFromFile(f *core.EnvFile) map[string]string {
	m := make(map[string]string)
	for _, line := range f.Lines {
		if line.Type == core.LineEnvVar {
			m[line.Key] = line.Value
		}
	}
	return m
}
