package table_test

import (
	"bytes"
	"testing"

	"safeagent/table"
)

func TestPrintPadsColumnsToLongestValue(t *testing.T) {
	var buf bytes.Buffer
	err := table.Print(&buf, "No stacks.",
		[]string{"NAME", "ID", "NODE"},
		[][]string{
			{"web", "a1b2c3", "lts"},
			{"data-science", "d4e5f6", "22"},
		})
	if err != nil {
		t.Fatal(err)
	}

	want := "" +
		"NAME          ID      NODE\n" +
		"web           a1b2c3  lts\n" +
		"data-science  d4e5f6  22\n"
	if got := buf.String(); got != want {
		t.Fatalf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestPrintUppercasesHeaders(t *testing.T) {
	var buf bytes.Buffer
	if err := table.Print(&buf, "None.", []string{"host path", "container path"}, [][]string{{"/a", "/b"}}); err != nil {
		t.Fatal(err)
	}

	want := "" +
		"HOST PATH  CONTAINER PATH\n" +
		"/a         /b\n"
	if got := buf.String(); got != want {
		t.Fatalf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestPrintShowsEmptyMessageForNoRows(t *testing.T) {
	var buf bytes.Buffer
	if err := table.Print(&buf, "No profiles.", []string{"NAME", "ID"}, nil); err != nil {
		t.Fatal(err)
	}

	if got := buf.String(); got != "No profiles.\n" {
		t.Fatalf("got %q, want %q", got, "No profiles.\n")
	}
}
