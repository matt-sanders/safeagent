package profile

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

var (
	nameStyle   = lipgloss.NewStyle().Bold(true)
	detailStyle = lipgloss.NewStyle().Faint(true)
)

// FormatOption returns a styled string for displaying a profile in lists and selects.
// Shows the name in bold and details (id, node version) in subdued colours.
func FormatOption(p Profile) string {
	name := nameStyle.Render(p.Name)
	details := detailStyle.Render(fmt.Sprintf("  id: %s  node: %s", p.ID, p.NodeVersion))
	return name + "\n" + details
}
