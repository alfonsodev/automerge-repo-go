# Guidelines for Codex agents

This repository contains a work-in-progress port of the Automerge Repo from Rust to Go. The original Rust code now lives under `./rust` and should be used for reference when adding features in Go.

## Next steps

1. Flesh out the Go implementation so it matches the capabilities of the Rust repo. Basic networking handshake prototyped, more work needed (see `rust/src/*.rs`).
2. Expand the Go example under `cmd/example` into a minimal CLI demonstrating document creation, loading and storage using `repo.FsStore`. **(done)**
3. Create unit tests for the Go code. Aim for similar coverage as the Rust tests in `rust/tests`. **(done)**
4. Update the root `README.md` as new functionality becomes available. **(done)**

## TODO

- [x] Implement Automerge document data type or integrate an existing library.
- [x] Persist repository data to disk using `FsStore`.
- [x] Add tests for `repo.Repo` and `repo.FsStore`.
- [x] Prototype networking support based on the Rust implementation.
- [x] Port or rewrite example programs from `rust/examples` in Go.
- [x] Integrate `RepoMessage` handling with connectors.
- [x] Implement `RepoHandle` style connection management and background sync.
  - Basic document sync added via Automerge's sync protocol.
- [x] Add connection lifecycle events via `RepoHandle.Events` channel.

## Part 2: Roadmap to Production

The current Go code covers document storage and a simple JSON-based handshake.
To reach feature parity with the Rust implementation, we still need to port
several subsystems:

1. **Advanced storage** – replicate `fs_store` from Rust, including incremental
   saves and snapshot compaction.
2. **Repo handle and document handles** – design a goroutine-based model for
   background syncing and document observers similar to `RepoHandle` and
   `DocHandle`.
3. **Sync protocol** – implement join/peer exchange and `RepoMessage` handling
   so documents sync over TCP/WebSocket connections.
4. **Share policy & ephemeral messages** – port the share policy traits and the
   ability to send ephemeral messages between peers.
5. **Network connectors** – provide TCP and WebSocket utilities mirroring the
   `tokio` helpers in `rust/src/tokio.rs`.
6. **Integration tests** – recreate the Rust network tests in Go to verify
   multi-peer sync and connection lifecycle events.
7. **CLI updates** – extend `cmd` examples to demonstrate network sync and
   expose configuration flags for storage paths and ports.

## Part 3: Production readiness checklist

After feature parity with the Rust code is reached, several additional steps are
required to ship a stable Go library and CLI. These tasks focus on polishing the
API, improving reliability and providing tooling for real world use:

- [x] **CBOR message format** – replace the temporary JSON encoding with the
  CBOR based protocol implemented in `rust/src/message.rs` so Go peers can
  interoperate with other Automerge Repo implementations.
2. [x] **Robust error handling** – audit the goroutine based sync loops and
   connectors to ensure failures are surfaced correctly and connections are
   cleaned up. Mirror the `ConnComplete` semantics from the Rust code.
3. **Documentation & examples** – write package level docs and expand the CLI
   programs to demonstrate document sharing across multiple peers. **(done)**
4. [x] **Continuous integration** – configure a CI workflow that builds and runs
   tests on Linux, macOS and Windows. Include `go vet` and coverage reporting.
5. [x] **Versioned releases** – once the API stabilises, tag releases and provide
   instructions for importing via Go modules.

## Part 4: Review-based improvements

Following an initial code review the project should address several issues
around network reliability, testing and documentation. The tasks below outline
work needed in this phase:

- [x] **Add context-aware handshake helpers**
  - Modify `repo.Handshake`, `repo.Connect` and `repo.DialWebSocket` to accept
    a `context.Context`.
  - Use deadlines or context cancellation to abort the handshake if it times
    out.
  - Update callers and unit tests to pass a context with timeout.
- [x] **Implement connection completion/retry logic**
  - [x] Introduce a `ConnComplete`-style mechanism in `repo/handle.go` to signal
    when connection goroutines end.
  - [x] Allow optional reconnection attempts via callbacks or a retry policy.
  - [x] Extend tests in `repo/handle_events_test.go` to verify reconnection or
    completion events.
- [x] **Add multi-peer integration tests**
  - Create tests under `repo/` that run multiple `RepoHandle` instances
    connected via TCP or WebSockets.
  - Verify document changes propagate across all peers and lifecycle events are
    fired.
  - Use the Rust `tests/` directory as a guide for expected behaviours.
- [ ] **Reach feature parity with the Rust implementation**
  - Implement snapshot compaction, document handles, share policy and remaining
    network utilities.
- [ ] **Improve package documentation**
  - Provide higher level docs explaining repository usage and integration
    patterns.

## Maintaining this file

Whenever you make progress on the project, update the **Next steps** and **TODO** lists above. When you commit changes that complete an item, mark it as done with an `x` (e.g. `- [x]`). Add any new tasks or notes that future agents should be aware of. Keep this file concise and focused on actionable items.

See [ROADMAP_PROGRESS.md](ROADMAP_PROGRESS.md) for a log of work completed on
the production roadmap and notes on remaining tasks.
