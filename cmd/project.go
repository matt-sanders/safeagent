package cmd

import (
	"errors"
	"fmt"

	"safeagent/profile"
	"safeagent/project"
	"safeagent/stack"

	"charm.land/huh/v2"
	"github.com/spf13/cobra"
)

func newProjectCommand(deps Deps) *cobra.Command {
	c := &cobra.Command{
		Use:   "project",
		Short: "Manage Projects (the directories sessions work in)",
		Args:  cobra.NoArgs,
	}
	c.AddCommand(newProjectCreateCommand(deps), newProjectRmCommand(deps))
	c.AddCommand(newProjectLsCommand(deps), newProjectInspectCommand(deps))
	c.AddCommand(newProjectUseProfileCommand(deps), newProjectUseStackCommand(deps))
	return c
}

func newProjectCreateCommand(deps Deps) *cobra.Command {
	var path, profileName, stackName string
	c := &cobra.Command{
		Use:   "create",
		Short: "Create a Project in the current directory (or at -p)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := projectPath(deps, path)
			if err != nil {
				return err
			}
			p, err := pickProfile(cmd, deps, profileName)
			if err != nil {
				return err
			}
			st, err := pickStack(cmd, deps, stackName)
			if err != nil {
				return err
			}
			err = project.NewStore(deps.ConfigRoot).Create(project.Project{Path: dir, ProfileID: p.ID, StackID: st.ID})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created project %s (profile %s, stack %s).\n", dir, p.Name, st.Name)
			return nil
		},
	}
	addProjectFlag(c, &path)
	c.Flags().StringVar(&profileName, "profile", "", "Profile name (prompted if omitted)")
	c.Flags().StringVar(&stackName, "stack", "", "Stack name (prompted if omitted)")
	return c
}

func newProjectRmCommand(deps Deps) *cobra.Command {
	var path string
	c := &cobra.Command{
		Use:   "rm",
		Short: "Forget a Project (files on disk are untouched)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := resolveProject(deps, path)
			if err != nil {
				return err
			}
			if err := project.NewStore(deps.ConfigRoot).Remove(p.Path); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Removed project %s. Files on disk were not touched.\n", p.Path)
			return nil
		},
	}
	addProjectFlag(c, &path)
	return c
}

// pickProfile returns the Profile called name, or asks the user to choose an
// existing one when name is empty.
func pickProfile(cmd *cobra.Command, deps Deps, name string) (profile.Profile, error) {
	store := profile.NewStore(deps.ConfigRoot)
	if name != "" {
		return store.FindByName(name)
	}
	profiles, err := store.List()
	if err != nil {
		return profile.Profile{}, err
	}
	if len(profiles) == 0 {
		return profile.Profile{}, errors.New("no profiles exist yet; create one with `safeagent profile create`")
	}
	options := make([]huh.Option[int], len(profiles))
	for i, p := range profiles {
		options[i] = huh.NewOption(p.Name, i)
	}
	choice, err := selectOne(cmd, "Profile", options)
	if err != nil {
		return profile.Profile{}, err
	}
	return profiles[choice], nil
}

// pickStack returns the Stack called name, or asks the user to choose an
// existing one when name is empty.
func pickStack(cmd *cobra.Command, deps Deps, name string) (stack.Stack, error) {
	store := stack.NewStore(deps.ConfigRoot)
	if name != "" {
		return store.FindByName(name)
	}
	stacks, err := store.List()
	if err != nil {
		return stack.Stack{}, err
	}
	if len(stacks) == 0 {
		return stack.Stack{}, errors.New("no stacks exist yet; create one with `safeagent stack create`")
	}
	options := make([]huh.Option[int], len(stacks))
	for i, st := range stacks {
		options[i] = huh.NewOption(st.Name, i)
	}
	choice, err := selectOne(cmd, "Stack", options)
	if err != nil {
		return stack.Stack{}, err
	}
	return stacks[choice], nil
}

// selectOne shows a huh select on the command's streams and returns the
// chosen option's value.
func selectOne(cmd *cobra.Command, title string, options []huh.Option[int]) (int, error) {
	var choice int
	err := huh.NewForm(huh.NewGroup(
		huh.NewSelect[int]().Title(title).Options(options...).Value(&choice),
	)).WithInput(cmd.InOrStdin()).WithOutput(cmd.ErrOrStderr()).Run()
	return choice, err
}
