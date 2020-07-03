package usecase

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"platform-backend/models"
	"platform-backend/server/api/ws_interface"
	"time"
)

type Subscription struct {
	uuid uuid.UUID
	user *models.User
	send chan []byte
}

type SubscriptionUseCase struct {
	subscriptions []*Subscription
}

func NewSubscriptionUseCase() *SubscriptionUseCase {
	return &SubscriptionUseCase{
		subscriptions: make([]*Subscription, 0),
	}
}

func (s *SubscriptionUseCase) AddSession(uuid uuid.UUID, user *models.User, send chan []byte) {
	s.subscriptions = append(s.subscriptions, &Subscription{
		uuid: uuid,
		user: user,
		send: send,
	})
}

func (s *SubscriptionUseCase) RemoveSession(uuid uuid.UUID) {
	for i := 0; i < len(s.subscriptions); i++ {
		if s.subscriptions[i].uuid == uuid {
			s.subscriptions = append(s.subscriptions[:i], s.subscriptions[i+1:]...)
			return
		}
	}
}

func (s *SubscriptionUseCase) Notify(user string, reason string, payload interface{}) {
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
