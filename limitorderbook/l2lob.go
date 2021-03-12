package limitorderbook

import "github.com/google/btree"

// L2LimitOrderBook
type L2LimitOrderBook struct {
	Bids                   *btree.BTree
	Asks                   *btree.BTree
	PricePrecision         float64
	CumulativeBidLimitsMap map[int64]float64
	CumulativeAskLimitsMap map[int64]float64
}

func NewL2LimitOrderBook(pricePrecision float64) *L2LimitOrderBook {
	return &L2LimitOrderBook{
		Bids:                   btree.New(2),
		Asks:                   btree.New(2),
		PricePrecision:         pricePrecision,
		CumulativeBidLimitsMap: make(map[int64]float64),
		CumulativeAskLimitsMap: make(map[int64]float64),
	}
}

func (l2lob *L2LimitOrderBook) UpdateOrAdd(price, quantity float64, side string) {
	adjustedPrice := int64(price * l2lob.PricePrecision)

	if side == "a" {

		if _, ok := l2lob.CumulativeAskLimitsMap[adjustedPrice]; !ok {
			l2lob.Asks.ReplaceOrInsert(btree.Int(adjustedPrice))
		}
		l2lob.CumulativeAskLimitsMap[adjustedPrice] = quantity

	} else if side == "b" {

		if _, ok := l2lob.CumulativeBidLimitsMap[adjustedPrice]; !ok {
			l2lob.Asks.ReplaceOrInsert(btree.Int(adjustedPrice))
		}
		l2lob.CumulativeBidLimitsMap[adjustedPrice] = quantity

	}
}

func (l2lob *L2LimitOrderBook) Remove(price float64, side string) {
	adjustedPrice := int64(price * l2lob.PricePrecision)

	if side == "a" {

		if _, ok := l2lob.CumulativeAskLimitsMap[adjustedPrice]; ok {
			l2lob.Asks.Delete(btree.Int(adjustedPrice))
		}
		delete(l2lob.CumulativeAskLimitsMap, adjustedPrice)

	} else if side == "b" {

		if _, ok := l2lob.CumulativeBidLimitsMap[adjustedPrice]; ok {
			l2lob.Asks.Delete(btree.Int(adjustedPrice))
		}
		delete(l2lob.CumulativeBidLimitsMap, adjustedPrice)

	}
}
