package blockchain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/rs/zerolog/log"
	"net/http"
	"platform-backend/config"
	"strings"
)

type ByteArray []byte

func (m ByteArray) MarshalJSON() ([]byte, error) {
	return []byte(strings.Join(strings.Fields(fmt.Sprintf("%d", m)), ",")), nil
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
}

type Blockchain struct {
	Api        *eos.API
	PubKeys    *PubKeys
	ChainId    eos.Checksum256
	sponsorUrl string
}

func Init(config *config.BlockchainConfig) (*Blockchain, error) {
	blockchain := new(Blockchain)
	blockchain.Api = eos.New(config.NodeUrl)
	blockchain.sponsorUrl = config.SponsorUrl

	info, err := blockchain.Api.GetInfo()

	if err != nil {
		return nil, err
	}

	blockchain.Api.EnableKeepAlives()
	blockchain.ChainId = info.ChainID

	keyBag := &eos.KeyBag{}
	if err := keyBag.ImportPrivateKey(config.Permissions.Deposit); err != nil {
		return nil, err
	}
	if err := keyBag.ImportPrivateKey(config.Permissions.GameAction); err != nil {
		return nil, err
	}
	blockchain.Api.SetSigner(keyBag)

	pubKeys := &PubKeys{
		Deposit:    keyBag.Keys[0].PublicKey(),
		GameAction: keyBag.Keys[1].PublicKey(),
	}
	blockchain.PubKeys = pubKeys

	log.Info().Msgf("Connected with blockchain with chaid id: %s", blockchain.ChainId.String())
	return blockchain, nil
}

func (b *Blockchain) GetSponsoredTrx(trx *eos.Transaction) (*eos.SignedTransaction, error) {
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

	sponsoredSignedTrx := eos.NewSignedTransaction(trx)

	for _, strSignature := range response.Signatures {
		sign, err := ecc.NewSignature(strSignature)
		if err != nil {
			return nil, errors.New("sponsored signature parsing error: " + err.Error())
		}
		sponsoredSignedTrx.Signatures = append(sponsoredSignedTrx.Signatures, sign)
	}

	return sponsoredSignedTrx, nil
}
