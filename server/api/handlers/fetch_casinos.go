package handlers

import (
	"context"
	"platform-backend/server/api/ws_interface"
)

func ProcessFetchCasinosRequest(context context.Context, req *ws_interface.ApiRequest) (interface{}, *ws_interface.HandlerError) {
	casinos, err := req.Repos.Casino.AllCasinos(context)
	if err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	return casinos, nil
}
