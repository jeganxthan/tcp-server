package tcp

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net"

	"gps-server/service"
	"gps-server/store"
)

func Start() {
	listener, err := net.Listen("tcp", ":5023")
	if err != nil {
		panic(err)
	}

	fmt.Println("🚀 TCP Server running on port 5023 (GT06 Protocol)")

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		fmt.Println("✅ Connected:", conn.RemoteAddr())
		go handleConnection(conn)
	}
}

type session struct {
	conn net.Conn
	imei string
}

func handleConnection(conn net.Conn) {
	defer func() {
		fmt.Println("❌ Disconnected:", conn.RemoteAddr())
		conn.Close()
	}()

	s := &session{conn: conn}
	reader := bufio.NewReader(conn)
	buffer := make([]byte, 0, 4096)

	for {
		tmp := make([]byte, 1024)
		n, err := reader.Read(tmp)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Read error:", err)
			}
			return
		}
		buffer = append(buffer, tmp[:n]...)

		// Frame decoding loop (Supports both GT06 and JT808)
		for len(buffer) >= 5 {
			var length int
			if buffer[0] == 0x7E {
				// 🎥 JT808 Frame decoding
				foundEnd := -1
				for i := 1; i < len(buffer); i++ {
					if buffer[i] == 0x7E {
						foundEnd = i
						break
					}
				}
				if foundEnd == -1 {
					break // Wait for more data
				}
				packet := buffer[:foundEnd+1]
				buffer = buffer[foundEnd+1:]
				s.processJT808(packet)
				continue
			} else if buffer[0] == 0x78 && buffer[1] == 0x78 {
				// 📦 GT06 Standard
				length = int(buffer[2]) + 5
			} else if buffer[0] == 0x79 && buffer[1] == 0x79 {
				// 📦 GT06 Extended
				if len(buffer) < 6 {
					break
				}
				length = int(binary.BigEndian.Uint16(buffer[2:4])) + 6
			} else {
				// Search for next possible header to recover from bad data
				found := false
				for i := 1; i < len(buffer)-1; i++ {
					if buffer[i] == 0x7E || (buffer[i] == 0x78 && buffer[i+1] == 0x78) || (buffer[i] == 0x79 && buffer[i+1] == 0x79) {
						buffer = buffer[i:]
						found = true
						break
					}
				}
				if !found {
					buffer = buffer[len(buffer)-1:]
				}
				continue
			}

			if len(buffer) < length {
				break
			}

			packet := buffer[:length]
			buffer = buffer[length:]
			s.processPacket(packet)
		}
	}
}

func (s *session) processJT808(data []byte) {
	decoded := service.DecodeJT808(data)
	if len(decoded) < 12 {
		return
	}

	// Dump raw hex for debugging
	fmt.Printf("🔬 RAW JT808 (%d bytes): %s\n", len(decoded), hex.EncodeToString(decoded))

	bodyOffset, phone, seq, msgID, _ := service.ParseJT808Header(decoded)
	if bodyOffset == 0 {
		fmt.Println("⚠️ Failed to parse JT808 header")
		return
	}

	fmt.Printf("🎥 JT808 Packet (%s) ID: 0x%04X, Phone: %s, Seq: %d\n", s.conn.RemoteAddr(), msgID, phone, seq)

	switch msgID {
	case 0x0100: // Terminal Register
		fmt.Println("📝 JT808 Registering:", phone)
		s.imei = phone
		response := service.GenerateJT808RegisterResponse(decoded[4:10], 0, seq, 0, "AUTH123")
		s.conn.Write(response)
		fmt.Println("📤 JT808 Register Response (0x8100) sent")

	case 0x0102: // Terminal Authentication
		fmt.Println("🔐 JT808 Authenticating:", phone)
		s.imei = phone
		s.sendJT808ACK(decoded, 0)

	case 0x0200: // Location Report
		gps, id := service.ParseJT808Location(decoded)
		if gps != nil {
			s.imei = id
			status := "📡 NO FIX"
			if gps.Valid {
				status = "🛰️ FIXED"
			}
			acc := "OFF"
			if gps.Ignition {
				acc = "ON"
			}
			alarmStr := ""
			if gps.Alarm != 0 {
				alarmStr = fmt.Sprintf(" | 🚨 ALARM: 0x%08X", gps.Alarm)
			}
			fmt.Printf("📍 JT808 GPS [%s] %s | ACC: %s | Lat: %.6f, Lng: %.6f | Speed: %.1f | Alt: %dm%s\n",
				s.imei, status, acc, gps.Latitude, gps.Longitude, gps.Speed, gps.RSSI, alarmStr)
		}
		s.sendJT808ACK(decoded, 0)

	case 0x0002: // Heartbeat
		fmt.Println("💓 JT808 Heartbeat from", phone)
		s.sendJT808ACK(decoded, 0)

	case 0x0704: // Batch Location Upload
		fmt.Printf("📦 JT808 Batch Location Upload from %s (%d bytes)\n", phone, len(decoded))
		s.sendJT808ACK(decoded, 0)

	case 0x0900: // Data Uplink (Multimedia / AI events)
		fmt.Printf("🤖 JT808 AI/Media Data from %s (%d bytes)\n", phone, len(decoded))
		s.sendJT808ACK(decoded, 0)

	case 0x0800, 0x0801: // Multimedia Event / Data Upload
		fmt.Printf("📸 JT808 Multimedia from %s (%d bytes)\n", phone, len(decoded))
		s.sendJT808ACK(decoded, 0)

	default:
		fmt.Printf("❓ Unknown JT808 ID: 0x%04X (%d bytes)\n", msgID, len(decoded))
		s.sendJT808ACK(decoded, 0)
	}
}

func (s *session) sendJT808ACK(decoded []byte, result byte) {
	_, _, seq, msgID, _ := service.ParseJT808Header(decoded)
	// Use raw phone bytes from the decoded packet
	var phone []byte
	prop := binary.BigEndian.Uint16(decoded[2:4])
	if (prop & 0x4000) != 0 {
		phone = decoded[5:15]
	} else {
		phone = decoded[4:10]
	}
	response := service.GenerateJT808Response(phone, 0, seq, msgID, result)
	s.conn.Write(response)
}


func (s *session) processPacket(data []byte) {
	if len(data) < 5 {
		return
	}

	// Protocol byte is at index 3 for 0x7878 and index 4 for 0x7979
	protocol := data[3]
	if data[0] == 0x79 {
		protocol = data[4]
	}

	fmt.Printf("📦 PACKET (%s) Type: 0x%02X\n", s.conn.RemoteAddr(), protocol)

	switch protocol {
	case 0x01: // Login
		s.imei = service.ParseIMEI(data)
		fmt.Println("🔐 Login attempt IMEI:", s.imei)

		if _, ok := store.GetDevice(s.imei); ok {
			fmt.Println("✅ Device authenticated:", s.imei)
			s.sendACK(0x01, service.GetSequenceIndex(data))
		} else {
			fmt.Println("❌ Unknown device:", s.imei)
			// Traccar still ACKs login even for unknown devices sometimes, 
			// but here we keep it strict or follow Traccar's lead.
			s.sendACK(0x01, service.GetSequenceIndex(data))
		}

	case 0x12, 0x22, 0x10, 0x27, 0x1A: // GPS Data
		if s.imei == "" {
			fmt.Println("⚠️ GPS data received before login - ignoring")
			return
		}
		gps, err := service.ParseGPS(data)
		if err != nil {
			fmt.Println("❌ Parse error:", err)
			return
		}
		status := "📡 OFFLINE"
		if gps.Valid {
			status = "🛰️ ONLINE"
		}
		fmt.Printf("📍 GPS [%s] %s: Lat: %.6f, Lng: %.6f, Speed: %.1f, Course: %d, Sats: %d\n", 
			s.imei, status, gps.Latitude, gps.Longitude, gps.Speed, gps.Course, gps.Satellites)

	case 0x13, 0x23, 0x16, 0x26: // Status / Heartbeat / LBS Status
		info := "💓 Heartbeat"
		if protocol == 0x16 || protocol == 0x26 {
			info = "📶 LBS Status"
		}
		fmt.Printf("%s received from %s\n", info, s.imei)
		s.sendACK(protocol, service.GetSequenceIndex(data))

	case 0x15, 0x21: // String / Command Response
		fmt.Printf("💬 Command response received from %s\n", s.imei)

	case 0x17, 0x18, 0x19: // LBS / WIFI / Multi-LBS
		fmt.Printf("📶 LBS/WIFI data received from %s\n", s.imei)
		s.sendACK(protocol, service.GetSequenceIndex(data))

	default:
		fmt.Printf("❓ Unknown protocol: 0x%02X, RAW: %s\n", protocol, hex.EncodeToString(data))
	}
}

func (s *session) sendACK(packetType byte, index uint16) {
	response := service.GenerateResponse(packetType, index)
	s.conn.Write(response)
	fmt.Printf("📤 ACK sent for 0x%02X (Index: %d)\n", packetType, index)
}
