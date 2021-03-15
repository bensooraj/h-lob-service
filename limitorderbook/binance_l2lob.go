package limitorderbook

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/google/btree"
	jsoniter "github.com/json-iterator/go"
	"github.com/robaho/fixed"
	"gitlab.com/hooklabs-backend/order-management-system-engine/h-lob-service/binancewebsocket"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type BinanceL2LimitOrderBook struct {
	*L2LimitOrderBook
	LastUpdateID             int64
	DepthUpdateBufferChannel chan binancewebsocket.DepthUpdate
}

func NewBinanceL2LimitOrderBook() *BinanceL2LimitOrderBook {
	l2lob := NewL2LimitOrderBook(0)
	bL2LoB := &BinanceL2LimitOrderBook{
		L2LimitOrderBook:         l2lob,
		LastUpdateID:             0,
		DepthUpdateBufferChannel: make(chan binancewebsocket.DepthUpdate, 100),
	}

	bL2LoB.Exchange = "binance"
	bL2LoB.Symbol = "BTCUSDT"
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

// InitOrderBookFromSnapshot ...
func (bL2LoB *BinanceL2LimitOrderBook) InitOrderBookFromSnapshot() error {
	depthSnapshotURL := url.URL{Scheme: "https", Host: "testnet.binancefuture.com", Path: "/fapi/v1/depth", RawQuery: "symbol=BTCUSDT&limit=1000"}
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

	bL2LoB.Lock()
	defer bL2LoB.Unlock()

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
	log.Printf("%s orderbook for %s initialised\n", bL2LoB.Exchange, bL2LoB.Symbol)

	return nil
}

func (bL2LoB *BinanceL2LimitOrderBook) UpdateOrderBook(doneChannel <-chan struct{}) {

	go func() {
		for {
			select {
			case <-doneChannel:
				log.Println("Exiting UpdateOrderBook")
				return
			case depthUpdate := <-bL2LoB.DepthUpdateBufferChannel:
				// If the LastUpdateID has been reset to 0, please wait for
				if bL2LoB.LastUpdateID == 0 {
					err := bL2LoB.InitOrderBookFromSnapshot()
					if err != nil {
						log.Println("Error initialising the orderbook from the depth snapshot: ", err)
						break
					}
				}

				//
				bL2LoB.Lock()

				// Update the bids
				for _, bid := range depthUpdate.BidDepthDelta {
					p := bid[0] // Price
					q := bid[1] // Quantity

					price := LoBFixed(fixed.NewS(p))
					quantity, err := strconv.ParseFloat(q, 64)
					if err != nil {
						continue
					}

					if quantity == 0 {
						bL2LoB.Remove(price, "b")
					} else {
						bL2LoB.UpdateOrAdd(price, quantity, "b")
					}
				}
				// Update the asks
				for _, ask := range depthUpdate.AskDepthDelta {
					p := ask[0] // Price
					q := ask[1] // Quantity

					price := LoBFixed(fixed.NewS(p))
					quantity, err := strconv.ParseFloat(q, 64)
					if err != nil {
						continue
					}

					if quantity == 0 {
						bL2LoB.Remove(price, "a")
					} else {
						bL2LoB.UpdateOrAdd(price, quantity, "a")
					}
				}
				defer bL2LoB.Unlock()
			default:
				// Just so the select is non-blocking

			}
		}
	}()

}
