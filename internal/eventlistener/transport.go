package eventlistener

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"time"
)

func (e *EventListener) readPump(parentContext context.Context) {
	defer func() {
		log.Debug().Msg("event listener read stop")
		e.conn.Close()
	}()

	log.Debug().Msg("event listener read start")

	e.conn.SetReadLimit(e.MaxMessageSize)
	e.conn.SetReadDeadline(time.Now().Add(e.PongWait))
	e.conn.SetPongHandler(func(string) error { return e.conn.SetReadDeadline(time.Now().Add(e.PongWait)) })

	for {
		select {
		case <-parentContext.Done():
			return
		default:
			_, message, err := e.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Error().Msgf("event listener socket error: %s", err.Error())
				}
				return
			}

			if err := e.processMessage(message); err != nil {
				// TODO: log error
			}
		}
	}
}

func (e *EventListener) responsePump(ctx context.Context, send <-chan *responseQueue) {
	process := make(map[string]chan *responseMessage)
	defer func() {
		for ID, ch := range process {
			if ch != nil {
				close(ch)
			}
			delete(process, ID)
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case message, ok := <-send:
			if !ok {
				return
			}
			if message.response != nil { // Add wait response
				process[message.ID] = message.response
			}

		case response, ok := <-e.response:
			if !ok {
				return
			}
			ID := *response.ID
			if ch, ok := process[ID]; ok {
				if ch != nil {
					ch <- response
					close(ch)
				}
				delete(process, ID)
			}
		}
	}
}

func (e *EventListener) writePump(parentContext context.Context) {
	ticker := time.NewTicker(e.PingPeriod)
	waitResponse := make(chan *responseQueue)

	go e.responsePump(parentContext, waitResponse)

	defer func() {
		close(waitResponse)
		ticker.Stop()
		e.conn.Close()

		log.Debug().Msg("event listener write stop")
	}()

	log.Debug().Msg("event listener write start")

	for {
		select {
		case <-parentContext.Done():
			return

		case message, ok := <-e.send:
			if !ok {
				// The session closed the channel.
				closeMessage(e.conn, e.WriteWait)
				return
			}
			err := writeMessage(e.conn, e.WriteWait, message.message)
			if err != nil {
				// TODO: log
				return
			}
			if message.response != nil {
				waitResponse <- message
			}
		case <-ticker.C:
			pingMessage(e.conn, e.WriteWait)
		}
	}
}

func pingMessage(conn *websocket.Conn, writeWait time.Duration) error {
	if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		return err
	}
	return conn.WriteMessage(websocket.PingMessage, nil)
}

func closeMessage(conn *websocket.Conn, writeWait time.Duration) error {
	if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		return err
	}
	return conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}

func writeMessage(conn *websocket.Conn, writeWait time.Duration, message []byte) error {
	if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		return err
	}

	w, err := conn.NextWriter(websocket.TextMessage)
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
