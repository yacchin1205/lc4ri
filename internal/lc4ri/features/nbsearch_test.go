package features

import (
	"bytes"
	"strings"
	"testing"
)

func TestConfirmNBSearchVolumeRemovalAcceptsYes(t *testing.T) {
	for _, input := range []string{"y\n", "yes\n", "Y\n", "YES\n"} {
		t.Run(strings.TrimSpace(input), func(t *testing.T) {
			var output bytes.Buffer
			if err := confirmNBSearchVolumeRemoval(strings.NewReader(input), &output); err != nil {
				t.Fatalf("confirmNBSearchVolumeRemoval: %v", err)
			}
			if output.String() != "Remove nbsearch volumes? [y/N] " {
				t.Fatalf("output = %q", output.String())
			}
		})
	}
}

func TestConfirmNBSearchVolumeRemovalRejectsNonYes(t *testing.T) {
	for _, input := range []string{"\n", "n\n", "no\n"} {
		t.Run(strings.TrimSpace(input), func(t *testing.T) {
			var output bytes.Buffer
			if err := confirmNBSearchVolumeRemoval(strings.NewReader(input), &output); err == nil {
				t.Fatal("confirmNBSearchVolumeRemoval succeeded")
			}
		})
	}
}

func TestConfirmNBSearchVolumeRemovalRejectsEOF(t *testing.T) {
	var output bytes.Buffer
	if err := confirmNBSearchVolumeRemoval(strings.NewReader(""), &output); err == nil {
		t.Fatal("confirmNBSearchVolumeRemoval succeeded")
	}
}

func TestNBSearchVolumeNames(t *testing.T) {
	got := nbsearchVolumeNames("example")
	want := []string{
		"example_nbsearch_solr_data",
		"example_nbsearch_minio_data",
	}
	if len(got) != len(want) {
		t.Fatalf("len(got) = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
