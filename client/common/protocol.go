package common

import (
	"encoding/binary"
	"io"
	"net"
)

// sendMessage sends a length-prefixed message through the connection, avoiding short writes.
func sendMessage(conn net.Conn, message string) error {
	payload := []byte(message)
	length := uint16(len(payload))

	header := make([]byte, 2)
	binary.BigEndian.PutUint16(header, length)

	data := append(header, payload...)

	totalSent := 0
	for totalSent < len(data) {
		n, err := conn.Write(data[totalSent:])
		if err != nil {
			return err
		}
		totalSent += n
	}
	return nil
}

// recvMessage receives a length-prefixed message from the connection, avoiding short reads.
func recvMessage(conn net.Conn) (string, error) {
	header := make([]byte, 2)
	if _, err := io.ReadFull(conn, header); err != nil {
		return "", err
	}
	length := binary.BigEndian.Uint16(header)

	payload := make([]byte, length)
	if _, err := io.ReadFull(conn, payload); err != nil {
		return "", err
	}
	return string(payload), nil
}