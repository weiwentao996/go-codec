package codec

import (
	"bytes"
	"strings"
)

func lengthAlignRightPadding(d []byte, length int) []byte {
	if len(d) >= length {
		return d
	}
	padding := make([]byte, length-len(d))
	return append(d, padding...)
}

func decodeFixedString(strBytes []byte) string {
	newByte := bytes.ReplaceAll(strBytes, []byte{0}, []byte{32})
	return strings.TrimSpace(string(newByte))
}
