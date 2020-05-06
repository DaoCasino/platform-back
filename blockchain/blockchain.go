package blockchain

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/rs/zerolog/log"
	"net/http"
	"platform-backend/config"
	"strings"
	"sync"
	"time"
)

const (
	txOptsCacheTTL = 2 //seconds
)

type ByteArray []byte

func (m ByteArray) MarshalJSON() ([]byte, error) {
	return []byte(strings.Join(strings.Fields(fmt.Sprintf("%d", m)), ",")), nil
}

func (m *ByteArray) UnmarshalJSON(data []byte) error {
	str := ""
	err := json.Unmarshal(data, &str)
	if err != nil {
		return err
	}

	b, err := hex.DecodeString(str)
	if err != nil {
		return err
	}

	*m = b
	return nil
}

type sponsorRequest struct {
	SerializedTransaction ByteArray `json:"serializedTransaction"`
}

type sponsorResponse struct {
	SerializedTransaction []byte   `json:"serializedTransaction"`
	Signatures            []string `json:"signatures"`
}

type PubKeys struct {
	Deposit    ecc.PublicKey
	GameAction ecc.PublicKey
	SigniDice  ecc.PublicKey
}

type Blockchain struct {
	Api                 *eos.API
	PubKeys             *PubKeys
	ChainId             eos.Checksum256
	PlatformAccountName string
	sponsorUrl          string
	disableSponsor      bool

	optsMutex       sync.Mutex
	lastInfoTime    time.Time
	lastHeadBlockID eos.Checksum256
}

func Init(config *config.BlockchainConfig) (*Blockchain, error) {
	blockchain := new(Blockchain)
	blockchain.Api = eos.New(config.NodeUrl)
	blockchain.sponsorUrl = config.SponsorUrl
	blockchain.PlatformAccountName = config.Contracts.Platform

	info, err := blockchain.Api.GetInfo()

	if err != nil {
		return nil, err
	}

	blockchain.Api.EnableKeepAlives()
	blockchain.ChainId = info.ChainID
	blockchain.lastInfoTime = time.Now()
	blockchain.lastHeadBlockID = info.HeadBlockID

	keyBag := &eos.KeyBag{}
	if err := keyBag.ImportPrivateKey(config.Permissions.Deposit); err != nil {
		return nil, err
	}
	if err := keyBag.ImportPrivateKey(config.Permissions.GameAction); err != nil {
		return nil, err
	}
	if err := keyBag.ImportPrivateKey(config.Permissions.SigniDice); err != nil {
		return nil, err
	}
	blockchain.Api.SetSigner(keyBag)

	pubKeys := &PubKeys{
		Deposit:    keyBag.Keys[0].PublicKey(),
		GameAction: keyBag.Keys[1].PublicKey(),
		SigniDice:  keyBag.Keys[2].PublicKey(),
	}
	blockchain.PubKeys = pubKeys

	blockchain.disableSponsor = config.DisableSponsor

	log.Info().Msgf("Connected with blockchain with chaid id: %s", blockchain.ChainId.String())
	return blockchain, nil
}

func (b *Blockchain) GetSponsoredTrx(trx *eos.Transaction) (*eos.SignedTransaction, error) {
	if b.disableSponsor {
		return eos.NewSignedTransaction(trx), nil
	}

	packedTrx, err := eos.MarshalBinary(trx)
	if err != nil {
		return nil, err
	}

	reqBody, err := json.Marshal(&sponsorRequest{
		SerializedTransaction: packedTrx,
	})
	if err != nil {
		return nil, errors.New("request body marshal error")
	}

	httpResp, err := http.Post(b.sponsorUrl+"/sponsor", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, errors.New("sponsorship provider request error: " + err.Error())
	}
	if httpResp.StatusCode != http.StatusOK {
		return nil, errors.New("sponsorship provider respond with error: " + httpResp.Status)
	}

	var response sponsorResponse
	err = json.NewDecoder(httpResp.Body).Decode(&response)
	if err != nil {
		return nil, errors.New("sponsorship response parsing error: " + err.Error())
	}

	var sponsoredTrx eos.Transaction
	err = eos.UnmarshalBinary(response.SerializedTransaction, &sponsoredTrx)
	if err != nil {
		return nil, errors.New("sponsored transaction parsing error: " + err.Error())
	}

	sponsoredSignedTrx := eos.NewSignedTransaction(&sponsoredTrx)

	for _, strSignature := range response.Signatures {
		sign, err := ecc.NewSignature(strSignature)
		if err != nil {
			return nil, errors.New("sponsored signature parsing error: " + err.Error())
		}
		sponsoredSignedTrx.Signatures = append(sponsoredSignedTrx.Signatures, sign)
	}

	for _, action := range sponsoredSignedTrx.Actions {
		action.SetToServer(true)
	}

	return sponsoredSignedTrx, nil
}

func (b *Blockchain) GetTrxOpts() *eos.TxOptions {
	b.optsMutex.Lock()
	defer b.optsMutex.Unlock()

	if b.lastInfoTime.Unix()+txOptsCacheTTL < time.Now().Unix() {
		resp, err := b.Api.GetInfo()
		if err != nil {
			b.lastHeadBlockID = nil
		} else {
			b.lastHeadBlockID = resp.HeadBlockID
		}
	}

	return &eos.TxOptions{
		ChainID:     b.ChainId,
		HeadBlockID: b.lastHeadBlockID,
	}
}
