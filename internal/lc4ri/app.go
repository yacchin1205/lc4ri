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
		open, err := parseUpArgs(args[1:])
		if err != nil {
			return err
		}
		return app.Up(open)
	case "open":
		return app.Open()
	case "down":
		return app.Down()
	case "status":
		return app.Status()
	case "ssh":
		return app.SSH(args[1:])
	default:
		return fmt.Errorf("unknown command %q\n\n%s", args[0], usageText())
	}
}

func usageError() error {
	return errors.New(usageText())
}

func parseUpArgs(args []string) (bool, error) {
	switch len(args) {
	case 0:
		return true, nil
	case 1:
		if args[0] == "--no-open" {
			return false, nil
		}
	}

	return false, fmt.Errorf("invalid up arguments %q\n\n%s", args, usageText())
}

func usageText() string {
	return `usage: lc4ri <command>

commands:
  up [--no-open]
  open
  down
  status
  ssh keygen [ssh-keygen-options...]`
}
