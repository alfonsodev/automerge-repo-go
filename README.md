# Automerge Repo (Go)

This repository is a work in progress port of the Rust project
`automerge-repo-rs` to Go. The original Rust implementation has
been moved to the `rust/` directory for reference.

The goal of this port is to provide a Go implementation that offers
the same high level API for working with Automerge documents and
network peers. Development is at an early stage and the API should be
considered unstable. A prototype networking handshake is now
available via `repo.Handshake`.

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

