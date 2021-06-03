package maxmind

import (
	"net"
	"platform-backend/models"

	"github.com/oschwald/geoip2-golang"
	"github.com/rs/zerolog/log"
)

type LocationMaxmindRepo struct {
	db *geoip2.Reader
}

func NewLocationMaxmindRepo(file string) (*LocationMaxmindRepo, error) {
	db, err := geoip2.Open(file)
	if err != nil {
		return nil, err
	}
	return &LocationMaxmindRepo{db: db}, nil
}

func (r *LocationMaxmindRepo) GetLocation(ip net.IP) (*models.Location, error) {
	record, err := r.db.City(ip)
	if err != nil {
		return nil, err
	}

	var result models.Location
	result.Country.IsoCode = record.Country.IsoCode
	result.Location.Latitude = record.Location.Latitude
	result.Location.Longitude = record.Location.Longitude

	return &result, nil
}

func DeleteLocationMaxmindRepo(r *LocationMaxmindRepo) error {
	log.Info().Msg("Location database closed")
	return r.db.Close()
}
