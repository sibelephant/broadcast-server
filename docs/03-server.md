# 3. Implement the WebSocket Server

The server has two types:

```go
type client struct {
    conn    *websocket.Conn
    writeMu sync.Mutex
}

type broadcastServer struct {
    mu       sync.RWMutex
    clients  map[*client]struct{}
    closed   bool
    upgrader websocket.Upgrader
}
```

## Why Two Mutexes?

`broadcastServer.mu` protects shared server state:

- Adding a client.
- Removing a client.
- Counting clients.
- Taking a snapshot for a broadcast.
- Closing the server.

`client.writeMu` protects one WebSocket connection from concurrent writes. Gorilla WebSocket permits one concurrent reader and one concurrent writer, but two goroutines must not write to the same connection at once.

Initialize the registry and upgrader:

```go
func newBroadcastServer() *broadcastServer {
    return &broadcastServer{
        clients: make(map[*client]struct{}),
        upgrader: websocket.Upgrader{
            CheckOrigin: func(*http.Request) bool { return true },
        },
    }
}
```

For a production service, replace the permissive origin check with an allowlist appropriate to the application.

## Accept a Connection

Make the server implement `http.Handler` by defining `ServeHTTP`:

```go
func (s *broadcastServer) ServeHTTP(
    writer http.ResponseWriter,
    request *http.Request,
) {
    connection, err := s.upgrader.Upgrade(writer, request, nil)
    if err != nil {
        return
    }

    connectedClient := &client{conn: connection}
    s.mu.Lock()
    if s.closed {
        s.mu.Unlock()
        _ = connection.Close()
        return
    }
    s.clients[connectedClient] = struct{}{}
    s.mu.Unlock()

    defer func() {
        s.removeClient(connectedClient)
        _ = connection.Close()
    }()

    // Read messages here.
}
```

The deferred cleanup is important. It runs when the client disconnects or when reading returns an error.

## Read and Broadcast

Read messages until the connection closes:

```go
for {
    messageType, message, err := connection.ReadMessage()
    if err != nil {
        return
    }
    if messageType != websocket.TextMessage &&
        messageType != websocket.BinaryMessage {
        continue
    }
    s.Broadcast(messageType, message)
}
```

The server preserves the incoming message type, so text remains text and binary remains binary.

## Broadcast Safely

Do not hold the server lock while writing to the network. A slow client could otherwise block registration and removal of every other client. Take a snapshot first:

```go
func (s *broadcastServer) Broadcast(messageType int, message []byte) {
    s.mu.RLock()
    clients := make([]*client, 0, len(s.clients))
    for connectedClient := range s.clients {
        clients = append(clients, connectedClient)
    }
    s.mu.RUnlock()

    for _, connectedClient := range clients {
        connectedClient.writeMu.Lock()
        err := connectedClient.conn.WriteMessage(messageType, message)
        connectedClient.writeMu.Unlock()
        if err != nil {
            s.removeClient(connectedClient)
        }
    }
}
```

A failed write removes the client from the registry. The next chapter adds the CLI commands around this server.

Continue with [the CLI](04-cli.md).
