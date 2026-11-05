package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/gocarina/gocsv"
	"gopkg.in/yaml.v2"
)

var modFile = `
# yaml file
description: modify bom
remove:
 - cmpName: Test point 2
 - ref: R11
 - ref: D13 D14
 - ref: R1,R2
add:
 - cmpName: "screw #4 2"
   ref: S3
   ipn: SCR-002-0002
 - cmpName: "led green"
   ref: D22 D20 D21
   ipn: DIO-033-000G
 - cmpName: "led blue"
   ref: D32,D30,D31
   ipn: DIO-033-000B
 - ipn: PCB-009-0013
`

var bomIn = `
Ref,Qty,Value,Cmp name,Footprint,Description,Vendor,IPN,Datasheet
TP4 TP5,2,,Test point 2,,,,,
R1 R2,2,,100K_100mw,,,,RES-006-0232,
D1 D2 D13 D14,4,,diode,,,,DIO-023-0023,
"R11","1","2010_500mW_1%_3000V_10M","2010_500mW_1%_3000V_10M","Resistor_SMD:R_2010_5025Metric","","","RES-008-1005","https://www.bourns.com/docs/Product-Datasheets/CHV.pdf"
`

var bomExp = `
Ref,Qty,Value,Cmp name,Footprint,Description,Vendor,IPN,Datasheet
D1 D2,2,,diode,,,,DIO-023-0023,
D30 D31 D32,3,,led blue,,,,DIO-033-000B,
D20 D21 D22,3,,led green,,,,DIO-033-000G,
,1,,,,,,PCB-009-0013,
S3,1,,screw #4 2,,,,SCR-002-0002,
`

func TestRelScript(t *testing.T) {
	initCSV()
	bIn := bom{}
	err := gocsv.UnmarshalBytes([]byte(bomIn), &bIn)
	if err != nil {
		t.Errorf("error parsing bomIn: %v", err)
	}

	bExp := bom{}
	err = gocsv.UnmarshalBytes([]byte(bomExp), &bExp)
	if err != nil {
		t.Errorf("error parsing bomExp: %v", err)
	}

	rs := relScript{}

	err = yaml.Unmarshal([]byte(modFile), &rs)
	if err != nil {
		t.Errorf("error parsing yaml: %v", err)
	}

	bModified, err := rs.processBom(bIn)
	if err != nil {
		t.Errorf("error processing bom: %v", err)
	}

	if reflect.DeepEqual(bExp, bModified) != true {
		fmt.Printf("bExp: %v", bExp)
		fmt.Printf("bModified: %v", bModified)
		t.Error("bExp not the same as bModified")
	}
}

// setupReleaseTree writes a minimal PCA source tree into a fresh directory
// and makes it the working directory for the rest of the test. It returns
// the partmaster directory.
func setupReleaseTree(t *testing.T, relScript string) string {
	t.Helper()
	initCSV()

	root := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	pmDir := filepath.Join(root, "partmaster")
	if err := os.Mkdir(pmDir, 0755); err != nil {
		t.Fatal(err)
	}

	files := map[string]string{
		"partmaster/res.csv": "IPN,Description,Footprint,Value,Manufacturer,MPN,Datasheet,Priority,Checked\n" +
			"RES-0000-1002,10K 0603,R_0603,10K,Yageo,RC0603FR-0710KL,,,\n",
		"PCA-001.csv": "Ref,Qty,Value,Cmp name,Footprint,Description,Vendor,IPN,Datasheet\n" +
			"R1 R2,2,10K,res,,,,RES-0000-1002,\n",
		"PCA-001.yml":  relScript,
		"CHANGELOG.md": "## [PCA-001-0001]\n\n- first release\n",
	}
	for name, content := range files {
		if err := os.WriteFile(name, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	return pmDir
}

// TestPostHooksSeeReleaseBOM checks that a post hook runs after the release
// BOM is written, so it can read the merged partmaster data, and that a
// file it generates satisfies a required entry.
func TestPostHooksSeeReleaseBOM(t *testing.T) {
	script := `
hooks:
  - test ! -e {{ .RelDir }}/{{ .IPN }}.csv
postHooks:
  - cp {{ .RelDir }}/{{ .IPN }}.csv {{ .RelDir }}/{{ .IPN }}-bom.txt
required:
  - PCA-001-0001-bom.txt
`
	pmDir := setupReleaseTree(t, script)

	var relLog strings.Builder
	_, err := processRelease("PCA-001-0001", &relLog, pmDir, io.Discard)
	if err != nil {
		t.Fatalf("processRelease: %v", err)
	}

	data, err := os.ReadFile(filepath.Join("PCA-001-0001", "PCA-001-0001-bom.txt"))
	if err != nil {
		t.Fatalf("post hook output missing: %v", err)
	}

	if !strings.Contains(string(data), "RC0603FR-0710KL") {
		t.Errorf("post hook did not see the merged release BOM:\n%s", data)
	}
}

// TestRequiredCheckedAfterPostHooks checks that a required file missing
// after the post hooks fails the release.
func TestRequiredCheckedAfterPostHooks(t *testing.T) {
	script := `
postHooks:
  - "true"
required:
  - never-generated.txt
`
	pmDir := setupReleaseTree(t, script)

	var relLog strings.Builder
	_, err := processRelease("PCA-001-0001", &relLog, pmDir, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "never-generated.txt") {
		t.Fatalf("expected missing required file error, got: %v", err)
	}
}

// TestPostHookFailureStopsRelease checks that a failing post hook is
// reported as an error.
func TestPostHookFailureStopsRelease(t *testing.T) {
	script := `
postHooks:
  - exit 3
`
	pmDir := setupReleaseTree(t, script)

	var relLog strings.Builder
	_, err := processRelease("PCA-001-0001", &relLog, pmDir, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "postHooks") {
		t.Fatalf("expected postHooks error, got: %v", err)
	}
}

// TestReleaseWithoutBOMCopiesAssets checks that a part with no BOM, such as a
// PCB, gets MFG.md and CHANGELOG.md in its release directory, in time to
// satisfy a required entry.
func TestReleaseWithoutBOMCopiesAssets(t *testing.T) {
	pmDir := setupReleaseTree(t, "")

	files := map[string]string{
		"PCB-001.yml":  "required:\n  - MFG.md\n  - CHANGELOG.md\n",
		"MFG.md":       "manufacturing notes\n",
		"CHANGELOG.md": "## [PCB-001-0001]\n\n- first release\n",
	}
	for name, content := range files {
		if err := os.WriteFile(name, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	var relLog strings.Builder
	_, err := processRelease("PCB-001-0001", &relLog, pmDir, io.Discard)
	if err != nil {
		t.Fatalf("processRelease: %v", err)
	}

	for _, name := range []string{"MFG.md", "CHANGELOG.md"} {
		if _, err := os.Stat(filepath.Join("PCB-001-0001", name)); err != nil {
			t.Errorf("%v not copied into the release directory: %v", name, err)
		}
	}
}

// writeFiles creates each file under dir, making parent directories as needed.
func writeFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

// TestCopyGlob checks that a copy entry can be a glob pattern, that * leaves
// the release directory out of its own copy, and that a pattern matching
// nothing is an error.
func TestCopyGlob(t *testing.T) {
	src := t.TempDir()
	dest := filepath.Join(src, "PCA-001-0001")
	writeFiles(t, src, map[string]string{
		"docs/a.pdf":           "a",
		"docs/b.pdf":           "b",
		"docs/c.txt":           "c",
		"PCA-001-0001/old.txt": "old",
	})

	rs := relScript{Copy: []string{"docs/*.pdf"}}
	if err := rs.copy(src, dest); err != nil {
		t.Fatalf("copy: %v", err)
	}
	for name, want := range map[string]bool{"docs/a.pdf": true, "docs/b.pdf": true, "docs/c.txt": false} {
		if got := fileExists(filepath.Join(dest, name)); got != want {
			t.Errorf("%v in release directory: got %v, want %v", name, got, want)
		}
	}

	rs = relScript{Copy: []string{"*"}}
	if err := rs.copy(src, dest); err != nil {
		t.Fatalf("copy *: %v", err)
	}
	if fileExists(filepath.Join(dest, "PCA-001-0001")) {
		t.Error("* copied the release directory into itself")
	}
	if !fileExists(filepath.Join(dest, "docs", "c.txt")) {
		t.Error("* did not copy the docs directory")
	}

	rs = relScript{Copy: []string{"*.step"}}
	if err := rs.copy(src, dest); err == nil || !strings.Contains(err.Error(), "*.step") {
		t.Errorf("expected no match error, got: %v", err)
	}
}

// TestMove checks that move takes files out of the source directory, replaces
// a stale file in the release directory, fails when a file was not generated
// again, and refuses a path outside the source directory.
func TestMove(t *testing.T) {
	src := t.TempDir()
	dest := filepath.Join(src, "PCB-001-0001")
	writeFiles(t, src, map[string]string{
		"fab/drill.pdf":              "new drill",
		"fab/panel.pdf":              "new panel",
		"notes.txt":                  "notes",
		"PCB-001-0001/fab/drill.pdf": "stale drill",
	})

	rs := relScript{Move: []string{"fab/*.pdf", "notes.txt"}}
	if err := rs.move(src, dest); err != nil {
		t.Fatalf("move: %v", err)
	}

	for _, name := range []string{"fab/drill.pdf", "fab/panel.pdf", "notes.txt"} {
		if fileExists(filepath.Join(src, name)) {
			t.Errorf("%v is still in the source directory", name)
		}
		if !fileExists(filepath.Join(dest, name)) {
			t.Errorf("%v is not in the release directory", name)
		}
	}

	data, err := os.ReadFile(filepath.Join(dest, "fab", "drill.pdf"))
	if err != nil || string(data) != "new drill" {
		t.Errorf("stale drill.pdf not replaced: %q, %v", data, err)
	}

	if err := rs.move(src, dest); err == nil || !strings.Contains(err.Error(), "please generate") {
		t.Errorf("expected an error moving files that were not generated again, got: %v", err)
	}

	rs = relScript{Move: []string{"notes.txt"}}
	if err := rs.move(src, dest); err == nil || !strings.Contains(err.Error(), "notes.txt") {
		t.Errorf("expected an error moving a missing file, got: %v", err)
	}

	rs = relScript{Move: []string{"../outside.txt"}}
	if err := rs.move(src, dest); err == nil || !strings.Contains(err.Error(), "outside") {
		t.Errorf("expected an error moving from outside the source directory, got: %v", err)
	}
}

// TestRequiredGlob checks that a required entry can be a glob pattern.
func TestRequiredGlob(t *testing.T) {
	dest := t.TempDir()
	writeFiles(t, dest, map[string]string{"PCA-001-0001_ibom.html": "html"})

	rs := relScript{Required: []string{"*_ibom.html"}}
	if err := rs.required(dest); err != nil {
		t.Errorf("required *_ibom.html: %v", err)
	}

	rs = relScript{Required: []string{"*.pdf"}}
	if err := rs.required(dest); err == nil || !strings.Contains(err.Error(), "*.pdf") {
		t.Errorf("expected missing required file error, got: %v", err)
	}
}

// TestCopyMarkdownAssets checks that copying a Markdown file brings along the
// files it links to, resolved from the file's own directory.
func TestCopyMarkdownAssets(t *testing.T) {
	src := t.TempDir()
	dest := filepath.Join(src, "ASY-001-0001")
	writeFiles(t, src, map[string]string{
		"docs/assembly.md": "![step 1](img/step1.png)\n" +
			"[spec](../spec.pdf)\n" +
			"[missing](img/missing.png)\n" +
			"[outside](../../outside.pdf)\n",
		"docs/img/step1.png": "png",
		"spec.pdf":           "pdf",
	})

	rs := relScript{Copy: []string{"docs/assembly.md"}}
	if err := rs.copy(src, dest); err != nil {
		t.Fatalf("copy: %v", err)
	}

	for _, name := range []string{"docs/assembly.md", "docs/img/step1.png", "spec.pdf"} {
		if !fileExists(filepath.Join(dest, name)) {
			t.Errorf("%v not copied into the release directory", name)
		}
	}
}

// TestReleaseCopiesMFGAssets checks that the images MFG.md embeds are copied
// into the release directory with it.
func TestReleaseCopiesMFGAssets(t *testing.T) {
	pmDir := setupReleaseTree(t, "")
	writeFiles(t, ".", map[string]string{
		"MFG.md":           "![panel](images/panel.png)\n",
		"images/panel.png": "png",
	})

	var relLog strings.Builder
	_, err := processRelease("PCA-001-0001", &relLog, pmDir, io.Discard)
	if err != nil {
		t.Fatalf("processRelease: %v", err)
	}

	if !fileExists(filepath.Join("PCA-001-0001", "images", "panel.png")) {
		t.Error("image embedded in MFG.md not copied into the release directory")
	}
}

// TestExistingReleaseDir checks that the release directory is reported only
// once it exists.
func TestExistingReleaseDir(t *testing.T) {
	setupReleaseTree(t, "")

	dir, err := existingReleaseDir("PCA-001-0001")
	if err != nil || dir != "" {
		t.Fatalf("before release: got %q, %v", dir, err)
	}

	if err := os.Mkdir("PCA-001-0001", 0755); err != nil {
		t.Fatal(err)
	}

	dir, err = existingReleaseDir("PCA-001-0001")
	if err != nil || dir != "PCA-001-0001" {
		t.Errorf("after release: got %q, %v", dir, err)
	}
}
