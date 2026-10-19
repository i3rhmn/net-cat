package tcpchat

import "testing"

func TestConnectionLimit(t *testing.T) {
	s := newServer()

	for i := 0; i < maxConnections; i++ {
		if !s.reserveConnection() {
			t.Fatalf("connection %d rejected", i+1)
		}
	}

	if s.reserveConnection() {
		t.Fatal("accepted 11th connection")
	}

	s.releaseConnection()

	if !s.reserveConnection() {
		t.Fatal("slot not released")
	}
}

func TestNewServer(t *testing.T) {
	s := newServer()

	if s.clients == nil || len(s.history) != 0 || s.connections != 0 {
		t.Fatal("invalid initial state")
	}
}