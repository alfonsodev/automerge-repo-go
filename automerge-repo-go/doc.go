/*
Package repo provides a Go implementation of the Automerge Repo protocol.
It allows for creating, managing, and synchronizing Automerge documents between
multiple peers over a network.

The main components of this package are:

- Repo: A collection of documents that can be persisted to storage.

- RepoHandle: Manages peer connections and handles the synchronization of
documents. It uses a pluggable networking model, with TCP and WebSocket
connectors provided.

- DocumentHandle: The primary way to interact with a single Automerge document.
It provides methods for reading, mutating, and saving the document, as well as
a mechanism to watch for changes.

A typical usage pattern involves:
 1. Creating a Repo, optionally with an FsStore for persistence.
 2. Creating or loading a DocumentHandle from the Repo.
 3. Using the DocumentHandle to read or modify the document.
 4. Creating a RepoHandle to manage network connections.
 5. Connecting to peers and using the RepoHandle to synchronize documents.
*/
package repo