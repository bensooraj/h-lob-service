package binancewebsocket

import (
	"time"

	"gitlab.com/hooklabs-backend/order-management-system-engine/h-lob-service/hwebsocket"
)

// BinanceWebsocket ...
type BinanceWebsocket struct {
	BaseURL string
	Conn    *hwebsocket.WebsocketConnection
}

// NewBinanceWebsocket ...
func NewBinanceWebsocket() *BinanceWebsocket {
	binanceWebsocket := &BinanceWebsocket{}

	return binanceWebsocket
}

// Open ...
func (bws *BinanceWebsocket) Open(url string, messageHandleFunc func([]byte) error, errorHandleFunc func(error)) {
	bws.BaseURL = url

	bws.Conn = hwebsocket.
		New().
		SetWebsocketURL(url).
		SetMessageHandleFunc(messageHandleFunc).
		SetErrorHandleFunc(errorHandleFunc).
		SetAutoReconnect(true).
		SetConnectionRetryLimit(10).
		Build()

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-bws.Conn.CloseChannel:
				return
			case tick := <-ticker.C:
				bws.Conn.SendPingMessage([]byte(tick.String()))
			}
		}
	}()
	return
}
