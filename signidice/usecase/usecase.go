package usecase

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"github.com/eoscanada/eos-go"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"platform-backend/blockchain"
)

type SignidiceUseCase struct {
	bc                  *blockchain.Blockchain
	rsaKey              *rsa.PrivateKey
	platformAccountName string
}

func NewSignidiceUseCase(
	bc *blockchain.Blockchain,
	platformAccountName string,
	rsaKeyPath string,
) *SignidiceUseCase {
	data, err := ioutil.ReadFile(rsaKeyPath)
	if err != nil {
		log.Panic().Msgf("Cannot read rsa key file: %s", err.Error())
		return nil
	}
	block, _ := pem.Decode(data)
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		log.Panic().Msgf("Cannot parse PKCS1 private key: %s", err.Error())
		return nil
	}

	return &SignidiceUseCase{
		bc:                  bc,
		rsaKey:              key,
		platformAccountName: platformAccountName,
	}
}


func (a *SignidiceUseCase) rsaSign(digest eos.Checksum256) (string, error) {
	sign, err := rsa.SignPKCS1v15(rand.Reader, a.rsaKey, crypto.SHA256, digest)
	if err != nil {
		return "", err
	}

	// contract require base64 string
	return base64.StdEncoding.EncodeToString(sign), nil
}

func (a *SignidiceUseCase)PerformSignidice(ctx context.Context, gameName string, digest []byte, bcSessionID uint64) error {
	rsaSign, err := a.rsaSign(digest)
	if err != nil {
		return err
	}

	action := eos.Action{
		Account: eos.AN(gameName),
		Name:    eos.ActN("sgdicefirst"),
		Authorization: []eos.PermissionLevel{
			{Actor: eos.AN(a.platformAccountName), Permission: eos.PN("signidice")},
		},
		ActionData: eos.NewActionData(struct {
			SessionId uint64 `json:"ses_id"`
			Signature string `json:"sign"`
		}{
			SessionId: bcSessionID,
			Signature: rsaSign,
		}),
	}

	txOpts := &eos.TxOptions{}
	if err := txOpts.FillFromChain(a.bc.Api); err != nil {
		return err
	}

	tx := eos.NewTransaction([]*eos.Action{&action}, txOpts)

	signedTrx, err := a.bc.Api.Signer.Sign(eos.NewSignedTransaction(tx), a.bc.ChainId, a.bc.PubKeys.SigniDice)
	if err != nil {
		return err
	}

	packedTrx, err := signedTrx.Pack(eos.CompressionNone)
	if err != nil {
		return err
	}

	_, err = a.bc.Api.PushTransaction(packedTrx)

	return err
}
