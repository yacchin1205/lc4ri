package features

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const mynervaName = "mynerva"

type Mynerva struct {
	RuntimeDir string
	Repo       string
}

func (f Mynerva) Enable() error {
	if err := validateMynervaRepo(f.Repo); err != nil {
		return err
	}

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
	if err := writeFile(filepath.Join(dir, "build.sh"), mynervaNotebookBuildScript(f.Repo), 0o644); err != nil {
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

func mynervaNotebookBuildScript(repo string) string {
	install := ""
	if repo != "" {
		install = fmt.Sprintf(`if ! command -v npm; then
  mamba install -y nodejs=22
  mamba clean -a -f -y
fi
python -m pip install --upgrade --force-reinstall --no-cache-dir %s
`, shellQuote(mynervaPipSpec(repo)))
	}

	return fmt.Sprintf(`set -eu
%sjupyter labextension enable jupyter-mynerva --level=system
fix-permissions "${CONDA_DIR}"
fix-permissions "/home/${NB_USER}"
`, install)
}

func mynervaPipSpec(repo string) string {
	return "git+" + repo
}

func validateMynervaRepo(repo string) error {
	if repo == "" {
		return nil
	}
	if strings.TrimSpace(repo) != repo || strings.ContainsAny(repo, " \t\n\r") {
		return fmt.Errorf("mynerva repo must not contain whitespace: %s", repo)
	}
	if !strings.HasPrefix(repo, "https://") && !strings.HasPrefix(repo, "ssh://") && !strings.HasPrefix(repo, "git@") {
		return fmt.Errorf("mynerva repo must be a git URL with an explicit ref: %s", repo)
	}
	if !hasGitRef(repo) {
		return fmt.Errorf("mynerva repo must include an explicit ref with @: %s", repo)
	}
	return nil
}

func hasGitRef(repo string) bool {
	if strings.HasPrefix(repo, "git@") {
		index := strings.LastIndex(repo, "@")
		return index > 0 && index < len(repo)-1 && strings.Count(repo, "@") >= 2
	}
	index := strings.LastIndex(repo, "@")
	return index > 0 && index < len(repo)-1
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
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
