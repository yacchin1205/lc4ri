package lc4ri

import (
	"fmt"
	"strings"

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
	if _, err := a.loadEnv(); err != nil {
		return err
	}
	composeFiles, err := a.activeComposeFiles()
	if err != nil {
		return err
	}
	runningServices, err := a.composeRunner.RunningServices(a.workdir, composeFiles, a.envFile, a.projectName)
	if err != nil {
		return err
	}
	if len(runningServices) > 0 {
		return fmt.Errorf("cannot disable while services are running: %s; run `lc4ri down` first", strings.Join(runningServices, ", "))
	}

	feature, err := a.feature(args)
	if err != nil {
		return err
	}
	return feature.Disable(features.DisableOptions{ProjectName: a.projectName})
}

func (a *App) feature(args []string) (features.Feature, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("invalid feature arguments %q\n\n%s", args, usageText())
	}

	switch args[0] {
	case "nbsearch":
		if len(args) != 1 {
			return nil, fmt.Errorf("invalid nbsearch arguments %q\n\n%s", args[1:], usageText())
		}
		return features.NBSearch{RuntimeDir: a.runtimeDir}, nil
	case "mynerva":
		options, err := parseMynervaArgs(args[1:])
		if err != nil {
			return nil, err
		}
		return features.Mynerva{RuntimeDir: a.runtimeDir, Repo: options.Repo}, nil
	default:
		return nil, fmt.Errorf("unknown feature %q\n\n%s", args[0], usageText())
	}
}

type mynervaOptions struct {
	Repo string
}

func parseMynervaArgs(args []string) (*mynervaOptions, error) {
	options := &mynervaOptions{}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--repo":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("missing value for --repo\n\n%s", usageText())
			}
			if options.Repo != "" {
				return nil, fmt.Errorf("duplicate --repo\n\n%s", usageText())
			}
			i++
			options.Repo = args[i]
		default:
			return nil, fmt.Errorf("invalid mynerva argument %q\n\n%s", args[i], usageText())
		}
	}
	return options, nil
}
