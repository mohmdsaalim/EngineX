package engine

import (
	"github.com/google/btree"
)

const btreeDegree = 32 
// OrderBook holds all bids and asks for one symbol.
// Bids: descending — highest price first (best bid on top)
// Asks: ascending  — lowest price first (best ask on top)
// Uses B-tree for O(log n) insert/delete with natural sort order.
type OrderBook struct {
	Symbol string
	Bids *btree.BTreeG[*PriceLevel]
	Asks *btree.BTreeG[*PriceLevel]
	Orders map[string]*Order // orderID -> Order (for cancel lookup )
}

// NewOrderBook creates a fresh order book for a symbol. 
func NewOrderBook(symbol string) *OrderBook {
	return &OrderBook{
		Symbol: symbol,
		// Bids - descening : highst price has priority
		Bids: btree.NewG(btreeDegree, func(a, b *PriceLevel)bool{
				return a.Price > b.Price
		}),
		// Asks - ascending : lowest price has prirty
		Asks: btree.NewG(btreeDegree, func(a, b *PriceLevel) bool {
			return a.Price < b.Price
		}),
		Orders: make(map[string]*Order),
	}
}

// addTobook places an unfilled/partial order onto the book.
func (ob *OrderBook) addToBook(o *Order) {
	ob.Orders[o.ID] = o

	if o.Side == Buy{
		ob.addToBids(o)
	}else{
		ob.addToAsks(o)
	}
}

// AddToBids
func (ob *OrderBook) addToBids(o *Order) {
	key := &PriceLevel{Price: o.Price}
	if existing, ok := ob.Bids.Get(key); ok{
		existing.Add(o)
	}else{
		pl := &PriceLevel{Price: o.Price}
		pl.Add(o)
		ob.Bids.ReplaceOrInsert(pl)
	}
}

// AddToAsks
func (ob *OrderBook) addToAsks(o *Order) {
	key := &PriceLevel{Price: o.Price}
	if existing, ok := ob.Asks.Get(key); ok {
		existing.Add(o)
	} else{
		pl := &PriceLevel{Price: o.Price}
		pl.Add(o)
		ob.Asks.ReplaceOrInsert(pl)
	}
}

// Cancel removes an order from the book. 
func (ob *OrderBook) Cancel(orderID string) bool {
	o, exists := ob.Orders[orderID]
	if !exists{
		return false
	}

	tree := ob.Bids
	if o.Side == Sell{
		tree = ob.Asks
	}

	key := &PriceLevel{Price: o.Price}
	if pl, ok := tree.Get(key); ok {
		// remove from price level FIFO queue
		for i, ord := range pl.Orders {
			if ord.ID == orderID {
				pl.Orders = append(pl.Orders[:i], pl.Orders[i+1:]...)
			}
		}
		if pl.IsEmpty(){
			tree.Delete(key)
		}
	}
	delete(ob.Orders, orderID)
	return true
}



// Snapshot returns top 5 price levels for bids and asks.
// Send to WS Hub via kafka orderbook.updated topic.
func (ob *OrderBook) Snapshot(depth int) DepthSnapshot{
	snap := DepthSnapshot{Symbol: ob.Symbol}

	count := 0
	ob.Bids.Ascend(func(pl *PriceLevel) bool {
		if count >= depth {
			return false
		}
		snap.Bids = append(snap.Bids, DepthLevel{
			Price: pl.Price,
			Quantity: pl.TotalQuantity(),
		})
		count++
		return true
	})
	ob.Asks.Ascend(func(pl *PriceLevel) bool {
		if count >= depth {
			return false
		}
		snap.Asks = append(snap.Asks, DepthLevel{
			Price: pl.Price,
			Quantity: pl.TotalQuantity(),
		})
		count++
		return true
	})
	return snap
}