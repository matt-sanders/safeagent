package cmd

import (
	"fmt"
	"os"
	"strings"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
)

var (
	exclusionStyle = lipgloss.NewStyle().Faint(true)
)

var excludeCmd = &cobra.Command{
	Use:   "exclude",
	Short: "Manage directory exclusions for the current project",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("could not determine working directory: %w", err)
		}

		// Ensure project exists
		proj, _, err := projectService.GetOrCreate(cwd)
		if err != nil {
			return err
		}

		for {
			exclusions, err := projectService.GetExclusions(cwd)
			if err != nil {
				return err
			}

			// Build menu options
			options := []huh.Option[string]{}
			for _, e := range exclusions {
				options = append(options, huh.NewOption(
					fmt.Sprintf("Remove %s", exclusionStyle.Render(e)),
					"remove:"+e,
				))
			}
			options = append(options, huh.NewOption("Add new exclusion", "add"))
			options = append(options, huh.NewOption("Done", "done"))

			// Show current state
			if len(exclusions) == 0 {
				fmt.Println("No exclusions configured for this project.")
			} else {
				fmt.Println("Current exclusions:")
				for _, e := range exclusions {
					fmt.Printf("  %s\n", exclusionStyle.Render(e))
				}
			}
			fmt.Println()

			var choice string
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("Manage exclusions").
						Options(options...).
						Value(&choice),
				),
			)

			if err := form.Run(); err != nil {
				return nil
			}

			switch {
			case choice == "done":
				return nil
			case choice == "add":
				var newExclusion string
				addForm := huh.NewForm(
					huh.NewGroup(
						huh.NewInput().
							Title("Directory to exclude").
							Value(&newExclusion).
							Validate(func(s string) error {
								if strings.TrimSpace(s) == "" {
									return fmt.Errorf("path cannot be empty")
								}
								return nil
							}),
					),
				)
				if err := addForm.Run(); err != nil {
					return nil
				}
				if err := projectService.AddExclusion(cwd, strings.TrimSpace(newExclusion)); err != nil {
					fmt.Printf("Error: %s\n", err)
				} else {
					fmt.Printf("Excluded %q\n", strings.TrimSpace(newExclusion))
					if err := removeProjectContainer(cwd, proj.ContainerID); err != nil {
						return err
					}
					proj.ContainerID = ""
				}
			case strings.HasPrefix(choice, "remove:"):
				exclusion := strings.TrimPrefix(choice, "remove:")
				if err := projectService.RemoveExclusion(cwd, exclusion); err != nil {
					fmt.Printf("Error: %s\n", err)
				} else {
					fmt.Printf("Removed exclusion %q\n", exclusion)
					if err := removeProjectContainer(cwd, proj.ContainerID); err != nil {
						return err
					}
					proj.ContainerID = ""
				}
			}
			fmt.Println()
		}
	},
}

func init() {
	rootCmd.AddCommand(excludeCmd)
}
