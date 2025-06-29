package main

import (
	"fmt"
	"io"
	"net"
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

		// Handle the connection in a separate goroutine
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
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
	}
}