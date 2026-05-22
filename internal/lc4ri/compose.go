package lc4ri

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type ComposeRunner struct{}

func (ComposeRunner) Up(workdir string, composeFiles []string, envFile, projectName, pullPolicy string) error {
	cmd := composeCommand(workdir, composeFiles, envFile, projectName, "up", "-d", "--build", "--pull", pullPolicy)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run docker compose up: %w", err)
	}

	return nil
}

func (ComposeRunner) Down(workdir string, composeFiles []string, envFile, projectName string) error {
	cmd := composeCommand(workdir, composeFiles, envFile, projectName, "down")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run docker compose down: %w", err)
	}

	return nil
}

func (ComposeRunner) Status(workdir string, composeFiles []string, envFile, projectName string) error {
	cmd := composeCommand(workdir, composeFiles, envFile, projectName, "ps")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run docker compose ps: %w", err)
	}

	return nil
}

func (ComposeRunner) IsRunning(workdir string, composeFiles []string, envFile, projectName, service string) (bool, error) {
	cmd := composeCommand(workdir, composeFiles, envFile, projectName, "ps", "--status", "running", "--services")

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

func (ComposeRunner) Exec(workdir string, composeFiles []string, envFile, projectName, service string, args ...string) error {
	return runComposeExec(workdir, composeFiles, envFile, projectName, service, nil, args...)
}

func (ComposeRunner) ExecAsUser(workdir string, composeFiles []string, envFile, projectName, service, username string, args ...string) error {
	home := "/home/" + username
	options := []string{
		"--user", username,
		"--env", "HOME=" + home,
	}
	return runComposeExec(workdir, composeFiles, envFile, projectName, service, options, args...)
}

func runComposeExec(workdir string, composeFiles []string, envFile, projectName, service string, options []string, args ...string) error {
	composeArgs := append([]string{"exec"}, options...)
	composeArgs = append(composeArgs, service)
	composeArgs = append(composeArgs, args...)

	cmd := composeCommand(workdir, composeFiles, envFile, projectName, composeArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run docker compose exec: %w", err)
	}

	return nil
}

func composeCommand(workdir string, composeFiles []string, envFile, projectName string, args ...string) *exec.Cmd {
	base := []string{
		"compose",
		"--project-directory", workdir,
	}
	for _, composeFile := range composeFiles {
		base = append(base, "-f", composeFile)
	}
	base = append(base, "--env-file", envFile, "-p", projectName)

	return exec.Command("docker", append(base, args...)...)
}
