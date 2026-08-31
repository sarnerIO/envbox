package cli

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:     "envbox",
	Short:   "Developer-first CLI tool for .env & .env.example management",
	Long:    "envbox helps teams keep .env and .env.example in sync, prevents secret leaks, and automates developer onboarding.",
	Version: "0.1.0",
}

func Execute() error {
	return rootCmd.Execute()
}
