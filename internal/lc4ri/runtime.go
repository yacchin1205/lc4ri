package lc4ri

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const defaultImage = "niicloudoperation/notebook:feature-lab"
const serverReadyTimeout = 45 * time.Second

func (a *App) Up(options *UpOptions) error {
	if err := a.ensureRuntimeFiles(); err != nil {
		return err
	}

	composeFiles, err := a.activeComposeFiles()
	if err != nil {
		return err
	}
	if err := a.composeRunner.Up(a.workdir, composeFiles, a.envFile, a.projectName, options.PullPolicy); err != nil {
		return err
	}

	if options.Open {
		env, err := a.loadEnv()
		if err != nil {
			return err
		}
		if err := waitForServer(env); err != nil {
			return err
		}
		return a.Open()
	}

	return nil
}

func (a *App) Open() error {
	if err := a.requireComposeFile(); err != nil {
		return err
	}
	env, err := a.loadEnv()
	if err != nil {
		return err
	}

	composeFiles, err := a.activeComposeFiles()
	if err != nil {
		return err
	}
	running, err := a.composeRunner.IsRunning(a.workdir, composeFiles, a.envFile, a.projectName, "notebook")
	if err != nil {
		return err
	}
	if !running {
		return errors.New("notebook is not running")
	}

	url := fmt.Sprintf("http://localhost:%s/tree?token=%s", env.Port, env.Token)
	return a.browser.Open(url)
}

func (a *App) Down() error {
	if err := a.requireComposeFile(); err != nil {
		return err
	}
	if _, err := a.loadEnv(); err != nil {
		return err
	}

	composeFiles, err := a.activeComposeFiles()
	if err != nil {
		return err
	}
	return a.composeRunner.Down(a.workdir, composeFiles, a.envFile, a.projectName)
}

func (a *App) Status() error {
	if err := a.requireComposeFile(); err != nil {
		return err
	}
	if _, err := a.loadEnv(); err != nil {
		return err
	}

	composeFiles, err := a.activeComposeFiles()
	if err != nil {
		return err
	}
	return a.composeRunner.Status(a.workdir, composeFiles, a.envFile, a.projectName)
}

func (a *App) ensureRuntimeFiles() error {
	if err := os.MkdirAll(a.runtimeDir, 0o755); err != nil {
		return fmt.Errorf("create runtime directory %s: %w", a.runtimeDir, err)
	}

	if err := a.ensureComposeFile(); err != nil {
		return err
	}
	if err := a.ensureNotebookDockerfile(); err != nil {
		return err
	}
	if err := a.ensureEnvFile(); err != nil {
		return err
	}

	if _, err := a.loadEnv(); err != nil {
		return err
	}

	if err := os.MkdirAll(a.notebookDir, 0o755); err != nil {
		return fmt.Errorf("create notebook directory %s: %w", a.notebookDir, err)
	}

	return nil
}

func (a *App) ensureComposeFile() error {
	if _, err := os.Stat(a.composeFile); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat compose file %s: %w", a.composeFile, err)
	}

	content := composeTemplate()
	if err := os.WriteFile(a.composeFile, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write compose file %s: %w", a.composeFile, err)
	}

	return nil
}

func (a *App) ensureNotebookDockerfile() error {
	path := filepath.Join(a.runtimeDir, "notebook", "Dockerfile")
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat notebook Dockerfile %s: %w", path, err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create notebook build directory %s: %w", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(notebookDockerfile()), 0o644); err != nil {
		return fmt.Errorf("write notebook Dockerfile %s: %w", path, err)
	}
	return nil
}

func (a *App) ensureEnvFile() error {
	if _, err := os.Stat(a.envFile); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat env file %s: %w", a.envFile, err)
	}

	port, err := choosePort()
	if err != nil {
		return err
	}

	token, err := newToken()
	if err != nil {
		return err
	}

	username, err := currentUsername()
	if err != nil {
		return err
	}

	content := fmt.Sprintf(
		"LC4RI_PORT=%s\nLC4RI_TOKEN=%s\nLC4RI_IMAGE=%s\nLC4RI_USER=%s\nLC4RI_UID=%d\nLC4RI_GID=%d\n",
		port,
		token,
		defaultImage,
		username,
		os.Getuid(),
		os.Getgid(),
	)
	if err := os.WriteFile(a.envFile, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write env file %s: %w", a.envFile, err)
	}

	return nil
}

func (a *App) loadEnv() (*Env, error) {
	if err := a.requireEnvFile(); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(a.envFile)
	if err != nil {
		return nil, fmt.Errorf("read env file %s: %w", a.envFile, err)
	}

	envMap := make(map[string]string)
	for lineNo, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			return nil, fmt.Errorf("parse env file %s: line %d is invalid", a.envFile, lineNo+1)
		}
		envMap[key] = value
	}

	env := &Env{
		Port:     envMap["LC4RI_PORT"],
		Token:    envMap["LC4RI_TOKEN"],
		Image:    envMap["LC4RI_IMAGE"],
		Username: envMap["LC4RI_USER"],
		UID:      envMap["LC4RI_UID"],
		GID:      envMap["LC4RI_GID"],
	}

	if env.Port == "" {
		return nil, fmt.Errorf("env file %s does not define LC4RI_PORT", a.envFile)
	}
	if env.Token == "" {
		return nil, fmt.Errorf("env file %s does not define LC4RI_TOKEN", a.envFile)
	}
	if env.Image == "" {
		return nil, fmt.Errorf("env file %s does not define LC4RI_IMAGE", a.envFile)
	}
	if env.Username == "" {
		return nil, fmt.Errorf("env file %s does not define LC4RI_USER", a.envFile)
	}
	if env.UID == "" {
		return nil, fmt.Errorf("env file %s does not define LC4RI_UID", a.envFile)
	}
	if env.GID == "" {
		return nil, fmt.Errorf("env file %s does not define LC4RI_GID", a.envFile)
	}

	if _, err := strconv.Atoi(env.Port); err != nil {
		return nil, fmt.Errorf("env file %s has invalid LC4RI_PORT %q: %w", a.envFile, env.Port, err)
	}
	if _, err := strconv.Atoi(env.UID); err != nil {
		return nil, fmt.Errorf("env file %s has invalid LC4RI_UID %q: %w", a.envFile, env.UID, err)
	}
	if _, err := strconv.Atoi(env.GID); err != nil {
		return nil, fmt.Errorf("env file %s has invalid LC4RI_GID %q: %w", a.envFile, env.GID, err)
	}

	return env, nil
}

func (a *App) requireComposeFile() error {
	if _, err := os.Stat(a.composeFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("compose file not found: %s", a.composeFile)
		}
		return fmt.Errorf("stat compose file %s: %w", a.composeFile, err)
	}
	return nil
}

func (a *App) requireEnvFile() error {
	if _, err := os.Stat(a.envFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("env file not found: %s", a.envFile)
		}
		return fmt.Errorf("stat env file %s: %w", a.envFile, err)
	}
	return nil
}

func (a *App) activeComposeFiles() ([]string, error) {
	composeFiles := []string{a.composeFile}
	matches, err := filepath.Glob(filepath.Join(a.runtimeDir, "compose.*.yaml"))
	if err != nil {
		return nil, fmt.Errorf("find compose feature fragments: %w", err)
	}
	sort.Strings(matches)
	composeFiles = append(composeFiles, matches...)
	return composeFiles, nil
}

type Env struct {
	Port     string
	Token    string
	Image    string
	Username string
	UID      string
	GID      string
}

func composeTemplate() string {
	return `services:
  notebook:
    build:
      context: ./.lc4ri/notebook
      args:
        LC4RI_BASE_IMAGE: ${LC4RI_IMAGE}
    user: root
    working_dir: /home/${LC4RI_USER}/notebooks
    environment:
      NB_USER: ${LC4RI_USER}
      NB_UID: ${LC4RI_UID}
      NB_GID: ${LC4RI_GID}
      GRANT_SUDO: "yes"
    command:
      - start-notebook.sh
      - --ServerApp.ip=0.0.0.0
      - --ServerApp.port=8888
      - --ServerApp.root_dir=/home/${LC4RI_USER}/notebooks
      - --ServerApp.open_browser=False
      - --IdentityProvider.token=${LC4RI_TOKEN}
    ports:
      - 127.0.0.1:${LC4RI_PORT}:8888
    volumes:
      - ./:/home/${LC4RI_USER}
`
}

func notebookDockerfile() string {
	return `ARG LC4RI_BASE_IMAGE
FROM ${LC4RI_BASE_IMAGE}

USER root

COPY . /tmp/lc4ri-notebook
RUN set -eu; \
    find /tmp/lc4ri-notebook -mindepth 2 -maxdepth 2 -type f -name build.sh | sort | while IFS= read -r script; do sh "$script"; done; \
    rm -rf /tmp/lc4ri-notebook
`
}

func choosePort() (string, error) {
	const minPort = 10000
	const span = 50000

	value, err := rand.Int(rand.Reader, big.NewInt(span))
	if err != nil {
		return "", fmt.Errorf("choose port: %w", err)
	}

	return strconv.FormatInt(value.Int64()+minPort, 10), nil
}

func newToken() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	return hex.EncodeToString(buf), nil
}

func waitForServer(env *Env) error {
	url := fmt.Sprintf("http://localhost:%s/tree?token=%s", env.Port, env.Token)
	deadline := time.Now().Add(serverReadyTimeout)
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	var lastErr error
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode < http.StatusInternalServerError {
				return nil
			}
			lastErr = fmt.Errorf("server returned %s", resp.Status)
		} else {
			lastErr = err
		}

		time.Sleep(500 * time.Millisecond)
	}

	if lastErr != nil {
		return fmt.Errorf("wait for notebook server: %w", lastErr)
	}
	return errors.New("wait for notebook server timed out")
}

func currentUsername() (string, error) {
	if user := os.Getenv("USER"); user != "" {
		return user, nil
	}
	return "", errors.New("USER is not set")
}
