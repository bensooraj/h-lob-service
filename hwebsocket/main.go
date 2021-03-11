/*
  Heavily taken from alexey-ernest's go-binance-websocket: https://github.com/alexey-ernest/go-binance-websocket
  I didn't like how the variables were named, besides I wanted this to be both a learning experience and
  be completely aware of what I am writing so I can fix it (if any issues arise).
*/

package hwebsocket

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/gorilla/websocket"
)

// WebsocketConfiguration ...
type WebsocketConfiguration struct {
	WebsocketURL         string
	RequestHeader        map[string][]string
	MessageHandleFunc    func([]byte) error
	ErrorHandleFunc      func(error)
	IsAutoReconnect      bool
	IsDump               bool
	ReadDeadlineTime     time.Duration
	ConnectionRetryLimit int
}

type bufferChannel chan []byte

// WebsocketConnection ...
type WebsocketConnection struct {
	Conn                      *websocket.Conn
	WriteBufferChannel        bufferChannel
	PingMessageBufferChannel  bufferChannel
	CloseMessageBufferChannel bufferChannel // For sending close signal to the ws server
	Subscriptions             []interface{}
	CloseChannel              chan struct{} // For closing the connection locally

	WebsocketConfiguration
}

// WebsocketBuilder ...
type WebsocketBuilder struct {
	wsConfig *WebsocketConfiguration
}

// New returns a fresh copy of a websocket conenction builder
func New() *WebsocketBuilder {
	return &WebsocketBuilder{
		wsConfig: &WebsocketConfiguration{
			RequestHeader:        make(map[string][]string, 1),
			ConnectionRetryLimit: 10,
		},
	}
}

// Build ...
func (wsb *WebsocketBuilder) Build() *WebsocketConnection {
	wsc := &WebsocketConnection{
		WebsocketConfiguration: *wsb.wsConfig,
	}
	return wsc.InitialiseConnection()
}

// SetWebsocketURL ...
func (wsb *WebsocketBuilder) SetWebsocketURL(url string) *WebsocketBuilder {
	wsb.wsConfig.WebsocketURL = url
	return wsb
}

// SetRequestHeader ...
func (wsb *WebsocketBuilder) SetRequestHeader(key, value string) *WebsocketBuilder {
	wsb.wsConfig.RequestHeader[key] = append(wsb.wsConfig.RequestHeader[key], value)
	return wsb
}

// SetMessageHandleFunc ...
func (wsb *WebsocketBuilder) SetMessageHandleFunc(f func([]byte) error) *WebsocketBuilder {
	wsb.wsConfig.MessageHandleFunc = f
	return wsb
}

// SetErrorHandleFunc ...
func (wsb *WebsocketBuilder) SetErrorHandleFunc(f func(error)) *WebsocketBuilder {
	wsb.wsConfig.ErrorHandleFunc = f
	return wsb
}

// SetAutoReconnect ...
func (wsb *WebsocketBuilder) SetAutoReconnect(isAutoReconnect bool) *WebsocketBuilder {
	wsb.wsConfig.IsAutoReconnect = isAutoReconnect
	return wsb
}

// SetDump ...
func (wsb *WebsocketBuilder) SetDump(isDump bool) *WebsocketBuilder {
	wsb.wsConfig.IsDump = isDump
	return wsb
}

// SetReadDeadlineTime ...
func (wsb *WebsocketBuilder) SetReadDeadlineTime(d time.Duration) *WebsocketBuilder {
	wsb.wsConfig.ReadDeadlineTime = d
	return wsb
}

// SetConnectionRetryLimit ...
func (wsb *WebsocketBuilder) SetConnectionRetryLimit(crl int) *WebsocketBuilder {
	wsb.wsConfig.ConnectionRetryLimit = crl
	return wsb
}

// InitialiseConnection ...
func (wsc *WebsocketConnection) InitialiseConnection() *WebsocketConnection {
	wsc.ReadDeadlineTime = time.Minute

	err := wsc.Connect()
	if err != nil {
		log.Panic("Error establishing websocket connection", err)
	}

	// This is like the done channel to exit all active go routines
	wsc.CloseChannel = make(chan struct{}, 2)

	// For sending messages to the remote websocket server
	wsc.CloseMessageBufferChannel = make(bufferChannel, 1)
	wsc.PingMessageBufferChannel = make(bufferChannel, 10)
	wsc.WriteBufferChannel = make(bufferChannel, 10)

	go wsc.WriteRequest()
	go wsc.ReceiveMessage()

	return wsc
}

// Connect ...
func (wsc *WebsocketConnection) Connect() error {
	c, resp, err := websocket.DefaultDialer.Dial(wsc.WebsocketURL, http.Header(wsc.RequestHeader))
	if err != nil {
		log.Printf("[ws][%s] %s", wsc.WebsocketURL, err.Error())
		if wsc.IsDump && resp != nil {
			dumpData, _ := httputil.DumpResponse(resp, true)
			log.Printf("[ws][dump][%s] %s", wsc.WebsocketURL, string(dumpData))
		}
		return err
	}

	wsc.Conn = c

	if wsc.IsDump && resp != nil {
		dumpData, _ := httputil.DumpResponse(resp, true)
		log.Printf("[ws][dump][%s] %s", wsc.WebsocketURL, string(dumpData))
	}

	return nil
}

// Close ...
func (wsc *WebsocketConnection) Close() error {
	close(wsc.CloseChannel)

	err := wsc.Conn.Close()
	if err != nil {
		log.Printf("[ws][%s] Error closing the websocket connection: %s", wsc.WebsocketURL, err.Error())
		return err
	}

	if wsc.IsDump {
		log.Printf("[ws][%s] connection closed", wsc.WebsocketURL)
	}
	return nil
}

// WriteRequest ...
func (wsc *WebsocketConnection) WriteRequest() {
	var err error
	for {
		select {
		case <-wsc.CloseChannel:
			log.Printf("[ws][%s] Exiting the WriteRequest go routine", wsc.WebsocketURL)
			return

		case msg := <-wsc.WriteBufferChannel:
			err = wsc.Conn.WriteMessage(websocket.TextMessage, msg)

		case msg := <-wsc.PingMessageBufferChannel:
			err = wsc.Conn.WriteMessage(websocket.PingMessage, msg)

		case msg := <-wsc.CloseMessageBufferChannel:
			err = wsc.Conn.WriteMessage(websocket.CloseMessage, msg)
		}

		if err != nil {
			log.Printf("[ws][%s] Error writing message: %s", wsc.WebsocketURL, err.Error())
			time.Sleep(1 * time.Second)
		}
	}
}

// ReceiveMessage ...
func (wsc *WebsocketConnection) ReceiveMessage() {
	// CLOSE
	wsc.Conn.SetCloseHandler(func(code int, text string) error {
		log.Printf("[ws][%s] websocket exiting [code=%d, text=%s]", wsc.WebsocketURL, code, text)
		err := wsc.Close()
		return err
	})

	// PONG
	wsc.Conn.SetPongHandler(func(appData string) error {
		log.Printf("[ws][%s] PONG RECEIVE %s", wsc.WebsocketURL, appData)
		wsc.Conn.SetReadDeadline(time.Now().Add(wsc.ReadDeadlineTime))
		return nil
	})

	// PING
	wsc.Conn.SetPingHandler(func(appData string) error {
		wsc.Conn.SetReadDeadline(time.Now().Add(wsc.ReadDeadlineTime))
		err := wsc.Conn.WriteMessage(websocket.PongMessage, nil)
		if err != nil {
			log.Printf("[ws][%s] PING RECEIVE ERROR %s", wsc.WebsocketURL, err.Error())
			return err
		}
		log.Printf("[ws][%s] PING RECEIVED %s", wsc.WebsocketURL, appData)
		return nil
	})

	for {
		select {
		case <-wsc.CloseChannel:
			log.Printf("[ws][%s] Exiting the ReceiveMessage goroutine", wsc.WebsocketURL)
			return
		default:
			msgType, msg, err := wsc.Conn.ReadMessage()
			if err != nil {
				log.Printf("[ws][%s] Error receiving message from the websocket: %s", wsc.WebsocketURL, err.Error())

				if wsc.IsAutoReconnect {
					log.Printf("[ws][%s] Attempting reconnect", wsc.WebsocketURL)
					wsc.Reconnect()
					continue
				}
				// Should this come before the reconnect attempt
				if wsc.ErrorHandleFunc != nil {
					wsc.ErrorHandleFunc(err)
				}

				return
			}

			wsc.Conn.SetReadDeadline(time.Now().Add(wsc.ReadDeadlineTime * time.Second))

			switch msgType {
			case websocket.BinaryMessage:
			case websocket.TextMessage:
				err = wsc.MessageHandleFunc(msg)
				if err != nil {
					log.Printf("[ws][%s] Error processing the message: %s", wsc.WebsocketURL, err.Error())
				}
			case websocket.CloseMessage:
				wsc.Close()
			default:
				log.Printf("[ws][%s] unhandled message type %d. Message: %s", wsc.WebsocketURL, msgType, msg)
			}

		}
	}
}

// Subscribe ...
func (wsc *WebsocketConnection) Subscribe(subscription interface{}) error {
	jsonData, err := json.Marshal(subscription)
	if err != nil {
		log.Printf("[ws][%s] error marshalling subscription info: %s", wsc.WebsocketURL, err.Error())
		return err
	}

	wsc.WriteBufferChannel <- jsonData
	wsc.Subscriptions = append(wsc.Subscriptions, subscription)

	return nil
}

// Reconnect ...
func (wsc *WebsocketConnection) Reconnect() {
	wsc.Conn.Close()

	var err error

	sleep := 0
	for i := 0; i < wsc.ConnectionRetryLimit; i++ {
		time.Sleep(time.Duration(sleep) * time.Second)
		err = wsc.Connect()
		if err != nil {
			log.Printf("[ws][%s] Failed to reconnect: %s", wsc.WebsocketURL, err.Error())
		} else {
			break
		}

		sleep <<= 1
	}

	if err != nil {
		log.Printf("[ws][%s] Failed to reconnect: %s", wsc.WebsocketURL, err.Error())
		wsc.Close()
		if wsc.ErrorHandleFunc != nil {
			wsc.ErrorHandleFunc(errors.New("Failed To Reconnect"))
		}
	} else {
		var subscriptions []interface{}

		copy(subscriptions, wsc.Subscriptions)
		wsc.Subscriptions = []interface{}{}

		for _, subscription := range subscriptions {
			_ = wsc.Subscribe(subscription)
		}
	}

}

// SendMessage ...
func (wsc *WebsocketConnection) SendMessage(msg []byte) {
	wsc.WriteBufferChannel <- msg
}

// SendPingMessage ...
func (wsc *WebsocketConnection) SendPingMessage(msg []byte) {
	wsc.PingMessageBufferChannel <- msg
}

// SendCloseMessage ...
func (wsc *WebsocketConnection) SendCloseMessage(msg []byte) {
	wsc.CloseMessageBufferChannel <- msg
}

// SendJSONMessage ...
func (wsc *WebsocketConnection) SendJSONMessage(msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[ws][%s] Failed to marshal the msg to JSON: %s", wsc.WebsocketURL, err.Error())
		return err
	}
	wsc.WriteBufferChannel <- data
	return nil
}
