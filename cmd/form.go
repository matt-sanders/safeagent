package cmd

import (
	"charm.land/huh/v2"
	"github.com/spf13/cobra"
)

// runForm shows fields as one huh form group on the command's streams (input
// from stdin, drawn on stderr so stdout stays clean for output).
func runForm(cmd *cobra.Command, fields ...huh.Field) error {
	return huh.NewForm(huh.NewGroup(fields...)).
		WithInput(cmd.InOrStdin()).
		WithOutput(cmd.ErrOrStderr()).
		Run()
}
