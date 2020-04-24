package handlers

import (
	"context"
	"encoding/json"
	"platform-backend/server/api/interfaces"
)

type AuthPayload struct {
	Token string `json:"token"`
}

func ProcessAuthRequest(context context.Context, req *interfaces.ApiRequest) (interface{}, *interfaces.HandlerError) {
	var payload AuthPayload
	if err := json.Unmarshal(req.Data.Payload, &payload); err != nil {
		return nil, interfaces.NewHandlerError(interfaces.InternalError, err)
	}

	user, err := req.UseCases.Auth.SignIn(context, payload.Token)
	if err != nil {
		return nil, interfaces.NewHandlerError(interfaces.AuthCheckError, err)
	}

	return user, nil
}
