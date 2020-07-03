package handlers

import (
	"context"
	"platform-backend/server/api/ws_interface"
)

func ProcessSubscribeRequest(context context.Context, req *ws_interface.ApiRequest) (interface{}, *ws_interface.HandlerError) {
	return struct{}{}, nil
}
