package lc4ri

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type ComposeRunner struct{}

func (ComposeRunner) Up(workdir, composeFile, envFile, projectName string) error {
	cmd := composeCommand(workdir, composeFile, envFile, projectName, "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run docker compose up: %w", err)
	}

	return nil
}

func (ComposeRunner) Down(workdir, composeFile, envFile, projectName string) error {
	cmd := composeCommand(workdir, composeFile, envFile, projectName, "down")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run docker compose down: %w", err)
	}

	return nil
}

func (ComposeRunner) Status(workdir, composeFile, envFile, projectName string) error {
	cmd := composeCommand(workdir, composeFile, envFile, projectName, "ps")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run docker compose ps: %w", err)
	}

	return nil
}

func (ComposeRunner) IsRunning(workdir, composeFile, envFile, projectName, service string) (bool, error) {
	cmd := composeCommand(workdir, composeFile, envFile, projectName, "ps", "--status", "running", "--services")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("run docker compose ps --status running: %w", err)
	}

	for _, line := range strings.Split(stdout.String(), "\n") {
		if strings.TrimSpace(line) == service {
			return true, nil
		}
	}

	return false, nil
}

func composeCommand(workdir, composeFile, envFile, projectName string, args ...string) *exec.Cmd {
	base := []string{
		"compose",
		"--project-directory", workdir,
		"-f", composeFile,
		"--env-file", envFile,
		"-p", projectName,
	}

	return exec.Command("docker", append(base, args...)...)
}
