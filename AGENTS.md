# Guidelines for Gemini agents

This repository is a Go implementation of the Automerge Repo. The canonical implementation is the [TypeScript version](https://github.com/automerge/automerge-repo/tree/main/packages/automerge-repo), and this project aims to replicate its structure and functionality in Go.

The project is structured as a Go workspace with the following modules:

*   `automerge-repo-go`: The core `Repo` implementation.
*   `automerge-repo-network-websocket-go`: A WebSocket network adapter.
*   `automerge-repo-storage-fs-go`: A file system storage adapter.

## Next steps

1.  Flesh out the `automerge-repo-go` implementation so it matches the capabilities of the TypeScript repo.
2.  Implement the WebSocket network adapter in `automerge-repo-network-websocket-go`.
3.  Implement the file system storage adapter in `automerge-repo-storage-fs-go`.
4.  Create unit tests for all modules, aiming for similar coverage as the TypeScript tests.
5.  Update the root `README.md` as new functionality becomes available.

## TODO

- [ ] Implement the `Repo` class in `automerge-repo-go`.
- [ ] Implement the `DocHandle` class in `automerge-repo-go`.
- [ ] Implement the `Synchronizer` classes in `automerge-repo-go`.
- [ ] Implement the `NetworkSubsystem` in `automerge-repo-go`.
- [ ] Implement the `StorageSubsystem` in `automerge-repo-go`.
- [ ] Implement the `WebSocketClientAdapter` in `automerge-repo-network-websocket-go`.
- [ ] Implement the `WebSocketServerAdapter` in `automerge-repo-network-websocket-go`.
- [ ] Implement the `NodeFSStorageAdapter` in `automerge-repo-storage-fs-go`.

## Maintaining this file

Whenever you make progress on the project, update the **Next steps** and **TODO** lists above. When you commit changes that complete an item, mark it as done with an `x` (e.g. `- [x]`). Add any new tasks or notes that future agents should be aware of. Keep this file concise and focused on actionable items.

See [ROADMAP_PROGRESS.md](ROADMAP_PROGRESS.md) for a log of work completed on the roadmap and notes on remaining tasks.