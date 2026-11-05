package main

import (
	"reflect"
	"testing"
)

func TestMarkdownAssetPaths(t *testing.T) {
	content := "# Assembly\n\n" +
		"![top](images/top.png)\n" +
		"![bottom](./images/bottom.png \"Bottom side\")\n" +
		"See the [drawing](drawings/enclosure.pdf#page=2) and [notes](<test notes.txt>).\n" +
		"<IMG width=\"200\" src=\"images/detail.jpg\">\n" +
		"![top again](images/top.png)\n" +
		"[site](https://example.com/x.png), [mail](mailto:a@example.com), [section](#assembly), [root](/etc/hosts)\n" +
		"![harness](images/wire%20harness.png)\n"

	want := []string{
		"images/top.png",
		"images/bottom.png",
		"drawings/enclosure.pdf",
		"test notes.txt",
		"images/detail.jpg",
		"images/wire harness.png",
	}

	got := markdownAssetPaths(content)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("markdownAssetPaths:\n got %q\nwant %q", got, want)
	}
}
