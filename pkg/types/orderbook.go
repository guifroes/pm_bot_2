package types

// OrderBook represents a market's order book.
type OrderBook struct {
	MarketID string
	TokenID  string
	Bids     []Level
	Asks     []Level
}

// Level represents a price level in the order book.
type Level struct {
	Price float64
	Size  float64
}

// BestBid returns the highest bid price, or 0 if no bids.
func (o *OrderBook) BestBid() float64 {
	if len(o.Bids) == 0 {
		return 0
	}
	return o.Bids[0].Price
}

// BestAsk returns the lowest ask price, or 0 if no asks.
func (o *OrderBook) BestAsk() float64 {
	if len(o.Asks) == 0 {
		return 0
	}
	return o.Asks[0].Price
}

// Spread returns the difference between best ask and best bid.
func (o *OrderBook) Spread() float64 {
	bid := o.BestBid()
	ask := o.BestAsk()
	if bid == 0 || ask == 0 {
		return 0
	}
	return ask - bid
}

// MidPrice returns the midpoint between best bid and ask.
func (o *OrderBook) MidPrice() float64 {
	bid := o.BestBid()
	ask := o.BestAsk()
	if bid == 0 || ask == 0 {
		return 0
	}
	return (bid + ask) / 2
}
