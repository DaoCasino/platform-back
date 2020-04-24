package eventprocessor

import (
    "context"
    "encoding/json"
    eventlistener "github.com/DaoCasino/platform-action-monitor-client"
    "github.com/eoscanada/eos-go"

    "github.com/rs/zerolog/log"
)

func onGameStarted(ctx context.Context, p *EventProcessor, event *eventlistener.Event) error {
    return nil
}

func onActionRequest(ctx context.Context, p *EventProcessor, event *eventlistener.Event) error {
    return nil
}

func onSignidicePartOneRequest(ctx context.Context, p *EventProcessor, event *eventlistener.Event) error {
    gs, err := p.repos.GameSession.GetGameSession(ctx, event.RequestID)
    if err != nil {
        return err
    }

    game, err := p.repos.Casino.GetGame(ctx, gs.GameID)
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

    log.Debug().Msgf("Successfully signed and sent signidice for session: %d", gs.ID)

    return nil
}

func onSignidicePartTwoRequest(ctx context.Context, p *EventProcessor, event *eventlistener.Event) error {
    return nil
}

func onGameFinished(ctx context.Context, p *EventProcessor, event *eventlistener.Event) error {
    return nil
}

func onGameFailed(ctx context.Context, p *EventProcessor, event *eventlistener.Event) error {
    return nil
}

func onGameMessage(ctx context.Context, p *EventProcessor, event *eventlistener.Event) error {
    return nil
}
