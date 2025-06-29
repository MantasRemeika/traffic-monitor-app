package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

var (
	clients   = make(map[net.Conn]bool)
	clientsMu sync.Mutex
)


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
		clientsMu.Lock()
		clients[conn] = true
		clientsMu.Unlock()

		// Handle the connection in a separate goroutine
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer func() {
		// Remove the client from the clients map when done
		clientsMu.Lock()
		delete(clients, conn)
		clientsMu.Unlock()

		conn.Close()
		fmt.Println("Disconnected:", conn.RemoteAddr())
	}()
	// Handle the connection
	fmt.Println("Handling connection from", conn.RemoteAddr())
	buffer := make([]byte, 1024)
	
	for {
		// Read data from the connection
		message, err := conn.Read(buffer)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error reading from connection:", err)
			}
			return
		}

		// Print the received data
		fmt.Println("Received data:", string(buffer[:message]))

		// Broadcast the message to all other clients
		broadcastMessage(buffer[:message], conn)
	}
	
}

func broadcastMessage(message []byte, conn net.Conn) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	// Send the message to all clients except the sender
	for client := range clients {
		if client != conn {
			_, err := client.Write(message)
				if err != nil {
					fmt.Println("Error writing to client:", err)
				}
			}
	}
}