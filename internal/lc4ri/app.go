package lc4ri

import (
	"errors"
	"fmt"
)

func Run(args []string) error {
	if len(args) == 0 {
		return usageError()
	}

	app, err := NewApp()
	if err != nil {
		return err
	}

	switch args[0] {
	case "up":
		options, err := parseUpArgs(args[1:])
		if err != nil {
			return err
		}
		return app.Up(options)
	case "open":
		return app.Open()
	case "down":
		return app.Down()
	case "status":
		return app.Status()
	case "logs":
		return app.Logs(args[1:])
	case "enable":
		return app.Enable(args[1:])
	case "disable":
		return app.Disable(args[1:])
	case "ssh":
		return app.SSH(args[1:])
	default:
		return fmt.Errorf("unknown command %q\n\n%s", args[0], usageText())
	}
}

func usageError() error {
	return errors.New(usageText())
}

type UpOptions struct {
	Open       bool
	PullPolicy string
}

func parseUpArgs(args []string) (*UpOptions, error) {
	options := &UpOptions{
		Open:       true,
		PullPolicy: "missing",
	}

	for _, arg := range args {
		switch arg {
		case "--no-open":
			options.Open = false
		case "--pull":
			options.PullPolicy = "always"
		default:
			return nil, fmt.Errorf("invalid up argument %q\n\n%s", arg, usageText())
		}
	}

	return options, nil
}

func usageText() string {
	return `usage: lc4ri <command>

commands:
  up [--no-open] [--pull]
  open
  down
  status
  logs [docker-compose-logs-options...]
  enable nbsearch
  disable nbsearch
  ssh keygen [ssh-keygen-options...]
  ssh authorize [--pubkey public_key_file] [-i identity_file] [-p port] target`
}
