package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"safeagent/project"

	"github.com/spf13/cobra"
)

// Project resolution shared by every command that acts on a Project.
//
// Register the flag with addProjectFlag, then call resolveProject (an
// existing Project: exact match for -p, closest ancestor of the cwd without
// it) or projectPath (just the target path, e.g. for project create).

// addProjectFlag registers -p/--project on c, storing the raw value in target.
func addProjectFlag(c *cobra.Command, target *string) {
	c.Flags().StringVarP(target, "project", "p", "", "Project path (default: the Project for the current directory)")
}

// cwd returns the directory commands resolve relative paths against.
func (d Deps) cwd() (string, error) {
	if d.Cwd != "" {
		return filepath.Clean(d.Cwd), nil
	}
	return os.Getwd()
}

// projectPath returns the absolute, cleaned path that flag names, resolved
// relative to the current directory. An empty flag means the current
// directory itself.
func projectPath(deps Deps, flag string) (string, error) {
	cwd, err := deps.cwd()
	if err != nil {
		return "", err
	}
	if flag == "" {
		return cwd, nil
	}
	if filepath.IsAbs(flag) {
		return filepath.Clean(flag), nil
	}
	return filepath.Join(cwd, flag), nil
}

// resolveProject finds the Project a command acts on. With -p it is the
// Project at exactly that path; without it, the Project containing the
// current directory, innermost first. Errors explain how to fix things.
func resolveProject(deps Deps, flag string) (project.Project, error) {
	path, err := projectPath(deps, flag)
	if err != nil {
		return project.Project{}, err
	}
	store := project.NewStore(deps.ConfigRoot)
	var p project.Project
	if flag != "" {
		p, err = store.Get(path)
	} else {
		p, err = store.FindForDir(path)
	}
	if errors.Is(err, project.ErrNotFound) {
		return project.Project{}, fmt.Errorf("%w (create one with `safeagent project create`)", err)
	}
	return p, err
}
