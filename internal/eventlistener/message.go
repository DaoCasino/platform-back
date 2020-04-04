package eventlistener

import "encoding/json"

type requestMessage struct {
	ID     *string         `json:"id"`
	Method *string         `json:"method"`
	Params json.RawMessage `json:"params"`
}

type responseErrorMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type responseMessage struct {
	ID     *string               `json:"id"`
	Result json.RawMessage       `json:"result"`
	Error  *responseErrorMessage `json:"error"`
}

type eventMessage struct {
	Offset uint64   `json:"offset"` // last event.offset
	Events []*Event `json:"events"`
}

func (e *EventListener) processMessage(message []byte) error {
	return nil
}
