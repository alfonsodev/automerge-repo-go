package main

import (
	"context"
	"log"
	"net/http"

	automerge_echo "github.com/alfonsodev/automerge-repo-go/adapters/echo"
	"github.com/alfonsodev/automerge-repo-go/repo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Create a new Automerge repo.
	r := repo.New()

	// Create a handle for the repo, which manages connections.
	handle := repo.NewRepoHandle(r)

	// Run the repo's message handling loop in the background.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-handle.Inbox:
				// In a real application, you would process incoming messages.
				// For this demo, we just log them.
				log.Printf("received message: type=%s doc=%s", msg.Type, msg.DocumentID)
			}
		}
	}()

	// Create a new Echo instance.
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Serve the static frontend files.
	e.Static("/", "public")

	// Add the Automerge repo WebSocket handler.
	e.GET("/ws", automerge_echo.AutomergeRepoHandler(handle))

	// Start the server.
	log.Println("Starting server on :1323...")
	if err := e.Start(":1323"); err != nil && err != http.ErrServerClosed {
		e.Logger.Fatal(err)
	}
}
