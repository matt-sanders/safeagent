package image_test

import (
	"testing"

	"safeagent/image"
)

func TestRefAndName(t *testing.T) {
	if got := image.Ref("ab12cd34", "ef56ab78"); got != "safeagent/ab12cd34:ef56ab78" {
		t.Fatalf("Ref = %q", got)
	}
	if got := image.Name("node22", "work"); got != "node22:work" {
		t.Fatalf("Name = %q", got)
	}
}

func TestParseName(t *testing.T) {
	stack, profile, err := image.ParseName("node22:work")
	if err != nil || stack != "node22" || profile != "work" {
		t.Fatalf("got %q %q %v", stack, profile, err)
	}
	for _, bad := range []string{"node22", ":work", "node22:", "Node22:work", "a:b:c"} {
		if _, _, err := image.ParseName(bad); err == nil {
			t.Errorf("ParseName(%q) accepted", bad)
		}
	}
}
