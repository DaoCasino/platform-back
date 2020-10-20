package server

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/rs/zerolog/log"
	"net/http"
	"platform-backend/auth"
	"platform-backend/models"
)

const (
	TokenExpired = 401
)

type HTTPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func wsClientHandler(app *App, w http.ResponseWriter, r *http.Request) {
	log.Debug().Msgf("New connect request")

	c, err := app.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Err(err)
		return
	}

	log.Info().Msgf("Client with ip %q connected", c.RemoteAddr())

	app.smRepo.AddSession(context.Background(), c, app.wsApi)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, HTTPError{Code: code, Message: message})
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

func authHandler(app *App, w http.ResponseWriter, r *http.Request) {
	var user *models.User
	if app.developmentMode {
		user = &models.User{
			AccountName: "testuserever",
			Email:       "test@user.ever",
			AffiliateID: "afff",
		}
	} else {
		var req AuthRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			log.Debug().Msgf("Http body parse error, %s", err.Error())
			return
		}

		log.Debug().Msgf("New auth request with token %s", req.TmpToken)

		user, err = app.useCases.Auth.ResolveUser(context.Background(), req.TmpToken)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			log.Debug().Msgf("Token validate error: %s", err.Error())
			return
		}
		user.AffiliateID = req.AffiliateID
	}
	refreshToken, accessToken, err := app.useCases.Auth.SignUp(context.Background(), user)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		log.Debug().Msgf("SignUp error: %s", err.Error())
		return
	}

	response, err := json.Marshal(JsonResponse{
		"refreshToken": refreshToken,
		"accessToken":  accessToken,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		log.Debug().Msgf("Response marshal error: %s", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(response)
}

func logoutHandler(app *App, w http.ResponseWriter, r *http.Request) {
	log.Debug().Msgf("New logout request")

	var req LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		log.Debug().Msgf("Http body parse error, %s", err.Error())
		return
	}

	err := app.useCases.Auth.Logout(context.Background(), req.AccessToken)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		log.Debug().Msgf("RefreshToken error: %s", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func refreshTokensHandler(app *App, w http.ResponseWriter, r *http.Request) {
	log.Debug().Msgf("New refresh_token request")

	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		log.Debug().Msgf("Http body parse error, %s", err.Error())
		return
	}

	refreshToken, accessToken, err := app.useCases.Auth.RefreshToken(context.Background(), req.RefreshToken)
	if err != nil {
		log.Debug().Msgf("RefreshToken error: %s", err.Error())
		if errors.Is(err, auth.ErrExpiredToken) {
			respondWithError(w, TokenExpired, err.Error())
			return
		}
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	response, err := json.Marshal(JsonResponse{
		"refreshToken": refreshToken,
		"accessToken":  accessToken,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		log.Debug().Msgf("Response marshal error: %s", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(response)
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
