package lc4ri

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewAppUsesAncestorRuntimeDirectory(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".lc4ri"), 0o755); err != nil {
		t.Fatalf("create runtime directory: %v", err)
	}
	nested := filepath.Join(root, "notebooks", "example")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("create nested directory: %v", err)
	}
	t.Chdir(nested)

	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp: %v", err)
	}

	if app.workdir != root {
		t.Fatalf("workdir = %q, want %q", app.workdir, root)
	}
	if app.runtimeDir != filepath.Join(root, ".lc4ri") {
		t.Fatalf("runtimeDir = %q", app.runtimeDir)
	}
	if app.notebookDir != filepath.Join(root, "notebooks") {
		t.Fatalf("notebookDir = %q", app.notebookDir)
	}
}

func TestNewAppUsesCurrentDirectoryWithoutAncestorRuntimeDirectory(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "notebooks", "example")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("create nested directory: %v", err)
	}
	t.Chdir(nested)

	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp: %v", err)
	}

	if app.workdir != nested {
		t.Fatalf("workdir = %q, want %q", app.workdir, nested)
	}
	if app.runtimeDir != filepath.Join(nested, ".lc4ri") {
		t.Fatalf("runtimeDir = %q", app.runtimeDir)
	}
}
