package main

import (
	"fmt"
	"net"
	"sync"
)

const byteLimit = 100 // Limit for upload/download in bytes
var (
	clients   = make(map[net.Conn]*Client)
	clientsMu sync.Mutex
)

// StartServer starts the TCP server and listens for incoming connections.
func StartServer(port string) {
	// Start the server and listen for incoming connections on the specified port
	listener, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Server is listening on port", port)

	// Accept incoming connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// Add the new client to the clients map
		client := NewClient(conn)
		clientsMu.Lock()
		clients[conn] = client
		clientsMu.Unlock()

		// Handle the connection in a separate goroutine
		go client.HandleConnection()
	}
}

// BroadcastMessage sends a message to all connected clients except the sender
// and checks if any client has reached the byte limit.
func BroadcastMessage(message []byte, sender *Client) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	// Send the message to all clients except the sender
	for _, client := range clients {
		if client == sender {
			continue
		}

		_, err := client.conn.Write([]byte("Message sent from " + sender.address + ": " + string(message)))
		if err == nil {
			client.currentByteSum += len(message)
		}

		// Check if the client has reached the byte limit
		// If so, disconnect the client and remove it from the clients map
		if client.currentByteSum >= byteLimit {
			fmt.Println("Client", client.address, "has reached the byte limit. Disconnecting...")
			client.conn.Write([]byte("Traffic limit reached. Disconnecting.\n"))
			client.conn.Close()
			delete(clients, client.conn)
		}
	}
}
