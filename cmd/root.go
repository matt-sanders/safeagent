package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"safe-claude/config"
	"safe-claude/profile"
	"safe-claude/project"

	"github.com/spf13/cobra"
)

var (
	profileService *profile.Service
	projectService *project.Service

	rootCmd = &cobra.Command{
		Use:   "safe-claude",
		Short: "Run Claude Code safely inside Docker",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("could not determine home directory: %w", err)
			}
			configDir, err := config.EnsureConfigDir(homeDir)
			if err != nil {
				return err
			}
			store := profile.NewStore(filepath.Join(configDir, "profiles.json"))
			profileService = profile.NewService(store)
			projectStore := project.NewStore(filepath.Join(configDir, "projects.json"))
			projectService = project.NewService(projectStore)
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return startSession()
		},
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
