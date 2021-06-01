package models

type Location struct {
	Country struct {
		IsoCode string `json:"iso_code"`
	} `json:"country"`
	Location struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"location"`
}
