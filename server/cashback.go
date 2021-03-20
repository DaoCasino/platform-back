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
		log.Debug().Msgf("Set cashback claim state error: %s", err.Error())
		return
	}

	respondOK(w, true)
}

func cashbacksHandler(app *App, w http.ResponseWriter, r *http.Request) {
	log.Debug().Msgf("New cashbacks request")
	ctx := r.Context()
	cashbacks, err := app.useCases.Cashback.GetCashbacksForClaimed(ctx)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		log.Debug().Msgf("get cashbacks error: %s", err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, cashbacks)
}