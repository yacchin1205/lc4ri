package lc4ri

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (a *App) Image(args []string) error {
	if len(args) != 1 || args[0] != "edit" {
		return fmt.Errorf("invalid image arguments %q\n\n%s", args, usageText())
	}
	return a.EditNotebookDockerfile()
}

func (a *App) EditNotebookDockerfile() error {
	if err := a.ensureNotebookDockerfile(); err != nil {
		return err
	}
	return runEditor(a.notebookDockerfilePath())
}

func (a *App) notebookDockerfilePath() string {
	return filepath.Join(a.runtimeDir, "notebook", "Dockerfile")
}

func runEditor(path string) error {
	editor := os.Getenv("VISUAL")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = "vi"
	}

	parts := strings.Fields(editor)
	if len(parts) == 0 {
		return errors.New("editor command is empty")
	}

	cmd := exec.Command(parts[0], append(parts[1:], path)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run editor %q: %w", editor, err)
	}
	return nil
}
