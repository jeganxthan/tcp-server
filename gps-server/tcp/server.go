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

		// Frame decoding loop (inspired by Traccar's Gt06FrameDecoder)
		for len(buffer) >= 5 {
			var length int
			if buffer[0] == 0x78 && buffer[1] == 0x78 {
				length = int(buffer[2]) + 5
			} else if buffer[0] == 0x79 && buffer[1] == 0x79 {
				if len(buffer) < 6 {
					break
				}
				length = int(binary.BigEndian.Uint16(buffer[2:4])) + 6
			} else {
				// Search for next possible header to recover from bad data
				found := false
				for i := 1; i < len(buffer)-1; i++ {
					if (buffer[i] == 0x78 && buffer[i+1] == 0x78) || (buffer[i] == 0x79 && buffer[i+1] == 0x79) {
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
				// Wait for more data
				break
			}

			packet := buffer[:length]
			buffer = buffer[length:]

			s.processPacket(packet)
		}
	}
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

	case 0x12, 0x22, 0x10, 0x27: // GPS Data
		if s.imei == "" {
			fmt.Println("⚠️ GPS data received before login - ignoring")
			return
		}
		gps, err := service.ParseGPS(data)
		if err != nil {
			fmt.Println("❌ Parse error:", err)
			return
		}
		status := "OFFLINE"
		if gps.Valid {
			status = "ONLINE"
		}
		fmt.Printf("📍 GPS [%s] %s: Lat: %.6f, Lng: %.6f, Speed: %.1f, Course: %d, Sats: %d\n", 
			s.imei, status, gps.Latitude, gps.Longitude, gps.Speed, gps.Course, gps.Satellites)

	case 0x13, 0x23: // Status (Heartbeat)
		fmt.Println("💓 Heartbeat received")
		s.sendACK(protocol, service.GetSequenceIndex(data))

	case 0x15: // String (Command Response)
		fmt.Println("💬 String/Command response received")

	default:
		fmt.Printf("❓ Unknown protocol: 0x%02X, RAW: %s\n", protocol, hex.EncodeToString(data))
	}
}

func (s *session) sendACK(packetType byte, index uint16) {
	response := service.GenerateResponse(packetType, index)
	s.conn.Write(response)
	fmt.Printf("📤 ACK sent for 0x%02X (Index: %d)\n", packetType, index)
}
