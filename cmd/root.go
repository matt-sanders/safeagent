package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"safeagent/auth"
	"safeagent/config"
	"safeagent/docker"
	"safeagent/profile"
	"safeagent/project"

	"github.com/spf13/cobra"
)

var (
	profileService *profile.Service
	projectService *project.Service
	dockerService  *docker.Service
	authService    *auth.Service
	configDir      string
	shellMode      bool

	rootCmd = &cobra.Command{
		Use:   "safeagent",
		Short: "Run agents safely inside Docker",
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
			if err := config.MigrateAuthLayout(configDir); err != nil {
				return err
			}
			authService = auth.NewService(filepath.Join(configDir, "auth"))
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return startSession(shellMode)
		},
	}
)

func init() {
	rootCmd.Flags().BoolVar(&shellMode, "shell", false, "Open a shell instead of starting Claude")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
