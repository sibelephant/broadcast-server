package main

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

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

func newBroadcastServer() *broadcastServer {
	return &broadcastServer{
		clients: make(map[*client]struct{}),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(*http.Request) bool { return true },
		},
	}
}

func (s *broadcastServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
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

	for {
		messageType, message, err := connection.ReadMessage()
		if err != nil {
			return
		}
		if messageType != websocket.TextMessage && messageType != websocket.BinaryMessage {
			continue
		}
		s.Broadcast(messageType, message)
	}
}

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

func (s *broadcastServer) removeClient(connectedClient *client) {
	s.mu.Lock()
	delete(s.clients, connectedClient)
	s.mu.Unlock()
}

func (s *broadcastServer) ClientCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.clients)
}

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
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "server shutting down"),
		)
		_ = connectedClient.conn.Close()
		connectedClient.writeMu.Unlock()
	}
	return nil
}
