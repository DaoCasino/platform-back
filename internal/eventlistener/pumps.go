package eventlistener

import (
	"context"
	"github.com/gorilla/websocket"
	"time"
)

const (
	msgPumpStopped       = "pump stopped"
	msgPumpRunning       = "pump running"
	msgParentContextDone = "parent context done"
)

func (e *EventListener) readPump(parentContext context.Context) {
	log := logger.With().Str("gorutine", "readPump").Logger()

	defer func() {
		if err := e.conn.Close(); err != nil {
			log.Error().Err(err).Str("func", "conn.Close").Send()
		}
		log.Info().Msg(msgPumpStopped)
	}()

	log.Info().Msg(msgPumpRunning)

	e.conn.SetReadLimit(e.MaxMessageSize)
	if err := e.conn.SetReadDeadline(time.Now().Add(e.PongWait)); err != nil {
		log.Error().Err(err).Str("func", "conn.SetReadDeadline").Send()
	}
	e.conn.SetPongHandler(func(string) error { return e.conn.SetReadDeadline(time.Now().Add(e.PongWait)) })

	for {
		select {
		case <-parentContext.Done():
			log.Debug().Msg(msgParentContextDone)
			return
		default:
			_, message, err := e.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Error().Err(err).Str("func", "conn.ReadMessage").Send()
				}
				return
			}

			if err := e.processMessage(message); err != nil {
				log.Error().Err(err).Str("func", "processMessage").Send()
			}
		}
	}
}

func (e *EventListener) responsePump(ctx context.Context, send <-chan *responseQueue) {
	log := logger.With().Str("gorutine", "responsePump").Logger()

	process := make(map[string]chan *responseMessage)
	defer func() {
		for ID, ch := range process {
			if ch != nil {
				close(ch)
			}
			delete(process, ID)
		}

		log.Info().Msg(msgPumpStopped)
	}()

	log.Info().Msg(msgPumpStopped)
	for {
		select {
		case <-ctx.Done():
			log.Debug().Msg(msgParentContextDone)
			return
		case message, ok := <-send:
			if !ok {
				log.Debug().Msg("close send channel")
				return
			}
			if message.response != nil { // Add wait response
				process[message.ID] = message.response
			}

		case response, ok := <-e.response:
			if !ok {
				log.Debug().Msg("close response channel")
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
	log := logger.With().Str("gorutine", "writePump").Logger()

	ticker := time.NewTicker(e.PingPeriod)
	waitResponse := make(chan *responseQueue)

	go e.responsePump(parentContext, waitResponse)

	defer func() {
		close(waitResponse)

		if e.event != nil {
			close(e.event) // <- events can not be expected
			log.Debug().Msg("close event channel")
		}

		ticker.Stop()
		e.conn.Close()

		log.Info().Msg(msgPumpStopped)
	}()

	log.Info().Msg(msgPumpRunning)

	for {
		select {
		case <-parentContext.Done():
			log.Debug().Msg(msgParentContextDone)
			return

		case message, ok := <-e.send:
			if !ok {
				// The session closed the channel.
				log.Debug().Msg("close send channel")
				if err := closeMessage(e.conn, e.WriteWait); err != nil {
					log.Error().Err(err).Str("func", "closeMessage").Send()
				}
				return
			}
			err := writeMessage(e.conn, e.WriteWait, message.message)
			if err != nil {
				log.Error().Err(err).Str("func", "writeMessage").Send()
				return
			}
			if message.response != nil {
				waitResponse <- message
			}
		case <-ticker.C:
			if err := pingMessage(e.conn, e.WriteWait); err != nil {
				log.Error().Err(err).Str("func", "pingMessage").Send()
			}
		}
	}
}
