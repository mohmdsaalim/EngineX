package websocket

import (
	"context"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mohmdsaalim/EngineX/pkg/logger"
)

const (
	sendBufferSize  = 256
	writeTimeout    = 10 * time.Second
	pingInterval    = 30 * time.Second
)

// Client represents one connected WebSocket client.
type Client struct {
	conn   *websocket.Conn
	symbol string
	sendCh chan []byte // buffered — slow clients dropped
	log    *slog.Logger
}

func NewClient(conn *websocket.Conn, symbol string) *Client {
	return &Client{
		conn:   conn,
		symbol: symbol,
		sendCh: make(chan []byte, sendBufferSize),
		log:    logger.New("wshub-client"),
	}
}

// Send queues message to client. Non-blocking.
// Drops message if buffer full — slow client protection.
func (c *Client) Send(msg []byte) {
	select {
	case c.sendCh <- msg:
	default:
		// Buffer full — drop message, client is too slow
		c.log.Warn("client buffer full — dropping message", "symbol", c.symbol)
	}
}

// WritePump reads from sendCh and writes to WebSocket.
// Runs in its own goroutine per client.
func (c *Client) WritePump(ctx context.Context, onDone func()) {
	defer func() {
		c.conn.Close()
		onDone()
	}()

	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case msg, ok := <-c.sendCh:
			c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}