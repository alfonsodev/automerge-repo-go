# Guidelines for Codex agents

This repository contains a work-in-progress port of the Automerge Repo from Rust to Go. The original Rust code now lives under `./rust` and should be used for reference when adding features in Go.

## Next steps

1. Flesh out the Go implementation so it matches the capabilities of the Rust repo. Basic networking handshake prototyped, more work needed (see `rust/src/*.rs`).
2. Expand the Go example under `cmd/example` into a minimal CLI demonstrating document creation, loading and storage using `repo.FsStore`. **(done)**
3. Create unit tests for the Go code. Aim for similar coverage as the Rust tests in `rust/tests`. **(done)**
4. Update the root `README.md` as new functionality becomes available.

## TODO

- [x] Implement Automerge document data type or integrate an existing library.
- [x] Persist repository data to disk using `FsStore`.
- [x] Add tests for `repo.Repo` and `repo.FsStore`.
- [x] Prototype networking support based on the Rust implementation.
- [x] Port or rewrite example programs from `rust/examples` in Go.
- [x] Integrate `RepoMessage` handling with connectors.
- [ ] Implement `RepoHandle` style connection management and background sync.
  - Basic connection handling implemented, but document sync logic still pending.

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

## Maintaining this file

Whenever you make progress on the project, update the **Next steps** and **TODO** lists above. When you commit changes that complete an item, mark it as done with an `x` (e.g. `- [x]`). Add any new tasks or notes that future agents should be aware of. Keep this file concise and focused on actionable items.

See [ROADMAP_PROGRESS.md](ROADMAP_PROGRESS.md) for a log of work completed on
the production roadmap and notes on remaining tasks.
