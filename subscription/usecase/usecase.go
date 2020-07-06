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
	subscriptions map[uuid.UUID]*Subscription
	sync.Mutex
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
	s.Lock()
	defer s.Unlock()

	for _, subscription := range s.subscriptions {
		if subscription.user.AccountName != user {
			continue
		}

		resp := &ws_interface.WsUpdate{
			Type:    "update",
			Reason:  reason,
			Time:    time.Now().Unix(),
			Payload: payload,
		}

		if marshal, err := json.Marshal(resp); err != nil {
			log.Debug().Msgf("Websocket answer marshal error, %s", err.Error())
			return
		} else {
			subscription.send <- marshal
		}
	}
}
