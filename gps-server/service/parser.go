package service

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"time"
)

type GpsData struct {
	Time       time.Time
	Latitude   float64
	Longitude  float64
	Speed      float64
	Course     uint16
	Valid      bool
	Satellites int
	// Extended Attributes (Traccar-style)
	Ignition   bool
	Charging   bool
	PowerCut   bool
	Battery    float64 // Percentage or Voltage
	RSSI       int     // Signal strength
}

// ParseIMEI extracts the IMEI from a login packet (protocol 0x01)
func ParseIMEI(data []byte) string {
	if len(data) < 12 {
		return ""
	}
	imeiHex := hex.EncodeToString(data[4:12])
	if len(imeiHex) > 15 && imeiHex[0] == '0' {
		return imeiHex[1:]
	}
	return imeiHex
}

// ParseStatus parses terminal information from protocol 0x13, 0x16, 0x23
func ParseStatus(data []byte) (ignition, charging, powerCut bool, battery int) {
	if len(data) < 5 {
		return
	}
	// The status byte is usually after the protocol type
	statusByte := data[4]
	ignition = (statusByte & 0x02) != 0
	charging = (statusByte & 0x04) != 0
	powerCut = (statusByte & 0x08) == 0 // 0 means power cut in some variants
	
	// Battery level is often the next byte
	if len(data) > 5 {
		battery = int(data[5])
	}
	return
}

// ParseGPS extracts GPS data and terminal attributes
func ParseGPS(data []byte) (*GpsData, error) {
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
	
	sats := int(data[offset] & 0x0F)
	offset += 1
	
	latRaw := binary.BigEndian.Uint32(data[offset : offset+4])
	latitude := float64(latRaw) / 60.0 / 30000.0
	offset += 4
	
	lonRaw := binary.BigEndian.Uint32(data[offset : offset+4])
	longitude := float64(lonRaw) / 60.0 / 30000.0
	offset += 4
	
	speed := float64(data[offset])
	offset += 1
	
	courseStatus := binary.BigEndian.Uint16(data[offset : offset+2])
	course := courseStatus & 0x3FF
	
	if (courseStatus & 0x0400) == 0 { latitude = -latitude }
	if (courseStatus & 0x0800) != 0 { longitude = -longitude }
	valid := (courseStatus & 0x1000) != 0

	// Check for optional terminal info at the end of some GPS packets
	// Usually: [GPS Data] [LBS Data] [Terminal Info] ...
	// This varies by model, but we can try to extract if packet is long enough
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
