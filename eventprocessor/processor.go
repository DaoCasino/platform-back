package eventprocessor

import (
	"context"
	"fmt"
	eventlistener "github.com/DaoCasino/platform-action-monitor-client"
	gamesessions "platform-backend/game_sessions"
	"platform-backend/models"
	"time"
	"github.com/rs/zerolog/log"
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

type SessionState = uint16

// session states
const (

)

type Processor struct {
	gsRepo gamesessions.GameSessionRepository
}

func New(gsRepo gamesessions.GameSessionRepository) *Processor {
	return &Processor{gsRepo}
}

func (p *Processor) Process(ctx context.Context, event *eventlistener.Event) {
	gsRepo := p.gsRepo
	bcSession, err := gsRepo.GetSessionByBlockChainID(ctx, event.RequestID)
	if err != nil {
		log.Warn().Msgf("Couldn't find session with requestID %v", event.RequestID)
		return
	}
	gsRepo.AddGameSessionUpdate(ctx, &models.GameSessionUpdate{
		SessionID: bcSession.ID,
		UpdateType: uint16(event.EventType),
		Timestamp: time.Now(),
		Data: event.Data,
	})
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

	handleError := handler(p, event)
	if handleError != nil {
		log.Warn().Msgf("Failed to handle event, %+v, reason: %s", event, handleError.Error())
		return
	}

	if gsRepo.UpdateSessionState(ctx, bcSession.ID, nextState) != nil {
		log.Warn().Msgf("Failed to update session state, id: %d", bcSession.ID)
	}

}

func GetHandler(eventType eventlistener.EventType) (func(*Processor, *eventlistener.Event) error, error) {
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

func GetNextState(currentState uint16, eventType eventlistener.EventType) (SessionState, error) {
	// TODO
	return 0, nil
}