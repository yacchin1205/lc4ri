package lc4ri

import "testing"

func TestParseMynervaArgsRepo(t *testing.T) {
	options, err := parseMynervaArgs([]string{"--repo", "https://github.com/NII-cloud-operation/jupyter-mynerva.git@main"})
	if err != nil {
		t.Fatalf("parseMynervaArgs: %v", err)
	}
	if options.Repo != "https://github.com/NII-cloud-operation/jupyter-mynerva.git@main" {
		t.Fatalf("Repo = %q", options.Repo)
	}
}

func TestParseMynervaArgsRejectsMissingRepoValue(t *testing.T) {
	if _, err := parseMynervaArgs([]string{"--repo"}); err == nil {
		t.Fatal("parseMynervaArgs succeeded")
	}
}

func TestParseMynervaArgsRejectsDuplicateRepo(t *testing.T) {
	args := []string{
		"--repo", "https://github.com/NII-cloud-operation/jupyter-mynerva.git@main",
		"--repo", "https://github.com/NII-cloud-operation/jupyter-mynerva.git@dev",
	}
	if _, err := parseMynervaArgs(args); err == nil {
		t.Fatal("parseMynervaArgs succeeded")
	}
}
