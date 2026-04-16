package engine

import "time"

// side represends order ditect
type Side int8

const (
	Buy Side = 1
	Sell Side = 2
)

// OrderType represends order execution type 
type OrderType int8

const (
	Limit OrderType = 1
	Market OrderType = 2
)

// Status represends current order state
type Status int8

const (
	StatusOpen      Status = 1
	StatusPartial   Status = 2
	StatusFilled    Status = 3
	StatusCancelled Status = 4
)

// Order represents a single order in the engine.
// ₹1500.50 → 150050
type Order struct {
	ID        string
	UserID    string
	Symbol    string
	Side      Side
	Type      OrderType
// Price and Quantity are scaled int64 
	Price     int64
	Quantity  int64
	Filled    int64
	Status    Status
	CreatedAt time.Time
}

// Remaining returns unfilled quantity.
func (o *Order) Remaining() int64 {
	return o.Quantity - o.Filled
}

// Isfilled returns true if order id fully matched
func (o *Order ) IsFilled() bool {
	return o.Filled >= o.Quantity
}

// Trade represents an executed match between two orders.
type Trade struct {
	ID          string
	BuyOrderID  string
	SellOrderID string
	BuyerID     string
	SellerID    string
	Symbol      string
	Price       int64
	Quantity    int64
	ExecutedAt  time.Time
}

// DepthLevel represents one price level in order book snapshot.
type DepthLevel struct {
	Price    int64
	Quantity int64
}

// DepthSnapshot is the order book state sent to WS Hub via Kafka.
type DepthSnapshot struct {
	Symbol string
	Bids   []DepthLevel // descending price
	Asks   []DepthLevel // ascending price
}