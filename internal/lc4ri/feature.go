package lc4ri

import (
	"fmt"

	"github.com/yacchin1205/lc4ri/internal/lc4ri/features"
)

func (a *App) Enable(args []string) error {
	if err := a.ensureRuntimeFiles(); err != nil {
		return err
	}

	feature, err := a.feature(args)
	if err != nil {
		return err
	}
	return feature.Enable()
}

func (a *App) Disable(args []string) error {
	if err := a.requireComposeFile(); err != nil {
		return err
	}

	feature, err := a.feature(args)
	if err != nil {
		return err
	}
	return feature.Disable()
}

func (a *App) feature(args []string) (features.Feature, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("invalid feature arguments %q\n\n%s", args, usageText())
	}

	switch args[0] {
	case "nbsearch":
		return features.NBSearch{RuntimeDir: a.runtimeDir}, nil
	default:
		return nil, fmt.Errorf("unknown feature %q\n\n%s", args[0], usageText())
	}
}
