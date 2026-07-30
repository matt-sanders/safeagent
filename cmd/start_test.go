package cmd

import "testing"

func TestBuildMounts_IncludesProjectAndIdentityClaudeDir(t *testing.T) {
	claudeDir := "/home/user/.safeagent/auth/work/.claude"
	mounts := buildMounts("/home/user/myapp", "/home/user/myapp", claudeDir, nil)

	if len(mounts) != 2 {
		t.Fatalf("mounts len = %d, want 2", len(mounts))
	}
	if mounts[0].Source != "/home/user/myapp" || mounts[0].Target != "/home/user/myapp" {
		t.Errorf("project mount = %+v, want source=target=/home/user/myapp", mounts[0])
	}
	if mounts[1].Source != claudeDir || mounts[1].Target != "/claude" {
		t.Errorf("claude mount = %+v, want source=%q target=/claude", mounts[1], claudeDir)
	}
}

func TestBuildMounts_AddsExclusionAnonymousVolumes(t *testing.T) {
	claudeDir := "/x/.claude"
	mounts := buildMounts("/home/user/myapp", "/home/user/myapp", claudeDir, []string{"node_modules", "build"})

	if len(mounts) != 4 {
		t.Fatalf("mounts len = %d, want 4", len(mounts))
	}
	// Exclusions are anonymous volumes: empty Source, Target rooted at cwd.
	if mounts[2].Source != "" || mounts[2].Target != "/home/user/myapp/node_modules" {
		t.Errorf("exclusion[0] = %+v, want empty source, target /home/user/myapp/node_modules", mounts[2])
	}
	if mounts[3].Source != "" || mounts[3].Target != "/home/user/myapp/build" {
		t.Errorf("exclusion[1] = %+v, want empty source, target /home/user/myapp/build", mounts[3])
	}
}
