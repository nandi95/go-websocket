# 🔌 Go WebSocket Learning Project

## 📝 Overview

This project is a simple WebSocket server implementation in Go, created as a learning exercise to understand the
WebSocket protocol. It's **not intended for production use** but rather serves as an educational tool to explore how
WebSockets work under the hood.

## ✨ Features

- 🤝 WebSocket handshake implementation
- ⚡ Every connection is handled in a separate goroutine
- 📦 Frame parsing and construction
- 📊 Support for various opcodes:
    - Text messages
    - Binary messages
    - Ping/Pong for keepalive
    - Close frames for connection termination

## ⚠️ Limitations

- 🚫 No support for continuation frames (fragmented messages)
- 🔍 Basic error handling
- 🧪 Created solely for learning purposes
- 💓 No health checks to see if the connection is still alive
- 📌 Missing some protocol features required for production use
- 📋 And probably many more

## 🚀 Getting Started

1. Clone the repository
2. Run the server:
   ```
   go run main.go
   ```
3. Open the `index.html` file in your browser

## 🔧 How It Works

The server handles WebSocket connections by:

1. Performing the initial HTTP upgrade handshake
2. Parsing incoming WebSocket frames according to the RFC 6455 specification (including extended length payloads)
3. Responding appropriately based on the frame opcode

## 📚 Learning Resources

- [RFC 6455 - The WebSocket Protocol](https://tools.ietf.org/html/rfc6455)
- [MDN WebSocket Documentation](https://developer.mozilla.org/en-US/docs/Web/API/WebSockets_API)

## 📣 Feedback

Feel free to experiment with this code, modify it, and use it as a starting point for your own WebSocket
implementations!