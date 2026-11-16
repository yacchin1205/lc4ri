package lc4ri

import "testing"

func TestMynervaRejectsExtraArguments(t *testing.T) {
	app := &App{runtimeDir: t.TempDir()}
	if _, err := app.feature([]string{"mynerva", "--repo", "https://github.com/NII-cloud-operation/jupyter-mynerva.git@main"}); err == nil {
		t.Fatal("feature accepted unsupported mynerva arguments")
	}
}
