package features

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMynervaEnableWritesNotebookBuildScript(t *testing.T) {
	runtimeDir := t.TempDir()
	feature := Mynerva{RuntimeDir: runtimeDir}

	if err := feature.Enable(); err != nil {
		t.Fatalf("Enable: %v", err)
	}

	path := filepath.Join(runtimeDir, "notebook", "mynerva", "build.sh")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read build script: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "jupyter labextension enable jupyter-mynerva --level=system") {
		t.Fatalf("build script does not enable jupyter-mynerva:\n%s", content)
	}
	if strings.Contains(content, "command -v") {
		t.Fatalf("build script contains preflight command checks:\n%s", content)
	}
	if strings.Contains(content, "/dev/null") {
		t.Fatalf("build script redirects output to /dev/null:\n%s", content)
	}

	composePath := filepath.Join(runtimeDir, "compose.mynerva.yaml")
	compose, err := os.ReadFile(composePath)
	if err != nil {
		t.Fatalf("read compose fragment: %v", err)
	}
	secretKey := readMynervaSecretKey(t, string(compose))
	decoded, err := base64.URLEncoding.DecodeString(secretKey)
	if err != nil {
		t.Fatalf("decode secret key: %v", err)
	}
	if len(decoded) != 32 {
		t.Fatalf("decoded secret key length = %d, want 32", len(decoded))
	}
}

func TestMynervaEnableFailsWhenAlreadyEnabled(t *testing.T) {
	runtimeDir := t.TempDir()
	feature := Mynerva{RuntimeDir: runtimeDir}

	if err := feature.Enable(); err != nil {
		t.Fatalf("Enable: %v", err)
	}
	if err := feature.Enable(); err == nil {
		t.Fatal("Enable succeeded")
	}
}

func TestMynervaDisableRemovesNotebookBuildDirectory(t *testing.T) {
	runtimeDir := t.TempDir()
	feature := Mynerva{RuntimeDir: runtimeDir}

	if err := feature.Enable(); err != nil {
		t.Fatalf("Enable: %v", err)
	}
	if err := feature.Disable(DisableOptions{}); err != nil {
		t.Fatalf("Disable: %v", err)
	}

	if _, err := os.Stat(filepath.Join(runtimeDir, "notebook", "mynerva")); !os.IsNotExist(err) {
		t.Fatalf("mynerva build directory still exists: %v", err)
	}
	if _, err := os.Stat(filepath.Join(runtimeDir, "compose.mynerva.yaml")); !os.IsNotExist(err) {
		t.Fatalf("mynerva compose file still exists: %v", err)
	}
}

func TestMynervaComposeTemplateSetsSecretKey(t *testing.T) {
	compose := mynervaComposeTemplate("secret")
	if !strings.Contains(compose, "MYNERVA_SECRET_KEY: secret") {
		t.Fatalf("compose template does not set MYNERVA_SECRET_KEY:\n%s", compose)
	}
}

func TestMynervaEnableWithRepoWritesPipInstall(t *testing.T) {
	runtimeDir := t.TempDir()
	feature := Mynerva{
		RuntimeDir: runtimeDir,
		Repo:       "https://github.com/NII-cloud-operation/jupyter-mynerva.git@main",
	}

	if err := feature.Enable(); err != nil {
		t.Fatalf("Enable: %v", err)
	}

	path := filepath.Join(runtimeDir, "notebook", "mynerva", "build.sh")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read build script: %v", err)
	}
	want := "python -m pip install --upgrade --force-reinstall --no-cache-dir 'git+https://github.com/NII-cloud-operation/jupyter-mynerva.git@main'"
	content := string(data)
	if !strings.Contains(content, "if ! command -v npm; then") {
		t.Fatalf("build script does not check npm before installing repo:\n%s", content)
	}
	if !strings.Contains(content, "mamba install -y nodejs=22") {
		t.Fatalf("build script does not install nodejs before installing repo:\n%s", content)
	}
	if !strings.Contains(content, want) {
		t.Fatalf("build script does not install repo; want %q in:\n%s", want, content)
	}
	if strings.Contains(content, "/dev/null") {
		t.Fatalf("build script redirects output to /dev/null:\n%s", content)
	}
}

func TestMynervaEnableRejectsRepoWithoutRef(t *testing.T) {
	feature := Mynerva{
		RuntimeDir: t.TempDir(),
		Repo:       "https://github.com/NII-cloud-operation/jupyter-mynerva.git",
	}

	if err := feature.Enable(); err == nil {
		t.Fatal("Enable succeeded")
	}
}

func readMynervaSecretKey(t *testing.T, compose string) string {
	t.Helper()
	for _, line := range strings.Split(compose, "\n") {
		line = strings.TrimSpace(line)
		const prefix = "MYNERVA_SECRET_KEY: "
		if strings.HasPrefix(line, prefix) {
			return strings.TrimPrefix(line, prefix)
		}
	}
	t.Fatalf("MYNERVA_SECRET_KEY not found:\n%s", compose)
	return ""
}
