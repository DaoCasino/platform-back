package subscription

import (
	"github.com/google/uuid"
	"platform-backend/models"
)

type UseCase interface {
	AddSession(uuid uuid.UUID, user *models.User, send chan []byte)
	RemoveSession(uuid uuid.UUID)
	Notify(user string, reason string, payload interface{})
}
