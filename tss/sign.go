// Copyright © 2019-2020 Binance
//
// This file is part of Binance. The full Binance copyright notice, including
// terms governing use, modification, and redistribution, is contained in the
// file LICENSE at the root of the source code distribution tree.

package tss

import (
	"encoding/hex"
	"github.com/bnb-chain/tss-lib/common"
	"github.com/bnb-chain/tss-lib/ecdsa/keygen"
	"github.com/bnb-chain/tss-lib/ecdsa/signing"
	"github.com/bnb-chain/tss-lib/tss"
	"log"
	"math/big"
	"runtime"
	"strconv"
)

func Sign(msg string, partyNo1 int, key1 keygen.LocalPartySaveData, partyNo2 int, key2 keygen.LocalPartySaveData) string {
	log.Printf("go start sign msg: %v", msg)
	outCh := make(chan tss.Message)
	endCh := make(chan common.SignatureData)
	msgDigest := []byte(msg)
	parties := loadPartyByBytes(msgDigest, outCh, endCh, partyNo1, key1, partyNo2, key2)
	partyMap := make(map[string]tss.Party)
	partyMap[parties[0].PartyID().Id] = parties[0]
	partyMap[parties[1].PartyID().Id] = parties[1]
	startParty(parties)
	var signData common.SignatureData
signing:
	for {
		log.Printf("ACTIVE GOROUTINES: %d\n", runtime.NumGoroutine())
		select {
		case msg := <-outCh:
			dest := msg.GetTo()
			if dest == nil {
				for _, P := range parties {
					if P.PartyID().Id == msg.GetFrom().Id {
						continue
					}
					go partyUpdate(P, msg)
				}
			} else {
				if dest[0].Id == msg.GetFrom().Id {
					log.Printf("party %s tried to send a message to itself (%s)", dest[0].Id, msg.GetFrom().Id)
				}
				go partyUpdate(partyMap[dest[0].Id], msg)
			}
		case signData = <-endCh:
			log.Println("GetSignatureRecovery = ", hex.EncodeToString(signData.GetSignatureRecovery()))
			log.Println("S = ", hex.EncodeToString(signData.GetS()))
			log.Println("R = ", hex.EncodeToString(signData.GetR()))
			log.Println("message = ", string(signData.GetM()))
			log.Println("Sign finish ", hex.EncodeToString(signData.GetS()), hex.EncodeToString(signData.GetR()), hex.EncodeToString(signData.GetM()))
			break signing
		}
	}

	result := hex.EncodeToString(signData.GetSignatureRecovery()) + hex.EncodeToString(signData.GetS()) + hex.EncodeToString(signData.GetR())
	log.Println("Sign result ", result)
	return result
}

func loadPartyByBytes(digest []byte, outCh chan tss.Message, endCh chan common.SignatureData, index1 int,
	key1 keygen.LocalPartySaveData, index2 int, key2 keygen.LocalPartySaveData) [2]tss.Party {
	parties := tss.SortPartyIDs(tss.UnSortedPartyIDs{tss.NewPartyID(strconv.Itoa(index1), " ", big.NewInt(int64(index1))),
		tss.NewPartyID(strconv.Itoa(index2), " ", big.NewInt(int64(index2)))})

	ctx := tss.NewPeerContext(parties)
	curve := tss.S256()
	msg := &big.Int{}
	msg.SetBytes(digest)

	partyId1 := parties[0]
	log.Println("Core ", runtime.GOMAXPROCS(0))
	params1 := tss.NewParameters(curve, ctx, partyId1, 3, 1)
	params1.SetConcurrency(4)
	party1 := signing.NewLocalParty(msg, params1, key1, outCh, endCh)

	partyId2 := parties[1]
	params2 := tss.NewParameters(curve, ctx, partyId2, 3, 1)
	params2.SetConcurrency(4)
	party2 := signing.NewLocalParty(msg, params2, key2, outCh, endCh)

	return [2]tss.Party{party1, party2}
}

func startParty(parties [2]tss.Party) {
	for _, party := range parties {
		currentParty := party
		go func() {
			log.Println("------> start party begin: ", currentParty.PartyID().Id)
			err := currentParty.Start()
			if err == nil {
				log.Println("------> start party successfully: ", currentParty.PartyID().Id)
			} else {
				log.Println("------> start party error: ", currentParty.PartyID().Id, err)
			}
		}()
	}
}

func partyUpdate(party tss.Party, msg tss.Message) {
	// do not send a message from this party back to itself
	log.Printf("Party id: %v , message from: %v , message destination %v , broadcast %v", party.PartyID().Id, msg.GetFrom().Id, msg.GetTo(), msg.IsBroadcast())
	if party.PartyID().Id == msg.GetFrom().Id {
		return
	}
	bz, _, err := msg.WireBytes()
	if err != nil {
		log.Println("Message error", err)
		return
	}
	pMsg, err := tss.ParseWireMessage(bz, msg.GetFrom(), msg.IsBroadcast())
	if err != nil {
		log.Println("Pare Message error", err)
		return
	}
	if _, err := party.Update(pMsg); err != nil {
		log.Println("Update Message error", err)
	}
}
