// Copyright © 2019-2020 Binance
//
// This file is part of Binance. The full Binance copyright notice, including
// terms governing use, modification, and redistribution, is contained in the
// file LICENSE at the root of the source code distribution tree.

package main

import (
	"github.com/bnb-chain/tss-lib/ecdsa/keygen"
	"log"
	"syscall/js"
	"tss-web/tss"
)

var keyById map[int]keygen.LocalPartySaveData

func main() {
	log.Println("Hello, WebAssembly!")
	js.Global().Set("keyNum", "11111")
	js.Global().Set("sign", JsSign())
	js.Global().Set("parseKey", JsParseKey())
	js.Global().Set("setKey", JsSetKey())
	js.Global().Set("clearKey", JsClearKey())
	select {}
}

func JsSetKey() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		keyById = make(map[int]keygen.LocalPartySaveData)

		//log.Println("Type of args[1] is ", args[1].Type())
		//log.Println("Length of args[1] is ", args[1].Length())
		//log.Println("partyNo1 is ", args[0])
		//log.Println("partyNo2 is ", args[2])

		partyNo1 := args[0].Int()
		key1Bytes := make([]byte, args[1].Length())
		js.CopyBytesToGo(key1Bytes, args[1])
		key1 := tss.LoadPartyDataByBytes(key1Bytes)
		keyById[partyNo1] = key1

		partyNo2 := args[2].Int()
		key2Bytes := make([]byte, args[3].Length())
		js.CopyBytesToGo(key2Bytes, args[3])
		key2 := tss.LoadPartyDataByBytes(key2Bytes)
		keyById[partyNo2] = key2
		return true
	})
}

func JsClearKey() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		keyById = make(map[int]keygen.LocalPartySaveData)
		return true
	})
}

func JsSign() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		msgStr := args[0].String()
		var partyNo1, partyNo2 int
		var key1, key2 keygen.LocalPartySaveData
		iter := 0
		for partyNo, key := range keyById {
			if iter == 0 {
				partyNo1 = partyNo
				key1 = key
			} else {
				partyNo2 = partyNo
				key2 = key
			}
			iter++
		}
		return tss.Sign(msgStr, partyNo1, key1, partyNo2, key2)
	})
}

func JsParseKey() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		var key1, key2 keygen.LocalPartySaveData
		iter := 0
		for _, key := range keyById {
			if iter == 0 {
				key1 = key
			} else {
				key2 = key
			}
			iter++
		}
		addrAndPriKey := tss.ParseKey(key1, key2)
		return []interface{}{addrAndPriKey[0], addrAndPriKey[1]}
	})
}
