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
		return app.Up()
	case "open":
		return app.Open()
	case "down":
		return app.Down()
	case "status":
		return app.Status()
	default:
		return fmt.Errorf("unknown command %q\n\n%s", args[0], usageText())
	}
}

func usageError() error {
	return errors.New(usageText())
}

func usageText() string {
	return `usage: lc4ri <command>

commands:
  up
  open
  down
  status`
}
