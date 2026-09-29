# 5. Graceful Shutdown and Tests

## Close Connected Clients

The server should stop accepting new clients and close existing connections:

```go
func (s *broadcastServer) Close() error {
    s.mu.Lock()
    if s.closed {
        s.mu.Unlock()
        return nil
    }
    s.closed = true

    clients := make([]*client, 0, len(s.clients))
    for connectedClient := range s.clients {
        clients = append(clients, connectedClient)
    }
    s.clients = make(map[*client]struct{})
    s.mu.Unlock()

    for _, connectedClient := range clients {
        connectedClient.writeMu.Lock()
        _ = connectedClient.conn.WriteMessage(
            websocket.CloseMessage,
            websocket.FormatCloseMessage(
                websocket.CloseNormalClosure,
                "server shutting down",
            ),
        )
        _ = connectedClient.conn.Close()
        connectedClient.writeMu.Unlock()
    }
    return nil
}
```

In `runStart`, handle Ctrl+C and SIGTERM:

```go
shutdown := make(chan os.Signal, 1)
signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
defer signal.Stop(shutdown)

case <-shutdown:
    ctx, cancel := context.WithTimeout(
        context.Background(),
        5*time.Second,
    )
    defer cancel()
    _ = broadcaster.Close()
    return server.Shutdown(ctx)
```

`broadcaster.Close()` closes WebSocket clients. `server.Shutdown(ctx)` stops accepting HTTP requests and waits for handlers to finish, subject to the timeout.

## Test Message Fan-Out

Use `httptest.NewServer` so the test does not need a real TCP port:

```go
func TestBroadcastsMessageToAllConnectedClients(t *testing.T) {
    broadcaster := newBroadcastServer()
    httpServer := httptest.NewServer(broadcaster)
    defer httpServer.Close()
    defer broadcaster.Close()

    websocketURL := "ws" + httpServer.URL[len("http"):]
    first, _, err := websocket.DefaultDialer.Dial(websocketURL, nil)
    if err != nil {
        t.Fatal(err)
    }
    defer first.Close()

    second, _, err := websocket.DefaultDialer.Dial(websocketURL, nil)
    if err != nil {
        t.Fatal(err)
    }
    defer second.Close()

    if !waitForClientCount(
        broadcaster,
        2,
        time.Now().Add(time.Second),
    ) {
        t.Fatalf("connected clients = %d, want 2", broadcaster.ClientCount())
    }

    if err := first.WriteMessage(
        websocket.TextMessage,
        []byte("hello"),
    ); err != nil {
        t.Fatal(err)
    }

    for _, connection := range []*websocket.Conn{first, second} {
        _, message, err := connection.ReadMessage()
        if err != nil {
            t.Fatal(err)
        }
        if string(message) != "hello" {
            t.Fatalf("message = %q, want hello", message)
        }
    }
}
```

This test verifies the central requirement: two connected clients receive the same message.

## Test Shutdown

The second test connects a client, calls `Close`, and checks that the registry becomes empty. This verifies that shutdown removes active clients instead of leaving stale entries.

Run the tests:

```bash
go test ./...
```

Then run the race detector:

```bash
go test -race ./...
```

The race detector is especially useful here because clients connect, disconnect, read, and write concurrently.

Continue with [running and extending the project](06-run-and-extend.md).
