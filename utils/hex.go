package utils

import "encoding/hex"

func BytesToHexPrefixed(b []byte) string {
	return "0x" + hex.EncodeToString(b)
}
