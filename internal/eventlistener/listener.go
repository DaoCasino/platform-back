package eventlistener

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"net/url"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type EventListener struct {
	Addr           string        // TCP address to listen.
	MaxMessageSize int64         // Maximum message size allowed from client.
	WriteWait      time.Duration // Time allowed to write a message to the client.
	PongWait       time.Duration // Time allowed to read the next pong message from the peer.
	PingPeriod     time.Duration // Send pings to peer with this period. Must be less than pongWait.

	conn  *websocket.Conn
	event chan<- *Event
	send  chan []byte // Buffered channel of outbound messages.
}

func NewEventListener(addr string, event chan<- *Event) *EventListener {
	return &EventListener{
		Addr:           addr,
		MaxMessageSize: maxMessageSize,
		WriteWait:      writeWait,
		PongWait:       pongWait,
		PingPeriod:     pingPeriod,
		event:          event,
		send:           make(chan []byte, 512),
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

func (e *EventListener) Subscribe(eventType int, offset uint64) error {
	// if error need cancel listen
	return nil
}

func (e *EventListener) Unsubscribe(eventType int) error {
	// if error need cancel listen
	return nil
}
