package eventprocessor

import (
	eventlistener "github.com/DaoCasino/platform-action-monitor-client"
)

func onGameStarted(p *Processor, event *eventlistener.Event) error {
	// TODO
	return nil
}

func onActionRequest(p *Processor, event *eventlistener.Event) error {
	return nil
}

func onSignidicePartOneRequest(p *Processor, event *eventlistener.Event) error {
	return nil
}

func onSignidicePartTwoRequest(p *Processor, event *eventlistener.Event) error {
	return nil
}

func onGameFinished(p *Processor, event *eventlistener.Event) error {
	return nil
}

func onGameFailed(p *Processor, event *eventlistener.Event) error {
	return nil
}

func onGameMessage(p *Processor, event *eventlistener.Event) error {
	return nil
}
