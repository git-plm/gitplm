package main

import (
	"log"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/otiai10/copy"
)

// reMarkdownTarget matches the target of a Markdown link or image, such as
// [text](target) or ![alt](<target with spaces> "title"), and the src of an
// HTML <img> tag.
var reMarkdownTarget = regexp.MustCompile(`\]\(\s*(?:<([^>]*)>|([^)\s]+))|(?i:<img\s[^>]*?src\s*=\s*["']([^"']+)["'])`)

// markdownAssetPaths returns the local files a Markdown document links to or
// embeds, as paths relative to the document's directory. Web links, links
// within the page, and absolute paths are left out.
func markdownAssetPaths(content string) []string {
	var paths []string
	seen := make(map[string]bool)

	for _, m := range reMarkdownTarget.FindAllStringSubmatch(content, -1) {
		// only one of the groups matched
		u, err := url.Parse(m[1] + m[2] + m[3])
		if err != nil || u.Scheme != "" || u.Host != "" || u.Path == "" || strings.HasPrefix(u.Path, "/") {
			continue
		}

		p := path.Clean(u.Path)
		if !seen[p] {
			seen[p] = true
			paths = append(paths, p)
		}
	}

	return paths
}

// copyMarkdownAssets copies the local files that the Markdown document name,
// relative to srcDir, links to or embeds into the same relative location under
// destDir, so the links still resolve in the release directory. Links that
// leave srcDir or name a missing file are logged and skipped.
func copyMarkdownAssets(srcDir, destDir, name string) error {
	docPath := path.Join(srcDir, name)
	info, err := os.Stat(docPath)
	if err != nil || info.IsDir() {
		return err
	}

	content, err := os.ReadFile(docPath)
	if err != nil {
		return err
	}

	for _, target := range markdownAssetPaths(string(content)) {
		rel := path.Join(path.Dir(name), target)
		if escapesDir(rel) {
			log.Printf("%v: %v is outside the source directory, not copied\n", name, target)
			continue
		}

		srcPath := path.Join(srcDir, rel)
		// files already in the release directory need no copy
		if containsPath(destDir, srcPath) {
			continue
		}

		info, err := os.Stat(srcPath)
		if err != nil {
			log.Printf("%v: %v not found, not copied\n", name, target)
			continue
		}
		if info.IsDir() {
			continue
		}

		err = copy.Copy(srcPath, path.Join(destDir, rel), copyOptions)
		if err != nil {
			return err
		}

		log.Printf("%v copied to release dir\n", rel)
	}

	return nil
}
