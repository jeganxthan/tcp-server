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

// DecodeJT808 handles the unescaping logic for JT808 protocol
func DecodeJT808(data []byte) []byte {
	if len(data) < 2 {
		return data
	}
	// Remove start/end 0x7E
	body := data[1 : len(data)-1]
	decoded := make([]byte, 0, len(body))
	for i := 0; i < len(body); i++ {
		if body[i] == 0x7D && i+1 < len(body) {
			if body[i+1] == 0x01 {
				decoded = append(decoded, 0x7D)
				i++
			} else if body[i+1] == 0x02 {
				decoded = append(decoded, 0x7E)
				i++
			}
		} else {
			decoded = append(decoded, body[i])
		}
	}
	return decoded
}

// ParseJT808Location extracts GPS from a 0x0200 message
func ParseJT808Location(data []byte) (*GpsData, string) {
	// Header: ID(2) + Property(2) + Phone(6 BCD) + Seq(2) = 12 bytes
	if len(data) < 40 {
		return nil, ""
	}
	
	phone := hex.EncodeToString(data[4:10])
	body := data[12:]
	
	// Message Body for 0x0200:
	// Alarm(4) + Status(4) + Lat(4) + Lon(4) + Alt(2) + Speed(2) + Course(2) + Time(6)
	statusRaw := binary.BigEndian.Uint32(body[4:8])
	latRaw := binary.BigEndian.Uint32(body[8:12])
	lonRaw := binary.BigEndian.Uint32(body[12:16])
	alt := binary.BigEndian.Uint16(body[16:18])
	speedRaw := binary.BigEndian.Uint16(body[18:20])
	course := binary.BigEndian.Uint16(body[20:22])
	
	// JT808 Status Bits:
	// Bit 0: 0=ACC Off, 1=ACC On
	// Bit 1: 0=Not Fixed, 1=Fixed
	ignition := (statusRaw & 0x00000001) != 0
	fixed := (statusRaw & 0x00000002) != 0
	
	return &GpsData{
		Time:      time.Now(),
		Latitude:  float64(latRaw) / 1000000.0,
		Longitude: float64(lonRaw) / 1000000.0,
		Speed:     float64(speedRaw) / 10.0,
		Course:    course,
		Valid:     fixed,
		Ignition:  ignition,
		RSSI:      int(alt),
	}, phone
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
