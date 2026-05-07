package service

import (
	"encoding/binary"
)

// CRC16_X25 calculates the CRC16-X25 checksum used in GT06 protocol
func CRC16_X25(data []byte) uint16 {
	crc := uint16(0xFFFF)
	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if crc&1 != 0 {
				crc = (crc >> 1) ^ 0x8408
			} else {
				crc >>= 1
			}
		}
	}
	return crc ^ 0xFFFF
}

// GenerateResponse creates a GT06 response packet
func GenerateResponse(packetType byte, index uint16) []byte {
	response := make([]byte, 10)
	response[0] = 0x78
	response[1] = 0x78
	response[2] = 0x05 // Length
	response[3] = packetType
	binary.BigEndian.PutUint16(response[4:6], index)
	
	crc := CRC16_X25(response[2:6])
	binary.BigEndian.PutUint16(response[6:8], crc)
	
	response[8] = 0x0D // \r
	response[9] = 0x0A // \n
	
	return response
}
