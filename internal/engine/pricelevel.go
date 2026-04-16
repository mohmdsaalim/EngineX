package engine

// PriceLevel holds all orders at a single price point.
// Orders are stored in FIFO order — front is matched first.
// This is the core of price-time priority matching.
type PriceLevel struct{
	Price int64
	Orders []*Order
}

// pushing (appnd) order to back of queu - time priority
func (pl *PriceLevel) Add(o *Order) {
	pl.Orders = append(pl.Orders, o)
}

// Front func retirn 1st order with out removing it.
func (pl *PriceLevel) Front() *Order{
	if len(pl.Orders) == 0{
		return nil
	}
	return pl.Orders[0]
}

// RemoveFront remove the first order after it is fully filled
func (pl *PriceLevel) RemoveFront() {
	if len(pl.Orders) == 0{
		return
	}
	pl.Orders = pl.Orders[1:]
}

// IsEmpty returs true if no orders remian at this price level
func (pl *PriceLevel) IsEmpty() bool{
	return len(pl.Orders) == 0
}

// TotalQuatity returns sum of remiaing quatity at this level
func (pl *PriceLevel) TotalQuantity() int64 {
	var total int64

	for _, o := range pl.Orders{
		total += o.Remaining()
	}
	return total
}