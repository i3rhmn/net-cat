package tcpchat

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

type Client struct {
	conn    net.Conn
	name    string
	writeMu sync.Mutex
}

func (s *Server) handleConnection(conn net.Conn) {
	defer s.releaseConnection()
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	if _, err := fmt.Fprint(conn, welcomeMessage); err != nil {
		return
	}

	client := &Client{conn: conn}

	// Keep asking until the client chooses an available, nonempty name.
	for scanner.Scan() {
		name := strings.TrimSpace(scanner.Text())
		if name == "" {
			if _, err := fmt.Fprint(conn, "[ENTER YOUR NAME]: "); err != nil {
				return
			}
			continue
		}

		client.name = name
		joined, err := s.join(client)
		if err != nil {
			return
		}
		if !joined {
			if _, err := fmt.Fprint(conn, "Username already taken. Please choose another name.\n[ENTER YOUR NAME]: "); err != nil {
				return
			}
			continue
		}

		defer s.leave(client)
		s.chatLoop(client, scanner)
		return
	}
}

func (s *Server) chatLoop(client *Client, scanner *bufio.Scanner) {
	for scanner.Scan() {
		text := scanner.Text()

		if strings.TrimSpace(text) == "" {
			if err := client.send(client.prompt()); err != nil {
				return
			}
			continue
		}

		// A command is not stored as a normal chat message.
		if text == "/name" || strings.HasPrefix(text, "/name ") {
			newName := strings.TrimSpace(strings.TrimPrefix(text, "/name"))
			response := s.rename(client, newName)
			if err := client.send("\n" + response + client.prompt()); err != nil {
				return
			}
			continue
		}

		if err := s.publish(client, text); err != nil {
			return
		}
		if err := client.send(client.prompt()); err != nil {
			return
		}
	}
}

func (c *Client) prompt() string {
	return fmt.Sprintf("[%s][%s]:", time.Now().Format("2006-01-02 15:04:05"), c.name)
}

// send protects writes to this client's connection and bounds slow writes.
func (c *Client) send(message string) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	if err := c.conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
		return err
	}
	_, err := fmt.Fprint(c.conn, message)
	return err
}

func (c *Client) sendWithPrompt(message string) error {
	return c.send("\n" + message + c.prompt())
}
