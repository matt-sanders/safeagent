package cmd

import (
	"fmt"

	"safeagent/project"
	"safeagent/table"

	"charm.land/huh/v2"
	"github.com/spf13/cobra"
)

// exclusionsNote explains how Exclusion paths apply; shown under listings.
const exclusionsNote = "Paths are relative to the directory the session starts in."

func newProjectExclusionsCommand(deps Deps) *cobra.Command {
	c := &cobra.Command{
		Use:   "exclusions",
		Short: "Manage the paths hidden from Sessions (e.g. node_modules)",
	}
	c.AddCommand(newExclusionsLsCommand(deps), newExclusionsAddCommand(deps), newExclusionsRmCommand(deps))
	return c
}

// printExclusions writes the Exclusions table followed by the note, or an
// empty-state sentence.
func printExclusions(cmd *cobra.Command, exclusions []string) error {
	out := cmd.OutOrStdout()
	rows := make([][]string, len(exclusions))
	for i, e := range exclusions {
		rows[i] = []string{e}
	}
	if err := table.Print(out, "No exclusions.", []string{"PATH"}, rows); err != nil {
		return err
	}
	if len(rows) > 0 {
		fmt.Fprintf(out, "\n%s\n", exclusionsNote)
	}
	return nil
}

func newExclusionsLsCommand(deps Deps) *cobra.Command {
	var path string
	c := &cobra.Command{
		Use:   "ls",
		Short: "List the Project's Exclusions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := resolveProject(deps, path)
			if err != nil {
				return err
			}
			return printExclusions(cmd, p.Exclusions)
		},
	}
	addProjectFlag(c, &path)
	return c
}

func newExclusionsAddCommand(deps Deps) *cobra.Command {
	var path string
	c := &cobra.Command{
		Use:   "add [<path>]",
		Short: "Hide a path (relative to the session's start directory) from Sessions",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := resolveProject(deps, path)
			if err != nil {
				return err
			}
			var exclusion string
			if len(args) == 1 {
				exclusion = args[0]
			} else {
				err := runForm(cmd, huh.NewInput().
					Title("Path to exclude").
					Description(exclusionsNote).
					Value(&exclusion).
					Validate(func(s string) error {
						_, err := project.CleanExclusion(s)
						return err
					}))
				if err != nil {
					return err
				}
			}
			added, err := project.NewStore(deps.ConfigRoot).AddExclusion(p.Path, exclusion)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Excluded %s in project %s.\n", added, p.Path)
			return nil
		},
	}
	addProjectFlag(c, &path)
	return c
}

func newExclusionsRmCommand(deps Deps) *cobra.Command {
	var path string
	c := &cobra.Command{
		Use:   "rm <path>",
		Short: "Stop hiding a path from Sessions",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := resolveProject(deps, path)
			if err != nil {
				return err
			}
			removed, err := project.NewStore(deps.ConfigRoot).RemoveExclusion(p.Path, args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Removed exclusion %s from project %s.\n", removed, p.Path)
			return nil
		},
	}
	addProjectFlag(c, &path)
	return c
}
