package cmd

import (
	"fmt"
	"strconv"

	"safeagent/ident"
	"safeagent/project"
	"safeagent/stack"
	"safeagent/table"

	"charm.land/huh/v2"
	"github.com/spf13/cobra"
)

func newStackCommand(deps Deps) *cobra.Command {
	c := &cobra.Command{
		Use:   "stack",
		Short: "Manage Stacks (the software a session runs with)",
		Args:  cobra.NoArgs,
	}
	c.AddCommand(newStackCreateCommand(deps), newStackLsCommand(deps))
	return c
}

func newStackCreateCommand(deps Deps) *cobra.Command {
	var name, node string
	c := &cobra.Command{
		Use:   "create",
		Short: "Create a Stack",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var fields []huh.Field
			if !cmd.Flags().Changed("name") {
				fields = append(fields, huh.NewInput().Title("Stack name").Value(&name).Validate(ident.ValidateName))
			}
			if !cmd.Flags().Changed("node") {
				fields = append(fields, huh.NewInput().Title("Node version").Value(&node))
			}
			if len(fields) > 0 {
				err := huh.NewForm(huh.NewGroup(fields...)).WithInput(cmd.InOrStdin()).WithOutput(cmd.ErrOrStderr()).Run()
				if err != nil {
					return err
				}
			}
			st, err := stack.NewStore(deps.ConfigRoot).Create(name, node)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created stack %s (%s, node %s)\n", st.Name, st.ID, st.Node)
			return nil
		},
	}
	c.Flags().StringVar(&name, "name", "", "Name of the Stack")
	c.Flags().StringVar(&node, "node", stack.DefaultNode, "Node version (fixed once created)")
	return c
}

func newStackLsCommand(deps Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "List Stacks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			stacks, err := stack.NewStore(deps.ConfigRoot).List()
			if err != nil {
				return err
			}
			counts, err := project.NewStore(deps.ConfigRoot).CountByStack()
			if err != nil {
				return err
			}
			rows := make([][]string, len(stacks))
			for i, st := range stacks {
				rows[i] = []string{st.Name, st.ID, st.Node, strconv.Itoa(counts[st.ID])}
			}
			return table.Print(cmd.OutOrStdout(), "No stacks.", []string{"NAME", "ID", "NODE", "PROJECTS"}, rows)
		},
	}
}
