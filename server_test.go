package main

import (
	"net"
	"strings"
	"testing"
	"time"
)

// TestBroadcastMessage tests if the server correctly broadcasts messages to all clients except the sender.
func TestBroadcastMessage(t *testing.T) {
	clientsMu.Lock()
	clients = make(map[net.Conn]*Client)
	clientsMu.Unlock()

	c1Conn, s1 := net.Pipe()
	c2Conn, s2 := net.Pipe()

	defer func() {
		c1Conn.Close()
		s1.Close()
		c2Conn.Close()
		s2.Close()
	}()

	client1 := &Client{conn: s1, address: "client1"}
	client2 := &Client{conn: s2, address: "client2"}

	clientsMu.Lock()
	clients[s1] = client1
	clients[s2] = client2
	clientsMu.Unlock()

	message := []byte("test")
	go BroadcastMessage(message, client1)

	buf := make([]byte, 1024)
	receivedMessage, err := c2Conn.Read(buf)
	if err != nil {
		t.Fatalf("client2 did not receive broadcast: %v", err)
	}
	got := string(buf[:receivedMessage])
	want := "Message sent from client1: test"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}

	// client1 should not receive its own message
	c1Conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond)) // Set a short deadline to avoid blocking
	receivedMessage, err = c1Conn.Read(buf)
	if err == nil && receivedMessage > 0 {
		t.Errorf("client1 should not receive its own broadcast, but got: %q", string(buf[:receivedMessage]))
	}
}

// TestBroadcastDisconnectsOnTrafficLimit tests if the server disconnects clients that exceed the traffic limit.
func TestBroadcastDisconnectsOnTrafficLimit(t *testing.T) {
	clientsMu.Lock()
	clients = make(map[net.Conn]*Client)
	clientsMu.Unlock()

	c1Conn, s1 := net.Pipe()
	c2Conn, s2 := net.Pipe()

	defer func() {
		c1Conn.Close()
		s1.Close()
		c2Conn.Close()
		s2.Close()
	}()

	client1 := &Client{conn: s1, address: "client1", currentByteSum: 0}
	client2 := &Client{conn: s2, address: "client2", currentByteSum: byteLimit - 2} // Set traffic limit to exceed

	clientsMu.Lock()
	clients[s1] = client1
	clients[s2] = client2
	clientsMu.Unlock()

	message := []byte("123456")
	go BroadcastMessage(message, client1)

	buf := make([]byte, 1024)
	// Read and discard the first message
	_, err := c2Conn.Read(buf)
	if err != nil {
		t.Fatalf("client2 did not receive first message: %v", err)
	}

	// Read the second message (traffic limit warning)
	receivedMessage, err := c2Conn.Read(buf)
	if err != nil {
		t.Fatalf("client2 did not receive second message: %v", err)
	}

	// Check if the server responded with the expected traffic limit message
	if !strings.Contains(string(buf[:receivedMessage]), "Traffic limit reached") {
		t.Errorf("Expected traffic limit message, got: %s", string(buf[:receivedMessage]))
	}

	// Expect client2 to be removed
	clientsMu.Lock()
	_, exists1 := clients[s1]
	_, exists2 := clients[s2]
	clientsMu.Unlock()
	if !exists1 {
		t.Errorf("client1 should still be connected")
	}
	if exists2 {
		t.Errorf("client2 should have been disconnected after exceeding traffic limit")
	}
}
