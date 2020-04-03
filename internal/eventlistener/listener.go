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

	conn         *websocket.Conn
	cancelListen context.CancelFunc
	ctx          context.Context
}

func NewEventListener(addr string, event <-chan *Event) *EventListener {
	return &EventListener{
		Addr:           addr,
		MaxMessageSize: maxMessageSize,
		WriteWait:      writeWait,
		PongWait:       pongWait,
		PingPeriod:     pingPeriod,
	}
}

func (e *EventListener) ListenAndServe(ctx context.Context) (err error) {
	u := url.URL{Scheme: "ws", Host: e.Addr, Path: "/"}
	log.Info().Msgf("event listener connecting to %s", u.String())

	e.conn, _, err = websocket.DefaultDialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return
	}

	e.ctx, e.cancelListen = context.WithCancel(ctx)
	ticker := time.NewTicker(e.PingPeriod)

	defer func() {
		ticker.Stop()
		e.cancelListen()

		closeError := e.close()
		if err == nil {
			err = closeError
		}
	}()

	e.conn.SetReadLimit(e.MaxMessageSize)
	err = e.conn.SetReadDeadline(time.Now().Add(e.PongWait))
	if err != nil {
		return
	}
	e.conn.SetPongHandler(func(string) error { return e.conn.SetReadDeadline(time.Now().Add(e.PongWait)) })

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			err = e.conn.SetWriteDeadline(time.Now().Add(e.WriteWait))
			if err != nil {
				return
			}
			err = e.conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				return
			}
		default:
			var message []byte
			_, message, err = e.conn.ReadMessage()
			if err != nil {
				return
			}
			err = e.processMessage(message)
			if err != nil {
				return
			}
		}
	}
}

func (e *EventListener) processMessage(message []byte) error {
	return nil
}

func (e *EventListener) close() error {
	err := e.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return err
	}

	return e.conn.Close()
}

func (e *EventListener) write(message []byte) error {
	if err := e.conn.SetWriteDeadline(time.Now().Add(e.WriteWait)); err != nil {
		return err
	}

	w, err := e.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}

	_, errWrite := w.Write(message)
	err = w.Close()

	if errWrite != nil {
		return errWrite
	}

	return err
}

func (e *EventListener) Subscribe(eventType int, offset uint64) error {
	// if error need cancel listen
	return nil
}

func (e *EventListener) Unsubscribe(eventType int) error {
	// if error need cancel listen
	return nil
}
