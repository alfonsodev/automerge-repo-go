module github.com/automerge/automerge-repo-storage-fs-go

go 1.24.5

require (
	github.com/automerge/automerge-go v0.0.0-20241030180337-6fb4f2d08244
	github.com/automerge/automerge-repo-go v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.6.0
)

require (
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
)

replace github.com/automerge/automerge-repo-go => ../automerge-repo-go
