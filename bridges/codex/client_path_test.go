package codex

import (
	"path/filepath"
	"testing"
)

func TestResolveCodexWorkingDirectoryExpandsTilde(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	got, err := resolveCodexWorkingDirectory("~/workspace/project")
	if err != nil {
		t.Fatalf("resolveCodexWorkingDirectory returned error: %v", err)
	}

	want := filepath.Join(home, "workspace", "project")
	if got != want {
		t.Fatalf("resolveCodexWorkingDirectory returned %q, want %q", got, want)
	}
}

func TestResolveCodexWorkingDirectoryExpandsBareTilde(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	got, err := resolveCodexWorkingDirectory("~")
	if err != nil {
		t.Fatalf("resolveCodexWorkingDirectory returned error: %v", err)
	}
	if got != home {
		t.Fatalf("resolveCodexWorkingDirectory returned %q, want %q", got, home)
	}
}

func TestResolveCodexWorkingDirectoryAcceptsAbsolutePath(t *testing.T) {
	want := filepath.Join(string(filepath.Separator), "tmp", "workspace")

	got, err := resolveCodexWorkingDirectory(want)
	if err != nil {
		t.Fatalf("resolveCodexWorkingDirectory returned error: %v", err)
	}
	if got != want {
		t.Fatalf("resolveCodexWorkingDirectory returned %q, want %q", got, want)
	}
}

func TestResolveCodexWorkingDirectoryRejectsRelativePath(t *testing.T) {
	if _, err := resolveCodexWorkingDirectory("projects/labs"); err == nil {
		t.Fatal("expected relative path to be rejected")
	}
}
