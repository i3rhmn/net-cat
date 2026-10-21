package tcpchat

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

const maxConnections = 10
const writeTimeout = 3 * time.Second

type Server struct {
	mu          sync.Mutex
	clients     map[*Client]bool
	history     []string
	connections int
}

func newServer() *Server {
	return &Server{
		clients: make(map[*Client]bool),
	}
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
			fmt.Fprintf(conn, "Chat is full. Maximum %d connections allowed.\n", maxConnections)
			conn.Close()
			continue
		}

		go server.handleConnection(conn)
	}
}

// reserveConnection counts every open TCP connection, including users naming themselves.
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
	defer s.mu.Unlock()
	if s.connections > 0 {
		s.connections--
	}
}

// nameTaken checks registered clients; caller must hold s.mu.
func (s *Server) nameTaken(name string, except *Client) bool {
	for other := range s.clients {
		if other != except && strings.EqualFold(other.name, name) {
			return true
		}
	}
	return false
}

// join checks the name, replays history, registers the client and announces the join.
// Holding s.mu makes these steps consistent with concurrent messages and joins.
func (s *Server) join(client *Client) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.nameTaken(client.name, nil) {
		return false, nil
	}

	// Send history before registering so the client cannot receive it twice.
	client.writeMu.Lock()
	if err := client.conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
		client.writeMu.Unlock()
		return false, err
	}
	for _, message := range s.history {
		if _, err := fmt.Fprint(client.conn, message); err != nil {
			client.writeMu.Unlock()
			return false, err
		}
	}
	if _, err := fmt.Fprint(client.conn, client.prompt()); err != nil {
		client.writeMu.Unlock()
		return false, err
	}
	client.writeMu.Unlock()

	s.clients[client] = true
	s.notifyOthers(client, client.name+" has joined our chat...\n")
	return true, nil
}

// notifyOthers must be called with s.mu held.
func (s *Server) notifyOthers(sender *Client, message string) {
	for other := range s.clients {
		if other != sender {
			if err := other.sendWithPrompt(message); err != nil {
				other.conn.Close() // Unblock the disconnected client's reader.
			}
		}
	}
}

func (s *Server) leave(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.clients[client] {
		return
	}
	delete(s.clients, client)
	s.notifyOthers(client, client.name+" has left our chat...\n")
}

// publish saves a single formatted message and delivers it to all clients.
func (s *Server) publish(client *Client, text string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.clients[client] {
		return nil
	}

	message := fmt.Sprintf("[%s][%s]:%s\n", time.Now().Format("2006-01-02 15:04:05"), client.name, text)
	s.history = append(s.history, message)

	var senderError error
	for other := range s.clients {
		var err error
		if other == client {
			// Give the sender the same timestamped message as all other clients.
			err = other.send("\n" + message)
			senderError = err
		} else {
			err = other.sendWithPrompt(message)
		}
		if err != nil {
			other.conn.Close()
		}
	}
	return senderError
}

// rename changes a name only if no other active client uses it.
// It returns a message for the requester; the broadcast goes to other clients.
func (s *Server) rename(client *Client, newName string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if newName == "" {
		return "Usage: /name NewName\n"
	}
	if s.nameTaken(newName, client) {
		return "Username already taken. Please choose another name.\n"
	}
	if client.name == newName {
		return "You are already using that name.\n"
	}

	oldName := client.name
	client.name = newName
	s.notifyOthers(client, oldName+" changed their name to "+newName+"...\n")
	return "Your name is now " + newName + ".\n"
}
