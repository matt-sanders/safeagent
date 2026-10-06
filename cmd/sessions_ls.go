package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var sessionsLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List running sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		sessions, err := dockerService.ListRunningSessions()
		if err != nil {
			return err
		}

		if len(sessions) == 0 {
			fmt.Println("No running sessions.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "CONTAINER\tPROFILE\tDIRECTORY\tRUNNING FOR")
		for _, s := range sessions {
			profileName := resolveProfileName(s.ImageName)
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", s.ContainerID, profileName, s.WorkingDir, s.RunningFor)
		}
		w.Flush()
		return nil
	},
}

// resolveProfileName extracts the profile ID from an image name like
// "safeagent-profile-{id}:latest" and looks up the human-readable name.
func resolveProfileName(imageName string) string {
	id := strings.TrimPrefix(imageName, "safeagent-profile-")
	id = strings.TrimSuffix(id, ":latest")
	prof, err := profileService.GetByID(id)
	if err != nil {
		return id
	}
	return prof.Name
}

func init() {
	sessionsCmd.AddCommand(sessionsLsCmd)
}
