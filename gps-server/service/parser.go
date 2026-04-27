package service

import (
	"encoding/hex"
)

// GT06 IMEI is usually in login packet (protocol 0x01)
func ParseIMEI(data []byte) string {
	if len(data) < 12 {
		return ""
	}

	// IMEI bytes (simplified assumption)
	imeiBytes := data[4:12]

	return hex.EncodeToString(imeiBytes)
}
