package api

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"sync"
)

type SessionManager struct {
	sync.Mutex
	sessions map[string]*Session
	wsApi    *WsApi
}

func (s *SessionManager) removeSession(userName string) {
	s.Lock()
	defer s.Unlock()

	delete(s.sessions, userName)
}

func NewSessionManager(wsApi *WsApi) *SessionManager {
	manager := new(SessionManager)
	manager.wsApi = wsApi
	manager.sessions = make(map[string]*Session)

	return manager
}

func (s *SessionManager) NewConnection(userName string, wsConn *websocket.Conn) {
	s.Lock()
	defer s.Unlock()

	s.sessions[userName] = NewSession(context.Background(), wsConn, func() {
		s.removeSession(userName)
	}, s.wsApi)

	s.sessions[userName].Run()

	log.Debug().Msgf("New session started, user: %s", userName)
}

func (s *SessionManager) HasSession(userName string) bool {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.sessions[userName]; ok {
		return true
	}
	return false
}
