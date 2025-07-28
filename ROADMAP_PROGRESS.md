# Roadmap Progress

This file tracks the progress of the Go implementation of the Automerge Repo.

## 2025-07-28

*   Refactored the project to use a Go workspace.
*   The project is now divided into three modules:
    *   `automerge-repo-go`
    *   `automerge-repo-network-websocket-go`
    *   `automerge-repo-storage-fs-go`
*   The `automerge-repo-go` module contains the core `Repo` implementation, and its tests are passing.
*   The other two modules are currently empty.
*   The old `typescript` and `repo` directories have been removed.
*   The `AGENTS.md` file has been updated to reflect the new project structure and goals.