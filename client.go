package main

import (
	"fmt"
	"io"
	"net"
)

type Client struct {
	conn           net.Conn
	address        string
	currentByteSum int
}

func NewClient(conn net.Conn) *Client {
	return &Client{
		conn:    conn,
		address: conn.RemoteAddr().String(),
	}
}

// HandleConnection manages the client's connection, reading data and broadcasting messages.
// It also checks if the client has reached the byte limit and disconnects if necessary.
func (client *Client) HandleConnection() {
	defer func() {
		// Remove the client from the clients map when done
		clientsMu.Lock()
		delete(clients, client.conn)
		clientsMu.Unlock()
		client.conn.Close()
		fmt.Println("Disconnected:", client.conn.RemoteAddr())
	}()

	fmt.Println("Handling connection from", client.conn.RemoteAddr())
	buffer := make([]byte, 1024)

	for {
		// Read data from the connection
		message, err := client.conn.Read(buffer)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error reading from connection:", err)
			}
			return
		}

		// Print the received data
		fmt.Println("Received data from", client.address+":", string(buffer[:message]))
		BroadcastMessage(buffer[:message], client)

		// Check if the client has reached the byte limit
		// If so, disconnect the client and remove it from the clients map
		client.currentByteSum += message
		if client.currentByteSum >= byteLimit {
			fmt.Println("Client", client.address, "has reached the byte limit. Disconnecting...")
			client.conn.Write([]byte("Traffic limit reached. Disconnecting.\n"))
			return // Cleanup will be done in the deferred function
		}
	}
}
