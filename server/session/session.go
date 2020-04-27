package session

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"go.uber.org/atomic"
	"platform-backend/models"
	"platform-backend/server/api"
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
	// session Uuid
	Uuid uuid.UUID
	// session user (nil before auth)
	User *models.User
	// base context
	baseCtx context.Context
	// websocket connection
	wsConn *websocket.Conn
	// closing flag
	closing atomic.Bool
	// on session close callback
	onClose OnCloseCb
	// user ws api
	wsApi *api.WsApi

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
	ctx, cancel := context.WithCancel(s.baseCtx)
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

			// add session id into context
			ctx = context.WithValue(ctx, "suid", s.Uuid)

			// add user info into context
			ctx = context.WithValue(ctx, "user", s.User)

			resp, err := s.wsApi.ProcessRawRequest(ctx, messageType, message)
			if err != nil {
				log.Debug().Msgf("Websocket request fatal error, disconnection, %s", err.Error())
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

func NewSession(ctx context.Context, conn *websocket.Conn, wsApi *api.WsApi, onClose OnCloseCb) *Session {
	session := new(Session)

	session.Uuid, _ = uuid.NewRandom()
	session.baseCtx = ctx
	session.wsConn = conn
	session.onClose = onClose
	session.wsApi = wsApi
	session.User = nil
	session.Send = make(chan []byte)
	session.closing.Store(false)

	return session
}

func (s *Session) Run() {
	go s.readLoop()
	go s.writeLoop()
}
