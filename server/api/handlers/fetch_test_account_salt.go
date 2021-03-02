package handlers

import (
	"context"
	"platform-backend/server/api/ws_interface"
)

type TestAccountSaltResponse struct {
	Salt uint64 `json:"salt"`
}

func ProcessFetchTestAccountSaltRequest(context context.Context, req *ws_interface.ApiRequest) (interface{}, *ws_interface.HandlerError) {
	testAccountSalt := req.Repos.User.GetTestAccountSalt(context)
	return TestAccountSaltResponse{Salt: testAccountSalt}, nil
}
