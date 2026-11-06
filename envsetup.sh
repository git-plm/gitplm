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

# Run the version of prettier pinned in .prettier-version, so formatting
# locally reaches the same verdict as CI. A prettier release can add a rule and
# start failing a file that was passing.
gitplm_prettier() {
	local version
	version=$(cat "${GITPLM_BASE}/.prettier-version") || return 1
	(cd "${GITPLM_BASE}" && npx --yes "prettier@${version}" "$@") || return 1
}

gitplm_format() {
	(cd "${GITPLM_BASE}" && gofmt -s -w .) || return 1
	gitplm_prettier --write "**/*.md" || return 1
}

gitplm_format_check() {
	local unformatted
	unformatted=$(cd "${GITPLM_BASE}" && gofmt -s -l .)
	if [ -n "${unformatted}" ]; then
		echo "Go code is not formatted. Run gitplm_format:"
		echo "${unformatted}"
		return 1
	fi
	gitplm_prettier --check "**/*.md" || return 1
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
#   2. Move the [Unreleased] entries under a "## [X.Y.Z] - <date>" heading and
#      commit that to main. gitplm_prepare_release X.Y.Z does this for you and
#      also creates the tag locally; then push main and the tag by hand.
#   3. Otherwise, with the heading committed and pushed, gitplm_tag reads the
#      version from CHANGELOG.md, tags main, and pushes the tag.
#
# Promote the [Unreleased] changelog section to a version, commit it, and create
# the tag. This does not push; it prints the push commands to run.
gitplm_prepare_release() {
	(cd "${GITPLM_BASE}" && ./scripts/prepare-release.sh "$@") || return 1
}

# Read the version to release out of CHANGELOG.md.
#
# The version is the first numbered section. Entries still sitting under
# [Unreleased] mean the release commit that moves them under a version heading
# has not been written yet, so there is nothing to tag.
_gitplm_release_version() {
	if ! grep -q '^## \[Unreleased\]' CHANGELOG.md; then
		echo "no [Unreleased] section in CHANGELOG.md" >&2
		return 1
	fi

	local pending version
	pending=$(sed -n '/^## \[Unreleased\]/,/^## \[[0-9]/p' CHANGELOG.md |
		sed '1d;$d' | tr -d '[:space:]')
	if [ -n "${pending}" ]; then
		echo "CHANGELOG.md has entries under [Unreleased]; give them a version heading first" >&2
		return 1
	fi

	version=$(grep -m1 -oE '^## \[[0-9]+\.[0-9]+\.[0-9]+\]' CHANGELOG.md | tr -d '#[] ')
	if [ -z "${version}" ]; then
		echo "no released version section in CHANGELOG.md" >&2
		return 1
	fi

	printf '%s\n' "${version}"
}

# Everything gitplm_tag does while sitting on main. Split out so the caller
# restores the original branch on every exit path, success or not.
_gitplm_tag_on_main() {
	git merge --ff-only origin/main || return 1

	# A release that does not build is worse than a late one.
	gitplm_test || return 1

	local tag
	tag="v$(_gitplm_release_version)" || return 1

	if git rev-parse -q --verify "refs/tags/${tag}" >/dev/null; then
		echo "tag ${tag} already exists locally" >&2
		return 1
	fi
	if git ls-remote --exit-code --tags origin "refs/tags/${tag}" >/dev/null 2>&1; then
		echo "tag ${tag} already exists on origin" >&2
		return 1
	fi

	git tag -a "${tag}" -m "${tag}" || return 1
	if ! git push origin "${tag}"; then
		git tag -d "${tag}"
		return 1
	fi

	echo "=== tagged and pushed ${tag} ==="
}

# Tag the version named at the top of CHANGELOG.md on main and push the tag.
gitplm_tag() {
	(
		cd "${GITPLM_BASE}" || exit 1

		if ! (git diff --quiet && git diff --cached --quiet); then
			echo "gitplm_tag: uncommitted changes; commit or stash them first" >&2
			exit 1
		fi

		local branch
		branch=$(git symbolic-ref --quiet --short HEAD)
		if [ -z "${branch}" ]; then
			echo "gitplm_tag: HEAD is detached; check out a branch first" >&2
			exit 1
		fi

		# Not --tags: a stale local tag that disagrees with origin makes git exit
		# nonzero even though the fetch did its job. Nothing here reads local tag
		# refs for the origin-side check, so only origin/main needs updating.
		git fetch origin || exit 1
		git checkout main || exit 1

		_gitplm_tag_on_main
		local rc=$?

		git checkout "${branch}" || exit 1
		exit ${rc}
	) || return 1
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
