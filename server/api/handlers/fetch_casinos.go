package handlers

import (
	"context"
	"platform-backend/models"
	"platform-backend/server/api/ws_interface"
	"strconv"
)

type CasinoResponse struct {
	Id       string `json:"id"`
	Contract string `json:"contract"`
	Paused   bool   `json:"paused"`
}

func toCasinoResponse(c *models.Casino) *CasinoResponse {
	return &CasinoResponse{
		Id:       strconv.FormatUint(c.Id, 10),
		Contract: c.Contract,
		Paused:   c.Paused,
	}
}

func ProcessFetchCasinosRequest(context context.Context, req *ws_interface.ApiRequest) (interface{}, *ws_interface.HandlerError) {
	casinos, err := req.Repos.Contracts.AllCasinos(context)
	if err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	var response []*CasinoResponse
	for _, casino := range casinos {
		response = append(response, toCasinoResponse(casino))
	}

	return response, nil
}
