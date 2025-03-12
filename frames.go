package main

import (
	"fmt"
	"io"
	"net"
)

type Opcode = int

// using hexadecimal values for the opcodes, to seem clever...
const (
	Continuation = 0x0
	Text         = 0x1
	Binary       = 0x2
	Close        = 0x8
	Ping         = 0x9
	Pong         = 0xA
)

// readFrame reads a single WebSocket frame from the connection.
// Returns:
//   - fin: true if this is the final frame of a message
//   - opcode: low 4 bits of first byte
//   - payload: the unmasked data
//   - err: any I/O or parsing error
func readFrame(conn net.Conn) (fin bool, opcode Opcode, payload []byte, err error) {
	// read only the first two bytes
	header := make([]byte, 2)
	_, err = io.ReadFull(conn, header)

	if err != nil {
		return false, -1, nil, err
	}

	byte1, byte2 := header[0], header[1]

	// & is the bitwise AND operator used for masking

	// if the left most bit is 1, then this is the final frame
	fin = byte1&0x80 != 0
	// get the integer value of the last 4 bits only
	opcode = Opcode(byte1 & 0x0F)

	// if the left most bit is 1, then the payload is masked
	isMasked := byte2&0x80 != 0
	// the last 7 bits of the second byte is the payload length
	length := int(byte2 & 0x7F)

	// the value of length determines the real length of the payload
	length, err = readPayloadLength(conn, length)

	if err != nil {
		return false, -1, nil, err
	}

	var maskKey [4]byte

	if isMasked {
		// turn the mask key into a byte slice of length 4 and read the next 4 bytes
		_, err = io.ReadFull(conn, maskKey[:])
		if err != nil {
			return false, -1, nil, err
		}
	}

	if length > 10*1024*1024 {
		return false, -1, nil, fmt.Errorf("payload too large: %d bytes, max allowed: %d bytes", length, 10*1024*1024)

	}

	payload = make([]byte, length)
	_, err = io.ReadFull(conn, payload)

	if err != nil {
		return false, -1, nil, err
	}

	// unmask the payload
	if isMasked {
		// loop over the bytes
		for i := range payload {
			// XOR the payload:
			// at the index we check that the XOR the byte against the given mask key
			// modulo ensures that the mask key is repeated
			payload[i] ^= maskKey[i%4]
		}
	}

	return fin, opcode, payload, nil
}

func readPayloadLength(conn net.Conn, length int) (int, error) {
	if length < 126 {
		return length, nil
	}

	// the payload is a big endian number (byte has to be multiplied by 256*position from the right most byte)

	// if length == 126, the next 2 bytes are the payload length
	if length == 126 {
		// read the next 2 bytes
		payloadLength := make([]byte, 2)
		// the previous read already moved the pointer on the connection 2 bytes ahead
		_, err := io.ReadFull(conn, payloadLength)

		if err != nil {
			return 0, err
		}

		// shift the first byte 8 bits to the left and OR it with the second byte (creating a 16-bit integer)
		return int(payloadLength[0])<<8 | int(payloadLength[1]), nil
	}

	// length == 127
	// read the next 8 bytes
	payloadLength := make([]byte, 8)
	_, err := io.ReadFull(conn, payloadLength)

	if err != nil {
		return 0, err
	}

	// loop over the byte slice, multiple each byte by the position from the right most byte and add it to the length as an integer
	for i := 0; i < 8; i++ {
		length |= int(payloadLength[i]) << (56 - i*8)
	}

	return length, nil
}

func sendFrame(conn net.Conn, opcode Opcode, payload []byte) error {
	var header []byte

	// the first byte is made of the fin bit (leftmost bit) 3 RSV bits (unused), and the opcode (4 bits)
	// merges bits with the OR operator
	byte1 := byte(0x80) | byte(opcode)

	// the second byte is the mask bit (leftmost bit) and the payload length (7 bits)
	// no mask for server frames
	byte2 := byte(0)

	payloadLength := len(payload)

	if payloadLength < 126 {
		header = []byte{byte1, byte2 | byte(payloadLength)}
	} else if payloadLength < 0xFFFF+1 {
		// if the payload length is greater than 125, and less than 16 bit unsigned int (represented in hex)
		// the next 2 bytes are the payload length:
		// we shift the payload length 8 bits to the right (effectively discarding the last 8 bits)
		// when converting a 16-bit integer to a byte, the default behaviour is to take the least significant byte
		header = []byte{byte1, 126, byte(payloadLength >> 8), byte(payloadLength)}
	} else {
		header = []byte{byte1, 127}
		// again big-endian number...
		// so as we loop over, we add a new byte with the value of the payload length shifted 8 bits to the right
		for i := 0; i < 8; i++ {
			// we shift bits to the right by the multiples of 8 starting with 56
			header = append(header, byte(payloadLength>>(56-i*8)))
		}
	}

	// assign error inline
	if _, err := conn.Write(header); err != nil {
		return err
	}

	if payloadLength > 0 {
		if _, err := conn.Write(payload); err != nil {
			return err
		}
	}

	return nil
}
