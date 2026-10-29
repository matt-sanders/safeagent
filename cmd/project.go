package cmd

import (
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

// createNew is the picker value of the "Create new…" option.
const createNew = -1

// pickProfile returns the Profile called name, or asks the user to choose one
// when name is empty. The picker lists every Profile plus "Create new…" (which
// runs the `profile create` flow) and never chooses on the user's behalf, even
// when only one option exists.
func pickProfile(cmd *cobra.Command, deps Deps, name string) (profile.Profile, error) {
	store := profile.NewStore(deps.ConfigRoot)
	if name != "" {
		return store.FindByName(name)
	}
	profiles, err := store.List()
	if err != nil {
		return profile.Profile{}, err
	}
	options := make([]huh.Option[int], 0, len(profiles)+1)
	for i, p := range profiles {
		options = append(options, huh.NewOption(p.Name, i))
	}
	options = append(options, huh.NewOption("Create new…", createNew))
	choice, err := selectOne(cmd, "Profile", options)
	if err != nil {
		return profile.Profile{}, err
	}
	if choice == createNew {
		return createProfile(cmd, deps, "", true)
	}
	return profiles[choice], nil
}

// pickStack returns the Stack called name, or asks the user to choose one
// when name is empty. The picker lists every Stack plus "Create new…" (which
// runs the `stack create` flow) and never chooses on the user's behalf, even
// when only one option exists.
func pickStack(cmd *cobra.Command, deps Deps, name string) (stack.Stack, error) {
	store := stack.NewStore(deps.ConfigRoot)
	if name != "" {
		return store.FindByName(name)
	}
	stacks, err := store.List()
	if err != nil {
		return stack.Stack{}, err
	}
	options := make([]huh.Option[int], 0, len(stacks)+1)
	for i, st := range stacks {
		options = append(options, huh.NewOption(st.Name, i))
	}
	options = append(options, huh.NewOption("Create new…", createNew))
	choice, err := selectOne(cmd, "Stack", options)
	if err != nil {
		return stack.Stack{}, err
	}
	if choice == createNew {
		return createStack(cmd, deps, "", stack.DefaultNode, true, true)
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
