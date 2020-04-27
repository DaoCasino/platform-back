package localstorage

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"platform-backend/models"
	"platform-backend/server/api"
	"platform-backend/server/session"
	"sync"
)

type LocalRepository struct {
	sync.Mutex
	// main sessions registry
	sessionById map[uuid.UUID]*session.Session
	// session by account name map
	sessionByUser map[string]*session.Session
}

func NewLocalRepository() *LocalRepository {
	return &LocalRepository{
		sessionById:   make(map[uuid.UUID]*session.Session),
		sessionByUser: make(map[string]*session.Session),
	}
}

func (r *LocalRepository) AddSession(context context.Context, wsConn *websocket.Conn, wsApi *api.WsApi) {
	var sess *session.Session

	defer func() { // prevent potential dead-lock
		r.Unlock()
		sess.Run()
	}()

	sess = session.NewSession(context, wsConn, wsApi, func() {
		r.removeSession(sess.Uuid)
	})

	r.Lock()
	r.sessionById[sess.Uuid] = sess

	log.Debug().Msgf("New sess started, uid: %s", sess.Uuid.String())
}

func (r *LocalRepository) removeSession(uid uuid.UUID) {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.sessionById[uid]; !ok {
		// session doesn't exists
		return
	}

	// remove from by user map
	if r.sessionById[uid].User != nil {
		delete(r.sessionByUser, r.sessionById[uid].User.AccountName)
	}

	// remove from main map
	delete(r.sessionById, uid)

	log.Debug().Msgf("Session closed, uid: %s", uid)
}

func (r *LocalRepository) HasSessionByUser(accountName string) bool {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.sessionByUser[accountName]; ok {
		return true
	}
	return false
}

func (r *LocalRepository) SetUser(uid uuid.UUID, user *models.User) error {
	if _, ok := r.sessionById[uid]; !ok {
		return errors.New("session not found")
	}

	// set user info
	r.sessionById[uid].User = user
	return nil
}
