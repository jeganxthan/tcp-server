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
	Ignition bool
	Charging bool
	PowerCut bool
	Battery  float64 // Percentage or Voltage
	RSSI     int     // Signal strength
	Alarm    uint32  // JT808 alarm flags
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

// ParseGPS extracts GPS data and terminal attributes (GT06 protocol)
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

	if (courseStatus & 0x0400) == 0 {
		latitude = -latitude
	}
	if (courseStatus & 0x0800) != 0 {
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

// ParseJT808Header parses the JT808 header and returns body offset, phone, seq
// Handles both 2013 (6-byte phone) and 2019 (10-byte phone + version) formats
func ParseJT808Header(decoded []byte) (bodyOffset int, phone string, seq uint16, msgID uint16, bodyLen int) {
	if len(decoded) < 12 {
		return 0, "", 0, 0, 0
	}

	msgID = binary.BigEndian.Uint16(decoded[0:2])
	prop := binary.BigEndian.Uint16(decoded[2:4])
	bodyLen = int(prop & 0x03FF) // Bits 0-9 = body length

	// Check if bit 14 is set (JT808-2019 version flag)
	is2019 := (prop & 0x4000) != 0

	if is2019 {
		// 2019 format: ID(2) + Prop(2) + Version(1) + Phone(10 BCD) + Seq(2) = 17 bytes header
		if len(decoded) < 17 {
			return 0, "", 0, 0, 0
		}
		phone = hex.EncodeToString(decoded[5:15])
		seq = binary.BigEndian.Uint16(decoded[15:17])
		bodyOffset = 17
	} else {
		// 2013 format: ID(2) + Prop(2) + Phone(6 BCD) + Seq(2) = 12 bytes header
		phone = hex.EncodeToString(decoded[4:10])
		seq = binary.BigEndian.Uint16(decoded[10:12])
		bodyOffset = 12
	}

	return bodyOffset, phone, seq, msgID, bodyLen
}

// ParseJT808Location extracts GPS from a 0x0200 message
// Supports both JT808-2013 and JT808-2019 header formats
func ParseJT808Location(decoded []byte) (*GpsData, string) {
	bodyOffset, phone, _, _, _ := ParseJT808Header(decoded)
	if bodyOffset == 0 {
		return nil, ""
	}

	body := decoded[bodyOffset:]

	// Location body: Alarm(4) + Status(4) + Lat(4) + Lon(4) + Alt(2) + Speed(2) + Course(2) + Time(6) = 28 bytes
	if len(body) < 28 {
		fmt.Printf("⚠️ JT808 body too short: %d bytes, RAW: %s\n", len(body), hex.EncodeToString(body))
		return nil, ""
	}

	alarmRaw := binary.BigEndian.Uint32(body[0:4])
	statusRaw := binary.BigEndian.Uint32(body[4:8])
	latRaw := binary.BigEndian.Uint32(body[8:12])
	lonRaw := binary.BigEndian.Uint32(body[12:16])
	alt := binary.BigEndian.Uint16(body[16:18])
	speedRaw := binary.BigEndian.Uint16(body[18:20])
	course := binary.BigEndian.Uint16(body[20:22])

	// Parse BCD time (6 bytes: YY MM DD HH MM SS)
	var gpsTime time.Time
	if len(body) >= 28 {
		year := 2000 + int(body[22]>>4)*10 + int(body[22]&0x0F)
		month := time.Month(int(body[23]>>4)*10 + int(body[23]&0x0F))
		day := int(body[24]>>4)*10 + int(body[24]&0x0F)
		hour := int(body[25]>>4)*10 + int(body[25]&0x0F)
		minute := int(body[26]>>4)*10 + int(body[26]&0x0F)
		second := int(body[27]>>4)*10 + int(body[27]&0x0F)
		gpsTime = time.Date(year, month, day, hour, minute, second, 0, time.UTC)
	}

	// Status bits
	ignition := (statusRaw & 0x00000001) != 0
	fixed := (statusRaw & 0x00000002) != 0

	// Parse Additional Information (TLV items) starting at body[28]
	sats := 0
	rssi := int(alt)
	extra := body[28:]
	for i := 0; i+2 <= len(extra); {
		id := extra[i]
		length := int(extra[i+1])
		if i+2+length > len(extra) {
			break
		}
		value := extra[i+2 : i+2+length]
		
		switch id {
		case 0x30: // Signal Strength
			if length >= 1 { rssi = int(value[0]) }
		case 0x31: // Satellites
			if length >= 1 { sats = int(value[0]) }
		}
		i += 2 + length
	}

	// Debug: show raw values
	fmt.Printf("   🔍 RAW: Alarm=0x%08X Status=0x%08X Lat=%d Lon=%d Sats=%d RSSI=%d\n",
		alarmRaw, statusRaw, latRaw, lonRaw, sats, rssi)

	return &GpsData{
		Time:       gpsTime,
		Latitude:   float64(latRaw) / 1000000.0,
		Longitude:  float64(lonRaw) / 1000000.0,
		Speed:      float64(speedRaw) / 10.0,
		Course:     course,
		Valid:      fixed,
		Ignition:   ignition,
		Alarm:      alarmRaw,
		Satellites: sats,
		RSSI:       rssi,
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
