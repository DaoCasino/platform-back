package eventlistener

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"net/url"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024 * 4
)

type EventListener struct {
	Addr           string        // TCP address to listen.
	MaxMessageSize int64         // Maximum message size allowed from client.
	WriteWait      time.Duration // Time allowed to write a message to the client.
	PongWait       time.Duration // Time allowed to read the next pong message from the peer.
	PingPeriod     time.Duration // Send pings to peer with this period. Must be less than pongWait.

	conn    *websocket.Conn
	event   chan<- *EventMessage
	send    chan *responseQueue
	process map[string]chan *responseMessage
}

func NewEventListener(addr string, event chan<- *EventMessage) *EventListener {
	return &EventListener{
		Addr:           addr,
		MaxMessageSize: maxMessageSize,
		WriteWait:      writeWait,
		PongWait:       pongWait,
		PingPeriod:     pingPeriod,

		event:   event,
		send:    make(chan *responseQueue, 512),
		process: make(map[string]chan *responseMessage),
	}
}

func (e *EventListener) ListenAndServe(ctx context.Context) error {
	u := url.URL{Scheme: "ws", Host: e.Addr, Path: "/"}
	log.Info().Msgf("event listener connecting to %s", u.String())

	var err error
	e.conn, _, err = websocket.DefaultDialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return err
	}

	go e.readPump(ctx)
	go e.writePump(ctx)
	return nil
}

func (e *EventListener) Subscribe(eventType EventType, offset uint64) (bool, error) {
	params := struct {
		Topic  string
		Offset uint64
	}{
		eventType.ToString(),
		offset,
	}

	request := newRequestMessage(methodSubscribe, params)
	response, err := e.sendRequest(request)
	if err != nil {
		return false, err
	}

	if response.Error != nil {
		return false, errors.New(response.Error.Message)
	}

	result := false
	err = json.Unmarshal(response.Result, &result)
	return result, err
}

func (e *EventListener) Unsubscribe(eventType EventType) (bool, error) {
	params := struct {
		Topic string
	}{
		eventType.ToString(),
	}

	request := newRequestMessage(methodUnsubscribe, params)
	response, err := e.sendRequest(request)
	if err != nil {
		return false, err
	}

	if response.Error != nil {
		return false, errors.New(response.Error.Message)
	}

	result := false
	err = json.Unmarshal(response.Result, &result)
	return result, err
}
