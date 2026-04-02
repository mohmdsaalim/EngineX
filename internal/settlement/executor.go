package settlement

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	repository "github.com/mohmdsaalim/EngineX/internal/repository/generated"
)

type Executor struct {
    pool *pgxpool.Pool
    q    *repository.Queries
}

// converts string UUID to pgtype.UUID
func toPgUUID(s string) pgtype.UUID {
    var u pgtype.UUID
    u.Scan(s)
    return u
}

func NewExecutor(pool *pgxpool.Pool) *Executor {
    return &Executor{
        pool: pool,
        q:    repository.New(pool),
    }
}

type TradeParams struct {
    BuyOrderID  string
    SellOrderID string
    BuyerID     string
    SellerID    string
    Symbol      string
    Price       int64
    Quantity    int64
}

func (e *Executor) Settle(ctx context.Context, trade TradeParams) error {
    tx, err := e.pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }
    defer tx.Rollback(ctx)

    qtx := e.q.WithTx(tx)

    _, err = qtx.CreateTrade(ctx, repository.CreateTradeParams{
        BuyOrderID:  toPgUUID(trade.BuyOrderID),
        SellOrderID: toPgUUID(trade.SellOrderID),
        BuyerID:     toPgUUID(trade.BuyerID),
        SellerID:    toPgUUID(trade.SellerID),
        Symbol:      trade.Symbol,
        Price:       trade.Price,
        Quantity:    trade.Quantity,
    })
    if err != nil {
        return fmt.Errorf("insert trade: %w", err)
    }
	
// debit balance 
	_, err = qtx.DebitBalance(ctx, repository.DebitBalanceParams{
		DebitAmount: trade.Price * trade.Quantity,
		UserID:      toPgUUID(trade.BuyerID),
	})
	if err != nil {
		return fmt.Errorf("debit buyer: %w", err)
	}

// CreditBalance
	_, err = qtx.CreditBalance(ctx, repository.CreditBalanceParams{
		CreditAmount:  trade.Price * trade.Quantity,
		LockedRelease: trade.Price * trade.Quantity,
		UserID:        toPgUUID(trade.SellerID),
	})
	if err != nil {
		return fmt.Errorf("credit seller: %w", err)
	}

    return tx.Commit(ctx)
}