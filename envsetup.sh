# download goreleaser from https://github.com/goreleaser/goreleaser/releases/
# and put in /usr/local/bin
# This can be useful to test/debug the release process locally
gitplm_goreleaser_build() {
  goreleaser build --skip-validate --rm-dist
}

# before releasing, you need to tag the release
# you need to provide GITHUB_TOKEN in env or ~/.config/goreleaser/github_token
# generate tokens: https://github.com/settings/tokens/new
# enable repo and workflow sections
gitplm_goreleaser_release() {
  goreleaser release --rm-dist
}
