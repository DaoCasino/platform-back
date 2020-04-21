package handlers

import (
	"context"
	"platform-backend/server/api/interfaces"
)

func ProcessFetchSessionsRequest(context context.Context, req *interfaces.ApiRequest) (*interfaces.WsResponse, error) {
	return &interfaces.WsResponse{
		Type:    "response",
		Id:      req.Data.Id,
		Status:  "ok",
		Payload: struct{}{},
	}, nil
}
