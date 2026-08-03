package cmd

import (
	"fmt"
	"strings"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
)

var mountStyle = lipgloss.NewStyle().Faint(true)

var authMountsCmd = &cobra.Command{
	Use:   "mounts <name>",
	Short: "Manage bind mounts for an auth identity",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		exists, err := authService.Exists(name)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("identity %q not found", name)
		}

		for {
			cfg, err := authService.LoadConfig(name)
			if err != nil {
				return err
			}

			options := []huh.Option[string]{}
			for _, m := range cfg.Mounts {
				label := fmt.Sprintf("Remove %s → %s", mountStyle.Render(m.Source), mountStyle.Render(m.Target))
				options = append(options, huh.NewOption(label, "remove:"+m.Target))
			}
			options = append(options, huh.NewOption("Add new mount", "add"))
			options = append(options, huh.NewOption("Done", "done"))

			if len(cfg.Mounts) == 0 {
				fmt.Printf("No mounts configured for identity %q.\n", name)
			} else {
				fmt.Printf("Mounts for %q:\n", name)
				for _, m := range cfg.Mounts {
					fmt.Printf("  %s → %s\n", mountStyle.Render(m.Source), mountStyle.Render(m.Target))
				}
			}
			fmt.Println()

			var choice string
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("Manage mounts").
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
				var source, target string
				addForm := huh.NewForm(
					huh.NewGroup(
						huh.NewInput().
							Title("Host path (source)").
							Value(&source).
							Validate(func(s string) error {
								if strings.TrimSpace(s) == "" {
									return fmt.Errorf("path cannot be empty")
								}
								return nil
							}),
						huh.NewInput().
							Title("Container path (target)").
							Value(&target).
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
				source = strings.TrimSpace(source)
				target = strings.TrimSpace(target)
				if err := authService.AddMount(name, source, target); err != nil {
					fmt.Printf("Error: %s\n", err)
				} else {
					fmt.Printf("Added mount %s → %s\n", source, target)
				}
			case strings.HasPrefix(choice, "remove:"):
				target := strings.TrimPrefix(choice, "remove:")
				if err := authService.RemoveMount(name, target); err != nil {
					fmt.Printf("Error: %s\n", err)
				} else {
					fmt.Printf("Removed mount with target %q\n", target)
				}
			}
			fmt.Println()
		}
	},
}

func init() {
	authCmd.AddCommand(authMountsCmd)
}
