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
}

// NewRootCommand builds a fresh command tree. It holds no global state, so
// tests can build one per run. Commands must write through cmd.OutOrStdout()
// so output can be captured.
func NewRootCommand(deps Deps) *cobra.Command {
	root := &cobra.Command{
		Use:           "safeagent",
		Short:         "Run agents safely inside Docker",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		// Placeholder until starting a Session lands; makes the root runnable so
		// NoArgs rejects unknown commands instead of silently printing help.
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	_ = deps
	return root
}
