package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestReleaseLogWritten checks that a release writes a log file in the
// release directory holding the messages logged during the run, the hook
// output, and the final result.
func TestReleaseLogWritten(t *testing.T) {
	script := `
hooks:
  - echo "hook says hello"
`
	pmDir := setupReleaseTree(t, script)

	// The BOM references a part missing from the partmaster so the run logs
	// an error line through the log package.
	bomCSV := "Ref,Qty,Value,Cmp name,Footprint,Description,Vendor,IPN,Datasheet\n" +
		"R1 R2,2,10K,res,,,,RES-0000-1002,\n" +
		"C1,1,1uF,cap,,,,CAP-0000-0001,\n"
	if err := os.WriteFile("PCA-001.csv", []byte(bomCSV), 0644); err != nil {
		t.Fatal(err)
	}

	origOut := log.Writer()
	log.SetOutput(io.Discard)
	t.Cleanup(func() { log.SetOutput(origOut) })

	logPath, err := processRelease("PCA-001-0001", pmDir, io.Discard)
	if err != nil {
		t.Fatalf("processRelease: %v", err)
	}

	if log.Writer() != io.Discard {
		t.Error("processRelease did not restore the log output")
	}

	if want := filepath.Join("PCA-001-0001", "PCA-001-0001.log"); logPath != want {
		t.Errorf("log path = %q, want %q", logPath, want)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("release log missing: %v", err)
	}

	got := string(data)
	for _, want := range []string{
		"release PCA-001-0001 started",
		"hook says hello",
		"CAP-0000-0001",
		"release PCA-001-0001 updated",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("release log missing %q:\n%s", want, got)
		}
	}
}

// TestReleaseLogRecordsError checks that a failed release still leaves a
// log file that names the error.
func TestReleaseLogRecordsError(t *testing.T) {
	script := `
required:
  - never-generated.txt
`
	pmDir := setupReleaseTree(t, script)

	origOut := log.Writer()
	log.SetOutput(io.Discard)
	t.Cleanup(func() { log.SetOutput(origOut) })

	_, err := processRelease("PCA-001-0001", pmDir, io.Discard)
	if err == nil {
		t.Fatal("expected missing required file error")
	}

	data, err := os.ReadFile(filepath.Join("PCA-001-0001", "PCA-001-0001.log"))
	if err != nil {
		t.Fatalf("release log missing: %v", err)
	}

	if !strings.Contains(string(data), "release error:") ||
		!strings.Contains(string(data), "never-generated.txt") {
		t.Errorf("release log does not record the error:\n%s", data)
	}
}

// TestFailedHookIsFramed checks that the script of a failing hook is set
// apart from the output it printed.
func TestFailedHookIsFramed(t *testing.T) {
	script := `
hooks:
  - |
    echo "something went wrong"
    exit 1
`
	pmDir := setupReleaseTree(t, script)

	origOut := log.Writer()
	log.SetOutput(io.Discard)
	t.Cleanup(func() { log.SetOutput(origOut) })

	var out strings.Builder
	_, err := processRelease("PCA-001-0001", pmDir, &out)
	if err == nil {
		t.Fatal("expected hook error")
	}

	got := out.String()
	want := "something went wrong\n\n" + hookRule + "\nHook failed (exit status 1). The output above came from this hook:\n\n" +
		"echo \"something went wrong\"\nexit 1\n" + hookRule + "\n"
	if got != want {
		t.Errorf("hook failure output:\n%s\nwant:\n%s", got, want)
	}
}

// TestReleaseLogSurvivesHookCleaningDir checks that a hook which empties the
// release directory does not lose the log.
func TestReleaseLogSurvivesHookCleaningDir(t *testing.T) {
	script := `
hooks:
  - |
    rm -rf {{ .RelDir }}/*
    echo "cleaned"
`
	pmDir := setupReleaseTree(t, script)

	origOut := log.Writer()
	log.SetOutput(io.Discard)
	t.Cleanup(func() { log.SetOutput(origOut) })

	logPath, err := processRelease("PCA-001-0001", pmDir, io.Discard)
	if err != nil {
		t.Fatalf("processRelease: %v", err)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("release log missing: %v", err)
	}
	if !strings.Contains(string(data), "cleaned") {
		t.Errorf("release log missing hook output:\n%s", data)
	}
}
