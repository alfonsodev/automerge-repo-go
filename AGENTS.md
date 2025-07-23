# Guidelines for Codex agents

This repository contains a work-in-progress port of the Automerge Repo from Rust to Go. The original Rust code now lives under `./rust` and should be used for reference when adding features in Go.

## Next steps

1. Flesh out the Go implementation so it matches the capabilities of the Rust repo. Networking layer still needs porting (see `rust/src/*.rs`).
2. Expand the Go example under `cmd/example` into a minimal CLI demonstrating document creation, loading and storage using `repo.FsStore`.
3. Create unit tests for the Go code. Aim for similar coverage as the Rust tests in `rust/tests`. **(done)**
4. Update the root `README.md` as new functionality becomes available.

## TODO

- [ ] Implement Automerge document data type or integrate an existing library.
- [x] Persist repository data to disk using `FsStore`.
- [x] Add tests for `repo.Repo` and `repo.FsStore`.
- [ ] Prototype networking support based on the Rust implementation.
- [ ] Port or rewrite example programs from `rust/examples` in Go.

## Maintaining this file

Whenever you make progress on the project, update the **Next steps** and **TODO** lists above. When you commit changes that complete an item, mark it as done with an `x` (e.g. `- [x]`). Add any new tasks or notes that future agents should be aware of. Keep this file concise and focused on actionable items.
