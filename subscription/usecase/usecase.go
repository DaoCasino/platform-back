package usecase

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"platform-backend/models"
	"platform-backend/server/api/ws_interface"
	"sync"
	"time"
)

type Subscription struct {
	uuid uuid.UUID
	user *models.User
	send chan []byte
}

type SubscriptionUseCase struct {
	sync.Mutex
	subscriptions map[uuid.UUID]*Subscription
}

func NewSubscriptionUseCase() *SubscriptionUseCase {
	return &SubscriptionUseCase{
		subscriptions: make(map[uuid.UUID]*Subscription),
	}
}

func (s *SubscriptionUseCase) AddSession(uuid uuid.UUID, user *models.User, send chan []byte) {
	s.Lock()
	defer s.Unlock()

	s.subscriptions[uuid] = &Subscription{
		uuid: uuid,
		user: user,
		send: send,
	}
}

func (s *SubscriptionUseCase) RemoveSession(uuid uuid.UUID) {
	s.Lock()
	defer s.Unlock()

	delete(s.subscriptions, uuid)
}

func (s *SubscriptionUseCase) Notify(user string, reason string, payload interface{}) {
	resp := &ws_interface.WsUpdate{
		Type:    "update",
		Reason:  reason,
		Time:    time.Now().Unix(),
		Payload: payload,
	}

	marshal, err := json.Marshal(resp)

	if err != nil {
		log.Debug().Msgf("Websocket answer marshal error, %s", err.Error())
		return
	}

	s.Lock()
	defer s.Unlock()

	for _, subscription := range s.subscriptions {
		if subscription.user.AccountName != user {
			continue
		}

		// Select to prevent lock when send channel is not listening
		select {
		case subscription.send <- marshal:
		default:
			log.Info().Msgf("Subscribe: notify called when send channel is dead")
		}
	}
}
