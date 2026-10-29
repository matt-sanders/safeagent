// Package cmd builds safeagent's cobra command tree.
package cmd

import (
	"safeagent/docker"

	"github.com/spf13/cobra"
)

// Deps are the external dependencies the command tree is built from.
type Deps struct {
	// ConfigRoot is the v2 data folder (normally ~/.safeagent/v2).
	ConfigRoot string
	// Docker is the Docker port (docker.CLI in the binary, a fake in tests).
	Docker docker.Docker
	// Cwd is the directory relative paths and Project lookup start from.
	// Empty means the process's working directory.
	Cwd string
}

// NewRootCommand builds a fresh command tree. It holds no global state, so
// tests can build one per run. Commands must write through cmd.OutOrStdout()
// so output can be captured.
func NewRootCommand(deps Deps) *cobra.Command {
	var opts sessionOptions
	root := &cobra.Command{
		Use:           "safeagent",
		Short:         "Run agents safely inside Docker",
		Long:          "Run agents safely inside Docker.\n\nWith no subcommand, starts a Session in the current directory's Project,\ncreating the Project first if there isn't one.",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return startSession(cmd, deps, opts)
		},
	}
	addSessionFlags(root, &opts)
	root.AddCommand(newProfileCommand(deps), newProjectCommand(deps), newStackCommand(deps))
	return root
}
