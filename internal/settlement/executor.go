package settlement

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mohmdsaalim/EngineX/internal/cache"
	repository "github.com/mohmdsaalim/EngineX/internal/repository/generated"
	"github.com/mohmdsaalim/EngineX/pkg/logger"
)

// TradeMessage is what Engine publishes to trades.executed.
type TradeMessage struct {
	ID          string    `json:"id"`
	BuyOrderID  string    `json:"buy_order_id"`
	SellOrderID string    `json:"sell_order_id"`
	BuyerID     string    `json:"buyer_id"`
	SellerID    string    `json:"seller_id"`
	Symbol      string    `json:"symbol"`
	Price       int64     `json:"price"`
	Quantity    int64     `json:"quantity"`
	ExecutedAt  time.Time `json:"executed_at"`
}

type Executor struct {
	pool  *pgxpool.Pool
	q     *repository.Queries
	redis *cache.RedisClient
	log   *slog.Logger
}

func NewExecutor(pool *pgxpool.Pool, redis *cache.RedisClient) *Executor {
	return &Executor{
		pool:  pool,
		q:     repository.New(pool),
		redis: redis,
		log:   logger.New("executor"),
	}
}

// ProcessTrade settles one trade atomically.
// Called for each message from trades.executed Kafka topic.
func (e *Executor) ProcessTrade(ctx context.Context, raw []byte) error {
	// 1. Deserialize trade message
	var trade TradeMessage
	if err := json.Unmarshal(raw, &trade); err != nil {
		return fmt.Errorf("unmarshal trade: %w", err)
	}

	// 2. Idempotency check — skip if already processed
	processed, err := e.redis.IsTradeProcessed(ctx, trade.ID)
	if err != nil {
		return fmt.Errorf("idempotency check: %w", err)
	}
	if processed {
		e.log.Info("trade already processed — skipping", "trade_id", trade.ID)
		return nil
	}

	// 3. Atomic Postgres transaction
	if err := e.settle(ctx, trade); err != nil {
		return fmt.Errorf("settle trade: %w", err)
	}

	// 4. Mark as processed in Redis AFTER successful DB commit
	if err := e.redis.MarkTradeProcessed(ctx, trade.ID); err != nil {
		// Log but don't fail — trade is settled in DB
		e.log.Error("failed to mark trade processed", "trade_id", trade.ID, "error", err)
	}

	e.log.Info("trade settled",
		"trade_id", trade.ID,
		"symbol", trade.Symbol,
		"price", trade.Price,
		"quantity", trade.Quantity,
	)

	return nil
}

// settle runs the atomic Postgres transaction.
// All 4 operations succeed or all rollback — no partial state.
func (e *Executor) settle(ctx context.Context, trade TradeMessage) error {
	tx, err := e.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := e.q.WithTx(tx)

	buyerUUID := toUUID(trade.BuyerID)
	sellerUUID := toUUID(trade.SellerID)
	buyOrderUUID := toUUID(trade.BuyOrderID)
	sellOrderUUID := toUUID(trade.SellOrderID)
	tradeValue := trade.Price * trade.Quantity

	// 1. INSERT trade record
	_, err = qtx.CreateTrade(ctx, repository.CreateTradeParams{
		BuyOrderID:  buyOrderUUID,
		SellOrderID: sellOrderUUID,
		BuyerID:     buyerUUID,
		SellerID:    sellerUUID,
		Symbol:      trade.Symbol,
		Price:       trade.Price,
		Quantity:    trade.Quantity,
	})
	if err != nil {
		return fmt.Errorf("insert trade: %w", err)
	}

	// 2. DEBIT buyer cash balance
	_, err = qtx.DebitBalance(ctx, repository.DebitBalanceParams{
		UserID:      buyerUUID,
		DebitAmount: tradeValue,
	})
	if err != nil {
		return fmt.Errorf("debit buyer: %w", err)
	}

	// 3. CREDIT seller cash balance
	_, err = qtx.CreditBalance(ctx, repository.CreditBalanceParams{
		UserID:        sellerUUID,
		CreditAmount:  tradeValue,
		LockedRelease: tradeValue,
	})
	if err != nil {
		return fmt.Errorf("credit seller: %w", err)
	}

	// 4. UPDATE order statuses
	_, err = qtx.UpdateOrderStatus(ctx, repository.UpdateOrderStatusParams{
		ID:        buyOrderUUID,
		Status:    "FILLED",
		FilledQty: trade.Quantity,
	})
	if err != nil {
		return fmt.Errorf("update buy order: %w", err)
	}

	_, err = qtx.UpdateOrderStatus(ctx, repository.UpdateOrderStatusParams{
		ID:        sellOrderUUID,
		Status:    "FILLED",
		FilledQty: trade.Quantity,
	})
	if err != nil {
		return fmt.Errorf("update sell order: %w", err)
	}

	// 5. COMMIT — only reaches here if all 4 succeed
	return tx.Commit(ctx)
}

func toUUID(s string) pgtype.UUID {
	var u pgtype.UUID
	u.Scan(s)
	return u
}