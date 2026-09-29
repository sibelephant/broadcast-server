package main

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

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

	if deadline := time.Now().Add(time.Second); !waitForClientCount(broadcaster, 2, deadline) {
		t.Fatalf("connected clients = %d, want 2", broadcaster.ClientCount())
	}
	if err := first.WriteMessage(websocket.TextMessage, []byte("hello")); err != nil {
		t.Fatal(err)
	}
	for _, connection := range []*websocket.Conn{first, second} {
		if err := connection.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
			t.Fatal(err)
		}
		messageType, message, err := connection.ReadMessage()
		if err != nil {
			t.Fatal(err)
		}
		if messageType != websocket.TextMessage || string(message) != "hello" {
			t.Fatalf("received (%d, %q), want text hello", messageType, message)
		}
	}
}

func TestCloseDisconnectsClients(t *testing.T) {
	broadcaster := newBroadcastServer()
	httpServer := httptest.NewServer(broadcaster)
	defer httpServer.Close()

	websocketURL := "ws" + httpServer.URL[len("http"):]
	connection, _, err := websocket.DefaultDialer.Dial(websocketURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !waitForClientCount(broadcaster, 1, time.Now().Add(time.Second)) {
		t.Fatal("client did not connect")
	}
	if err := broadcaster.Close(); err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	if !waitForClientCount(broadcaster, 0, time.Now().Add(time.Second)) {
		t.Fatalf("connected clients = %d, want 0", broadcaster.ClientCount())
	}
}

func waitForClientCount(broadcaster *broadcastServer, expected int, deadline time.Time) bool {
	for time.Now().Before(deadline) {
		if broadcaster.ClientCount() == expected {
			return true
		}
		time.Sleep(time.Millisecond)
	}
	return broadcaster.ClientCount() == expected
}
