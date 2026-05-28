package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type Alert struct {
	Type      string `json:"type"`
	Message   string `json:"message"`
	Severity  string `json:"severity"`
	Timestamp string `json:"timestamp"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var clients = make(map[*websocket.Conn]bool)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	clients[conn] = true
	log.Println("New App Client Connected")
}

func broadcast(alert Alert) {
	for client := range clients {
		err := client.WriteJSON(alert)
		if err != nil {
			log.Println("Client disconnected:", err)
			client.Close()
			delete(clients, client)
		}
	}
}

func triggerAlert(w http.ResponseWriter, alertType, message, severity, status string) {
	alert := Alert{
		Type:      alertType,
		Message:   message,
		Severity:  severity,
		Timestamp: time.Now().Format(time.RFC3339),
	}
	broadcast(alert)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": status})
}

// createHandler helps generate endpoints quickly
func createHandler(alertType, message, severity, status string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Triggering Alert: %s (%s)", alertType, severity)
		triggerAlert(w, alertType, message, severity, status)
	}
}

func main() {
	// --- ADAS ALERTS (Advanced Driver Assistance Systems) ---
	http.HandleFunc("/alert/pedestrian", createHandler("PEDESTRIAN_ALERT", "Pedestraian Alert", "CRITICAL", "Pedestraian Alert triggered"))
	http.HandleFunc("/alert/theft", createHandler("THEFT_ALERT", "Theft Alert", "CRITICAL", "Theft Alert triggered"))
	http.HandleFunc("/alert/forward-collision", createHandler("FORWARD_COLLISION", "Forward Collision Warning!", "CRITICAL", "Forward collision alert triggered"))
	http.HandleFunc("/alert/collision-recording", createHandler("COLLISION_RECORDING", "Collision Recording Started", "LOW", "Collision recording triggered"))

	// --- DMS ALERTS (Driver Monitoring System) ---
	http.HandleFunc("/alert/drowsy", createHandler("DROWSINESS", "Drowsiness Detected!", "CRITICAL", "Drowsiness alert triggered"))
	http.HandleFunc("/alert/distraction", createHandler("DRIVER_DISTRACTION", "Driver Distraction Detected", "HIGH", "Distraction alert triggered"))
	http.HandleFunc("/alert/mobile", createHandler("CELL_PHONE_USAGE", "Cell Phone Usage Detected", "HIGH", "Cell phone alert triggered"))
	http.HandleFunc("/alert/no-helmet", createHandler("HELMET_DETECTION", "No Helmet Detected", "HIGH", "Helmet alert triggered"))
	// CAMERA_HIDDEN maps to Camera View Blocked
	http.HandleFunc("/alert/camera-hidden", createHandler("CAMERA_HIDDEN", "Camera View Blocked!", "HIGH", "Camera hidden alert triggered"))
	http.HandleFunc("/alert/smoking", createHandler("SMOKING", "Smoking Detection", "MEDIUM", "Smoking alert triggered"))
	http.HandleFunc("/alert/driver-detected", createHandler("DRIVER_DETECTED", "Driver Detected Successfully", "LOW", "Driver detected event triggered"))

	// --- VEHICLE DYNAMICS ALERTS ---
	http.HandleFunc("/alert/speed", createHandler("SPEED_ALERT", "Speed Limit Exceeded!", "HIGH", "Speed alert triggered"))
	http.HandleFunc("/alert/harsh-brake", createHandler("HARSH_BRAKE", "Harsh Braking Detected", "MEDIUM", "Harsh brake alert triggered"))
	http.HandleFunc("/alert/harsh-acceleration", createHandler("HARSH_ACCELERATION", "Harsh Acceleration Detected", "MEDIUM", "Harsh acceleration alert triggered"))
	http.HandleFunc("/alert/sharp-turn", createHandler("SHARP_TURN", "Sharp Turning Detected", "MEDIUM", "Sharp turn alert triggered"))

	http.HandleFunc("/ws", wsHandler)

	// --- DRIVER MANAGEMENT ---
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Driver Login Attempted")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "Driver authenticated",
			"token":   "fake-jwt-token-12345",
		})
	})

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		log.Println("New Driver Registering via Face ID")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "Driver face biometrics enrolled",
		})
	})

	fmt.Println("ThirdEye Multi-Alert AI Server running on :8080")
	fmt.Println("Ready to receive triggers for 14 different event types + Driver Auth.")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
