package handlers

import (
	"context"
	"platform-backend/server/api/interfaces"
)

func ProcessFetchCasinosRequest(context context.Context, req *interfaces.ApiRequest) (interface{}, *interfaces.HandlerError) {
	casinos, err := req.Repos.Casino.AllCasinos(context)
	if err != nil {
		return nil, interfaces.NewHandlerError(interfaces.InternalError, err)
	}

	return casinos, nil
}
