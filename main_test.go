package main

import (
	"io"
	"strings"
	"testing"
)

func TestAskOverwrite(t *testing.T) {
	answers := map[string]bool{"y\n": true, "YES\n": true, "n\n": false, "\n": false, "": false}
	for answer, want := range answers {
		got := askOverwrite("PCA-001-0001", strings.NewReader(answer), io.Discard)
		if got != want {
			t.Errorf("answer %q: got %v, want %v", answer, got, want)
		}
	}
}
