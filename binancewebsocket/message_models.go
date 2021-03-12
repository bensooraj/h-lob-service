package binancewebsocket

// LiveRequest ...
type LiveRequest struct {
	ID     int64    `json:"id"`
	Method string   `json:"method"`
	Params []string `json:"params"`
}

// LiveResponse ...
type LiveResponse struct {
	ID           int64       `json:"id,omitempty"`
	Result       interface{} `json:"result,omitempty"`
	ErrorCode    int         `json:"code,omitempty"`
	ErrorMessage string      `json:"msg,omitempty"`
}

// DepthUpdate ...
type DepthUpdate struct {
	EventType            string     `json:"e"`
	EventTime            int64      `json:"E"`
	TransactionTime      int64      `json:"T"`
	Symbol               string     `json:"s"`
	FirstUpdateID        int64      `json:"U"`
	LastUpdateID         int64      `json:"u"`
	PreviousLastUpdateID int64      `json:"pu"`
	BidDepthDelta        [][]string `json:"b"`
	AskDepthDelta        [][]string `json:"a"`
}
