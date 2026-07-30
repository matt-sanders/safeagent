package auth

import "testing"

func TestValidateName(t *testing.T) {
	valid := []string{"default", "work", "personal", "a", "a1", "my-work", "team-2"}
	for _, n := range valid {
		if err := ValidateName(n); err != nil {
			t.Errorf("ValidateName(%q) = %v, want nil", n, err)
		}
	}

	invalid := []string{"", "-work", "Work", "my_work", "with space", "café", "a/b"}
	for _, n := range invalid {
		if err := ValidateName(n); err == nil {
			t.Errorf("ValidateName(%q) = nil, want error", n)
		}
	}
}

func TestResolve(t *testing.T) {
	if got := Resolve(""); got != DefaultName {
		t.Errorf("Resolve(\"\") = %q, want %q", got, DefaultName)
	}
	if got := Resolve("work"); got != "work" {
		t.Errorf("Resolve(\"work\") = %q, want %q", got, "work")
	}
}
