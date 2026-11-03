// Package table prints docker-style tables: uppercase headers, columns padded
// to their longest value, two-space gutters and no styling.
package table

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// Print writes headers and rows as a table to w. With no rows it writes the
// empty message (e.g. "No profiles.") on its own line instead.
func Print(w io.Writer, empty string, headers []string, rows [][]string) error {
	if len(rows) == 0 {
		_, err := fmt.Fprintln(w, empty)
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	upper := make([]string, len(headers))
	for i, h := range headers {
		upper[i] = strings.ToUpper(h)
	}
	fmt.Fprintln(tw, strings.Join(upper, "\t"))
	for _, row := range rows {
		fmt.Fprintln(tw, strings.Join(row, "\t"))
	}
	return tw.Flush()
}
