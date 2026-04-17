package engine

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/mohmdsaalim/EngineX/api/gen/gRPC_order"
	"github.com/mohmdsaalim/EngineX/internal/constants"
	"github.com/mohmdsaalim/EngineX/internal/kafka"
	"github.com/mohmdsaalim/EngineX/pkg/logger"
	"google.golang.org/protobuf/proto"
	// "google.golang.org/grpc/encoding/proto"
)

// Engine manages one orderbook per symbol.
// each symbol runs in its own goroutine - no locking needed.
type Engine struct {
	books    map[string]*OrderBook
	mu 		 sync.RWMutex
	producer *kafka.Producer
	log		 *slog.Logger
}
// New engine create the matching 
func NewEngine(producer *kafka.Producer) *Engine {
	return &Engine{
		books: make(map[string]*OrderBook),
		producer: producer,
		log: logger.New("engine"),
	}
}

// getOrCreateBook returns existing book or create new one for symbol
func (e *Engine) getOrCreateBook(symbol string) *OrderBook {
	e.mu.Lock()
	defer e.mu.Unlock()

	if book, exists := e.books[symbol]; exists {
		return book
	}
	book := NewOrderBook(symbol)
	e.books[symbol] = book
	e.log.Info("create order book", "symbol", symbol)
	return book
}

// ProcessOrder is called for each message from kafka orders.
// Runs Match() and publish results back to kafka 
func (e *Engine) ProcessOrder(ctx context.Context, msg *gRPC_order.OrderMessage) {
	book := e.getOrCreateBook(msg.Symbol)

	// Convert proto mesage to internal order
	order := &Order{
		ID: msg.OrderId,
		UserID: msg.UserId,
		Symbol: msg.Symbol,
		Side: paraseSide(msg.Side),
		Type: parseType(msg.Type),
		Price: msg.Price,
		Quantity: msg.Quantity,
		Status: StatusOpen,
	}

	trades := book.Match(order)

	e.log.Info("order processed", "order_id", order.ID, "symbol", order.Symbol, "trades", len(trades))
	// Publish each trade to trades.executed
	for _, trade := range trades {
		if err := e.publishTrade(ctx, trade); err != nil{
			e.log.Error("publish trade failed", "error", err)
		}
	}

	// Publish order book snapshot to order.updated
	if err := e.publishSnapshot(ctx, book); err != nil{
		e.log.Error("publish snapshot failed", "error", err)
	}
}

//publishtrade sends executed trade to trade.executed topic
func (e *Engine) publishTrade(ctx context.Context, trade Trade) error {
	payload, err := json.Marshal(trade)
	if err != nil{
		return err
	}
	return e.producer.Publish(ctx, constants.TopicTradesExecuted, trade.ID, payload)
}

// publish snap shot sends executed trade to trade.execute topic 
func (e *Engine) publishSnapshot(ctx context.Context, book *OrderBook) error {
	snap := book.Snapshot(5) // top 5 levels
	payload, err := json.Marshal(snap)
	if err != nil{
		return err
	}
	return  e.producer.Publish(ctx, constants.TopicOrderbookUpdates, book.Symbol, payload)
}

// parseSide convert proto string to internal side type. 
func paraseSide(s string) Side {
	if s == "SELL"{
		return Sell
	}
	return Buy
}
// ParseType converts proto string to internal Side type
func parseType(t string) OrderType {
	if t == "MARKET"{
		return Market
	}
	return Limit
}

// consume startes consuming form order.submitted from kafka topics and 
// Each message is processed syncly per sumbol gorotines
func (e *Engine) Consume(ctx context.Context, consumer *kafka.Consumer)  {
	e.log.Info("engine consuming", "topic", constants.TopicOrdersSubmitted)

	for {
		select{
		case <-ctx.Done():
			e.log.Info("engine shutting down")
			return
		default:
			msg, err := consumer.ReadMessage(ctx)
			if err != nil{
				e.log.Error("read message failed", "error", err)
				continue
			}

			var order gRPC_order.OrderMessage
			if err := proto.Unmarshal(msg.Value, &order); err != nil{
				e.log.Error("unmarshal failed", "error", err)
				continue
			}
			e.ProcessOrder(ctx, &order)
		}
	}
}