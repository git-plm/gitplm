#!/bin/bash
# Prepare a release: promote the CHANGELOG [Unreleased] section to a versioned
# section, commit it, and create the release tag.
#
# Usage: scripts/prepare-release.sh <version>     e.g. scripts/prepare-release.sh 0.8.13
#
# Pushing is left to you. Once the tag is pushed, the Release workflow builds
# the binaries with GoReleaser and publishes them with the changelog section as
# the release notes.

set -euo pipefail

REPO_URL="https://github.com/git-plm/gitplm"

VERSION=${1:-}

if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version>   (e.g. $0 0.8.13)" >&2
    exit 1
fi

# Accept either 0.8.13 or v0.8.13; normalize to both forms.
VERSION=${VERSION#v}
TAG="v${VERSION}"

if ! [[ "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "error: version must look like 1.2.3 (got '${VERSION}')" >&2
    exit 1
fi

cd "$(git rev-parse --show-toplevel)"

if [ -n "$(git status --porcelain)" ]; then
    echo "error: working tree has uncommitted changes. Commit or stash them first." >&2
    git status --short >&2
    exit 1
fi

if git rev-parse -q --verify "refs/tags/${TAG}" >/dev/null; then
    echo "error: tag ${TAG} already exists." >&2
    exit 1
fi

if grep -q "\[${VERSION}\]" CHANGELOG.md; then
    echo "error: CHANGELOG.md already has a section for ${VERSION}." >&2
    exit 1
fi

# The [Unreleased] section must describe something, otherwise the release would
# ship with empty notes.
UNRELEASED=$(awk '
/^## \[Unreleased\]/ { found=1; next }
found && /^## \[/ { exit }
found { print }
' CHANGELOG.md)

if [ -z "${UNRELEASED//[[:space:]]/}" ]; then
    echo "error: CHANGELOG.md [Unreleased] section is empty; nothing to release." >&2
    exit 1
fi

DATE=$(date +%Y-%m-%d)
HEADING="## [[${VERSION}] - ${DATE}](${REPO_URL}/releases/tag/${TAG})"

echo "Preparing ${TAG} with these notes:"
echo "$UNRELEASED"
echo

# Promote [Unreleased]: leave the heading in place for the next cycle and insert
# the new version heading directly above the entries it now covers.
awk -v heading="$HEADING" '
/^## \[Unreleased\]/ && !done {
    print
    print ""
    print heading
    done=1
    next
}
{ print }
' CHANGELOG.md > CHANGELOG.md.tmp
mv CHANGELOG.md.tmp CHANGELOG.md

# Keep the changelog conformant with the Markdown format check in CI.
if command -v prettier >/dev/null 2>&1; then
    prettier --write CHANGELOG.md >/dev/null
fi

# A release that does not build is worse than a late one.
go test ./... >/dev/null

git add CHANGELOG.md
git commit -q -m "${TAG}"
git tag -a "${TAG}" -m "${TAG}"

echo "Committed CHANGELOG and tagged ${TAG}."
echo
echo "Review with:  git show ${TAG}"
echo "Then publish: git push origin main && git push origin ${TAG}"
