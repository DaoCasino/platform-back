package models

type User struct {
	AccountName string `json:"accountName"`
	Email       string `json:"email"`
	AffiliateID string `json:"affiliateID,omitempty"`
}
