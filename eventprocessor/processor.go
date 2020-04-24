package eventprocessor

import (
	"context"
	"fmt"
	eventlistener "github.com/DaoCasino/platform-action-monitor-client"
	"github.com/rs/zerolog/log"
	"platform-backend/blockchain"
	"platform-backend/models"
	"platform-backend/repositories"
	"platform-backend/usecases"
	"time"
)

// game events
const (
	gameStarted eventlistener.EventType = iota
	actionRequest
	signidicePartOneRequest
	signidicePartTwoRequest
	gameFinished
	gameFailed
	gameMessage
)

type UpdateHandler = func(context.Context, *EventProcessor, *eventlistener.Event) error

type EventProcessor struct {
	repos      *repositories.Repos
	blockChain *blockchain.Blockchain
	useCases   *usecases.UseCases
}

func New(
	repos *repositories.Repos,
	blockchain *blockchain.Blockchain,
	useCases *usecases.UseCases,
) *EventProcessor {
	return &EventProcessor{repos, blockchain, useCases}
}

func (p *EventProcessor) Process(ctx context.Context, event *eventlistener.Event) {
	gsRepo := p.repos.GameSession

	bcSession, err := gsRepo.GetSessionByBlockChainID(ctx, event.RequestID)
	if err != nil {
		log.Warn().Msgf("Couldn't find session with requestID %v", event.RequestID)
		return
	}

	// already processed offset
	if bcSession.LastOffset >= event.Offset {
		log.Debug().Msgf("Skip already processed event for session: %d with offset: %d", bcSession.ID, event.Offset)
		return
	}

	err = gsRepo.AddGameSessionUpdate(ctx, &models.GameSessionUpdate{
		SessionID:  bcSession.ID,
		UpdateType: uint16(event.EventType),
		Timestamp:  time.Now(),
		Data:       event.Data,
	})

	if err != nil {
		log.Warn().Msgf("Failed to add game session update, reason: %s", err.Error())
		return
	}

	handler, err := GetHandler(event.EventType)

	if err != nil {
		log.Warn().Msgf("Failed to process event, reason: %s", err.Error())
		return
	}

	nextState, err := GetNextState(bcSession.State, event.EventType)

	if err != nil {
		log.Warn().Msgf("Failed to process event, reason: %s", err.Error())
		return
	}

	handleError := handler(ctx, p, event)

	if handleError != nil {
		log.Warn().Msgf("Failed to process event, %+v, reason: %s", event, handleError.Error())
		return
	}

	err = gsRepo.UpdateSessionStateAndOffset(ctx, bcSession.ID, nextState, event.Offset)

	if err != nil {
		log.Warn().Msgf("Failed to update session state, reason: %s", err.Error())
	}
}

func GetHandler(eventType eventlistener.EventType) (UpdateHandler, error) {
	switch eventType {
	case gameStarted:
		return onGameStarted, nil
	case actionRequest:
		return onActionRequest, nil
	case signidicePartOneRequest:
		return onSignidicePartOneRequest, nil
	case signidicePartTwoRequest:
		return onSignidicePartTwoRequest, nil
	case gameFinished:
		return onGameFinished, nil
	case gameFailed:
		return onGameFailed, nil
	case gameMessage:
		return onGameMessage, nil
	}
	return nil, fmt.Errorf("couldn't get dispatcher for event_type %v", eventType)
}

func GetNextState(currentState models.GameSessionState, eventType eventlistener.EventType) (models.GameSessionState, error) {
	switch eventType {
	case gameStarted:
		return models.GameStartedInBC, nil
	case actionRequest:
		return models.RequestedGameAction, nil
	case signidicePartOneRequest:
		return models.RequestedSignidicePartOne, nil
	case gameFinished:
		return models.GameFinished, nil
	case gameFailed:
		return models.GameFailed, nil
	default:
		return currentState, nil
	}
}

func GetEventsToSubscribe() []eventlistener.EventType {
	return []eventlistener.EventType{
		gameStarted,
		actionRequest,
		signidicePartOneRequest,
		gameMessage,
		gameFinished,
		gameFailed,
	}
}