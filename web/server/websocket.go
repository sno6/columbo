package server

import (
	"encoding/json"
	"log"
	"math"
	"net/http"

	"github.com/sno6/columbo/fingerprint"

	"github.com/gorilla/websocket"
)

// WS holds the dependencies for the Websocket HTTP handler.
type WS struct {
	Store  fingerprint.Store
	logger *log.Logger
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(*http.Request) bool { return true },
}

type WebsocketMessage struct {
	Action     Action `json:"action"`
	BufferSize int    `json:"bufferSize,omitempty"`
}

type Action int

const (
	Init Action = iota
	Starting
	Stopping
)

// WS handles the websocket connection.
func (ws *WS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		ws.logger.Printf("ws: unable to upgrade connection: %v\n", err)
		return
	}

	// Stream passes the audio input from the microphone from the frontend
	// through to the fingerprint analyser to make predictions on the data.
	stream := make(chan []float64)

	// TODO: Should this be here?
	// This should be some sort of shared configuration struct..
	bufferSize := 4096

	// Create a new audio analyser to analyse the audio buffers as it comes
	// in over the Websocket connection.
	analyser, output := fingerprint.NewAnalyser(stream, ws.Store)
	go analyser.Start()

	// Send through the output from the analyser to the websocket connection.
	go outputBridge(output, conn)

	for {
		t, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, 1005) {
				return
			}
			ws.logger.Printf("ws: error reading message: %v\n", err)
		}

		switch t {
		case websocket.TextMessage:
			// A TextMessage is one from the frontend with configuration information, or
			// state info such as whether the recorder is running or has been stopped.
			var wsm WebsocketMessage
			if err := json.Unmarshal(msg, &wsm); err != nil {
				ws.logger.Printf("ws: error reading text message: %v\n", err)
				continue
			}

			ws.logger.Printf("Received TextMessage: %v\n", wsm)

			switch wsm.Action {
			case Init:
				if wsm.BufferSize != 0 {
					bufferSize = wsm.BufferSize
				}
			}
		case websocket.BinaryMessage:
			// Received Audio data, we can just normalize this and send it through to the analyser.
			buf := make([]float64, bufferSize)
			for i := range buf {
				buf[i] = normalizeBytes(msg[2*i:])
			}
			stream <- buf
		}
	}
}

func outputBridge(output chan *fingerprint.Packet, conn *websocket.Conn) {
	for packet := range output {
		if err := conn.WriteJSON(packet); err != nil {
			log.Fatal(err)
		}
	}
}

func normalizeBytes(b []byte) float64 {
	i := int16(uint16(b[0]) | uint16(b[1])<<8)
	return newRange(float64(i), math.MinInt16, math.MaxInt16, -1, 1)
}

func newRange(value, oMin, oMax, nMin, nMax float64) float64 {
	return (((value - oMin) * (nMax - nMin)) / (oMax - oMin)) + nMin
}
