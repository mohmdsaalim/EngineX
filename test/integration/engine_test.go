package integration_test

import (
	"testing"

	"github.com/mohmdsaalim/EngineX/internal/engine"
	"github.com/stretchr/testify/assert"
)

func TestEngine_NewOrderBook(t *testing.T) {
	book := engine.NewOrderBook("TEST")
	assert.NotNil(t, book)
	assert.Equal(t, "TEST", book.Symbol)
}

func TestEngine_NewEngine(t *testing.T) {
	e := engine.NewEngine(nil)
	assert.NotNil(t, e)
}

func TestEngine_Order_Remaining(t *testing.T) {
	order := &engine.Order{
		ID:        "test-1",
		UserID:    "user-1",
		Symbol:   "A",
		Quantity: 100,
		Filled:   30,
	}

	remaining := order.Remaining()
	assert.Equal(t, int64(70), remaining)
}

func TestEngine_Order_IsFilled(t *testing.T) {
	tests := []struct {
		name     string
		order   *engine.Order
		isFilled bool
	}{
		{
			name: "not filled",
			order: &engine.Order{
				ID:        "o1",
				UserID:   "u1",
				Symbol:   "A",
				Quantity: 100,
				Filled:   50,
			},
			isFilled: false,
		},
		{
			name: "fully filled",
			order: &engine.Order{
				ID:        "o2",
				UserID:   "u2",
				Symbol:   "A",
				Quantity: 100,
				Filled:   100,
			},
			isFilled: true,
		},
		{
			name: "more than filled",
			order: &engine.Order{
				ID:        "o3",
				UserID:   "u3",
				Symbol:   "A",
				Quantity: 50,
				Filled:   100,
			},
			isFilled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isFilled, tt.order.IsFilled())
		})
	}
}

func TestEngine_Order_Side(t *testing.T) {
	buyOrder := &engine.Order{
		ID:     "buy-1",
		UserID: "u1",
		Symbol: "A",
		Side:   engine.Buy,
	}

	sellOrder := &engine.Order{
		ID:     "sell-1",
		UserID: "u2",
		Symbol: "A",
		Side:   engine.Sell,
	}

	assert.Equal(t, engine.Buy, buyOrder.Side)
	assert.Equal(t, engine.Sell, sellOrder.Side)
}

func TestEngine_Order_Status(t *testing.T) {
	order := &engine.Order{
		ID:     "o1",
		UserID: "u1",
		Symbol: "A",
		Status: engine.StatusOpen,
	}

	assert.Equal(t, engine.StatusOpen, order.Status)

	order.Status = engine.StatusFilled
	assert.Equal(t, engine.StatusFilled, order.Status)
}

func TestEngine_DepthSnapshot(t *testing.T) {
	snap := engine.DepthSnapshot{
		Symbol: "TEST",
		Bids: []engine.DepthLevel{
			{Price: 100000, Quantity: 50},
			{Price: 99000, Quantity: 30},
		},
		Asks: []engine.DepthLevel{
			{Price: 100050, Quantity: 20},
			{Price: 101000, Quantity: 10},
		},
	}

	assert.Equal(t, "TEST", snap.Symbol)
	assert.Len(t, snap.Bids, 2)
	assert.Len(t, snap.Asks, 2)
}

func TestEngine_DepthLevel(t *testing.T) {
	level := engine.DepthLevel{
		Price:    150000,
		Quantity: 100,
	}

	assert.Equal(t, int64(150000), level.Price)
	assert.Equal(t, int64(100), level.Quantity)
}

func TestEngine_Trade_Fields(t *testing.T) {
	trade := engine.Trade{
		ID:          "trade-1",
		BuyOrderID:  "buy-o1",
		SellOrderID: "sell-o1",
		BuyerID:    "buyer-1",
		SellerID:   "seller-1",
		Symbol:     "INFY",
		Price:      150000,
		Quantity:   50,
	}

	assert.Equal(t, "trade-1", trade.ID)
	assert.Equal(t, "buy-o1", trade.BuyOrderID)
	assert.Equal(t, "sell-o1", trade.SellOrderID)
	assert.Equal(t, "buyer-1", trade.BuyerID)
	assert.Equal(t, "seller-1", trade.SellerID)
	assert.Equal(t, "INFY", trade.Symbol)
	assert.Equal(t, int64(150000), trade.Price)
	assert.Equal(t, int64(50), trade.Quantity)
}

func TestEngine_OrderBook_Symbol(t *testing.T) {
	book := engine.NewOrderBook("SYMBOL")
	assert.Equal(t, "SYMBOL", book.Symbol)
}

func TestEngine_Side_Constants(t *testing.T) {
	assert.Equal(t, engine.Side(1), engine.Buy)
	assert.Equal(t, engine.Side(2), engine.Sell)
}

func TestEngine_OrderType_Constants(t *testing.T) {
	assert.Equal(t, engine.OrderType(1), engine.Limit)
	assert.Equal(t, engine.OrderType(2), engine.Market)
}

func TestEngine_Status_Constants(t *testing.T) {
	assert.Equal(t, engine.Status(1), engine.StatusOpen)
	assert.Equal(t, engine.Status(2), engine.StatusPartial)
	assert.Equal(t, engine.Status(3), engine.StatusFilled)
	assert.Equal(t, engine.Status(4), engine.StatusCancelled)
}

func TestEngine_Order_Fields(t *testing.T) {
	order := &engine.Order{
		ID:        "order-1",
		UserID:    "user-1",
		Symbol:   "INFY",
		Side:     engine.Buy,
		Type:     engine.Limit,
		Price:    150000,
		Quantity: 100,
		Filled:   0,
		Status:   engine.StatusOpen,
	}

	assert.Equal(t, "order-1", order.ID)
	assert.Equal(t, "user-1", order.UserID)
	assert.Equal(t, "INFY", order.Symbol)
	assert.Equal(t, engine.Buy, order.Side)
	assert.Equal(t, engine.Limit, order.Type)
	assert.Equal(t, int64(150000), order.Price)
	assert.Equal(t, int64(100), order.Quantity)
}

func TestEngine_OrderStatus_Update(t *testing.T) {
	order := &engine.Order{
		ID:     "o1",
		UserID: "u1",
		Symbol: "A",
		Status: engine.StatusOpen,
	}

	order.Status = engine.StatusPartial
	assert.Equal(t, engine.StatusPartial, order.Status)

	order.Status = engine.StatusFilled
	assert.Equal(t, engine.StatusFilled, order.Status)

	order.Status = engine.StatusCancelled
	assert.Equal(t, engine.StatusCancelled, order.Status)
}