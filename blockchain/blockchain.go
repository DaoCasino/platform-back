package blockchain

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/rs/zerolog/log"
	"net/http"
)

type sponsorRequest struct {
	SerializedTransaction []byte `json:"serializedTransaction"`
}
type sponsorResponse struct {
	SerializedTransaction []byte   `json:"serializedTransaction"`
	Signatures            []string `json:"signatures"`
}

type Blockchain struct {
	Api        *eos.API
	sponsorUrl string
}

func Init(url string, sponsorUrl string) (*Blockchain, error) {
	blockchain := new(Blockchain)
	blockchain.Api = eos.New(url)
	blockchain.sponsorUrl = sponsorUrl

	info, err := blockchain.Api.GetInfo()

	if err != nil {
		return nil, err
	}

	blockchain.Api.EnableKeepAlives()

	log.Info().Msgf("Connected with blockchain with chaid id: %s", info.ChainID.String())
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

	httpResp, err := http.Post(b.sponsorUrl, "application/json", bytes.NewReader(reqBody))
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

	var sponsoredTrx *eos.Transaction
	err = eos.UnmarshalBinary(response.SerializedTransaction, sponsoredTrx)
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
