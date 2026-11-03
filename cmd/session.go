package cmd

import (
	"errors"
	"fmt"
	"path/filepath"

	"safeagent/docker"
	"safeagent/image"
	"safeagent/profile"
	"safeagent/project"
	"safeagent/stack"

	"github.com/spf13/cobra"
)

// claudeCommand is what a Session runs unless --shell is given.
var claudeCommand = []string{"claude", "--dangerously-skip-permissions"}

// sessionOptions are the root command's flags.
type sessionOptions struct {
	shell       bool
	profileName string
	stackName   string
}

func addSessionFlags(c *cobra.Command, opts *sessionOptions) {
	c.Flags().BoolVar(&opts.shell, "shell", false, "Open a shell in the container instead of Claude")
	c.Flags().StringVar(&opts.profileName, "profile", "", "Profile for a new Project (prompted if omitted)")
	c.Flags().StringVar(&opts.stackName, "stack", "", "Stack for a new Project (prompted if omitted)")
}

// startSession is `safeagent` itself: find or create the Project for the
// current directory, make sure its Image exists, then run the container.
func startSession(cmd *cobra.Command, deps Deps, opts sessionOptions) error {
	cwd, err := deps.cwd()
	if err != nil {
		return err
	}
	p, err := findOrCreateProject(cmd, deps, cwd, opts)
	if err != nil {
		return err
	}
	prof, st, err := projectProfileAndStack(deps, p)
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	if exists, err := deps.Docker.ImageExists(image.Ref(st.ID, prof.ID)); err != nil {
		return err
	} else if !exists {
		fmt.Fprintf(out, "Building image %s (node %s)...\n", image.Name(st.Name, prof.Name), st.Node)
		if err := image.Build(deps.Docker, deps.ConfigRoot, st, prof); err != nil {
			return fmt.Errorf("building image %s: %w", image.Name(st.Name, prof.Name), err)
		}
	}

	req, err := sessionRunRequest(deps, p, cwd, opts.shell)
	if err != nil {
		return err
	}
	what := "Claude"
	if opts.shell {
		what = "a shell"
	}
	fmt.Fprintf(out, "Starting %s in %s (profile %s, stack %s)...\n", what, cwd, prof.Name, st.Name)
	return deps.Docker.Run(req)
}

// findOrCreateProject returns the Project containing cwd (innermost wins).
// With none, it creates one at cwd using --profile / --stack, or the pickers
// when those are missing.
func findOrCreateProject(cmd *cobra.Command, deps Deps, cwd string, opts sessionOptions) (project.Project, error) {
	store := project.NewStore(deps.ConfigRoot)
	p, err := store.FindForDir(cwd)
	if err == nil {
		if opts.profileName != "" || opts.stackName != "" {
			return project.Project{}, fmt.Errorf("project %s already exists; --profile and --stack only apply when creating one (use `safeagent project use-profile` / `use-stack` to switch)", p.Path)
		}
		return p, nil
	}
	if !errors.Is(err, project.ErrNotFound) {
		return project.Project{}, err
	}
	prof, err := pickProfile(cmd, deps, opts.profileName)
	if err != nil {
		return project.Project{}, err
	}
	st, err := pickStack(cmd, deps, opts.stackName)
	if err != nil {
		return project.Project{}, err
	}
	p = project.Project{Path: cwd, ProfileID: prof.ID, StackID: st.ID}
	if err := store.Create(p); err != nil {
		return project.Project{}, err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Created project %s (profile %s, stack %s).\n", cwd, prof.Name, st.Name)
	return p, nil
}

// projectProfileAndStack loads the Profile and Stack a Project references,
// explaining how to fix things when either has been removed.
func projectProfileAndStack(deps Deps, p project.Project) (profile.Profile, stack.Stack, error) {
	prof, err := profile.NewStore(deps.ConfigRoot).FindByID(p.ProfileID)
	if err != nil {
		return profile.Profile{}, stack.Stack{}, fmt.Errorf("project %s uses a profile that no longer exists (ID %s); choose another with `safeagent project use-profile <name>`", p.Path, p.ProfileID)
	}
	st, err := stack.NewStore(deps.ConfigRoot).FindByID(p.StackID)
	if err != nil {
		return profile.Profile{}, stack.Stack{}, fmt.Errorf("project %s uses a stack that no longer exists (ID %s); choose another with `safeagent project use-stack <name>`", p.Path, p.StackID)
	}
	return prof, st, nil
}

// sessionLabels are the labels on every Session container: the Project path
// and the Profile and Stack IDs. `session ls` reads them back.
func sessionLabels(p project.Project) map[string]string {
	return map[string]string{
		docker.LabelProject: p.Path,
		docker.LabelProfile: p.ProfileID,
		docker.LabelStack:   p.StackID,
	}
}

// sessionRunRequest assembles everything Docker needs to run a Session of p
// started from cwd: the Project's Image, cwd as the working directory, the
// session mounts, the labels and the command (Claude, or a shell when shell
// is true). Anything that depends on cwd (e.g. Exclusions) belongs here; mounts
// that depend only on the Project belong in sessionMounts.
func sessionRunRequest(deps Deps, p project.Project, cwd string, shell bool) (docker.RunRequest, error) {
	mounts, err := sessionMounts(deps, p)
	if err != nil {
		return docker.RunRequest{}, err
	}
	if err := checkProfileMountsAvoidProject(p, mounts); err != nil {
		return docker.RunRequest{}, err
	}
	req := docker.RunRequest{
		Image:   image.Ref(p.StackID, p.ProfileID),
		Workdir: cwd,
		Mounts:  mounts,
		Labels:  sessionLabels(p),
	}
	req.Mounts = append(req.Mounts, exclusionMounts(p, cwd)...)
	if !shell {
		req.Command = claudeCommand
	}
	return req, nil
}

// exclusionMounts hides each of p's Exclusions with an anonymous volume at
// that path relative to cwd (not the Project path), so they work per
// worktree or subdirectory.
func exclusionMounts(p project.Project, cwd string) []docker.Mount {
	mounts := make([]docker.Mount, len(p.Exclusions))
	for i, e := range p.Exclusions {
		mounts[i] = docker.Mount{Target: filepath.Join(cwd, e)}
	}
	return mounts
}

// checkProfileMountsAvoidProject fails when a Profile Mount targets the
// Project path, where the Project itself is mounted. mounts[0] is the Project
// mount (see sessionMounts).
func checkProfileMountsAvoidProject(p project.Project, mounts []docker.Mount) error {
	for _, m := range mounts[1:] {
		if m.Target == p.Path {
			return fmt.Errorf("the project's profile has a mount (host %s) at the project path %s; remove it with `safeagent profile mounts rm <profile> %s`", m.Source, p.Path, p.Path)
		}
	}
	return nil
}
