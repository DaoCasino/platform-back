package eventprocessor

import (
	"context"
	"encoding/json"
	eventlistener "github.com/DaoCasino/platform-action-monitor-client"
	"github.com/eoscanada/eos-go"
	"github.com/rs/zerolog/log"
	"platform-backend/blockchain"
	gamesessions "platform-backend/game_sessions"
	"platform-backend/models"
	"time"
)

type finishedEventData struct {
	Msg       blockchain.ByteArray `json:"msg"`
	PlayerWin eos.Asset            `json:"player_win_amount"`
}

type messageEventData struct {
	Msg       blockchain.ByteArray `json:"msg"`
}

type finishedUpdateData struct {
	Msg       []uint64  `json:"msg"`
	PlayerWin eos.Asset `json:"player_win_amount"`
}

func onGameStarted(ctx context.Context, p *EventProcessor, event *eventlistener.Event, session *models.GameSession) error {
	log.Debug().Msgf("Got started event for session: %d", session.ID)

	err := p.repos.GameSession.AddGameSessionUpdate(ctx, &models.GameSessionUpdate{
		SessionID:  session.ID,
		UpdateType: models.SessionStartedUpdate,
		Timestamp:  time.Now(),
		Data:       event.Data,
	})
	if err != nil {
		return err
	}

	err = p.repos.GameSession.UpdateSessionState(ctx, session.ID, models.GameStartedInBC)
	if err != nil {
		return err
	}

	return nil
}

func onActionRequest(ctx context.Context, p *EventProcessor, event *eventlistener.Event, session *models.GameSession) error {
	log.Debug().Msgf("Got action request event for session: %d", session.ID)

	// first action
	if session.State == models.GameStartedInBC {
		action, err := p.repos.GameSession.GetFirstAction(ctx, session.ID)
		if err != nil {
			// if first game session not saved action was already sent
			if err == gamesessions.ErrFirstGameActionNotFound {
				return nil
			}
			return err
		}

		log.Debug().Msgf("Try to perform first game action for session: %d", session.ID)
		err = p.useCases.GameSession.GameAction(ctx, session.ID, action.Type, action.Params)
		if err != nil {
			return err
		}

		err = p.repos.GameSession.DeleteFirstGameAction(ctx, session.ID)
		if err != nil {
			return err
		}

		return nil
	}

	err := p.repos.GameSession.AddGameSessionUpdate(ctx, &models.GameSessionUpdate{
		SessionID:  session.ID,
		UpdateType: models.GameActionRequestedUpdate,
		Timestamp:  time.Now(),
		Data:       event.Data,
	})
	if err != nil {
		return err
	}

	err = p.repos.GameSession.UpdateSessionState(ctx, session.ID, models.RequestedGameAction)
	if err != nil {
		return err
	}

	return nil
}

func onSignidicePartOneRequest(ctx context.Context, p *EventProcessor, event *eventlistener.Event, session *models.GameSession) error {
	log.Debug().Msgf("Got signidice one request event for session: %d", session.ID)

	gs, err := p.repos.GameSession.GetGameSession(ctx, event.RequestID)
	if err != nil {
		return err
	}

	game, err := p.repos.Contracts.GetGame(ctx, gs.GameID)
	if err != nil {
		return err
	}

	var data struct{ Digest eos.Checksum256 }
	err = json.Unmarshal(event.Data, &data)
	if err != nil {
		return err
	}

	err = p.useCases.Signidice.PerformSignidice(ctx, game.Contract, data.Digest, gs.BlockchainSesID)
	if err != nil {
		return err
	}

	err = p.repos.GameSession.UpdateSessionState(ctx, session.ID, models.SignidicePartOneTrxSent)
	if err != nil {
		return err
	}

	log.Debug().Msgf("Successfully signed and sent signidice for session: %d", gs.ID)

	return nil
}

func onGameFinished(ctx context.Context, p *EventProcessor, event *eventlistener.Event, session *models.GameSession) error {
	log.Debug().Msgf("Got finished event for session: %d", session.ID)

	var eventData finishedEventData
	err := json.Unmarshal(event.Data, &eventData)
	if err != nil {
		return err
	}

	var resultsArray []uint64
	err = eos.NewDecoder(eventData.Msg).Decode(&resultsArray)
	if err != nil {
		return err
	}

	updateData, err := json.Marshal(finishedUpdateData{
		PlayerWin: eventData.PlayerWin,
		Msg:       resultsArray,
	})
	if err != nil {
		return err
	}

	err = p.repos.GameSession.AddGameSessionUpdate(ctx, &models.GameSessionUpdate{
		SessionID:  session.ID,
		UpdateType: models.GameFinishedUpdate,
		Timestamp:  time.Now(),
		Data:       updateData,
	})
	if err != nil {
		return err
	}

	err = p.repos.GameSession.UpdateSessionState(ctx, session.ID, models.GameFinished)
	if err != nil {
		return err
	}

	return nil
}

func onGameFailed(ctx context.Context, p *EventProcessor, event *eventlistener.Event, session *models.GameSession) error {
	log.Debug().Msgf("Got failed event for session: %d", session.ID)

	err := p.repos.GameSession.AddGameSessionUpdate(ctx, &models.GameSessionUpdate{
		SessionID:  session.ID,
		UpdateType: models.GameFailedUpdate,
		Timestamp:  time.Now(),
		Data:       event.Data,
	})
	if err != nil {
		return err
	}

	err = p.repos.GameSession.UpdateSessionState(ctx, session.ID, models.GameFailed)
	if err != nil {
		return err
	}

	return nil
}

func onGameMessage(ctx context.Context, p *EventProcessor, event *eventlistener.Event, session *models.GameSession) error {
	log.Debug().Msgf("Got game message event for session: %d", session.ID)
	var eventData messageEventData
	err := json.Unmarshal(event.Data, &eventData)
	if err != nil {
		return err
	}

	var resultsArray []uint64
	err = eos.NewDecoder(eventData.Msg).Decode(&resultsArray)
	if err != nil {
		return err
	}

	updateData, err := json.Marshal(finishedUpdateData{
		Msg:       resultsArray,
	})
	if err != nil {
		return err
	}

	return p.repos.GameSession.AddGameSessionUpdate(ctx, &models.GameSessionUpdate{
		SessionID:  session.ID,
		UpdateType: models.GameMessageUpdate,
		Timestamp:  time.Now(),
		Data:       updateData,
	})
}
