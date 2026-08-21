package docker

import (
	_ "embed"
	"strings"
)

//go:embed Dockerfile
var Dockerfile string

// InjectExtensions splices extra Dockerfile lines at the {{PROFILE_EXTENSIONS}}
// marker. If extra is empty the marker line is removed cleanly.
func InjectExtensions(extra string) string {
	trimmed := strings.TrimSpace(extra)
	replacement := ""
	if trimmed != "" {
		replacement = trimmed + "\n"
	}
	return strings.Replace(Dockerfile, "# {{PROFILE_EXTENSIONS}}\n", replacement, 1)
}
