#!/usr/bin/env bash
set -euo pipefail

if [ $# -ne 1 ]; then
  echo "usage: $0 <version>" >&2
  exit 1
fi

VERSION=$1

echo "Tagging release $VERSION"

git tag -a "$VERSION" -m "Release $VERSION"

git push origin "$VERSION"

mkdir -p dist
for dir in cmd/*; do
  name=$(basename "$dir")
  echo "Building $name for linux-amd64"
  GOOS=linux GOARCH=amd64 go build -o dist/${name}-linux-amd64 "$dir"
  echo "Building $name for darwin-amd64"
  GOOS=darwin GOARCH=amd64 go build -o dist/${name}-darwin-amd64 "$dir"
  echo "Building $name for windows-amd64"
  GOOS=windows GOARCH=amd64 go build -o dist/${name}-windows-amd64.exe "$dir"
  echo "" >/dev/null
done

TARBALL="automerge-repo-go-${VERSION}.tar.gz"
tar -czf "$TARBALL" -C dist .

echo "Release artifacts written to $TARBALL"
