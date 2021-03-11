package binancewebsocket

// LiveRequest ...
type LiveRequest struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
	ID     int64    `json:"id"`
}
