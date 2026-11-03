package main

import (
	"fmt"
	"os"
	"path/filepath"

	"safeagent/cmd"
	"safeagent/docker"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not determine home directory: %v\n", err)
		os.Exit(1)
	}
	root := cmd.NewRootCommand(cmd.Deps{
		ConfigRoot: filepath.Join(home, ".safeagent", "v2"),
		Docker:     docker.CLI{},
	})
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
