package repo

// Package repo provides a minimal Automerge repository implementation in Go.
//
// The Repo type manages a set of Automerge documents. Documents can be
// persisted to disk using FsStore and synchronised with remote peers using
// RepoHandle and network connectors. Documents can also be accessed through
// DocumentHandle which exposes helper methods and a simple change
// notification mechanism.
//
// A simple TCP connector and WebSocket helpers are available to establish
// connections between repositories. Messages exchanged between peers use the
// CBOR encoding defined by the Automerge Repo specification.
//
// Example usage:
//
//  store := &repo.FsStore{Dir: "data"}
//  r := repo.NewWithStore(store)
//  doc := r.NewDoc()
//  _ = doc.Set("key", "value")
//  _ = r.SaveDoc(doc.ID)
//
// For peer-to-peer sync, create a RepoHandle for your Repo, then connect to
// another peer using repo.Connect or the WebSocket helpers. These functions
// now take a context so callers can enforce timeouts. The handle will forward
// received messages on its Inbox channel and can broadcast document updates via
// SyncAll.
//
// See the programs under cmd/ for concrete examples.
