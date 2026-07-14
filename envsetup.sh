#!/usr/bin/env bash
# this file should be sourced (.), not run as a script

GITPLM_BASE=$(readlink -f "$(dirname "${BASH_SOURCE[0]:-$0}")")

# Build the binary with the version stamped in, the same way CI and the release
# build do, so `gitplm version` reports something meaningful.
gitplm_build() {
	local version
	version=$(cd "${GITPLM_BASE}" && git describe --tags --always --dirty 2>/dev/null || echo "Development")
	(cd "${GITPLM_BASE}" && CGO_ENABLED=0 go build -ldflags "-X main.version=${version}" -o "${GITPLM_BASE}/gitplm" .) || return 1
}

gitplm_test() {
	(cd "${GITPLM_BASE}" && go test ./...) || return 1
}

gitplm_vet() {
	(cd "${GITPLM_BASE}" && go vet ./...) || return 1
}

gitplm_format() {
	(cd "${GITPLM_BASE}" && gofmt -s -w .) || return 1
	(cd "${GITPLM_BASE}" && prettier --write "**/*.md") || return 1
}

gitplm_format_check() {
	local unformatted
	unformatted=$(cd "${GITPLM_BASE}" && gofmt -s -l .)
	if [ -n "${unformatted}" ]; then
		echo "Go code is not formatted. Run gitplm_format:"
		echo "${unformatted}"
		return 1
	fi
	(cd "${GITPLM_BASE}" && prettier --check "**/*.md") || return 1
}

# Everything CI runs. Use this before pushing.
gitplm_check() {
	gitplm_test || return 1
	gitplm_build || return 1
	gitplm_format_check || return 1
	gitplm_vet || return 1
	echo "=== all checks passed ==="
}

# --- Releasing ----------------------------------------------------------------
#
# Releases are cut by pushing a tag. The Release workflow
# (.github/workflows/release.yaml) then runs GoReleaser, which builds every
# platform and publishes the CHANGELOG section for that version as the release
# notes.
#
#   1. Land the changes to release on main, each with a CHANGELOG entry under
#      [Unreleased].
#   2. gitplm_prepare_release 0.8.13   # promotes the changelog, commits, tags
#   3. git push origin main && git push origin v0.8.13
#
# Promote the [Unreleased] changelog section to a version, commit it, and create
# the tag. This does not push; it prints the push commands to run.
gitplm_prepare_release() {
	(cd "${GITPLM_BASE}" && ./scripts/prepare-release.sh "$@") || return 1
}

# Preview the release notes GoReleaser would publish for a version.
gitplm_release_notes() {
	(cd "${GITPLM_BASE}" && ./scripts/extract-changelog.sh "${1:-$(git describe --tags --abbrev=0)}") || return 1
}

# Build release artifacts locally without publishing, to check that every
# platform in .goreleaser.yml still compiles. Install goreleaser from
# https://github.com/goreleaser/goreleaser/releases into /usr/local/bin.
gitplm_goreleaser_build() {
	(cd "${GITPLM_BASE}" && goreleaser build --snapshot --clean) || return 1
}

# Regenerate the example releases in example/. Run after changing release
# processing so the checked-in example output stays current.
gitplm_update_examples() {
	gitplm_build || return 1
	local ipn
	for ipn in ASY-001-0000 PCA-019-0000; do
		echo "=== release ${ipn} ==="
		(cd "${GITPLM_BASE}/example" && "${GITPLM_BASE}/gitplm" release "${ipn}") || return 1
	done
}
