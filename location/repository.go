package location

import (
	"net"
	"platform-backend/models"
)

type Repository interface {
	GetLocation(ip net.IP) (*models.Location, error)
}
