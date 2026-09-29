# 2. Set Up the Go Project

Start with an empty directory and initialize a Go module:

```bash
mkdir broadcast-server
cd broadcast-server
go mod init github.com/example/broadcast-server
```

The module path can be changed to match your repository.

Add the WebSocket package:

```bash
go get github.com/gorilla/websocket
go mod tidy
```

Create these files:

```text
broadcast-server/
├── go.mod
├── main.go
├── server.go
└── server_test.go
```

The package is `main` because the result is an executable CLI:

```go
package main
```

## Why Use a WebSocket Package?

WebSockets begin as an HTTP request and then upgrade to a long-lived, two-way connection. The Gorilla package handles the protocol details, including framing, control messages, and close messages, so the application can focus on client management.

## First Check

Before implementing behavior, make sure the module builds:

```bash
go test ./...
```

At this stage there may be no tests yet. The important check is that the module and dependency setup are valid.

Continue with [the server data model](03-server.md).
