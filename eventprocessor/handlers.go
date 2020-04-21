package eventprocessor

import (
    "context"
    "encoding/json"
    eventlistener "github.com/DaoCasino/platform-action-monitor-client"
    "github.com/eoscanada/eos-go"
    "github.com/eoscanada/eos-go/ecc"
    "github.com/rs/zerolog/log"
)

func onGameStarted(ctx context.Context, p *Processor, event *eventlistener.Event) error {
    return nil
}

func onActionRequest(ctx context.Context, p *Processor, event *eventlistener.Event) error {
    return nil
}

func onSignidicePartOneRequest(ctx context.Context, p *Processor, event *eventlistener.Event) error {
    var data struct{ Digest []byte }
    parseError := json.Unmarshal(event.Data, &data)
    if parseError != nil {
        return parseError
    }
    blockchain, api := p.BlockChain, p.BlockChain.Api
    requiredKeys := []ecc.PublicKey{blockchain.PubKeys.SigniDice}
    keyBag, _ := api.Signer.(*eos.KeyBag)
    signature, err := keyBag.SignDigest(data.Digest, requiredKeys[0])
    if err != nil {
        return err
    }
    tx, packedTx, err := GetSigndiceTransaction(api, event.Sender, blockchain.PlatformAccountName, event.RequestID, signature)

    if err != nil {
        return err
    }

    log.Debug().Msgf("%+v", tx)
    result, err := api.PushTransaction(packedTx)
    if err != nil {
        return err
    }
    log.Debug().Msg("Successfully signed and sent txn, id: " + result.TransactionID)

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
