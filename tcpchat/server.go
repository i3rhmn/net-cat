package tcpchat

import (
	"fmt"
	"net"
	"sync"
	"time"
)

const maxConnections = 10

type Server struct {
	mu          sync.Mutex
	clients     map[*Client]bool
	history     []string
	connections int
}

func newServer() *Server {
	return &Server{clients: make(map[*Client]bool)}
}

func RunServer(port string) error {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	defer listener.Close()

	fmt.Printf("Listening on the port :%s\n", port)

	server := newServer()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		if !server.reserveConnection() {
			fmt.Fprintln(conn, "Chat is full. Maximum 10 connections allowed.")
			conn.Close()
			continue
		}

		go server.handleConnection(conn)
	}
}

func (s *Server) reserveConnection() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.connections >= maxConnections {
		return false
	}

	s.connections++
	return true
}

func (s *Server) releaseConnection() {
	s.mu.Lock()
	s.connections--
	s.mu.Unlock()
}

// Join and history delivery happen under the same lock as broadcasting.
func (s *Server) join(client *Client) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client.writeMu.Lock()
	defer client.writeMu.Unlock()

	for _, message := range s.history {
		if _, err := fmt.Fprint(client.conn, message); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprint(client.conn, client.prompt()); err != nil {
		return err
	}

	s.clients[client] = true

	for other := range s.clients {
		if other != client {
			other.sendWithPrompt(client.name + " has joined our chat...\n")
		}
	}

	return nil
}

func (s *Server) leave(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.clients[client] {
		return
	}

	delete(s.clients, client)

	for other := range s.clients {
		other.sendWithPrompt(client.name + " has left our chat...\n")
	}
}

func (s *Server) publish(client *Client, text string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.clients[client] {
		return
	}

	message := fmt.Sprintf(
		"[%s][%s]:%s\n",
		time.Now().Format("2006-01-02 15:04:05"),
		client.name,
		text,
	)

	s.history = append(s.history, message)

	for other := range s.clients {
    if other == client {
        other.send("\n" + message)
    } else {

        other.sendWithPrompt(message)
    }
}
}
