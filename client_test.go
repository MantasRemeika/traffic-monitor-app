package main

import (
	"net"
	"strings"
	"testing"
	"time"
)

// TestServerHandlesMultipleConnections tests if the server can handle multiple simultaneous connections.
func TestHandlesMultipleClientConnections(t *testing.T) {
	go StartServer(":9000")
	time.Sleep(200 * time.Millisecond) // Give server time to start

	const numClients = 5
	connections := make([]net.Conn, numClients)

	// Establish multiple client connections to the server
	for i := 0; i < numClients; i++ {
		connection, err := net.Dial("tcp", "localhost:9000")
		if err != nil {
			t.Fatalf("Client %d failed to connect: %v", i, err)
		}
		connections[i] = connection
	}

	message := "test"
	// Send a message from each client connection to the server
	for _, connection := range connections {
		_, err := connection.Write([]byte(message))
		if err != nil {
			t.Fatalf("Failed to write to server: %v", err)
		}
		defer connection.Close()
	}
}

// TestTrafficLimitReached tests if the server responds correctly when the traffic limit is exceeded.
func TestTrafficLimitReached(t *testing.T) {
	go StartServer(":9002")
	time.Sleep(200 * time.Millisecond) // Give server time to start

	client, err := net.Dial("tcp", "localhost:9002") // Connect to the server
	if err != nil {
		t.Fatalf("Client failed to connect: %v", err)
	}
	defer client.Close()

	message := strings.Repeat("x", byteLimit+1) // Create a message that exceeds the byte limit
	_, err = client.Write([]byte(message))
	if err != nil {
		t.Fatalf("Failed to write to server: %v", err)
	}

	buf := make([]byte, 1024)
	receivedMessage, err := client.Read(buf) // Read the server's response
	if err != nil {
		t.Fatalf("Failed to read from server: %v", err)
	}
	// Check if the server responded with the expected traffic limit message
	if !strings.Contains(string(buf[:receivedMessage]), "Traffic limit reached") {
		t.Errorf("Expected traffic limit message, got: %s", string(buf[:receivedMessage]))
	}

}
