package network

import (
	"io"

	"github.com/automerge/automerge-repo-go"
	"github.com/gorilla/websocket"
)

// WSConnAdapter wraps a WebSocket connection to implement the Conn interface
// and also satisfies io.ReadWriter for handshaking.
type WSConnAdapter struct {
	conn *websocket.Conn
	r    io.Reader
}

// NewWSConnAdapter creates a new WSConnAdapter.
func NewWSConnAdapter(conn *websocket.Conn) *WSConnAdapter {
	return &WSConnAdapter{conn: conn}
}

// SendMessage sends a message over the WebSocket connection.
func (c *WSConnAdapter) SendMessage(msg repo.RepoMessage) error {
	data, err := msg.Encode()
	if err != nil {
		return err
	}
	return c.conn.WriteMessage(websocket.BinaryMessage, data)
}

// RecvMessage receives a message from the WebSocket connection.
func (c *WSConnAdapter) RecvMessage() (repo.RepoMessage, error) {
	_, data, err := c.conn.ReadMessage()
	if err != nil {
		return repo.RepoMessage{}, err
	}
	return repo.DecodeRepoMessage(data)
}

// Close closes the WebSocket connection.
func (c *WSConnAdapter) Close() error {
	return c.conn.Close()
}

// Read reads data from the WebSocket connection for io.Reader.
func (c *WSConnAdapter) Read(p []byte) (n int, err error) {
	if c.r == nil {
		// The gorilla/websocket library returns a new reader for each message.
		// We buffer it until it's fully consumed.
		_, r, err := c.conn.NextReader()
		if err != nil {
			return 0, err
		}
		c.r = r
	}
	n, err = c.r.Read(p)
	if err == io.EOF {
		// The current message has been fully read, clear the reader
		// so the next call to Read() will get the next message.
		c.r = nil
		return n, nil // Return (n, nil) because EOF for one message is not EOF for the connection
	}
	return n, err
}

// Write writes data to the WebSocket connection for io.Writer.
func (c *WSConnAdapter) Write(p []byte) (n int, err error) {
	err = c.conn.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}
