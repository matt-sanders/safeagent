package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"safe-claude/config"
	"safe-claude/docker"
	"safe-claude/profile"
	"safe-claude/project"

	"github.com/spf13/cobra"
)

var (
	profileService *profile.Service
	projectService *project.Service
	dockerService  *docker.Service
	configDir      string

	rootCmd = &cobra.Command{
		Use:   "safe-claude",
		Short: "Run Claude Code safely inside Docker",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("could not determine home directory: %w", err)
			}
			var cfgErr error
			configDir, cfgErr = config.EnsureConfigDir(homeDir)
			if cfgErr != nil {
				return cfgErr
			}
			store := profile.NewStore(filepath.Join(configDir, "profiles.json"))
			profileService = profile.NewService(store)
			projectStore := project.NewStore(filepath.Join(configDir, "projects.json"))
			projectService = project.NewService(projectStore)
			dockerService = docker.NewService()
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
