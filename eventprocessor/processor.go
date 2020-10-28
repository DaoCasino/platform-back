package eventprocessor

import (
	"context"
	eventlistener "github.com/DaoCasino/platform-action-monitor-client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"platform-backend/blockchain"
	"platform-backend/models"
	"platform-backend/repositories"
	"platform-backend/usecases"
)

// game events
const (
	gameStarted             = 0
	actionRequest           = 1
	signidicePartOneRequest = 2
	gameFinished            = 4
	gameFailed              = 5
	gameMessage             = 6
)

type UpdateHandler = func(context.Context, *EventProcessor, *eventlistener.Event, *models.GameSession) error

var handlersMap = map[eventlistener.EventType]UpdateHandler{
	gameStarted:             onGameStarted,
	actionRequest:           onActionRequest,
	signidicePartOneRequest: onSignidicePartOneRequest,
	gameFinished:            onGameFinished,
	gameFailed:              onGameFailed,
	gameMessage:             onGameMessage,
}

type EventProcessor struct {
	repos                *repositories.Repos
	blockchain           *blockchain.Blockchain
	useCases             *usecases.UseCases
	failedSessionCounter *prometheus.CounterVec
}

func New(
	repos *repositories.Repos,
	blockchain *blockchain.Blockchain,
	useCases *usecases.UseCases,
	reg prometheus.Registerer,
) *EventProcessor {
	failedSessionCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "failed_session",
		}, []string{"prev_state"},
	)

	reg.MustRegister(failedSessionCounter)

	return &EventProcessor{
		repos:                repos,
		blockchain:           blockchain,
		useCases:             useCases,
		failedSessionCounter: failedSessionCounter,
	}
}

func (p *EventProcessor) Process(ctx context.Context, event *eventlistener.Event) {
	gsRepo := p.repos.GameSession

	bcSession, err := gsRepo.GetSessionByBlockChainID(ctx, event.RequestID)
	if err != nil {
		log.Debug().Msgf("Couldn't find session with requestID %v", event.RequestID)
		return
	}

	// already processed offset
	if bcSession.LastOffset >= event.Offset {
		log.Debug().Msgf("Skip already processed event for session: %d with offset: %d", bcSession.ID, event.Offset)
		return
	}

	handler, ok := handlersMap[event.EventType]
	if !ok { // should never happen
		log.Panic().Msgf("Got unknown event type: %d", event.EventType)
		return
	}

	handleError := handler(ctx, p, event, bcSession)
	if handleError != nil {
		log.Error().Msgf("Failed to process event, %+v, reason: %s", event, handleError.Error())
		return
	}

	err = gsRepo.UpdateSessionOffset(ctx, bcSession.ID, event.Offset)
	if err != nil {
		log.Error().Msgf("Failed to update session offset, reason: %s", err.Error())
	}
}

func GetEventsToSubscribe() []eventlistener.EventType {
	var events []eventlistener.EventType
	for key := range handlersMap {
		events = append(events, key)
	}
	return events
}
