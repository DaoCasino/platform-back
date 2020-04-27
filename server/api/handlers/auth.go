package handlers

import (
	"context"
	"encoding/json"
	"platform-backend/server/api/ws_interface"
)

type AuthPayload struct {
	Token string `json:"token"`
}

func ProcessAuthRequest(context context.Context, req *ws_interface.ApiRequest) (interface{}, *ws_interface.HandlerError) {
	var payload AuthPayload
	if err := json.Unmarshal(req.Data.Payload, &payload); err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	user, err := req.UseCases.Auth.SignIn(context, payload.Token)
	if err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.AuthCheckError, err)
	}

	return user, nil
}
