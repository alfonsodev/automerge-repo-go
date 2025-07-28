package echo

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"

	"github.com/alfonsodev/automerge-repo-go/repo"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// AutomergeRepoHandler returns an Echo handler function that upgrades the connection
// to a WebSocket and bridges it with the Automerge repository.
func AutomergeRepoHandler(handle *repo.RepoHandle) echo.HandlerFunc {
	return func(c echo.Context) error {
		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			log.Printf("failed to upgrade connection: %v", err)
			return err
		}
		defer ws.Close()

		conn := repo.NewWSConnAdapter(ws)
		remoteID, err := repo.Handshake(c.Request().Context(), conn, handle.Repo.ID, repo.Incoming)
		if err != nil {
			log.Printf("handshake failed: %v", err)
			return err
		}

		// Bridge the WebSocket connection with the repo's network adapter.
		// This will handle the Automerge sync protocol.
		complete := handle.AddConn(remoteID, conn)
		complete.Await()

		return nil
	}
}