package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"safeagent/image"
	"safeagent/profile"
	"safeagent/stack"

	"charm.land/huh/v2"
	"github.com/spf13/cobra"
)

// This file holds `profile|stack dockerfile` and `profile|stack rebuild`:
// editing a Dockerfile fragment and rebuilding the existing Images that
// include it.

func newProfileDockerfileCommand(deps Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "dockerfile <name>",
		Short: "Edit a Profile's Dockerfile fragment in $EDITOR",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store := profile.NewStore(deps.ConfigRoot)
			p, err := store.FindByName(args[0])
			if err != nil {
				return err
			}
			return editFragment(cmd, store.DockerfilePath(p.ID), "profile", p.Name, func() error {
				return rebuildProfile(cmd, deps, p)
			})
		},
	}
}

func newStackDockerfileCommand(deps Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "dockerfile <name>",
		Short: "Edit a Stack's Dockerfile fragment in $EDITOR",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store := stack.NewStore(deps.ConfigRoot)
			st, err := store.FindByName(args[0])
			if err != nil {
				return err
			}
			return editFragment(cmd, store.DockerfilePath(st.ID), "stack", st.Name, func() error {
				return rebuildStack(cmd, deps, st)
			})
		},
	}
}

func newProfileRebuildCommand(deps Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "rebuild <name>",
		Short: "Rebuild the existing Images that use a Profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := profile.NewStore(deps.ConfigRoot).FindByName(args[0])
			if err != nil {
				return err
			}
			return rebuildProfile(cmd, deps, p)
		},
	}
}

func newStackRebuildCommand(deps Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "rebuild <name>",
		Short: "Rebuild the existing Images that use a Stack",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			st, err := stack.NewStore(deps.ConfigRoot).FindByName(args[0])
			if err != nil {
				return err
			}
			return rebuildStack(cmd, deps, st)
		},
	}
}

// editFragment opens path in $EDITOR (falling back to vim). Afterwards it
// offers to rebuild, but only when stdin is a terminal. Otherwise (scripts,
// tests) it prints the rebuild command instead of prompting.
func editFragment(cmd *cobra.Command, path, kind, name string, rebuild func() error) error {
	if err := runEditor(cmd, path); err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	hint := fmt.Sprintf("Run `safeagent %s rebuild %s` to apply the change.\n", kind, name)
	if !isTerminal(cmd.InOrStdin()) {
		fmt.Fprintf(out, "Saved %s.\n%s", path, hint)
		return nil
	}
	yes := true
	err := huh.NewForm(huh.NewGroup(
		huh.NewConfirm().Title(fmt.Sprintf("Rebuild the existing images that use %s %s?", kind, name)).Value(&yes),
	)).WithInput(cmd.InOrStdin()).WithOutput(cmd.ErrOrStderr()).Run()
	if err != nil {
		return err
	}
	if !yes {
		fmt.Fprint(out, hint)
		return nil
	}
	return rebuild()
}

// runEditor runs $EDITOR (which may carry arguments, e.g. "code -w") on path,
// attached to the command's streams. An unset $EDITOR means vim.
func runEditor(cmd *cobra.Command, path string) error {
	editor := strings.Fields(os.Getenv("EDITOR"))
	if len(editor) == 0 {
		editor = []string{"vim"}
	}
	c := exec.Command(editor[0], append(editor[1:], path)...)
	c.Stdin = cmd.InOrStdin()
	c.Stdout = cmd.OutOrStdout()
	c.Stderr = cmd.ErrOrStderr()
	if err := c.Run(); err != nil {
		return fmt.Errorf("editor %q: %w", strings.Join(editor, " "), err)
	}
	return nil
}

func isTerminal(r io.Reader) bool {
	f, ok := r.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

// rebuildProfile rebuilds every existing Image containing p. It never builds a
// pairing that doesn't already exist.
func rebuildProfile(cmd *cobra.Command, deps Deps, p profile.Profile) error {
	pairs, err := image.ExistingWithProfile(deps.Docker, p.ID)
	if err != nil {
		return err
	}
	stacks := stack.NewStore(deps.ConfigRoot)
	return rebuildPairs(cmd, deps, pairs, "profile "+p.Name, func(pair image.Pair) (stack.Stack, profile.Profile, error) {
		st, err := stacks.FindByID(pair.StackID)
		return st, p, err
	})
}

// rebuildStack rebuilds every existing Image containing st. It never builds a
// pairing that doesn't already exist.
func rebuildStack(cmd *cobra.Command, deps Deps, st stack.Stack) error {
	pairs, err := image.ExistingWithStack(deps.Docker, st.ID)
	if err != nil {
		return err
	}
	profiles := profile.NewStore(deps.ConfigRoot)
	return rebuildPairs(cmd, deps, pairs, "stack "+st.Name, func(pair image.Pair) (stack.Stack, profile.Profile, error) {
		p, err := profiles.FindByID(pair.ProfileID)
		return st, p, err
	})
}

// rebuildPairs builds each pair with the same build as session start. A pair
// whose other half no longer has a record can't be built, so it is skipped
// with a note.
func rebuildPairs(cmd *cobra.Command, deps Deps, pairs []image.Pair, what string, resolve func(image.Pair) (stack.Stack, profile.Profile, error)) error {
	out := cmd.OutOrStdout()
	if len(pairs) == 0 {
		fmt.Fprintf(out, "No existing images use %s; nothing to rebuild.\n", what)
		return nil
	}
	for _, pair := range pairs {
		st, p, err := resolve(pair)
		if err != nil {
			fmt.Fprintf(out, "Skipped %s: %v.\n", image.Ref(pair.StackID, pair.ProfileID), err)
			continue
		}
		name := image.Name(st.Name, p.Name)
		fmt.Fprintf(out, "Rebuilding %s...\n", name)
		if err := image.Build(deps.Docker, deps.ConfigRoot, st, p); err != nil {
			return fmt.Errorf("rebuilding %s: %w", name, err)
		}
		fmt.Fprintf(out, "Rebuilt %s.\n", name)
	}
	return nil
}
