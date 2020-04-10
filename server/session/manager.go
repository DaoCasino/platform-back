package session

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"platform-backend/server/api"
	"sync"
)

type Manager struct {
	sync.Mutex
	sessions map[string]*Session
	wsApi    *api.WsApi
}

func (s *Manager) removeSession(userName string) {
	s.Lock()
	defer s.Unlock()

	delete(s.sessions, userName)
}

func NewSessionManager(wsApi *api.WsApi) *Manager {
	manager := new(Manager)
	manager.wsApi = wsApi
	manager.sessions = make(map[string]*Session)

	return manager
}

func (s *Manager) NewConnection(userName string, wsConn *websocket.Conn) {
	s.Lock()
	defer func() { // prevent potential dead-lock
		s.Unlock()
		s.sessions[userName].Run()
	}()

	s.sessions[userName] = NewSession(context.Background(), wsConn, s.wsApi, func() {
		s.removeSession(userName)
	})

	log.Debug().Msgf("New session started, user: %s", userName)
}

func (s *Manager) HasSession(userName string) bool {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.sessions[userName]; ok {
		return true
	}
	return false
}
