package tss

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"encoding/json"
	"github.com/bnb-chain/tss-lib/crypto/vss"
	"github.com/bnb-chain/tss-lib/ecdsa/keygen"
	"github.com/bnb-chain/tss-lib/tss"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
)

func ParseKeyByBytes(key1Bytes []byte, key2Bytes []byte) [2]string {
	keyData1 := LoadPartyDataByBytes(key1Bytes)
	keyData2 := LoadPartyDataByBytes(key2Bytes)
	priKey, _ := reconstruct(1, tss.S256(), [2]keygen.LocalPartySaveData{keyData1, keyData2})
	log.Println("Private Key: ", priKey.PublicKey)
	addr := crypto.PubkeyToAddress(priKey.PublicKey).String()
	strPriKey := hex.EncodeToString(priKey.D.Bytes())
	log.Printf("Wallet address is %v , private key is %v", addr, strPriKey)
	return [2]string{addr, strPriKey}
}

func ParseKey(keyData1 keygen.LocalPartySaveData, keyData2 keygen.LocalPartySaveData) [2]string {
	priKey, _ := reconstruct(1, tss.S256(), [2]keygen.LocalPartySaveData{keyData1, keyData2})
	log.Println("Private Key: ", priKey.PublicKey)
	addr := crypto.PubkeyToAddress(priKey.PublicKey).String()
	strPriKey := hex.EncodeToString(priKey.D.Bytes())
	log.Printf("Wallet address is %v , private key is %v", addr, strPriKey)
	return [2]string{addr, strPriKey}
}

func LoadPartyDataByBytes(keyBytes []byte) keygen.LocalPartySaveData {
	var key keygen.LocalPartySaveData
	err := json.Unmarshal(keyBytes, &key)
	if err != nil {
		log.Println("Cannot decode key ", err)
	}

	return key
}

func reconstruct(threshold int, ec elliptic.Curve, shares [2]keygen.LocalPartySaveData) (*ecdsa.PrivateKey, error) {
	var vssShares = make(vss.Shares, len(shares))
	for i, share := range shares {
		vssShare := &vss.Share{
			Threshold: threshold,
			ID:        share.ShareID,
			Share:     share.Xi,
		}
		vssShares[i] = vssShare
	}

	d, err := vssShares.ReConstruct(ec)
	if err != nil {
		return nil, err
	}

	x, y := ec.ScalarBaseMult(d.Bytes())

	privateKey := &ecdsa.PrivateKey{
		D: d,
		PublicKey: ecdsa.PublicKey{
			Curve: ec,
			X:     x,
			Y:     y,
		},
	}

	return privateKey, nil
}
