package cmd

import (
	"safeagent/docker"
	"safeagent/table"

	"github.com/spf13/cobra"
)

func newSessionCommand(deps Deps) *cobra.Command {
	c := &cobra.Command{
		Use:   "session",
		Short: "Inspect running Sessions",
	}
	c.AddCommand(newSessionLsCommand(deps))
	return c
}

func newSessionLsCommand(deps Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "List running Sessions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			containers, err := deps.Docker.ListContainers()
			if err != nil {
				return err
			}
			n, err := loadNames(deps)
			if err != nil {
				return err
			}
			var rows [][]string
			for _, c := range containers {
				projectPath, ok := c.Labels[docker.LabelProject]
				if !ok {
					continue
				}
				rows = append(rows, []string{
					c.ID,
					projectPath,
					n.profile(c.Labels[docker.LabelProfile]),
					n.stack(c.Labels[docker.LabelStack]),
					c.Workdir,
					c.RunningFor,
				})
			}
			return table.Print(cmd.OutOrStdout(), "No running sessions.",
				[]string{"CONTAINER", "PROJECT", "PROFILE", "STACK", "DIRECTORY", "RUNNING FOR"}, rows)
		},
	}
}
