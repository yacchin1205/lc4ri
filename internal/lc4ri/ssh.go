package lc4ri

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const sshKeygenScript = `set -euo pipefail

read -r -s -p "Enter passphrase: " passphrase
echo
if [[ -z "${passphrase}" ]]; then
  echo "passphrase must not be empty" >&2
  exit 1
fi

read -r -s -p "Enter same passphrase again: " confirm
echo
if [[ "${passphrase}" != "${confirm}" ]]; then
  echo "passphrases do not match" >&2
  exit 1
fi

mkdir -p "${HOME}/.ssh"
chmod 700 "${HOME}/.ssh"
exec ssh-keygen "$@" -N "${passphrase}"
`

const sshAuthorizeScript = `set -eu
umask 077
mkdir -p ~/.ssh
chmod 700 ~/.ssh
touch ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys
key=$(cat)
if ! grep -qxF "$key" ~/.ssh/authorized_keys; then
  printf '%s\n' "$key" >> ~/.ssh/authorized_keys
fi
`

func (a *App) SSH(args []string) error {
	if len(args) == 0 {
		return sshUsageError()
	}

	switch args[0] {
	case "keygen":
		return a.SSHKeygen(args[1:])
	case "authorize":
		return a.SSHAuthorize(args[1:])
	default:
		return fmt.Errorf("unknown ssh command %q\n\n%s", args[0], sshUsageText())
	}
}

func (a *App) SSHKeygen(args []string) error {
	if hasSSHKeygenPassphraseArg(args) {
		return errors.New("ssh keygen does not allow -N; passphrase is always prompted")
	}

	env, err := a.requireRunningNotebook()
	if err != nil {
		return err
	}

	execArgs := append([]string{"bash", "-lc", sshKeygenScript, "lc4ri-ssh-keygen"}, args...)
	return a.composeRunner.ExecAsUser(a.workdir, a.composeFile, a.envFile, a.projectName, "notebook", env.Username, execArgs...)
}

func (a *App) SSHAuthorize(args []string) error {
	options, target, err := parseSSHAuthorizeArgs(args)
	if err != nil {
		return err
	}

	pubkeyPath := options.pubkeyPath
	if pubkeyPath == "" {
		pubkeyPath, err = a.findPublicKey()
		if err != nil {
			return err
		}
	}

	pubkey, err := readPublicKey(pubkeyPath)
	if err != nil {
		return err
	}

	sshArgs := options.sshArgs()
	sshArgs = append(sshArgs, target, "sh -c "+shellQuote(sshAuthorizeScript))
	cmd := exec.Command("ssh", sshArgs...)
	cmd.Stdin = strings.NewReader(pubkey + "\n")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run ssh authorize: %w", err)
	}

	return nil
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

type sshAuthorizeOptions struct {
	pubkeyPath   string
	identityFile string
	port         string
}

func (o sshAuthorizeOptions) sshArgs() []string {
	var args []string
	if o.identityFile != "" {
		args = append(args, "-i", o.identityFile)
	}
	if o.port != "" {
		args = append(args, "-p", o.port)
	}
	return args
}

func parseSSHAuthorizeArgs(args []string) (*sshAuthorizeOptions, string, error) {
	var options sshAuthorizeOptions
	var target string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--pubkey":
			value, err := nextSSHAuthorizeArg(args, &i, arg)
			if err != nil {
				return nil, "", err
			}
			options.pubkeyPath = value
		case "-i":
			value, err := nextSSHAuthorizeArg(args, &i, arg)
			if err != nil {
				return nil, "", err
			}
			options.identityFile = value
		case "-p":
			value, err := nextSSHAuthorizeArg(args, &i, arg)
			if err != nil {
				return nil, "", err
			}
			options.port = value
		default:
			if strings.HasPrefix(arg, "-") {
				return nil, "", fmt.Errorf("unknown authorize option %q\n\n%s", arg, sshUsageText())
			}
			if target != "" {
				return nil, "", fmt.Errorf("authorize accepts exactly one target\n\n%s", sshUsageText())
			}
			target = arg
		}
	}

	if target == "" {
		return nil, "", fmt.Errorf("authorize requires a target\n\n%s", sshUsageText())
	}

	return &options, target, nil
}

func nextSSHAuthorizeArg(args []string, index *int, option string) (string, error) {
	*index = *index + 1
	if *index >= len(args) {
		return "", fmt.Errorf("%s requires a value\n\n%s", option, sshUsageText())
	}
	return args[*index], nil
}

func (a *App) findPublicKey() (string, error) {
	matches, err := filepath.Glob(filepath.Join(a.workdir, ".ssh", "id_*.pub"))
	if err != nil {
		return "", fmt.Errorf("find public keys: %w", err)
	}
	if len(matches) == 0 {
		return "", errors.New("no public key found; specify --pubkey")
	}
	if len(matches) > 1 {
		return "", errors.New("multiple public keys found; specify --pubkey")
	}
	return matches[0], nil
}

func readPublicKey(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read public key %s: %w", path, err)
	}

	key := string(bytes.TrimRight(data, "\r\n"))
	if key == "" {
		return "", fmt.Errorf("public key %s is empty", path)
	}
	if strings.ContainsAny(key, "\r\n") {
		return "", fmt.Errorf("public key %s must contain exactly one line", path)
	}
	if err := validatePublicKeyLine(key); err != nil {
		return "", fmt.Errorf("invalid public key %s: %w", path, err)
	}

	return key, nil
}

func validatePublicKeyLine(key string) error {
	fields := strings.Fields(key)
	if len(fields) < 2 {
		return errors.New("public key must contain an algorithm and key data")
	}
	if !isAllowedPublicKeyAlgorithm(fields[0]) {
		return fmt.Errorf("unsupported public key algorithm %q", fields[0])
	}
	if fields[1] == "" {
		return errors.New("public key data is empty")
	}
	return nil
}

func isAllowedPublicKeyAlgorithm(algorithm string) bool {
	switch algorithm {
	case "ssh-ed25519",
		"ssh-rsa",
		"ecdsa-sha2-nistp256",
		"ecdsa-sha2-nistp384",
		"ecdsa-sha2-nistp521",
		"sk-ssh-ed25519@openssh.com",
		"sk-ecdsa-sha2-nistp256@openssh.com":
		return true
	default:
		return false
	}
}

func hasSSHKeygenPassphraseArg(args []string) bool {
	for _, arg := range args {
		if arg == "-N" {
			return true
		}
	}
	return false
}

func (a *App) requireRunningNotebook() (*Env, error) {
	if err := a.requireComposeFile(); err != nil {
		return nil, err
	}
	env, err := a.loadEnv()
	if err != nil {
		return nil, err
	}

	running, err := a.composeRunner.IsRunning(a.workdir, a.composeFile, a.envFile, a.projectName, "notebook")
	if err != nil {
		return nil, err
	}
	if !running {
		return nil, errors.New("notebook is not running")
	}

	return env, nil
}

func sshUsageError() error {
	return errors.New(sshUsageText())
}

func sshUsageText() string {
	return `usage: lc4ri ssh <command>

commands:
  keygen [ssh-keygen-options...]
  authorize [--pubkey public_key_file] [-i identity_file] [-p port] target`
}
