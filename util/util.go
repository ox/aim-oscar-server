package util

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

// splitBy splits string in chunks of n
// taken from: https://stackoverflow.com/a/69403603
func SplitBy(s string, n int) []string {
	var ss []string
	for i := 1; i < len(s); i++ {
		if i%n == 0 {
			ss = append(ss, s[:i])
			s = s[i:]
			i = 1
		}
	}
	ss = append(ss, s)
	return ss
}

func PrettyBytes(bytes []byte) string {
	hexStr := hex.EncodeToString(bytes)
	rows := SplitBy(hexStr, 16)

	res := ""
	for _, row := range rows {
		byteGroups := SplitBy(row, 2)
		// Align string view to full 16 bytes + spaces
		res += fmt.Sprintf("%-23s", strings.Join(byteGroups, " "))

		res += " |"
		for _, r := range byteGroups {
			n, err := strconv.ParseInt(r, 16, 8)
			if err != nil || (n < 32 || n > 126) {
				res += "."
			} else {
				res += string(rune(n))
			}
		}
		res += "|\n"
	}

	return strings.TrimSpace(res)
}

func Word(x uint16) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, x)
	return b
}

func Dword(x uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, x)
	return b
}

func LPString(x string) []byte {
	return append([]byte{uint8(len(x))}, []byte(x)...)
}

func LPUint16String(x string) []byte {
	return append(Word(uint16(len(x))), []byte(x)...)
}
