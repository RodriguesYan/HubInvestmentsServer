package websocket

import (
	"time"

	"github.com/gorilla/websocket"
)

type GorillaWebsocket struct {
	conn *websocket.Conn
}

func NewGorillaWebsocket(conn *websocket.Conn) *GorillaWebsocket {
	return &GorillaWebsocket{
		conn: conn,
	}
}

func (w *GorillaWebsocket) ReadMessage() (messageType int, p []byte, err error) {
	return w.conn.ReadMessage()
}

func (w *GorillaWebsocket) WriteMessage(messageType int, data []byte) error {
	return w.conn.WriteMessage(messageType, data)
}

func (w *GorillaWebsocket) Close() error {
	return w.conn.Close()
}

func (w *GorillaWebsocket) SetReadDeadline(t time.Time) error {
	return w.conn.SetReadDeadline(t)
}

func (w *GorillaWebsocket) SetWriteDeadline(t time.Time) error {
	return w.conn.SetWriteDeadline(t)
}

func (w *GorillaWebsocket) SetPongHandler(h func(appData string) error) {
	w.conn.SetPongHandler(h)
}
