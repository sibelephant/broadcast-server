# 1. Understand the Problem

We are building a command-line application with two modes:

```text
broadcast-server start
broadcast-server connect
```

The server accepts WebSocket connections. When one client sends a message, the server writes that message to every connected client.

## Requirements

The implementation needs to:

1. Listen for client connections.
2. Keep track of connected clients.
3. Read messages from each client.
4. Broadcast messages to all clients.
5. Remove disconnected clients.
6. Let a command-line client send and receive messages.
7. Shut down cleanly when the process receives an interrupt.
8. Be safe when multiple connections are active at the same time.

## Design

The project uses:

- Go's `net/http` package for the HTTP server and lifecycle.
- `github.com/gorilla/websocket` for the WebSocket protocol.
- A `map[*client]struct{}` as the in-memory client registry.
- `sync.RWMutex` to protect the registry.
- A mutex on each client to prevent concurrent WebSocket writes.

The WebSocket endpoint is `/ws`. The server defaults to `127.0.0.1:8080`, but both commands accept `-host` and `-port` flags.

## Suggested Build Order

Implement the project in this order:

1. Create the Go module and add the WebSocket dependency.
2. Define the client and server types.
3. Upgrade HTTP requests to WebSocket connections.
4. Register and remove clients.
5. Read and broadcast messages.
6. Add the `start` command.
7. Add the `connect` command.
8. Add tests and run the race detector.

Continue with [project setup](02-project-setup.md).
