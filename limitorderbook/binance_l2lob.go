package limitorderbook

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/google/btree"
	jsoniter "github.com/json-iterator/go"
	"github.com/robaho/fixed"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type BinanceL2LimitOrderBook struct {
	*L2LimitOrderBook
	LastUpdateID int64
}

func NewBinanceL2LimitOrderBook() *BinanceL2LimitOrderBook {
	bL2LoB := &BinanceL2LimitOrderBook{}

	bL2LoB.Exchange = "binance"
	bL2LoB.Exchange = "BTCUSDT"
	bL2LoB.Bids = btree.New(2)
	bL2LoB.Asks = btree.New(2)
	bL2LoB.CumulativeBidLimitsMap = make(map[LoBFixed]float64)
	bL2LoB.CumulativeAskLimitsMap = make(map[LoBFixed]float64)

	return bL2LoB
}

// DepthSnapshot ...
type DepthSnapshot struct {
	LastUpdateID      int64       `json:"lastUpdateId"`
	MessageOutputTime int64       `json:"E"`
	TransactionTime   int64       `json:"T"`
	Bids              [][2]string `json:"bids"`
	Asks              [][2]string `json:"asks"`
}

// SetOrderBookFromSnapshot ...
func (bL2LoB *BinanceL2LimitOrderBook) SetOrderBookFromSnapshot() error {
	depthSnapshotURL := url.URL{Scheme: "https", Host: "testnet.binancefuture.com", Path: "/fapi/v1/depth?symbol=BTCUSDT&limit=1000"}
	response, err := http.Get(depthSnapshotURL.String())
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	var depthSnapshot DepthSnapshot
	err = json.Unmarshal(data, &depthSnapshot)
	if err != nil {
		return err
	}

	// Set the last update ID
	bL2LoB.LastUpdateID = depthSnapshot.LastUpdateID

	// Update the bids
	for _, bid := range depthSnapshot.Bids {
		p := bid[0] // Price
		q := bid[1] // Quantity

		price := LoBFixed(fixed.NewS(p))
		quantity, err := strconv.ParseFloat(q, 64)
		if err != nil {
			continue
		}

		bL2LoB.UpdateOrAdd(price, quantity, "b")
	}
	// Update the asks
	for _, ask := range depthSnapshot.Asks {
		p := ask[0] // Price
		q := ask[1] // Quantity

		price := LoBFixed(fixed.NewS(p))
		quantity, err := strconv.ParseFloat(q, 64)
		if err != nil {
			continue
		}

		bL2LoB.UpdateOrAdd(price, quantity, "a")
	}

	return nil
}
