package lc4ri

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type App struct {
	workdir       string
	runtimeDir    string
	notebookDir   string
	composeFile   string
	envFile       string
	projectName   string
	composeRunner ComposeRunner
	browser       Browser
}

func NewApp() (*App, error) {
	workdir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}

	runtimeDir := filepath.Join(workdir, ".lc4ri")

	return &App{
		workdir:       workdir,
		runtimeDir:    runtimeDir,
		notebookDir:   filepath.Join(workdir, "notebooks"),
		composeFile:   filepath.Join(runtimeDir, "compose.yaml"),
		envFile:       filepath.Join(runtimeDir, ".env"),
		projectName:   deriveProjectName(workdir),
		composeRunner: ComposeRunner{},
		browser:       Browser{},
	}, nil
}

func deriveProjectName(workdir string) string {
	base := filepath.Base(workdir)
	base = strings.ToLower(base)

	var builder strings.Builder
	lastDash := false
	for _, r := range base {
		valid := r >= 'a' && r <= 'z' || r >= '0' && r <= '9'
		if valid {
			builder.WriteRune(r)
			lastDash = false
			continue
		}

		if !lastDash {
			builder.WriteByte('-')
			lastDash = true
		}
	}

	name := strings.Trim(builder.String(), "-")
	if name == "" {
		name = "lc4ri"
	}

	sum := sha256.Sum256([]byte(workdir))
	suffix := hex.EncodeToString(sum[:4])

	return name + "-" + suffix
}
