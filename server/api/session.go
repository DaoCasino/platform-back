package api

import (
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"go.uber.org/atomic"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type OnCloseCb func()

type Session struct {
	// base context
	baseCtx context.Context
	// websocket connection
	wsConn *websocket.Conn
	// closing flag
	closing atomic.Bool
	// on session close callback
	onClose OnCloseCb

	wsApi *WsApi

	subscribed bool

	// send msg to socket chan
	Send chan []byte
}

func (s *Session) close() {
	if !s.closing.Load() {
		_ = s.wsConn.Close()
		s.onClose()
		s.closing.Store(true)
	}
}

func (s *Session) readLoop() {
	_, cancel := context.WithCancel(s.baseCtx)
	defer func() {
		cancel()
		s.close()
	}()

	_ = s.wsConn.SetReadDeadline(time.Now().Add(pongWait))
	s.wsConn.SetReadLimit(maxMessageSize)
	s.wsConn.SetPongHandler(func(string) error { _ = s.wsConn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		select {
		case <-s.baseCtx.Done():
			log.Debug().Msgf("Session readLoop canceled, ip: %s", s.wsConn.RemoteAddr().String())
			return

		default:
			messageType, message, err := s.wsConn.ReadMessage()
			if err != nil {
				log.Debug().Msgf("Websocket read error, disconnection, %s", err.Error())
				return
			}

			resp, err := ProcessRequest(s.baseCtx, s.wsApi, s, messageType, message)
			if err != nil {
				log.Debug().Msgf("Websocket request parsing error, disconnection, %s", err.Error())
				return
			}

			if marshal, err := json.Marshal(resp); err != nil {
				log.Debug().Msgf("Websocket answer marshal error, %s", err.Error())
				return
			} else {
				s.Send <- marshal
			}
		}
	}
}

func (s *Session) writeLoop() {
	_, cancel := context.WithCancel(s.baseCtx)
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		cancel()
		s.close()
	}()

	for {
		select {
		case <-s.baseCtx.Done():
			log.Debug().Msgf("Session writeLoop canceled, ip: %s", s.wsConn.RemoteAddr().String())
			return

		case rawMsg, ok := <-s.Send:
			_ = s.wsConn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				log.Debug().Msgf("Session send channel is closed, disconnection")
				return
			}

			if err := s.wsConn.WriteMessage(websocket.TextMessage, rawMsg); err != nil {
				log.Debug().Msgf("Session write error, disconnection, %s", err.Error())
				return
			}
		case <-ticker.C:
			_ = s.wsConn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := s.wsConn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func NewSession(ctx context.Context, conn *websocket.Conn, onClose OnCloseCb, wsApi *WsApi) *Session {
	session := new(Session)

	session.baseCtx = ctx
	session.wsConn = conn
	session.onClose = onClose
	session.wsApi = wsApi
	session.subscribed = false
	session.Send = make(chan []byte)
	session.closing.Store(false)

	return session
}

func (s *Session) Run() {
	go s.readLoop()
	go s.writeLoop()
}
