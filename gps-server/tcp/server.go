package tcp

import (
	"encoding/hex"
	"fmt"
	"net"

	"gps-server/service"
	"gps-server/store"
)

func Start() {
	listener, err := net.Listen("tcp", ":5023")
	if err != nil {
		panic(err)
	}

	fmt.Println("🚀 TCP Server running on port 5023")

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		fmt.Println("✅ Connected:", conn.RemoteAddr())

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer func() {
		fmt.Println("❌ Disconnected:", conn.RemoteAddr())
		conn.Close()
	}()

	buffer := make([]byte, 1024)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			return
		}

		data := buffer[:n]

		fmt.Printf("📦 RAW (%s): %s\n", conn.RemoteAddr(), hex.EncodeToString(data))

		if len(data) < 5 {
			continue
		}

		protocol := data[3]

		switch protocol {

		// 🔐 Login Packet
		case 0x01:
			fmt.Println("🔐 Login packet")

			imei := service.ParseIMEI(data)
			fmt.Println("📱 IMEI:", imei)

			if _, ok := store.GetDevice(imei); ok {
				fmt.Println("✅ Device authenticated")

				sendLoginACK(conn)
			} else {
				fmt.Println("❌ Unknown device:", imei)
			}

		// 📍 GPS Data Packet
		case 0x12:
			fmt.Println("📍 GPS data received")

			// TODO: parse lat/lng next step

		default:
			fmt.Println("❓ Unknown protocol:", protocol)
		}
	}
}

func sendLoginACK(conn net.Conn) {
	response := []byte{
		0x78, 0x78,
		0x05,
		0x01,
		0x00, 0x01,
		0x0D, 0x0A,
	}

	conn.Write(response)
	fmt.Println("📤 ACK sent")
}
