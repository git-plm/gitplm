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

gitplm_update_examples() {
  for bom in ASY-012-0012 ASY-002-0001 PCA-019-0000 ASY-001-0000; do
    go run . -bom $bom || return
  done
}
