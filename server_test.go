package main

import (
	"bytes"
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

// TestBroadcastMessage tests if broadcastMessage sends data to all clients except the sender.
func TestBroadcastMessage(t *testing.T) {
	// Setup dummy connections using net.Pipe
	sender, senderRemote := net.Pipe()
	receiver1, receiver1Remote := net.Pipe()
	receiver2, receiver2Remote := net.Pipe()

	// Add connections to clients map
	clientsMu.Lock()
	clients[senderRemote] = true
	clients[receiver1Remote] = true
	clients[receiver2Remote] = true
	clientsMu.Unlock()

	defer func() {
		sender.Close()
		senderRemote.Close()
		receiver1.Close()
		receiver1Remote.Close()
		receiver2.Close()
		receiver2Remote.Close()
		clientsMu.Lock()
		delete(clients, senderRemote)
		delete(clients, receiver1Remote)
		delete(clients, receiver2Remote)
		clientsMu.Unlock()
	}()

	message := []byte("Hello world")
	go broadcastMessage(message, senderRemote)

	// Sender should not receive the message
	sender.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	buf := make([]byte, 64)
	n, err := sender.Read(buf)
	if err == nil && n > 0 {
		t.Errorf("Sender should not receive its own message")
	}

	// Check receiver1 got the message
	receiver1.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	buffer1 := make([]byte, len(message))
	_, err = receiver1.Read(buffer1)
	if err != nil {
		t.Fatalf("receiver1 failed to read: %v", err)
	}
	if !bytes.Equal(buffer1, message) {
		t.Errorf("receiver1 got wrong message: %s", string(buffer1))
	}

	// Check receiver2 got the message
	receiver2.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	buffer2 := make([]byte, len(message))
	_, err = receiver2.Read(buffer2)
	if err != nil {
		t.Fatalf("receiver2 failed to read: %v", err)
	}
	if !bytes.Equal(buffer2, message) {
		t.Errorf("receiver2 got wrong message: %s", string(buffer2))
	}
}
