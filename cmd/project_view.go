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
	for _, prof := range profiles {
		n.profiles[prof.ID] = prof.Name
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
func imageState(deps Deps, proj project.Project) (string, error) {
	exists, err := deps.Docker.ImageExists(image.Ref(proj.StackID, proj.ProfileID))
	if err != nil {
		return "", err
	}
	if exists {
		return "built", nil
	}
	return "not built", nil
}

// sessionMounts returns the mounts a Session in proj gets: the Project at its
// own path, the Profile's Claude config folder at /claude, then the Profile's
// Mounts. It errors if the Profile no longer exists.
func sessionMounts(deps Deps, proj project.Project) ([]docker.Mount, error) {
	store := profile.NewStore(deps.ConfigRoot)
	prof, err := store.FindByID(proj.ProfileID)
	if err != nil {
		return nil, err
	}
	mounts := []docker.Mount{
		{Source: proj.Path, Target: proj.Path},
		{Source: store.ClaudeDir(prof.ID), Target: profile.ClaudeMountPath},
	}
	for _, m := range prof.Mounts {
		mounts = append(mounts, docker.Mount{Source: m.HostPath, Target: m.ContainerPath})
	}
	return mounts, nil
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
			for i, proj := range projects {
				state, err := imageState(deps, proj)
				if err != nil {
					return err
				}
				rows[i] = []string{proj.Path, n.profile(proj.ProfileID), n.stack(proj.StackID), state}
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
			proj, err := resolveProject(deps, path)
			if err != nil {
				return err
			}
			prof, st, err := projectProfileAndStack(deps, proj)
			if err != nil {
				return err
			}
			state, err := imageState(deps, proj)
			if err != nil {
				return err
			}
			mounts, err := sessionMounts(deps, proj)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
			fmt.Fprintf(tw, "Path:\t%s\n", proj.Path)
			fmt.Fprintf(tw, "Profile:\t%s\n", prof.Name)
			fmt.Fprintf(tw, "Stack:\t%s\n", st.Name)
			fmt.Fprintf(tw, "Image:\t%s (%s)\n", image.Name(st.Name, prof.Name), state)
			if err := tw.Flush(); err != nil {
				return err
			}
			fmt.Fprintln(out, "\nMounts:")
			rows := make([][]string, len(mounts))
			for i, m := range mounts {
				rows[i] = []string{m.Source, m.Target}
			}
			if err := printMounts(out, rows); err != nil {
				return err
			}

			fmt.Fprintln(out, "\nExclusions:")
			return printExclusions(cmd, proj.Exclusions)
		},
	}
	addProjectFlag(c, &path)
	return c
}
