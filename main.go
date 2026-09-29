package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

const defaultHost = "127.0.0.1"
const defaultPort = 8080

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
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", os.Args[1])
		printUsage()
		os.Exit(2)
	}
	if err != nil {
		log.Printf("error: %v", err)
		os.Exit(1)
	}
}

func runStart(arguments []string) error {
	flags := flag.NewFlagSet("start", flag.ContinueOnError)
	host := flags.String("host", defaultHost, "address to listen on")
	port := flags.Int("port", defaultPort, "TCP port to listen on")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if *port < 1 || *port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	broadcaster := newBroadcastServer()
	server := &http.Server{Addr: *host + ":" + strconv.Itoa(*port), Handler: broadcaster}
	serverErrors := make(chan error, 1)
	go func() { serverErrors <- server.ListenAndServe() }()

	log.Printf("broadcast server listening on ws://%s/ws", server.Addr)
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(shutdown)

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-shutdown:
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = broadcaster.Close()
		return server.Shutdown(ctx)
	}
}

func runConnect(arguments []string) error {
	flags := flag.NewFlagSet("connect", flag.ContinueOnError)
	host := flags.String("host", defaultHost, "server host")
	port := flags.Int("port", defaultPort, "server port")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if *port < 1 || *port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	connection, _, err := websocket.DefaultDialer.Dial("ws://"+*host+":"+strconv.Itoa(*port)+"/ws", nil)
	if err != nil {
		return fmt.Errorf("connect to server: %w", err)
	}
	defer connection.Close()

	readErrors := make(chan error, 1)
	go func() {
		for {
			_, message, readErr := connection.ReadMessage()
			if readErr != nil {
				readErrors <- readErr
				return
			}
			fmt.Println(string(message))
		}
	}()

	log.Println("connected; type messages and press Enter")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if err := connection.WriteMessage(websocket.TextMessage, []byte(scanner.Text())); err != nil {
			return fmt.Errorf("send message: %w", err)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	select {
	case err := <-readErrors:
		return err
	default:
		return nil
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  broadcast-server start   [-host 127.0.0.1] [-port 8080]")
	fmt.Println("  broadcast-server connect [-host 127.0.0.1] [-port 8080]")
}
