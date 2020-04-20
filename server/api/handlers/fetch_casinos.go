package handlers

import (
	"context"
	"platform-backend/server/api/interfaces"
)


func ProcessFetchCasinosRequest(context context.Context, req *interfaces.ApiRequest) (*interfaces.WsResponse, error) {
	casinos, err := req.Repos.Casino.AllCasinos(context)
	if err != nil {
		return &interfaces.WsResponse{
			Type:   "response",
			Id:     req.Data.Id,
			Status: "error",
			Payload: interfaces.WsError{
				Code:    5000,
				Message: "Casinos fetch error: " + err.Error(),
			},
		}, nil
	}

	return &interfaces.WsResponse{
		Type:    "response",
		Id:      req.Data.Id,
		Status:  "ok",
		Payload: casinos,
	}, nil
}
