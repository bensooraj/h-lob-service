package binancewebsocket

import (
	"time"

	"github.com/bensooraj/h-lob-service/hwebsocket"
)

// BinanceWebsocket ...
type BinanceWebsocket struct {
	BaseURL      string
	Conn         *hwebsocket.WebsocketConnection
	CloseChannel chan struct{}
}

// NewBinanceWebsocket ...
func NewBinanceWebsocket(closeChannel chan struct{}) *BinanceWebsocket {
	binanceWebsocket := &BinanceWebsocket{
		CloseChannel: closeChannel,
	}

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
			case <-bws.CloseChannel:
				bws.Conn.Close()
				return
			case tick := <-ticker.C:
				bws.Conn.SendPingMessage([]byte(tick.String()))
			}
		}
	}()
	return
}

// Subscribe ...
func (bws *BinanceWebsocket) Subscribe(ID int64, streamList []string) {
	subscribeRequest := LiveRequest{
		Method: "SUBSCRIBE",
		Params: streamList,
		ID:     ID,
	}
	bws.Conn.SendJSONMessage(subscribeRequest)
	return
}

// Unsubscribe ...
func (bws *BinanceWebsocket) Unsubscribe(ID int64, streamList []string) {
	unsubscribeRequest := LiveRequest{
		Method: "UNSUBSCRIBE",
		Params: streamList,
		ID:     ID,
	}
	bws.Conn.SendJSONMessage(unsubscribeRequest)
	return
}
