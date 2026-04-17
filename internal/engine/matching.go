package engine

import (
	"time"

	"github.com/google/uuid"
)
// pusher type for fullfilled func -> 
type PriceLevelTree interface {
    Delete(*PriceLevel) (*PriceLevel, bool)
}
// Match executes price-time priority matching for incoming order.
// Returns list of trades generated.
// Remaining unfilled quantity rests on the book (LIMIT only).
//
// Algorithm:
// BUY  order → walk Asks ascending  (lowest ask first)
// SELL order → walk Bids descending (highest bid first)
func (ob *OrderBook) Match(incoming *Order) []Trade {
	var trades []Trade

	if incoming.Side == Buy {
		trades = ob.matchBuy(incoming)
	}else{
		trades = ob.matchSell(incoming)
	}

	// rest unfilled LIMIT order on book
	if incoming.Type == Limit && !incoming.IsFilled(){
		incoming.Status = StatusOpen
		if incoming.Filled > 0{
			incoming.Status = StatusPartial
		}
		ob.addToBook(incoming)
	}
	return trades
}

// matchBuy walks asks from lowest price upward.
// Matches if lowest ask price <= buy price.
func (ob *OrderBook) matchBuy(incoming *Order) []Trade  {
	var trades []Trade 

	for incoming.Remaining() > 0 {
		var bestAsk *PriceLevel
		ob.Asks.Ascend(func(pl *PriceLevel) bool {
			bestAsk = pl
			return false // stop after first
		})

		if bestAsk == nil{
			break
		}

		// price cgeck = LIMIT: asl must be <= buy price
		// Market: mtach at any price
		if incoming.Type == Limit && bestAsk.Price > incoming.Price{
			break
		}

		trades = append(trades, ob.fillFromLevel(incoming, bestAsk, ob.Asks)...)
	}
	return trades
}

// ,atchsell walks bids from hihst price downward
// Matches if highest bid price >= sell price.
func (ob *OrderBook) matchSell(incoming *Order)[]Trade {
	var trades []Trade

	for incoming.Remaining() > 0 {
		// peek best bid
		var bestBid *PriceLevel
		ob.Bids.Ascend(func(pl *PriceLevel) bool {
			bestBid = pl
			return false // stop after first
		})

		if bestBid == nil{
			break
		}

		// price check - LIMIT: bid must be >= sell price
		if incoming.Type == Limit && bestBid.Price < incoming.Price{
			break
		}

		trades = append(trades, ob.fillFromLevel(incoming, bestBid, ob.Bids)...)
	}
	return trades
}


func (ob *OrderBook) fillFromLevel(incoming *Order, level *PriceLevel, tree PriceLevelTree )[]Trade {
	var trades []Trade

	for !level.IsEmpty() && incoming.Remaining() > 0{
		resting := level.Front()

		//Self-match prevention 
		if resting.UserID == incoming.UserID{
			break
		}

		fillQty := min(incoming.Remaining(), resting.Remaining())

		trade := ob.buildTrade(incoming, resting, level.Price, fillQty)
		trades = append(trades, trade)

		//Update filled quantitiies
		incoming.Filled += fillQty
		resting.Filled += fillQty

		// Update Status
		if resting.IsFilled(){
			resting.Status = StatusFilled
			level.RemoveFront()
			delete(ob.Orders, resting.ID)
		}else{
			resting.Status = StatusPartial
		}

		if incoming.IsFilled(){
			incoming.Status = StatusFilled
		}
	}

	// Remove empty price level from tree
	if level.IsEmpty(){
		tree.Delete(level)
	}
	return trades
}

func (ob *OrderBook) buildTrade(incoming, resting *Order, price, qty int64) Trade {
	buyOrder, sellOrder := incoming, resting
	if incoming.Side == Sell{
		buyOrder, sellOrder = resting, incoming
	}

	return Trade{
		ID: uuid.New().String(),
		BuyOrderID: buyOrder.ID,
		SellOrderID: sellOrder.ID,
		BuyerID: buyOrder.UserID,
		SellerID: sellOrder.UserID,
		Symbol: ob.Symbol,
		Price: price,
		Quantity: qty,
		ExecutedAt: time.Now(),
	}
}

func min(a, b int64) int64 {
	if a < b{
		return a
	}
	return b
}