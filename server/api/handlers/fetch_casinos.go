package handlers

import (
	"context"
	"platform-backend/server/api/interfaces"
)


func ProcessFetchCasinosRequest(context context.Context, req *interfaces.ApiRequest) (*interfaces.WsResponse, error) {
	casinos, err := req.UseCases.Casino.AllCasinos(context)
	if err != nil {
		return nil, err
	}

	return &interfaces.WsResponse{
		Type:    "response",
		Id:      req.Data.Id,
		Status:  "ok",
		Payload: casinos,
	}, nil
}
