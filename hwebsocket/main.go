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
	Subs                      []interface{}
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
	return &WebsocketConnection{
		WebsocketConfiguration: *wsb.wsConfig,
	}
	// PENDING
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
// func (wsc *WebsocketConnection) InitialiseConnection() *WebsocketConnection {
// 	wsc.ReadDeadlineTime = time.Minute

// }

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
func (wsc *WebsocketConnection) Close() {
	close(wsc.CloseChannel)

	err := wsc.Conn.Close()
	if err != nil {
		log.Printf("[ws][%s] Error closing the websocket connection: %s", wsc.WebsocketURL, err.Error())
	}

	if wsc.IsDump {
		log.Printf("[ws][%s] connection closed", wsc.WebsocketURL)
	}
}

// Subscribe ...
func (wsc *WebsocketConnection) Subscribe(sub interface{}) error {
	jsonData, err := json.Marshal(sub)
	if err != nil {
		log.Printf("[ws][%s] error encoding subscription info: %s", wsc.WebsocketURL, err.Error())
		return err
	}

	wsc.WriteBufferChannel <- jsonData
	wsc.Subs = append(wsc.Subs, sub)

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
		var subs []interface{}
		copy(subs, wsc.Subs)
		wsc.Subs = []interface{}{}
		for _, sub := range subs {
			_ = wsc.Subscribe(sub)
		}
	}

}
