package main

import (
	"encoding/hex"
)

func prettyBytes(bytes []byte) string {
	res := ""
	hexStr := hex.EncodeToString(bytes)
	for i := 0; i < len(hexStr); i++ {
		if i > 0 && i%16 == 0 {
			res += "\n"
		} else if i > 0 && i%2 == 0 {
			res += " "
		}
		res += string(hexStr[i])
	}
	return res
}

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}
