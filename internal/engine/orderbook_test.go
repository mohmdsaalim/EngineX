package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOrderBook(t *testing.T)  {
	book := NewOrderBook("INFY")
	assert.Equal(t, "INFY", book.Symbol)
	assert.NotNil(t, book.Bids)
	assert.NotNil(t, book.Asks)
	assert.Empty(t, book.Orders)
}


func TestAddToBook(t *testing.T)  {
	book := NewOrderBook("INFY")

	order := &Order{
		ID: "ord-1",
		UserID: "user-1",
		Symbol: "INFY",
		Side: Buy,
		Type: Limit,
		Price: 150000,
		Quantity: 100,
		Status: StatusOpen,
	}

	book.addToBook(order)
	assert.Len(t, book.Orders, 1)
	assert.Equal(t, 1, book.Bids.Len())
}

func TestCancelOrder(t *testing.T)  {
	book := NewOrderBook("INFY")

	order := &Order{
		ID: "ord-1",
		UserID: "user-1",
		Side: Buy,
		Type: Limit,
		Price: 150000,
		Quantity: 100,
	}
	book.addToBook(order)

	cancelled := book.Cancel("ord-1")
	assert.True(t, cancelled)
	assert.Empty(t, book.Orders)
	assert.Equal(t, 0, book.Bids.Len())
}

func TestCancelNonExistent(t *testing.T)  {
	book := NewOrderBook("INFY")
	cancelled := book.Cancel("non existent")
	assert.False(t, cancelled)
}

func TestSnapshot(t *testing.T)  {
	book := NewOrderBook("INFY")

	book.addToBook(&Order{ID: "b1", UserID: "u1", Side: Buy, Type: Limit, Price: 150000, Quantity: 100})
	book.addToBook(&Order{ID: "a1", UserID: "a1", Side: Sell, Type: Limit, Price: 151000, Quantity: 50})

	snap := book.Snapshot(5)
	assert.Len(t, snap.Bids, 1)
	assert.Len(t, snap.Asks, 1)
	assert.Equal(t, int64(150000), snap.Bids[0].Price)
	assert.Equal(t, int64(151000), snap.Asks[0].Price)
}