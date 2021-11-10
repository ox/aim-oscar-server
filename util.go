package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
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

func printBytes(bytes []byte) {
	fmt.Printf("%s\n", prettyBytes(bytes))
}

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

func readNBytes(buf *bytes.Buffer, n int) ([]byte, error) {
	res := make([]byte, n)
	_, err := io.ReadFull(buf, res)
	return res, err
}

func mustReadNBytes(buf *bytes.Buffer, n int) []byte {
	res, err := readNBytes(buf, n)
	panicIfError(err)
	return res
}
