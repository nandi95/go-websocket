package main

import (
	"crypto"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade") ||
		!strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
		http.Error(w, "Connection and Upgrade headers required", http.StatusBadRequest)
		return
	}

	if r.Header.Get("Sec-WebSocket-Key") == "" {
		http.Error(w, "Missing Sec-WebSocket-Key header key", http.StatusBadRequest)
	}

	magicString := "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

	h := crypto.SHA1.New()
	_, err := h.Write([]byte(r.Header.Get("Sec-WebSocket-Key") + magicString))

	if err != nil {
		http.Error(w, "¯\\_(ツ)_/¯", http.StatusInternalServerError)
	}

	// base64_encode(sha1(key + magicString))
	acceptedKey := base64.StdEncoding.EncodeToString(h.Sum(nil))

	w.Header().Set("Upgrade", "websocket")
	w.Header().Set("Connection", "Upgrade")
	w.Header().Set("Sec-WebSocket-Accept", acceptedKey)
	// respond with a 101 Switching Protocols status code
	w.WriteHeader(http.StatusSwitchingProtocols)
	fmt.Println("Handshake successful!")

	fmt.Println("Reading frames...")

	conn, _, err := w.(http.Hijacker).Hijack()

	if err != nil {
		http.Error(w, "Failed to hijack connection", http.StatusInternalServerError)
		return
	}

	// handle the connection in a separate goroutine
	go handleWebsocketConnection(conn)
}

func handleWebsocketConnection(conn net.Conn) {
	defer conn.Close()

	for {
		fin, o, p, err := readFrame(conn)

		if err != nil {
			if errors.Is(err, io.EOF) {
				// this could be just the initial handshake
				fmt.Println("Connection closed by client")
				break
			}

			fmt.Println("readFrame err:", err)
			break
		}

		// not implementing continuation frames
		switch o {
		case Text:
			fmt.Println("Text:", string(p))
		case Ping:
			sendFrame(conn, Pong, p)
		case Close:
			sendFrame(conn, Close, p)
			fmt.Println("Close:", string(p))
			return
		case Binary:
			// write to file
			os.Create("output.bin")
			fmt.Println("Binary: written to file")
		default:
			fmt.Println("Unsupported opcode:", o)
		}

		if !fin {
			fmt.Println("Continuation frames not supported")
			break
		}

	}
}

func main() {
	http.HandleFunc("/ws", websocketHandler)
	fmt.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
