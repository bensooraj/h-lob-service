package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	jsoniter "github.com/json-iterator/go"
	"gitlab.com/hooklabs-backend/order-management-system-engine/h-lob-service/binancewebsocket"
	"gitlab.com/hooklabs-backend/order-management-system-engine/h-lob-service/limitorderbook"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func main() {
	flag.Parse()
	log.SetFlags(1)

	doneChannel := make(chan struct{}, 0)

	binanceL2LoB := limitorderbook.NewBinanceL2LimitOrderBook()
	binanceL2LoB.UpdateOrderBook(doneChannel)

	wsConnectionURL := url.URL{Scheme: "wss", Host: "stream.binancefuture.com", Path: "/ws/"}

	binanceWebsocket := binancewebsocket.NewBinanceWebsocket(doneChannel)
	binanceWebsocket.Open(wsConnectionURL.String(), func(msg []byte) error {
		var err error
		// Check if it's a depth update event
		var depthUpdate binancewebsocket.DepthUpdate
		err = json.Unmarshal(msg, &depthUpdate)
		if err == nil && depthUpdate.EventType == "depthUpdate" {
			log.Println("DEPTH Update Received", depthUpdate.EventType)

			binanceL2LoB.DepthUpdateBufferChannel <- depthUpdate

			return nil
		}

		// Else check if it's a response to a live subscribe unsubscribe request
		var liveResponse binancewebsocket.LiveResponse
		err = json.Unmarshal(msg, &liveResponse)
		if err == nil && liveResponse.ID > 0 {
			log.Println("Live Response Received", liveResponse.ID, liveResponse.Result)

			return nil
		}

		return err
	}, func(err error) {
		log.Println("WALLA WALLA WALLA: ", err.Error())
	})

	streamList := []string{"btcusdt@depth"}
	binanceWebsocket.Subscribe(1, streamList)

	signalInterrupt := make(chan os.Signal, 1)
	signal.Notify(signalInterrupt, os.Interrupt)

	for {
		select {
		case <-signalInterrupt:

			binanceWebsocket.Unsubscribe(1, streamList)
			<-time.After(7 * time.Second)

			close(doneChannel)
			<-time.After(7 * time.Second)
			return
		}
	}
}
