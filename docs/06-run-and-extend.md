# 6. Run and Extend the Project

## Build the CLI

Build an executable:

```bash
go build -o broadcast-server .
```

Start the server:

```bash
./broadcast-server start
```

The default endpoint is:

```text
ws://127.0.0.1:8080/ws
```

Open two more terminals and connect clients:

```bash
./broadcast-server connect
```

Type a message in either client. Both clients should print it.

To use another port:

```bash
./broadcast-server start -port 18080
./broadcast-server connect -port 18080
```

Press Ctrl+C in the server terminal to close connected clients and stop the server.

## Verify the CLI with a Pipe

A client needs to remain alive long enough to receive a message. For a quick smoke test, keep standard input open briefly:

```bash
{
  printf 'hello from the CLI\n'
  tail -f /dev/null
} | timeout 2s ./broadcast-server connect
```

The output should contain:

```text
hello from the CLI
```

## Possible Extensions

The current implementation intentionally keeps messages in memory and has no authentication. Natural next steps include:

- Add usernames and include the sender in each message.
- Add authentication during the WebSocket handshake.
- Keep a bounded message history for newly connected clients.
- Add a `/healthz` HTTP endpoint.
- Replace the permissive origin check with an allowlist.
- Add server options for maximum message size and write deadlines.
- Add integration tests that start the compiled binary.

When adding features, preserve the concurrency rules: protect shared state with the server mutex and serialize writes to each WebSocket connection.

## Full Verification

From the project root:

```bash
go mod tidy
go test ./...
go test -race ./...
go vet ./...
```

You now have a small real-time system with a clear path for adding chat, live dashboards, or other event-driven behavior.
