package handlers

import (
	"context"
	"encoding/json"
	"platform-backend/server/api/ws_interface"
)

type AuthTestAccountPayload struct {
	AccountName           string `json:"accountName"`
	SaltedAccountNameHash string `json:"saltedAccountNameHash"`
}

func ProcessAuthTestAccountRequest(context context.Context, req *ws_interface.ApiRequest) (interface{}, *ws_interface.HandlerError) {
	var payload AuthTestAccountPayload
	if err := json.Unmarshal(req.Data.Payload, &payload); err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.RequestParseError, err)
	}

	user, err := req.UseCases.Auth.SignInTestAccount(context, payload.AccountName, payload.SaltedAccountNameHash)
	if err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.AuthCheckError, err)
	}

	return user, nil
}
