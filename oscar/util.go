package oscar

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

// splitBy splits string in chunks of n
// taken from: https://stackoverflow.com/a/69403603
func splitBy(s string, n int) []string {
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

func prettyBytes(bytes []byte) string {
	hexStr := hex.EncodeToString(bytes)
	rows := splitBy(hexStr, 16)

	res := ""
	for _, row := range rows {
		byteGroups := splitBy(row, 2)
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
