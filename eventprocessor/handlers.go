package eventprocessor

import (
	"context"
	eventlistener "github.com/DaoCasino/platform-action-monitor-client"
)

func onGameStarted(ctx context.Context, p *Processor, event *eventlistener.Event) error {
	return nil
}

func onActionRequest(ctx context.Context, p *Processor, event *eventlistener.Event) error {
	return nil
}

func onSignidicePartOneRequest(ctx context.Context, p *Processor, event *eventlistener.Event) error {
	return nil
}

func onSignidicePartTwoRequest(ctx context.Context, p *Processor, event *eventlistener.Event) error {
	return nil
}

func onGameFinished(ctx context.Context, p *Processor, event *eventlistener.Event) error {
	return nil
}

func onGameFailed(ctx context.Context, p *Processor, event *eventlistener.Event) error {
	return nil
}

func onGameMessage(ctx context.Context, p *Processor, event *eventlistener.Event) error {
	return nil
}
