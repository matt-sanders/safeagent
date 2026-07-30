package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"safeagent/auth"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage auth identities",
	RunE: func(cmd *cobra.Command, args []string) error {
		identities, err := authService.List()
		if err != nil {
			return err
		}
		if len(identities) == 0 {
			fmt.Println("No identities yet.")
			return nil
		}

		// Best-effort: mark the identity the current project resolves to.
		current := ""
		if cwd, err := os.Getwd(); err == nil {
			if proj, found, err := projectService.FindForCwd(cwd); err == nil && found {
				current = auth.Resolve(proj.AuthName)
			}
		}

		for _, id := range identities {
			marker := "  "
			if id.Name == current {
				marker = "* "
			}
			status := "logged in"
			if !id.LoggedIn {
				status = "not logged in"
			}
			fmt.Printf("%s%s %s\n", marker, id.Name, useAuthDetailStyle.Render("("+status+")"))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
}
