package limitorderbook

import (
	"github.com/google/btree"
)

// L2LimitOrderBook
type L2LimitOrderBook struct {
	Bids                   *btree.BTree
	Asks                   *btree.BTree
	PriceScale             float64
	CumulativeBidLimitsMap map[LoBFloat64]float64
	CumulativeAskLimitsMap map[LoBFloat64]float64
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

func NewL2LimitOrderBook(pricePrecision float64) *L2LimitOrderBook {
	return &L2LimitOrderBook{
		Bids:                   btree.New(2),
		Asks:                   btree.New(2),
		PriceScale:             pricePrecision,
		CumulativeBidLimitsMap: make(map[LoBFloat64]float64),
		CumulativeAskLimitsMap: make(map[LoBFloat64]float64),
	}
}

func (l2lob *L2LimitOrderBook) UpdateOrAdd(price LoBFloat64, quantity float64, side string) {
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

func (l2lob *L2LimitOrderBook) Remove(price LoBFloat64, side string) {
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
