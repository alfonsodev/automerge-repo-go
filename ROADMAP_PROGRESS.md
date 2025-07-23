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
- Added WebSocket connector helpers `DialWebSocket` and `AcceptWebSocket`.
  - Uses Gorilla WebSocket to upgrade HTTP connections and send JSON messages.
  - Unit test `TestWebSocketHandshake` verifies the WebSocket handshake.
- Introduced `RepoMessage` type and JSON encoding helpers in `repo/message.go`.
  - Supports `sync` and `ephemeral` message variants.
  - Added unit test `TestRepoMessageEncodeDecode` for round‑trip validation.
- Extended `LPConn` and `WSConn` with `SendMessage`/`RecvMessage` for
  transmitting `RepoMessage` values.
  - Added unit tests `TestLPConnSendRecvMessage` and `TestWSConnSendRecvMessage`
    verifying JSON message exchange.
- Introduced `RepoHandle` with basic peer management in `repo/handle.go`.
  - Supports adding connections that forward incoming messages on a channel.
  - Provides `SendMessage` and `Broadcast` helpers.
  - Added unit test `TestRepoHandleMessageForwarding` using an in-memory connection.
- Added basic document synchronization via `RepoHandle.SyncDocument` and
  `RepoHandle.SyncAll`.
  - Documents use Automerge's sync protocol to exchange changes.
  - Unit test `TestRepoHandleSync` verifies syncing between two handles.
- Updated `cmd/tcp-example` to use `RepoHandle` and connectors.
  - On connect or accept it performs the join/peer handshake and syncs all
    documents.
  - Messages received are printed to stdout.

## Missing / Next Steps
- Connection lifecycle management remains incomplete and should evolve toward
  the Rust `RepoHandle` design.
- Additional CLI features like editing documents over the network are still planned.
