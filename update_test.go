package main

import "testing"

// The expected names are the assets actually published for v0.8.12. If this
// test fails, `gitplm update` is downloading a URL that does not exist, so keep
// it in step with the archives name_template in .goreleaser.yml.
func TestBinaryName(t *testing.T) {
	tests := []struct {
		goos   string
		goarch string
		goarm  string
		want   string
	}{
		{"linux", "amd64", "", "gitplm-v0.8.12-linux-x86_64"},
		{"linux", "386", "", "gitplm-v0.8.12-linux-i386"},
		{"linux", "arm64", "", "gitplm-v0.8.12-linux-arm64"},
		{"linux", "arm", "6", "gitplm-v0.8.12-linux-arm6"},
		{"linux", "arm", "7", "gitplm-v0.8.12-linux-arm7"},
		{"darwin", "amd64", "", "gitplm-v0.8.12-macos-x86_64"},
		{"darwin", "arm64", "", "gitplm-v0.8.12-macos-arm64"},
		{"windows", "amd64", "", "gitplm-v0.8.12-windows-x86_64.exe"},
		{"windows", "386", "", "gitplm-v0.8.12-windows-i386.exe"},
		{"windows", "arm64", "", "gitplm-v0.8.12-windows-arm64.exe"},
	}

	for _, test := range tests {
		got := binaryName("v0.8.12", test.goos, test.goarch, test.goarm)
		if got != test.want {
			t.Errorf("binaryName(v0.8.12, %v, %v, %v) = %v; want %v",
				test.goos, test.goarch, test.goarm, got, test.want)
		}
	}
}
