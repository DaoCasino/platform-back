package handlers

import (
	"context"
	"github.com/rs/zerolog/log"
	"platform-backend/server/api/interfaces"
)

func ProcessAccountInfo(context context.Context, req *interfaces.ApiRequest) (*interfaces.WsResponse, error) {
	player, err := req.UseCases.Player.GetInfo(context, req.User.AccountName)

	if err != nil {
		log.Debug().Msgf("Account info fetch error: %s", err.Error())
		return &interfaces.WsResponse{
			Type:   "response",
			Id:     req.Data.Id,
			Status: "error",
			Payload: interfaces.WsError{
				Code:    5001,
				Message: "cannot fetch account info",
			},
		}, nil
	}



	return &interfaces.WsResponse{
		Type:   "response",
		Id:     req.Data.Id,
		Status: "ok",
		Payload: player,
	}, nil
}
