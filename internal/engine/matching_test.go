package engine

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func newOrder(id, userID string, side Side, typ OrderType, price, qty int64) *Order {
	return &Order{
		ID:        id,
		UserID:    userID,
		Symbol:    "INFY",
		Side:      side,
		Type:      typ,
		Price:     price,
		Quantity:  qty,
		Status:    StatusOpen,
		CreatedAt: time.Now(),
	}
}

func TestFullFill(t *testing.T) {
	book := NewOrderBook("INFY")

	// Resting sell at 150000
	sell := newOrder("s1", "seller", Sell, Limit, 150000, 100)
	trades := book.Match(sell)
	assert.Empty(t, trades) // no match — no bids yet

	// Incoming buy at 150000 — should fully match
	buy := newOrder("b1", "buyer", Buy, Limit, 150000, 100)
	trades = book.Match(buy)

	assert.Len(t, trades, 1)
	assert.Equal(t, int64(100), trades[0].Quantity)
	assert.Equal(t, int64(150000), trades[0].Price)
	assert.Equal(t, "s1", trades[0].SellOrderID)
	assert.Equal(t, "b1", trades[0].BuyOrderID)
	assert.True(t, buy.IsFilled())
	assert.True(t, sell.IsFilled())
}

func TestPartialFill(t *testing.T) {
	book := NewOrderBook("INFY")

	// Resting sell 50 units
	sell := newOrder("s1", "seller", Sell, Limit, 150000, 50)
	book.Match(sell)

	// Incoming buy 100 units — partial fill
	buy := newOrder("b1", "buyer", Buy, Limit, 150000, 100)
	trades := book.Match(buy)

	assert.Len(t, trades, 1)
	assert.Equal(t, int64(50), trades[0].Quantity)
	assert.Equal(t, int64(50), buy.Filled)
	assert.Equal(t, int64(50), buy.Remaining())
	assert.False(t, buy.IsFilled())
	assert.True(t, sell.IsFilled())

	// buy rests on book with remaining 50
	assert.Equal(t, 1, book.Bids.Len())
}

func TestNoMatch_PriceMismatch(t *testing.T) {
	book := NewOrderBook("INFY")

	// Sell at 155000
	sell := newOrder("s1", "seller", Sell, Limit, 155000, 100)
	book.Match(sell)

	// Buy at 150000 — price too low, no match
	buy := newOrder("b1", "buyer", Buy, Limit, 150000, 100)
	trades := book.Match(buy)

	assert.Empty(t, trades)
	assert.Equal(t, 1, book.Bids.Len())
	assert.Equal(t, 1, book.Asks.Len())
}

func TestMarketOrder(t *testing.T) {
	book := NewOrderBook("INFY")

	// Resting limit sell
	sell := newOrder("s1", "seller", Sell, Limit, 150000, 100)
	book.Match(sell)

	// Market buy — matches at any price
	buy := newOrder("b1", "buyer", Buy, Market, 0, 100)
	trades := book.Match(buy)

	assert.Len(t, trades, 1)
	assert.Equal(t, int64(100), trades[0].Quantity)
	assert.True(t, buy.IsFilled())
}

func TestSelfMatchPrevention(t *testing.T) {
	book := NewOrderBook("INFY")

	// Same user places both sides
	sell := newOrder("s1", "same-user", Sell, Limit, 150000, 100)
	book.Match(sell)

	buy := newOrder("b1", "same-user", Buy, Limit, 150000, 100)
	trades := book.Match(buy)

	// Must not match — self-match prevented
	assert.Empty(t, trades)
}

func TestPriceTimePriority(t *testing.T) {
	book := NewOrderBook("INFY")

	// Two sells at same price — first in, first matched
	sell1 := newOrder("s1", "seller1", Sell, Limit, 150000, 50)
	sell2 := newOrder("s2", "seller2", Sell, Limit, 150000, 50)
	book.Match(sell1)
	book.Match(sell2)

	buy := newOrder("b1", "buyer", Buy, Limit, 150000, 50)
	trades := book.Match(buy)

	// sell1 matched first — time priority
	assert.Len(t, trades, 1)
	assert.Equal(t, "s1", trades[0].SellOrderID)
}

func TestMultipleFills(t *testing.T) {
	book := NewOrderBook("INFY")

	// 3 sell orders at different prices
	book.Match(newOrder("s1", "seller1", Sell, Limit, 149000, 30))
	book.Match(newOrder("s2", "seller2", Sell, Limit, 150000, 30))
	book.Match(newOrder("s3", "seller3", Sell, Limit, 151000, 30))

	// Large buy sweeps multiple levels
	buy := newOrder("b1", "buyer", Buy, Limit, 150000, 60)
	trades := book.Match(buy)

	// Matches s1 (30) + s2 (30) = 60 total
	assert.Len(t, trades, 2)
	assert.True(t, buy.IsFilled())
}