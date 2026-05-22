package lc4ri

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEditNotebookDockerfileCreatesDefaultDockerfile(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	t.Setenv("VISUAL", "true")

	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp: %v", err)
	}
	if err := app.EditNotebookDockerfile(); err != nil {
		t.Fatalf("EditNotebookDockerfile: %v", err)
	}

	path := filepath.Join(root, ".lc4ri", "notebook", "Dockerfile")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read notebook Dockerfile: %v", err)
	}
	if string(data) != notebookDockerfile() {
		t.Fatalf("notebook Dockerfile = %q", string(data))
	}
}
