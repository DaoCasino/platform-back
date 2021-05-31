package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"platform-backend/auth"
	"platform-backend/models"
	"platform-backend/utils"

	"github.com/rs/zerolog/log"
)

const (
	TokenExpired = 401
)

type HTTPResponse struct {
	Response interface{} `json:"response"`
	Error    *HTTPError  `json:"error"`
}

type HTTPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func wsClientHandler(ctx context.Context, app *App, w http.ResponseWriter, r *http.Request) {
	log.Debug().Msgf("New connect request")

	c, err := app.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Err(err)
		return
	}

	log.Info().Msgf("Client with ip %s connected", utils.GetIPFromRequest(r))

	// Save IP into context
	ctx = utils.SetContextRemoteAddr(ctx, utils.GetIPFromRequest(r))

	app.smRepo.AddSession(ctx, c, app.wsApi)
}

func respondOK(w http.ResponseWriter, response interface{}) {
	respondWithJSON(w, http.StatusOK, HTTPResponse{
		Response: response,
		Error:    nil,
	})
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, HTTPResponse{
		Response: nil,
		Error:    &HTTPError{Code: code, Message: message},
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err := w.Write(response)
	if err != nil {
		log.Warn().Msg("Failed to respond to client")
	}
}

func authHandler(ctx context.Context, app *App, w http.ResponseWriter, r *http.Request) {
	var (
		user       *models.User
		casinoName string
	)
	if app.developmentMode {
		user = &models.User{
			AccountName: "testuserever",
			Email:       "test@user.ever",
			AffiliateID: "afff",
		}
		casinoName = "casino"
	} else {
		var req AuthRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			log.Debug().Msgf("Http body parse error, %s", err.Error())
			return
		}

		log.Debug().Msgf("New auth request with token %s", req.TmpToken)

		user, err = app.useCases.Auth.ResolveUser(ctx, req.TmpToken)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, err.Error())
			log.Warn().Msgf("Token validate error: %s", err.Error())
			return
		}
		user.AffiliateID = req.AffiliateID
		casinoName = req.CasinoName
	}
	refreshToken, accessToken, err := app.useCases.Auth.SignUp(ctx, user, casinoName)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		log.Warn().Msgf("SignUp error: %s", err.Error())
		return
	}

	response := JsonResponse{
		"refreshToken": refreshToken,
		"accessToken":  accessToken,
	}

	respondOK(w, response)
}

func logoutHandler(ctx context.Context, app *App, w http.ResponseWriter, r *http.Request) {
	log.Debug().Msgf("New logout request")

	var req LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		log.Debug().Msgf("Http body parse error, %s", err.Error())
		return
	}

	err := app.useCases.Auth.Logout(ctx, req.AccessToken)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		log.Debug().Msgf("RefreshToken error: %s", err.Error())
		return
	}

	respondOK(w, true)
}

func refreshTokensHandler(ctx context.Context, app *App, w http.ResponseWriter, r *http.Request) {
	log.Debug().Msgf("New refresh_token request")

	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		log.Debug().Msgf("Http body parse error, %s", err.Error())
		return
	}

	refreshToken, accessToken, err := app.useCases.Auth.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		log.Warn().Msgf("RefreshToken error: %s", err.Error())
		if errors.Is(err, auth.ErrExpiredToken) || errors.Is(err, auth.ErrExpiredTokenNonce) {
			respondWithError(w, TokenExpired, err.Error())
			return
		}
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	response := JsonResponse{
		"refreshToken": refreshToken,
		"accessToken":  accessToken,
	}

	respondOK(w, response)
}

func optOutHandler(ctx context.Context, app *App, w http.ResponseWriter, r *http.Request) {
	log.Debug().Msgf("New opt-out request")

	var req OptOutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		log.Debug().Msgf("Http body parse error, %s", err.Error())
		return
	}

	if err := app.useCases.Auth.OptOut(ctx, req.AccessToken); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		log.Debug().Msgf("Opt-out error: %s", err.Error())
		return
	}

	respondOK(w, true)
}

func setEthAddrHandler(ctx context.Context, app *App, w http.ResponseWriter, r *http.Request) {
	log.Debug().Msgf("New set eth addr request")

	var req SetEthAddrRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		log.Debug().Msgf("Http body parse error, %s", err.Error())
		return
	}
	accountName, err := app.useCases.Auth.AccountNameFromToken(ctx, req.AccessToken)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		log.Debug().Msgf("Parse token error: %s", err.Error())
		return
	}

	if err := app.useCases.Cashback.SetEthAddress(ctx, accountName, req.EthAddress); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		log.Debug().Msgf("Set eth addr error: %s", err.Error())
		return
	}

	respondOK(w, true)
}

func pingHandler(w http.ResponseWriter, _ *http.Request) {
	log.Debug().Msgf("New ping request")
	w.WriteHeader(http.StatusOK)
}

func whoHandler(w http.ResponseWriter, _ *http.Request) {
	log.Debug().Msgf("New who request")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(ServiceName))
	if err != nil {
		log.Debug().Msgf("Failed to response /who")
	}
}
