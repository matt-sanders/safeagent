package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"safeagent/stack"
	"safeagent/table"

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
			in := bufio.NewReader(cmd.InOrStdin())
			out := cmd.OutOrStdout()
			if !cmd.Flags().Changed("name") {
				name = prompt(in, out, "Name", "")
			}
			if !cmd.Flags().Changed("node") {
				node = prompt(in, out, "Node version", stack.DefaultNode)
			}
			st, err := stack.NewStore(deps.ConfigRoot).Create(name, node)
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "Created stack %s (%s, node %s)\n", st.Name, st.ID, st.Node)
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
			rows := make([][]string, len(stacks))
			for i, st := range stacks {
				// PROJECTS stays 0 until Projects exist.
				rows[i] = []string{st.Name, st.ID, st.Node, "0"}
			}
			return table.Print(cmd.OutOrStdout(), "No stacks.", []string{"NAME", "ID", "NODE", "PROJECTS"}, rows)
		},
	}
}

// prompt asks for one line of input, showing def in brackets. An empty
// answer (or end of input) returns def.
func prompt(in *bufio.Reader, out io.Writer, label, def string) string {
	if def == "" {
		fmt.Fprintf(out, "%s: ", label)
	} else {
		fmt.Fprintf(out, "%s [%s]: ", label, def)
	}
	line, _ := in.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return def
	}
	return line
}
