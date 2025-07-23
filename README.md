# Automerge Repo (Go)

This repository is a work in progress port of the Rust project
`automerge-repo-rs` to Go. The original Rust implementation has
been moved to the `rust/` directory for reference.

The goal of this port is to provide a Go implementation that offers
the same high level API for working with Automerge documents and
network peers. Development is at an early stage and the API should be
considered unstable. A prototype networking handshake using CBOR
encoding is now available via `repo.Handshake`.

Basic document persistence is available using `repo.FsStore` and documents
internally use the [automerge-go](https://github.com/automerge/automerge-go)
library. The program under `cmd/example` provides a small CLI for creating and
editing documents stored on disk.

Run it with:

```bash
go run ./cmd/example <command>
```

Available commands are:

* `new` - create a new document and print its ID
* `list` - list stored document IDs
* `set <id> <key> <value>` - set a key/value pair in a document
* `get <id> <key>` - print a value from a document

### Networking example

The program under `cmd/tcp-example` demonstrates establishing a networking
handshake between peers and synchronising a document. Run one instance in
server mode:

```bash
go run ./cmd/tcp-example -listen :9999
```

And another in client mode connecting to it:

```bash
go run ./cmd/tcp-example -connect localhost:9999
```

Each side prints the remote repository ID once the handshake completes. After
connecting you can issue `set <key> <value>` commands on either side and the
changes will be synced to all peers.

Messages and handshake data are encoded using CBOR for compatibility with
other Automerge Repo implementations.

WebSocket connections are supported via `repo.DialWebSocket` and
`repo.AcceptWebSocket`. They use the same join/peer handshake over a WebSocket
upgrade so repositories can communicate through standard HTTP servers or
browsers.

## Building

This project uses Go modules. To download dependencies and build run:

```bash
go build ./...
```

## Testing

Tests can be executed with:

```bash
go test ./...
```

## Continuous Integration

All pushes and pull requests are validated by a GitHub Actions workflow defined
in `.github/workflows/go.yml`. The workflow runs `go vet` and `go test` on
Linux, macOS and Windows using the latest Go releases.


## Release process

Releases are tagged using semantic versions. The `scripts/release.sh` helper
builds binaries for the example programs across supported platforms and
packages them into a tarball. To cut a new version run:

```bash
./scripts/release.sh v0.1.0
```

The script creates and pushes the git tag then writes the build artifacts to
`dist/`. Upload the resulting tarball when creating a GitHub release.
