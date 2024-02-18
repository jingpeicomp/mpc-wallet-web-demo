// Copyright © 2019-2020 Binance
//
// This file is part of Binance. The full Binance copyright notice, including
// terms governing use, modification, and redistribution, is contained in the
// file LICENSE at the root of the source code distribution tree.

package main

import (
	"log"
	"syscall/js"
	"tss-web/tss"
)

func main() {
	log.Println("Hello, WebAssembly!")
	js.Global().Set("keyNum", "11111")
	js.Global().Set("sign", JsSign())
	js.Global().Set("parseKey", JsParseKey())
	select {}
}

func JsSign() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		log.Println("Type of args[1] is ", args[1].Type())
		log.Println("Length of args[1] is ", args[1].Length())
		log.Println("arg[0] ", args[0])
		log.Println("arg[4] ", args[4])

		partyNo1 := args[0].Int()
		key1Bytes := make([]byte, args[1].Length())
		js.CopyBytesToGo(key1Bytes, args[1])

		partyNo2 := args[2].Int()
		key2Bytes := make([]byte, args[3].Length())
		js.CopyBytesToGo(key2Bytes, args[3])

		msgStr := args[4].String()

		return tss.Sign(msgStr, partyNo1, key1Bytes, partyNo2, key2Bytes)
	})
}

func JsParseKey() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		//log.Println("Type of args[0] is ", args[0].Type())
		//log.Println("Length of args[0] is ", args[0].Length())
		//log.Println("arg[0] ", args[0])

		key1Bytes := make([]byte, args[0].Length())
		js.CopyBytesToGo(key1Bytes, args[0])

		key2Bytes := make([]byte, args[1].Length())
		js.CopyBytesToGo(key2Bytes, args[1])

		addrAndPriKey := tss.ParseKey(key1Bytes, key2Bytes)
		return []interface{}{addrAndPriKey[0], addrAndPriKey[1]}
	})
}
