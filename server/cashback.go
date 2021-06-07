package server

import (
	"context"
	"encoding/json"
	"github.com/rs/zerolog/log"
	"net/http"
)

type CashbackClaimRequest struct {
	AccessToken string `json:"accessToken"`
}

type CashbackAccruedRequest struct {
	AccountName string `json:"accountName"`
}

func claimHandler(ctx context.Context, app *App, w http.ResponseWriter, r *http.Request) {
	log.Debug().Msgf("New cashback claim request")

	var req CashbackClaimRequest
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

	if err := app.useCases.Cashback.SetStateClaim(ctx, accountName); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		log.Debug().Msgf("Set cashback claim state error: %s", err.Error())
		return
	}

	respondOK(w, true)
}

func cashbacksHandler(ctx context.Context, app *App, w http.ResponseWriter, r *http.Request) {
	log.Debug().Msgf("New cashbacks request")
	cashbacks, err := app.useCases.Cashback.GetCashbacksForClaimed(ctx)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		log.Debug().Msgf("get cashbacks error: %s", err.Error())
		return
	}
	respondOK(w, cashbacks)
}

func cashbackApproveHandler(ctx context.Context, app *App, w http.ResponseWriter, r *http.Request) {
	log.Debug().Msgf("New cashback approve request")
	var req CashbackAccruedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		log.Debug().Msgf("Http body parse error, %s", err.Error())
		return
	}
	if err := app.useCases.Cashback.PayCashback(ctx, req.AccountName); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		log.Debug().Msgf("cashback approve error: %s", err.Error())
		return
	}

	info, err := app.useCases.Cashback.CashbackInfo(ctx, req.AccountName)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		log.Debug().Msgf("get cashback info error: %s", err.Error())
		return
	}

	respondOK(w, info)
}
