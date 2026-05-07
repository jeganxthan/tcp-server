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

// JT808_Checksum calculates the XOR checksum for JT808
func JT808_Checksum(data []byte) byte {
	var checksum byte
	for _, b := range data {
		checksum ^= b
	}
	return checksum
}

// GenerateJT808Response creates a standard 0x8001 General Response
func GenerateJT808Response(phone []byte, seq uint16, msgSeq uint16, msgID uint16, result byte) []byte {
	// Body: MsgSeq(2) + MsgID(2) + Result(1) = 5 bytes
	body := make([]byte, 5)
	binary.BigEndian.PutUint16(body[0:2], msgSeq)
	binary.BigEndian.PutUint16(body[2:4], msgID)
	body[4] = result

	// Header: ID(2) + Prop(2) + Phone(6) + Seq(2) = 12 bytes
	header := make([]byte, 12)
	binary.BigEndian.PutUint16(header[0:2], 0x8001)
	binary.BigEndian.PutUint16(header[2:4], uint16(len(body)))
	copy(header[4:10], phone)
	binary.BigEndian.PutUint16(header[10:12], seq)

	packet := append(header, body...)
	checksum := JT808_Checksum(packet)
	
	// Escape and wrap in 0x7E (Simplified for now)
	final := []byte{0x7E}
	final = append(final, packet...)
	final = append(final, checksum)
	final = append(final, 0x7E)
	
	return final
}
