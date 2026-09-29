# Broadcast Server

A small Go WebSocket broadcast server and CLI client. Every message sent by one connected client is delivered to all connected clients.

## Run It

```bash
go mod tidy
go run . start
```

In another terminal, connect one or more clients:

```bash
go run . connect
```

Type a message and press Enter. Each connected client receives it, including the sender.

The host and port can be changed with flags:

```bash
go run . start -host 127.0.0.1 -port 18080
go run . connect -host 127.0.0.1 -port 18080
```

Run the tests with:

```bash
go test ./...
go test -race ./...
```

## Tutorial

Follow [the step-by-step tutorial](docs/01-overview.md) to reimplement the project from an empty Go module.
