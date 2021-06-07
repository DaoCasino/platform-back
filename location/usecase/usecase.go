package usecase

import (
	"context"
	"errors"
	"net"
	"platform-backend/location"
	"platform-backend/models"
)

type LocationUseCase struct {
	locationRepo location.Repository
}

var errNotValidIP = errors.New("not valid IP address")

func NewLocationUseCase(repo location.Repository) *LocationUseCase {
	return &LocationUseCase{
		locationRepo: repo,
	}
}

func (c *LocationUseCase) GetLocationFromIP(ctx context.Context, ip string) (*models.Location, error) {
	IP := net.ParseIP(ip)
	if IP == nil {
		return nil, errNotValidIP
	}
	return c.locationRepo.GetLocation(IP)
}
