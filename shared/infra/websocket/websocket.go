package websocket

type Websocket interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	Close() error
}

type WebsocketConfig struct {
	URL string
}
