# Traffic Monitor

An application written in Go that supports multiple clients, relays messages, and tracks traffic usage per client connection.

Application characteristics:
- Listens on TCP port "9000".
- Sends messages to all clients (except sender).
- Tracks upload/download byte usage per client and gracefully disconnects it when 100 byte limit is reached.
