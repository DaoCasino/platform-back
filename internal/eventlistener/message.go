package eventlistener

import (
	"encoding/json"
	"github.com/lucsky/cuid"
	"github.com/rs/zerolog/log"
)

const (
	methodSubscribe   = "subscribe"
	methodUnsubscribe = "unsubscribe"
)

type requestMessage struct {
	ID     string      `json:"id"`
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

func (req *requestMessage) toJSON() ([]byte, error) {
	return json.Marshal(req)
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

type responseQueue struct {
	ID       string
	message  []byte
	response chan *responseMessage
}

func newResponseQueue(ID string, message []byte) *responseQueue {
	return &responseQueue{
		ID:       ID,
		message:  message,
		response: make(chan *responseMessage),
	}
}

func newRequestMessage(method string, params interface{}) *requestMessage {
	return &requestMessage{
		ID:     cuid.New(),
		Method: method,
		Params: params,
	}
}

func (e *EventListener) sendRequest(req *requestMessage) (*responseMessage, error) {
	message, err := req.toJSON()
	if err != nil {
		return nil, err
	}
	log.Debug().Msgf("%+v", req)

	wait := newResponseQueue(req.ID, message)
	e.send <- wait

	return <-wait.response, nil
}

func (e *EventListener) processMessage(message []byte) error {
	response := new(responseMessage)
	if err := json.Unmarshal(message, response); err != nil {
		return err
	}

	if response.ID != nil {
		if ch, ok := e.process[*response.ID]; ok {
			if ch != nil {
				ch <- response
				close(ch)
			}
			delete(e.process, *response.ID)
		}
	} else {
		eventMessage := new(EventMessage)
		if err := json.Unmarshal(response.Result, eventMessage); err != nil {
			return err
		}
		e.event <- eventMessage
	}
	return nil
}
