module github.com/automerge/automerge-repo-network-websocket-go

go 1.24.5

require (
	github.com/fxamacker/cbor/v2 v2.9.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.3
)

require github.com/x448/float16 v0.8.4 // indirect

replace github.com/automerge/automerge-repo-go => ../automerge-repo-go
