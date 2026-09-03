#!/bin/bash
# Extract the changelog section for a specific version from CHANGELOG.md.
# Usage: extract-changelog.sh <version>
#
# The release workflow feeds this script's output to GoReleaser as the release
# notes, so a missing section is treated as an error rather than silently
# producing an empty release body.

set -euo pipefail

VERSION=${1:-}

if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version>" >&2
    exit 1
fi

# Remove 'v' prefix if present
VERSION=${VERSION#v}

# Extract the section for this version. Version headings are written as
# "## [[0.8.9] - 2026-03-02](https://.../releases/tag/v0.8.9)", so match on the
# bracketed version and stop at the next "## [" heading.
NOTES=$(awk -v version="$VERSION" '
/^## \[/ {
    if (found) exit
    if ($0 ~ "\\[" version "\\]") {
        found=1
        next
    }
}
found && /^## \[/ { exit }
found { print }
' CHANGELOG.md)

if [ -z "${NOTES//[[:space:]]/}" ]; then
    cat >&2 <<EOF
error: CHANGELOG.md has no entry for version ${VERSION}.

Add a "## [[${VERSION}] - YYYY-MM-DD](...)" section describing this release
before tagging. scripts/prepare-release.sh promotes the [Unreleased] section
for you.
EOF
    exit 1
fi

echo "$NOTES"
