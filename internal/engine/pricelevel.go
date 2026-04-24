package engine

// PriceLevel holds all orders at a single price point.
// Orders are stored in FIFO order — front is matched first.
// This is the core of price-time priority matching.
type PriceLevel struct {
	Price  int64
	Orders []*Order
}

// Add appends order to back of queue - time priority
func (pl *PriceLevel) Add(o *Order) {
	pl.Orders = append(pl.Orders, o)
}

// Front returns 1st order without removing it.
func (pl *PriceLevel) Front() *Order {
	if len(pl.Orders) == 0 {
		return nil
	}
	return pl.Orders[0]
}

// RemoveFront removes the first order after it is fully filled
func (pl *PriceLevel) RemoveFront() {
	if len(pl.Orders) == 0 {
		return
	}
	pl.Orders = pl.Orders[1:]
}

// IsEmpty returns true if no orders remain at this price level
func (pl *PriceLevel) IsEmpty() bool {
	return len(pl.Orders) == 0
}

// TotalQuantity returns sum of remaining quantity at this level
func (pl *PriceLevel) TotalQuantity() int64 {
	var total int64

	for _, o := range pl.Orders {
		total += o.Remaining()
	}
	return total
}