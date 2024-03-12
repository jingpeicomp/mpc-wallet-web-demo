package tss

import (
	"encoding/json"
	"github.com/bnb-chain/tss-lib/ecdsa/keygen"
	"github.com/bnb-chain/tss-lib/tss"
	"log"
	"math/big"
	"os"
	"runtime"
	"strconv"
	"sync/atomic"
)

func Generate(count int, threshold int) {
	partyIds := make(tss.UnSortedPartyIDs, 0, count)
	for i := 1; i <= count; i++ {
		partyIds = append(partyIds, tss.NewPartyID(strconv.Itoa(i), " ", big.NewInt(int64(i))))
	}
	sortedPartyIds := tss.SortPartyIDs(partyIds)
	p2PCtx := tss.NewPeerContext(sortedPartyIds)
	committees := make([]*keygen.LocalParty, 0, count)

	outCh := make(chan tss.Message, count)
	endCh := make(chan keygen.LocalPartySaveData, count)

	for i := 0; i < count; i++ {
		params := tss.NewParameters(tss.S256(), p2PCtx, partyIds[i], count, threshold)
		P := keygen.NewLocalParty(params, outCh, endCh).(*keygen.LocalParty)
		committees = append(committees, P)
	}

	for _, P := range committees {
		go func(P *keygen.LocalParty) {
			log.Println("Begin start party ", P.PartyID().Id)
			if err := P.Start(); err != nil {
				log.Println("Start new party error", err)
			}
			log.Println("Finish start party ", P.PartyID().Id)
		}(P)
	}

	saveDataArray := make([]keygen.LocalPartySaveData, count)
	var keygenEnded int32
keygen:
	for {
		log.Printf("ACTIVE GOROUTINES: %d\n", runtime.NumGoroutine())
		select {
		case msg := <-outCh:
			if msg.IsBroadcast() {
				for _, P := range committees {
					if P.PartyID().Id != msg.GetFrom().Id {
						go partyUpdate(P, msg)
					}
				}
			} else {
				dest := msg.GetTo()
				if dest == nil {
					log.Fatal("did not expect a msg to have a nil destination during resharing")
				}
				for _, destItem := range dest {
					go partyUpdate(committees[destItem.Index], msg)
				}
			}

		case saveData := <-endCh:
			log.Println("------> receive save data")
			if saveData.Xi != nil {
				index, err := saveData.OriginalIndex()
				if err != nil {
					log.Println("should not be an error getting a party's index from save data", err)
				}
				saveDataArray[index] = saveData
			}
			atomic.AddInt32(&keygenEnded, 1)
			if atomic.LoadInt32(&keygenEnded) == int32(count) {
				log.Println("Mpc finished.......")
				break keygen
			}
		}
	}

	for i, value := range saveDataArray {
		doSaveKey(i+1, value)
	}
	log.Println("=====> Keygen finished.......")
}

func doSaveKey(index int, data keygen.LocalPartySaveData) {
	file, err := os.Create("data/key" + strconv.Itoa(index))
	if err != nil {
		log.Println("Cannot create file ", err)
		return
	}

	content, err := json.Marshal(data)
	if err != nil {
		log.Println("Cannot marshal data to json ", err)
	}
	_, err = file.Write(content)
	if err != nil {
		log.Println("Cannot write data into file ", err)
	}
}
