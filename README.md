# Automerge Repo (Go)

This repository is a work in progress port of the Rust project
`automerge-repo-rs` to Go. The original Rust implementation has
been moved to the `rust/` directory for reference.

The goal of this port is to provide a Go implementation that offers
the same high level API for working with Automerge documents and
network peers. Development is at an early stage and the API should be
considered unstable. A prototype networking handshake using CBOR
encoding is now available via `repo.Handshake`. The handshake helpers
now accept a `context.Context` which can be used to apply timeouts or
cancel the operation.

Basic document persistence is available using `repo.FsStore` and documents
internally use the [automerge-go](https://github.com/automerge/automerge-go)
library. Documents can be accessed through `DocumentHandle` which provides
a simple change notification API. When a repository has a store configured,
mutating a `DocumentHandle` will automatically save the updated document to
disk. Repositories may also be configured with a `SharePolicy` to control
which documents are synchronised with particular peers. Policies can also
decide whether documents should be announced to or requested from a peer.
The program under
`cmd/example` provides a small CLI for creating and editing documents stored on
disk.

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
browsers. `DialWebSocket` also requires a `context.Context` for applying
timeouts.

### Echo Web Framework Integration

For easy integration with the [Echo](https://echo.labstack.com/) web framework, you can use the provided handler in the `adapters/echo` package. This handler upgrades HTTP connections to WebSockets and connects them to the Automerge repo.

Here is a complete example of a simple Echo server that uses the handler:

```go
package main

import (
	"context"
	"log"

	"github.com/automerge/automerge-repo-go/adapters/echo"
	"github.com/automerge/automerge-repo-go/repo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Create a new Automerge repo.
	r := repo.New()

	// Create a handle for the repo.
	handle := repo.NewRepoHandle(r)

	// Run the repo in the background.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-handle.Inbox:
				// In a real application, you would handle incoming messages here.
				log.Printf("received message: %+v", msg)
			}
		}
	}()

	// Create a new Echo instance.
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Add the Automerge repo WebSocket handler.
	e.GET("/ws", echo.AutomergeRepoHandler(handle))

	// Start the server.
	e.Logger.Fatal(e.Start(":1323"))
}
```

#### Runnable Demo

A complete, runnable demo of this integration with a vanilla JavaScript frontend can be found in the `examples/echo-vanilla-js` directory. To run it:

1.  Navigate to the example directory:
    ```bash
    cd examples/echo-vanilla-js
    ```
2.  Run the Go server:
    ```bash
    go run main.go
    ```
3.  Open your browser to [http://localhost:1323](http://localhost:1323).

You can open the page in multiple browser tabs or windows to see changes sync in real-time.

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


## Documentation

The full documentation for the `repo` package is available on [pkg.go.dev](https://pkg.go.dev/github.com/automerge/automerge-repo-go/repo).

You can also view the documentation locally using the `godoc` tool.
If you don't have `godoc` installed, you can get it by running:
```bash
go install golang.org/x/tools/cmd/godoc@latest
```
*Note: Ensure that your Go bin directory (e.g., `$GOPATH/bin` or `$HOME/go/bin`) is in your system's `PATH`.*

Once installed, run the following command in the root of the repository to start the documentation server:
```bash
godoc -http=:6060
```

Then, open your browser to [http://localhost:6060/pkg/github.com/automerge/automerge-repo-go/repo](http://localhost:6060/pkg/github.com/automerge/automerge-repo-go/repo).

## Release process

Releases are tagged using semantic versions. The `scripts/release.sh` helper
builds binaries for the example programs across supported platforms and
packages them into a tarball. To cut a new version run:

```bash
./scripts/release.sh v0.1.0
```

The script creates and pushes the git tag then writes the build artifacts to
`dist/`. Upload the resulting tarball when creating a GitHub release.
