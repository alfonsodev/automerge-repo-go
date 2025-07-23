# Roadmap Progress

This document captures work done towards **Part 2: Roadmap to Production**.
It summarizes implemented functionality and notes remaining tasks to guide
future development.

## Implemented

- Added a basic TCP connector in `repo/connector.go`.
  - Provides `LPConn` for sending/receiving length‑prefixed JSON messages.
  - Includes `Connect` helper which performs the join/peer handshake and
    returns the remote repository ID along with a connection handle.
  - Unit test `TestConnect` exercises the handshake over a `net.Pipe`.

## Missing / Next Steps

- WebSocket support and integration with the connector API.
- Handling of `RepoMessage` types (sync and ephemeral messages).
- Connection lifecycle management and background goroutines similar to the
  Rust `RepoHandle` implementation.
- Integration of connectors with the example CLI for real networking.
