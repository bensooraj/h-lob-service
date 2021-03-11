package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"gitlab.com/hooklabs-backend/order-management-system-engine/h-lob-service/binancewebsocket"
)

func main() {
	flag.Parse()
	log.SetFlags(1)

	doneChannel := make(chan struct{}, 0)

	binanceWebsocket := binancewebsocket.NewBinanceWebsocket(doneChannel)

	wsConnectionURL := url.URL{Scheme: "wss", Host: "fstream.binance.com", Path: "/ws/"}
	binanceWebsocket.Open(wsConnectionURL.String(), nil, nil)

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
