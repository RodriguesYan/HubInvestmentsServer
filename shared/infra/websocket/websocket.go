package websocket

import (
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// Message types
const (
	TextMessage   = websocket.TextMessage
	BinaryMessage = websocket.BinaryMessage
	CloseMessage  = websocket.CloseMessage
	PingMessage   = websocket.PingMessage
	PongMessage   = websocket.PongMessage
)

// Close codes
const (
	CloseNormalClosure   = websocket.CloseNormalClosure
	CloseGoingAway       = websocket.CloseGoingAway
	CloseAbnormalClosure = websocket.CloseAbnormalClosure
)

type Websocket interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	Close() error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
	SetPongHandler(h func(appData string) error)
}

type WebsocketConfig struct {
	URL string
}

// IsUnexpectedCloseError checks if an error is an unexpected close error
func IsUnexpectedCloseError(err error, expectedCodes ...int) bool {
	if err == nil {
		return false
	}

	// Check for common connection closed errors
	errStr := err.Error()
	if strings.Contains(errStr, "connection closed") ||
		strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "connection reset") {
		return true
	}

	// Use Gorilla's error checking if available
	return websocket.IsUnexpectedCloseError(err, expectedCodes...)
}
