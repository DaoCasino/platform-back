package server

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
	"net/http"
)

type CashbackClaimRequest struct {
	AccessToken string `json:"accessToken"`
}

func claimHandler(app *App, w http.ResponseWriter, r *http.Request) {
	log.Debug().Msgf("New cashback claim request")

	var req CashbackClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		log.Debug().Msgf("Http body parse error, %s", err.Error())
		return
	}
	ctx := r.Context()
	accountName, err := app.useCases.Auth.AccountNameFromToken(ctx, req.AccessToken)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		log.Debug().Msgf("Parse token error: %s", err.Error())
		return
	}

	if err := app.useCases.Cashback.SetStateClaim(ctx, accountName); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		log.Debug().Msgf("Set eth addr error: %s", err.Error())
		return
	}

	respondOK(w, true)
}
