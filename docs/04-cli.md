# 4. Implement the CLI

The CLI dispatches on the first argument:

```go
func main() {
    if len(os.Args) < 2 {
        printUsage()
        os.Exit(2)
    }

    var err error
    switch os.Args[1] {
    case "start":
        err = runStart(os.Args[2:])
    case "connect":
        err = runConnect(os.Args[2:])
    case "help", "-h", "--help":
        printUsage()
        return
    default:
        printUsage()
        os.Exit(2)
    }
    if err != nil {
        log.Printf("error: %v", err)
        os.Exit(1)
    }
}
```

## Start the Server

Parse host and port flags, then use the broadcaster as the HTTP handler:

```go
broadcaster := newBroadcastServer()
server := &http.Server{
    Addr: *host + ":" + strconv.Itoa(*port),
    Handler: broadcaster,
}
```

Start listening in a goroutine so the main goroutine can wait for either a server error or an operating-system signal:

```go
serverErrors := make(chan error, 1)
go func() {
    serverErrors <- server.ListenAndServe()
}()

select {
case err := <-serverErrors:
    return err
case <-shutdown:
    // Close clients, then shut down HTTP.
}
```

Validate ports before starting:

```go
if *port < 1 || *port > 65535 {
    return fmt.Errorf("port must be between 1 and 65535")
}
```

## Connect as a Client

Dial the same WebSocket endpoint:

```go
connection, _, err := websocket.DefaultDialer.Dial(
    "ws://"+*host+":"+strconv.Itoa(*port)+"/ws",
    nil,
)
if err != nil {
    return fmt.Errorf("connect to server: %w", err)
}
defer connection.Close()
```

The client needs two independent activities:

1. Read messages from the server continuously.
2. Read lines from standard input and send them.

Run the network reader in a goroutine:

```go
go func() {
    for {
        _, message, err := connection.ReadMessage()
        if err != nil {
            readErrors <- err
            return
        }
        fmt.Println(string(message))
    }
}()
```

The main goroutine scans standard input:

```go
scanner := bufio.NewScanner(os.Stdin)
for scanner.Scan() {
    if err := connection.WriteMessage(
        websocket.TextMessage,
        []byte(scanner.Text()),
    ); err != nil {
        return fmt.Errorf("send message: %w", err)
    }
}
```

This means every line entered by the user becomes a text WebSocket message.

Continue with [graceful shutdown and testing](05-shutdown-and-tests.md).
