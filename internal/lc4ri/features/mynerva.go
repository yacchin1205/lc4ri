package features

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const mynervaName = "mynerva"

type Mynerva struct {
	RuntimeDir string
}

func (f Mynerva) Enable() error {
	dir := f.notebookBuildDir()
	if _, err := os.Stat(dir); err == nil {
		return fmt.Errorf("mynerva is already enabled: %s", dir)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat mynerva notebook build directory %s: %w", dir, err)
	}
	composeFile := f.composeFile()
	if _, err := os.Stat(composeFile); err == nil {
		return fmt.Errorf("mynerva is already enabled: %s", composeFile)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat mynerva compose file %s: %w", composeFile, err)
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create mynerva notebook build directory %s: %w", dir, err)
	}
	if err := writeFile(filepath.Join(dir, "build.sh"), mynervaNotebookBuildScript(), 0o644); err != nil {
		return err
	}
	secretKey, err := newFernetKey()
	if err != nil {
		return err
	}
	if err := writeFile(composeFile, mynervaComposeTemplate(secretKey), 0o644); err != nil {
		return err
	}

	return nil
}

func (f Mynerva) Disable(options DisableOptions) error {
	dir := f.notebookBuildDir()
	if _, err := os.Stat(dir); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("mynerva is not enabled: %s", dir)
		}
		return fmt.Errorf("stat mynerva notebook build directory %s: %w", dir, err)
	}
	composeFile := f.composeFile()
	if _, err := os.Stat(composeFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("mynerva compose file not found: %s", composeFile)
		}
		return fmt.Errorf("stat mynerva compose file %s: %w", composeFile, err)
	}

	if err := removePath(dir); err != nil {
		return err
	}
	if err := os.Remove(composeFile); err != nil {
		return fmt.Errorf("remove mynerva compose file %s: %w", composeFile, err)
	}

	return nil
}

func (f Mynerva) composeFile() string {
	return filepath.Join(f.RuntimeDir, "compose."+mynervaName+".yaml")
}

func (f Mynerva) notebookBuildDir() string {
	return filepath.Join(f.RuntimeDir, "notebook", mynervaName)
}

func mynervaNotebookBuildScript() string {
	return `set -eu
python -m pip install --upgrade --no-cache-dir 'jupyter_mynerva>=2026.10.1'
jupyter labextension enable jupyter-mynerva --level=system
fix-permissions "${CONDA_DIR}"
fix-permissions "/home/${NB_USER}"
`
}

func newFernetKey() (string, error) {
	data := make([]byte, 32)
	if _, err := rand.Read(data); err != nil {
		return "", fmt.Errorf("generate mynerva secret key: %w", err)
	}
	return base64.URLEncoding.EncodeToString(data), nil
}

func mynervaComposeTemplate(secretKey string) string {
	return fmt.Sprintf(`services:
  notebook:
    environment:
      MYNERVA_SECRET_KEY: %s
`, secretKey)
}
