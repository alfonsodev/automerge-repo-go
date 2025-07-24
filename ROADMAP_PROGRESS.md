# Roadmap Progress

This document captures work done towards **Part 2: Roadmap to Production**.
It summarizes implemented functionality and notes remaining tasks to guide
future development.

## Implemented

- Added a basic TCP connector in `repo/connector.go`.
  - Provides `LPConn` for sending/receiving length‑prefixed messages.
  - Includes `Connect` helper which performs the join/peer handshake and
    returns the remote repository ID along with a connection handle.
  - Unit test `TestConnect` exercises the handshake over a `net.Pipe`.
- Added WebSocket connector helpers `DialWebSocket` and `AcceptWebSocket`.
  - Uses Gorilla WebSocket to upgrade HTTP connections and send CBOR-encoded messages.
  - Unit test `TestWebSocketHandshake` verifies the WebSocket handshake.
- Introduced `RepoMessage` type and encoding helpers in `repo/message.go`.
  - Encoding switched from JSON to CBOR for interoperability.
  - Supports `sync` and `ephemeral` message variants.
  - Added unit test `TestRepoMessageEncodeDecode` for round‑trip validation.
- Extended `LPConn` and `WSConn` with `SendMessage`/`RecvMessage` for
  transmitting `RepoMessage` values.
  - Added unit tests `TestLPConnSendRecvMessage` and `TestWSConnSendRecvMessage`
    verifying CBOR message exchange.
- Introduced `RepoHandle` with basic peer management in `repo/handle.go`.
  - Supports adding connections that forward incoming messages on a channel.
  - Provides `SendMessage` and `Broadcast` helpers.
  - Added unit test `TestRepoHandleMessageForwarding` using an in-memory connection.
- Added basic document synchronization via `RepoHandle.SyncDocument` and
  `RepoHandle.SyncAll`.
  - Documents use Automerge's sync protocol to exchange changes.
  - Unit test `TestRepoHandleSync` verifies syncing between two handles.
- Added connection lifecycle events via `RepoHandle.Events`.
  - Supports `peer_connected` and `peer_disconnected` notifications.
  - Unit test `TestRepoHandleConnectionEvents` verifies event delivery.
- Switched handshake messages to CBOR encoding across network helpers.
  - `LPConn`, `WSConn` and the `Handshake` function now use CBOR.
  - Added unit test `TestRepoHandleConnErrorEvent` exercising connection error events.
- Introduced connection error events.
  - `RepoHandle` publishes `conn_error` when a connection closes unexpectedly.
- Updated `cmd/tcp-example` to use `RepoHandle` and connectors.
  - On connect or accept it performs the join/peer handshake and syncs all
    documents.
  - Messages received are printed to stdout and a simple REPL allows editing a
    document which syncs to all peers.
- Added handling for send failures in `RepoHandle`.
  - `SendMessage` and `Broadcast` now emit `conn_error` and remove the peer when
    a send operation fails.
- Configured GitHub Actions workflow `go.yml` for CI.
  - Runs `go vet` and `go test` on Linux, macOS and Windows.
- Added package level documentation in `repo/doc.go` describing the main types
  and how to use network connectors.
- Added `scripts/release.sh` and release instructions in `README.md`.
  - Provides a helper for tagging versions and building cross platform binaries.
- Handshake helpers now accept a `context.Context` to allow timeouts.
  - `Connect`, `DialWebSocket` and `Handshake` updated along with all callers
    and unit tests.
- Added `ConnComplete` type and retry support in `RepoHandle`.
  - `AddConn` now returns a completion handle and optional reconnection can be
    provided via `AddConnWithRetry`.
  - New test `TestRepoHandleConnComplete` verifies completion delivery.

## Missing / Next Steps
- Document handles remain to be implemented.
- Continue improving reconnection behaviour and parity with the Rust `ConnComplete` API.
- More comprehensive usage examples would be helpful.
- Consider automating GitHub releases in the future.
