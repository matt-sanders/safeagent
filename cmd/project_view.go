package cmd

import (
	"fmt"
	"text/tabwriter"

	"safeagent/docker"
	"safeagent/image"
	"safeagent/profile"
	"safeagent/project"
	"safeagent/stack"
	"safeagent/table"

	"github.com/spf13/cobra"
)

// names maps Profile and Stack IDs to their Names. An ID with no record (e.g.
// a deleted Profile) maps to itself, so output shows the raw ID rather than
// hiding it.
type names struct {
	profiles map[string]string
	stacks   map[string]string
}

func loadNames(deps Deps) (names, error) {
	n := names{profiles: map[string]string{}, stacks: map[string]string{}}
	profiles, err := profile.NewStore(deps.ConfigRoot).List()
	if err != nil {
		return n, err
	}
	for _, p := range profiles {
		n.profiles[p.ID] = p.Name
	}
	stacks, err := stack.NewStore(deps.ConfigRoot).List()
	if err != nil {
		return n, err
	}
	for _, st := range stacks {
		n.stacks[st.ID] = st.Name
	}
	return n, nil
}

func (n names) profile(id string) string {
	if name, ok := n.profiles[id]; ok {
		return name
	}
	return id
}

func (n names) stack(id string) string {
	if name, ok := n.stacks[id]; ok {
		return name
	}
	return id
}

// imageState returns "built" or "not built" for the Project's Image.
func imageState(deps Deps, p project.Project) (string, error) {
	exists, err := deps.Docker.ImageExists(image.Ref(p.StackID, p.ProfileID))
	if err != nil {
		return "", err
	}
	if exists {
		return "built", nil
	}
	return "not built", nil
}

// sessionMounts returns the mounts a Session in p gets: the Project at its own
// path and the Profile's Claude config folder at /claude.
func sessionMounts(deps Deps, p project.Project) []docker.Mount {
	return []docker.Mount{
		{Source: p.Path, Target: p.Path},
		{Source: profile.NewStore(deps.ConfigRoot).ClaudeDir(p.ProfileID), Target: "/claude"},
	}
}

func newProjectLsCommand(deps Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "List Projects",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			projects, err := project.NewStore(deps.ConfigRoot).List()
			if err != nil {
				return err
			}
			n, err := loadNames(deps)
			if err != nil {
				return err
			}
			rows := make([][]string, len(projects))
			for i, p := range projects {
				state, err := imageState(deps, p)
				if err != nil {
					return err
				}
				rows[i] = []string{p.Path, n.profile(p.ProfileID), n.stack(p.StackID), state}
			}
			return table.Print(cmd.OutOrStdout(), "No projects.", []string{"PATH", "PROFILE", "STACK", "IMAGE"}, rows)
		},
	}
}

func newProjectInspectCommand(deps Deps) *cobra.Command {
	var path string
	c := &cobra.Command{
		Use:   "inspect",
		Short: "Show a Project and the session setup that will apply",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := resolveProject(deps, path)
			if err != nil {
				return err
			}
			n, err := loadNames(deps)
			if err != nil {
				return err
			}
			state, err := imageState(deps, p)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
			fmt.Fprintf(tw, "Path:\t%s\n", p.Path)
			fmt.Fprintf(tw, "Profile:\t%s\n", n.profile(p.ProfileID))
			fmt.Fprintf(tw, "Stack:\t%s\n", n.stack(p.StackID))
			fmt.Fprintf(tw, "Image:\t%s (%s)\n", image.Name(n.stack(p.StackID), n.profile(p.ProfileID)), state)
			if err := tw.Flush(); err != nil {
				return err
			}
			fmt.Fprintln(out, "\nMounts:")
			var rows [][]string
			for _, m := range sessionMounts(deps, p) {
				rows = append(rows, []string{m.Source, m.Target})
			}
			if err := table.Print(out, "No mounts.", []string{"HOST", "CONTAINER"}, rows); err != nil {
				return err
			}

			fmt.Fprintln(out, "\nExclusions:")
			return printExclusions(cmd, p.Exclusions)
		},
	}
	addProjectFlag(c, &path)
	return c
}
