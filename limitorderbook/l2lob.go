package limitorderbook

import (
	"sync"

	"github.com/google/btree"
	"github.com/robaho/fixed"
)

// L2LimitOrderBook
type L2LimitOrderBook struct {
	Exchange               string
	Symbol                 string
	Bids                   *btree.BTree
	Asks                   *btree.BTree
	PriceScale             float64
	CumulativeBidLimitsMap map[LoBFixed]float64
	CumulativeAskLimitsMap map[LoBFixed]float64
	sync.Mutex
}

// LoBInt64 implements the Item interface for int64.
type LoBInt64 int64

// Less returns true if int64(a) < int64(b).
func (a LoBInt64) Less(b btree.Item) bool {
	return a < b.(LoBInt64)
}

// LoBInt64 implements the Item interface for int64.
type LoBFloat64 float64

// Less returns true if int64(a) < int64(b).
func (a LoBFloat64) Less(b btree.Item) bool {
	return a < b.(LoBFloat64)
}

// LoBInt64 implements the Item interface for int64.
type LoBFixed fixed.Fixed

// Less returns true if Fixed(a) < Fixed(b).
func (a LoBFixed) Less(b btree.Item) bool {

	return fixed.Fixed(a).LessThan(fixed.Fixed(b.(LoBFixed)))
}

func NewL2LimitOrderBook(pricePrecision float64) *L2LimitOrderBook {
	return &L2LimitOrderBook{
		Bids:                   btree.New(2),
		Asks:                   btree.New(2),
		PriceScale:             pricePrecision,
		CumulativeBidLimitsMap: make(map[LoBFixed]float64),
		CumulativeAskLimitsMap: make(map[LoBFixed]float64),
	}
}

// SetExchange ..
func (l2lob *L2LimitOrderBook) SetExchange(exchangeName string) *L2LimitOrderBook {
	l2lob.Exchange = exchangeName
	return l2lob
}

// SetSymbol ..
func (l2lob *L2LimitOrderBook) SetSymbol(symbolName string) *L2LimitOrderBook {
	l2lob.Symbol = symbolName
	return l2lob
}

func (l2lob *L2LimitOrderBook) UpdateOrAdd(price LoBFixed, quantity float64, side string) {
	// adjustedPrice := LoBInt64(price * math.Pow(10, l2lob.PriceScale))

	if side == "a" {

		if _, ok := l2lob.CumulativeAskLimitsMap[price]; !ok {
			l2lob.Asks.ReplaceOrInsert(price)
		}
		l2lob.CumulativeAskLimitsMap[price] = quantity

	} else if side == "b" {

		if _, ok := l2lob.CumulativeBidLimitsMap[price]; !ok {
			l2lob.Bids.ReplaceOrInsert(price)
		}
		l2lob.CumulativeBidLimitsMap[price] = quantity

	}
}

func (l2lob *L2LimitOrderBook) Remove(price LoBFixed, side string) {
	// adjustedPrice := LoBInt64(price * math.Pow(10, l2lob.PriceScale))

	if side == "a" {

		if _, ok := l2lob.CumulativeAskLimitsMap[price]; ok {
			l2lob.Asks.Delete(price)
		}
		delete(l2lob.CumulativeAskLimitsMap, price)

	} else if side == "b" {

		if _, ok := l2lob.CumulativeBidLimitsMap[price]; ok {
			l2lob.Bids.Delete(price)
		}
		delete(l2lob.CumulativeBidLimitsMap, price)

	}
}
