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
	IsInSync                 bool
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

	// Set the last update ID
	bL2LoB.ProcessBidsAndAsks(depthSnapshot.Bids, "b") // b => bids
	bL2LoB.ProcessBidsAndAsks(depthSnapshot.Asks, "a") // a => asks
	bL2LoB.LastUpdateID = depthSnapshot.LastUpdateID

	log.Printf("%s orderbook for %s initialised\n", bL2LoB.Exchange, bL2LoB.Symbol)

	return nil
}

func (bL2LoB *BinanceL2LimitOrderBook) UpdateOrderBook(doneChannel <-chan struct{}) {

	go func() {
		for {
			select {
			case <-doneChannel:
				log.Println("Exiting UpdateOrderBook goroutine")
				return
			case depthUpdate := <-bL2LoB.DepthUpdateBufferChannel:
				log.Printf("[ORDERBOOK] u: %d | U: %d | pu: %d | local: %d\n", depthUpdate.LastUpdateID, depthUpdate.FirstUpdateID, depthUpdate.PreviousLastUpdateID, bL2LoB.LastUpdateID)

				// If the LastUpdateID has been reset to 0, please wait for
				if bL2LoB.LastUpdateID == 0 {
					err := bL2LoB.InitOrderBookFromSnapshot()
					if err != nil {
						log.Println("Error initialising the orderbook from the depth snapshot: ", err)
						break
					}
				}

				// Drop any event where u is < lastUpdateId in the snapshot.
				if depthUpdate.LastUpdateID < bL2LoB.LastUpdateID {
					log.Printf("[ORDERBOOK] Skipping stale Depth Update ID. Received %d | local %d\n ", depthUpdate.LastUpdateID, bL2LoB.LastUpdateID)
					break
				}

				// The first processed event should have U <= lastUpdateId AND u >= lastUpdateId
				if depthUpdate.FirstUpdateID <= bL2LoB.LastUpdateID && depthUpdate.LastUpdateID >= bL2LoB.LastUpdateID {

					log.Println("[ORDERBOOK] Processing first depth update event")
					bL2LoB.ProcessBidsAndAsks(depthUpdate.BidDepthDelta, "b") // b => bids
					bL2LoB.ProcessBidsAndAsks(depthUpdate.AskDepthDelta, "a") // a => asks
					bL2LoB.LastUpdateID = depthUpdate.LastUpdateID

				} else if depthUpdate.PreviousLastUpdateID == bL2LoB.LastUpdateID {

					// While listening to the stream, each new event's pu should be equal to the previous event's u,
					// otherwise re-initialize the process
					log.Println("[ORDERBOOK] Processing in-sync depth update events")
					bL2LoB.ProcessBidsAndAsks(depthUpdate.BidDepthDelta, "b") // b => bids
					bL2LoB.ProcessBidsAndAsks(depthUpdate.AskDepthDelta, "a") // a => asks
					bL2LoB.LastUpdateID = depthUpdate.LastUpdateID

				} else {
					log.Println("[ORDERBOOK] Updates/local copy not in-sync. Re-initialising")
					bL2LoB.LastUpdateID = 0
					break
				}
				//
			default:
				// Just so the select is non-blocking

			}
		}
	}()

}

func (bL2LoB *BinanceL2LimitOrderBook) ProcessBidsAndAsks(priceQuantityPairs [][2]string, side string) error {
	bL2LoB.Lock()
	defer bL2LoB.Unlock()

	for _, pqPair := range priceQuantityPairs {
		p := pqPair[0] // Price
		q := pqPair[1] // Quantity

		price := LoBFixed(fixed.NewS(p))
		quantity, err := strconv.ParseFloat(q, 64)
		if err != nil {
			log.Printf("[ERROR] price %s with quantity %s. %s", p, q, err.Error())
			continue
		}
		// Remove the quantity if needed
		if quantity == 0 {
			bL2LoB.Remove(price, side)
		} else {
			bL2LoB.UpdateOrAdd(price, quantity, side)
		}
	}

	return nil
}
