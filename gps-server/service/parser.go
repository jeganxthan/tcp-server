package service

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"time"
)

type GpsData struct {
	Time      time.Time
	Latitude  float64
	Longitude float64
	Speed     float64
	Course    uint16
	Valid     bool
	Satellites int
}

// ParseIMEI extracts the IMEI from a login packet (protocol 0x01)
func ParseIMEI(data []byte) string {
	if len(data) < 12 {
		return ""
	}

	// GT06 IMEI is 8 bytes, starts at index 4 in a standard 0x7878 packet
	// It's often hex encoded, where the first digit might be a padding 0
	imeiHex := hex.EncodeToString(data[4:12])
	if len(imeiHex) > 15 && imeiHex[0] == '0' {
		return imeiHex[1:]
	}
	return imeiHex
}

// ParseGPS extracts GPS data from a packet (protocol 0x12 or 0x22 etc)
func ParseGPS(data []byte) (*GpsData, error) {
	// Header(2) + Length(1) + Protocol(1) = 4 bytes offset
	// Date Time (6 bytes)
	if len(data) < 30 {
		return nil, fmt.Errorf("packet too short")
	}
	
	offset := 4
	year := int(data[offset]) + 2000
	month := time.Month(data[offset+1])
	day := int(data[offset+2])
	hour := int(data[offset+3])
	minute := int(data[offset+4])
	second := int(data[offset+5])
	
	gpsTime := time.Date(year, month, day, hour, minute, second, 0, time.UTC)
	offset += 6
	
	// Satellites
	sats := int(data[offset] & 0x0F)
	offset += 1
	
	// Latitude
	latRaw := binary.BigEndian.Uint32(data[offset : offset+4])
	latitude := float64(latRaw) / 60.0 / 30000.0
	offset += 4
	
	// Longitude
	lonRaw := binary.BigEndian.Uint32(data[offset : offset+4])
	longitude := float64(lonRaw) / 60.0 / 30000.0
	offset += 4
	
	// Speed
	speed := float64(data[offset])
	offset += 1
	
	// Course & Status
	courseStatus := binary.BigEndian.Uint16(data[offset : offset+2])
	course := courseStatus & 0x3FF
	
	// Direction flags (matching Traccar logic)
	// Bit 10: 1 = North, 0 = South
	// Bit 11: 1 = West, 0 = East
	// Bit 12: 1 = Fixed (Valid), 0 = Not Fixed
	
	if (courseStatus & 0x0400) == 0 { // South latitude
		latitude = -latitude
	}
	if (courseStatus & 0x0800) != 0 { // West longitude
		longitude = -longitude
	}
	
	valid := (courseStatus & 0x1000) != 0
	
	return &GpsData{
		Time:       gpsTime,
		Latitude:   latitude,
		Longitude:  longitude,
		Speed:      speed,
		Course:     course,
		Valid:      valid,
		Satellites: sats,
	}, nil
}

// GetSequenceIndex extracts the sequence number from the end of the packet
func GetSequenceIndex(data []byte) uint16 {
	if len(data) < 6 {
		return 0
	}
	// The sequence number is the 2 bytes before the CRC and tail
	// Packet: [Header 2][Len 1][Type 1][Data...][Seq 2][CRC 2][Tail 2]
	// Index from end: Tail(2) + CRC(2) + Seq(2) = 6
	return binary.BigEndian.Uint16(data[len(data)-6 : len(data)-4])
}
