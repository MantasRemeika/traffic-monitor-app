package main

import (
	"net"
	"testing"
	"time"
)

// TestServerHandlesMultipleConnections tests if the server can handle multiple simultaneous connections.
func TestServerHandlesMultipleClientMessages(t *testing.T) {
	go StartServer(":9000")
	time.Sleep(200 * time.Millisecond) // Give server time to start

	const numClients = 5
	conns := make([]net.Conn, numClients)
	for i := 0; i < numClients; i++ {
		conn, err := net.Dial("tcp", "localhost:9000")
		if err != nil {
			t.Fatalf("Client %d failed to connect: %v", i, err)
		}
		conns[i] = conn
	}

	message := "Hello, server!"
	for _, conn := range conns {
		_, err := conn.Write([]byte(message))
		if err != nil {
			t.Fatalf("Failed to write to server: %v", err)
		}
		defer conn.Close()
	}
}


