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

func Word(b []byte) uint16 {
	var _ = b[1]
	return uint16(b[1]) | uint16(b[0])<<8
}

func DWord(b []byte) uint32 {
	var _ = b[3]
	return uint32(b[3]) | uint32(b[2])<<8 | uint32(b[1])<<16 | uint32(b[0])<<24
}
