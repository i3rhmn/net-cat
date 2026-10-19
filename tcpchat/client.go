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

	for scanner.Scan() {
		name := strings.TrimSpace(scanner.Text())

		if name == "" {
			if _, err := fmt.Fprint(conn, "[ENTER YOUR NAME]: "); err != nil {
				return
			}
			continue
		}

		client.name = name

		if err := s.join(client); err != nil {
			return
		}

		defer s.leave(client)

		for scanner.Scan() {
			text := scanner.Text()

			if strings.TrimSpace(text) != "" {
				s.publish(client, text)
			}

			if err := client.send(client.prompt()); err != nil {
				return
			}
		}

		return
	}
}

func (c *Client) prompt() string {
	return fmt.Sprintf(
		"[%s][%s]:",
		time.Now().Format("2006-01-02 15:04:05"),
		c.name,
	)
}

func (c *Client) send(message string) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	_, err := fmt.Fprint(c.conn, message)
	return err
}

// Caller holds Server.mu. This method only locks the recipient's writer.
func (c *Client) sendWithPrompt(message string) {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	fmt.Fprint(c.conn, "\n", message, c.prompt())
}
