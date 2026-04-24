package websocket

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	kafkapkg "github.com/mohmdsaalim/EngineX/internal/kafka"
	"github.com/mohmdsaalim/EngineX/pkg/logger"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == "" || origin == "http://localhost" || origin == "https://yourdomain.com"
	},
}

// Hub manages all WebSocket connections grouped by symbol.
const maxClientsPerSymbol = 1000

type Hub struct {
	mu               sync.RWMutex
	clients          map[string][]*Client
	log              *slog.Logger
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string][]*Client),
		log:     logger.New("wshub"),
	}
}

// HandleWS upgrades HTTP to WebSocket and registers client.
// URL: /ws?symbol=INFY
func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		http.Error(w, "symbol required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Error("upgrade failed", "error", err)
		return
	}

	client := NewClient(conn, symbol)
	h.register(symbol, client)

	h.log.Info("client connected", "symbol", symbol)

	ctx, cancel := context.WithCancel(r.Context())

	// WritePump runs in background — unregisters client on disconnect
	go client.WritePump(ctx, func() {
		cancel()
		h.unregister(symbol, client)
		h.log.Info("client disconnected", "symbol", symbol)
	})
}

// Broadcast sends message to all clients subscribed to symbol.
func (h *Hub) Broadcast(symbol string, msg []byte) {
	h.mu.RLock()
	clients := h.clients[symbol]
	h.mu.RUnlock()

	for _, c := range clients {
		c.Send(msg)
	}
}

func (h *Hub) register(symbol string, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.clients[symbol]) >= maxClientsPerSymbol {
		h.log.Error("max clients limit reached", "symbol", symbol)
		return
	}
	h.clients[symbol] = append(h.clients[symbol], c)
}

func (h *Hub) unregister(symbol string, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	clients := h.clients[symbol]
	for i, client := range clients {
		if client == c {
			h.clients[symbol] = append(clients[:i], clients[i+1:]...)
			break
		}
	}
}

// Consume reads orderbook.updates from Kafka and fans out to clients.
func (h *Hub) Consume(ctx context.Context, consumer *kafkapkg.Consumer) {
	h.log.Info("wshub consuming", "topic", "orderbook.updates")

	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := consumer.ReadMessage(ctx)
			if err != nil {
				h.log.Error("read error", "error", err)
				continue
			}

			symbol := string(msg.Key)
			if symbol == "" {
				consumer.CommitMessage(ctx, msg)
				continue
			}

			if len(symbol) < 1 || len(symbol) > 10 {
				h.log.Error("invalid symbol format", "symbol", symbol)
				consumer.CommitMessage(ctx, msg)
				continue
			}

			consumer.CommitMessage(ctx, msg)
			wrappedMsg := buildDepthMessage(symbol, msg.Value)
			if wrappedMsg != nil {
				h.Broadcast(symbol, wrappedMsg)
			}
		}
	}
}

// DepthMessage wraps snapshot for clients.
type DepthMessage struct {
	Type   string      `json:"type"`
	Symbol string      `json:"symbol"`
	Data   interface{} `json:"data"`
}

func buildDepthMessage(symbol string, raw []byte) []byte {
	var data interface{}
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil
	}
	if data == nil {
		return nil
	}
	msg, _ := json.Marshal(DepthMessage{
		Type:   "depth",
		Symbol: symbol,
		Data:   data,
	})
	return msg
}
