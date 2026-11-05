package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatCSVFile(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		changed bool
	}{
		{
			name:    "already formatted",
			in:      "IPN,Description\nRES-001-0001,10k\n",
			want:    "IPN,Description\nRES-001-0001,10k\n",
			changed: false,
		},
		{
			name:    "drops needless quotes and pads short rows",
			in:      "\"IPN\",\"Description\",\"Notes\"\n\"RES-001-0001\",\"10k\"\n",
			want:    "IPN,Description,Notes\nRES-001-0001,10k\n",
			changed: true,
		},
		{
			name:    "keeps quotes around commas and quotes",
			in:      "IPN,Description\nRES-001-0001,\"10k, 1%\"\nRES-001-0002,\"1\"\"\"\n",
			want:    "IPN,Description\nRES-001-0001,\"10k, 1%\"\nRES-001-0002,\"1\"\"\"\n",
			changed: false,
		},
		{
			name:    "trims header whitespace and adds final newline",
			in:      " IPN , Description\nRES-001-0001,10k",
			want:    "IPN,Description\nRES-001-0001,10k\n",
			changed: true,
		},
		{
			name:    "keeps a stray quote by quoting the field",
			in:      "IPN,Description\nRES-001-0001,1/4\" wide\n",
			want:    "IPN,Description\nRES-001-0001,\"1/4\"\" wide\"\n",
			changed: true,
		},
		{
			name:    "sorts rows by IPN and keeps duplicate order",
			in:      "IPN,Priority\nRES-001-0002,1\nCAP-001-0001,2\nCAP-001-0001,1\n",
			want:    "IPN,Priority\nCAP-001-0001,2\nCAP-001-0001,1\nRES-001-0002,1\n",
			changed: true,
		},
		{
			name:    "keeps row order without an IPN column",
			in:      "Ref,Value\nR2,10k\nR1,1k\n",
			want:    "Ref,Value\nR2,10k\nR1,1k\n",
			changed: false,
		},
		{
			name:    "preserves CRLF line endings",
			in:      "\"IPN\",Description\r\nRES-001-0001,10k\r\n",
			want:    "IPN,Description\r\nRES-001-0001,10k\r\n",
			changed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "parts.csv")
			if err := os.WriteFile(path, []byte(tt.in), 0644); err != nil {
				t.Fatal(err)
			}
			changed, err := formatCSVFile(path)
			if err != nil {
				t.Fatalf("formatCSVFile: %v", err)
			}
			if changed != tt.changed {
				t.Errorf("changed = %v, want %v", changed, tt.changed)
			}
			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatCSVPaths(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		filepath.Join(dir, "a.csv"):     "\"IPN\"\nRES-001-0001\n",
		filepath.Join(sub, "b.CSV"):     "IPN\nRES-001-0002\n",
		filepath.Join(sub, "c.csv"):     "\"IPN\"\nRES-001-0003\n",
		filepath.Join(sub, "notes.txt"): "\"not a csv\"\n",
	}
	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	var out bytes.Buffer
	if err := formatCSVPaths([]string{dir}, &out); err != nil {
		t.Fatalf("formatCSVPaths: %v", err)
	}

	changed := strings.Split(strings.TrimSpace(out.String()), "\n")
	want := []string{filepath.Join(dir, "a.csv"), filepath.Join(sub, "c.csv")}
	if strings.Join(changed, ",") != strings.Join(want, ",") {
		t.Errorf("changed files = %v, want %v", changed, want)
	}
	got, _ := os.ReadFile(filepath.Join(sub, "notes.txt"))
	if string(got) != files[filepath.Join(sub, "notes.txt")] {
		t.Error("non-CSV file was modified")
	}

	err := formatCSVPaths([]string{filepath.Join(dir, "missing.csv")}, &out)
	if err == nil {
		t.Error("expected an error for a missing file")
	}
}
