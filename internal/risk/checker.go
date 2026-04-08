package risk
// balance, duplicate, quantity,
import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mohmdsaalim/EngineX/internal/cache"
	repository "github.com/mohmdsaalim/EngineX/internal/repository/generated"
	"github.com/mohmdsaalim/EngineX/pkg/apperr"
)

// max allowed order value 10 lakh (100000000)
const maxOrderValue int64 = 100000000

type Checker struct {
	redis *cache.RedisClient
	queries *repository.Queries
}

func NewChecker(redis *cache.RedisClient, queries *repository.Queries) *Checker {
	return &Checker{
		redis: redis,
		queries: queries,
	}
}

type OrderRequest struct {
	UserID string
	orderID string
	Symbol string
	Side string // BUY OR SELL
	Price int64
	Quantity int64
}

// checkorder runs all 3 pre-trade risk checks
// Returns nil if apprvd, AppError if reject

func (c Checker) CheckOrder(ctx context.Context, req OrderRequest) error {
	// 1 = Order value check
	if err := c.checkOrderSize(req); err != nil{
		return err
	}
	// 2 balance check only for buy orders
	if req.Side == "BUY"{
	if err := c.checkBalance(ctx, req); err != nil{
		return err
	}
	}
	//  else {
    //     if err := c.checkPosition(ctx, req); err != nil {
    //         return err
    //     }
    // }

	// 3 duplicate order chek
	if err := c.checkDuplicate(ctx, req.orderID); err != nil{
		return err
	}
	return nil // temp
}
// 1
func (c *Checker) checkOrderSize(req OrderRequest) error {
	orderValue := req.Price * req.Quantity
	if orderValue <= 0{
		return apperr.New(apperr.CodeInvalidInput, "order value must be greater than zero")
	}
	if orderValue > maxOrderValue{
		return apperr.New(apperr.CodePositionLimit, fmt.Sprintf("order value %d exceeds max allowed %d", orderValue, maxOrderValue))
	}
	return nil
}

// 2
// checkBalance reads from Redis cache 
// falls back to postgres if cache empty or miss

func (c *Checker) checkBalance(ctx context.Context, req OrderRequest) error {
	orderValue := req.Price * req.Quantity

	// Fast way -> Redis chache
	available, err := c.redis.GetBalance(ctx, req.UserID)
	if err != nil || available == 0{
		// slow way -> goes to postgres 
		available, err = c.getBalanceFromDB(ctx, req.UserID)
		if err != nil{
			return apperr.Wrap(apperr.CodeInternal, " failed to fetch balance", err)
		}
	}
	if available < orderValue{
		return apperr.New(apperr.CodeInsufficientBal, fmt.Sprintf("insufficient balance: have %d need %d", available, orderValue))
	}
	return nil
}
// depend with 2 checkBalance 
func (c *Checker) getBalanceFromDB(ctx context.Context, userID string) (int64, error) {
	var pgUUID pgtype.UUID
	
	if err := pgUUID.Scan(userID); err != nil{
		return 0, fmt.Errorf("invalid userID: %w", err)
	}
	// parse userID str to pgtyp UUID
	balance, err := c.queries.GetBalanceByUserID(ctx, pgUUID)
	if err != nil {
		return 0, err
	}
	return balance.Available, nil
}

// checkduplicate ensure same orderID not submitted twice
func (c *Checker) checkDuplicate(ctx context.Context, orderID string) error{
	key := "order:seen" + orderID
	existing, err := c.redis.Get(ctx, key)
	if err != nil{
		return apperr.Wrap(apperr.CodeInternal, " duplicate check failed", err)
	}
	if existing != ""{
		return apperr.New(apperr.CodeDuplicateOrder, " duplicate orderID")
	}

	// mark thuis irderID as seen - exp in 24 hrs

	if err := c.redis.Set(ctx, key, "1", 24*60*60*1000000000); err != nil{
		return apperr.Wrap(apperr.CodeInternal, "failed to mark order", err)
	}
	return nil
}

// positin not setuped in DB
// // checkPosition verifies user holds enough quantity for SELL order.

// func (c *Checker) checkPosition(ctx context.Context, req OrderRequest) error {
//     var pgUUID pgtype.UUID
//     if err := pgUUID.Scan(req.UserID); err != nil {
//         return apperr.Wrap(apperr.CodeInternal, "invalid userID", err)
//     }

//     position, err := c.queries.GetPositionByUserAndSymbol(ctx, repository.GetPositionByUserAndSymbolParams{
//         UserID: pgUUID,
//         Symbol: req.Symbol,
//     })
//     if err != nil {
//         return apperr.Wrap(apperr.CodeInternal, "failed to fetch position", err)
//     }

//     available := position.Quantity - position.LockedQty
//     if available < req.Quantity {
//         return apperr.New(apperr.CodeInsufficientBal,
//             fmt.Sprintf("insufficient position: have %d need %d", available, req.Quantity))
//     }
//     return nil
// }