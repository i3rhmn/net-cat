package tcpchat

import (
	"sync"
	"testing"
)

func TestConnectionLimit(t *testing.T) {
	s := newServer()

	for i := 0; i < maxConnections; i++ {
		if !s.reserveConnection() {
			t.Fatalf("connection %d was rejected", i+1)
		}
	}

	if s.reserveConnection() {
		t.Fatal("accepted a connection beyond the limit")
	}

	s.releaseConnection()
	if !s.reserveConnection() {
		t.Fatal("did not free a connection slot after disconnect")
	}
}

func TestNewServer(t *testing.T) {
	s := newServer()
	if s.clients == nil || len(s.clients) != 0 || len(s.history) != 0 || s.connections != 0 {
		t.Fatal("new server should have no clients, history, or connections")
	}
}

func TestDuplicateNames(t *testing.T) {
	s := newServer()
	ali := &Client{name: "Ali"}
	s.clients[ali] = true

	if !s.nameTaken("ali", nil) {
		t.Fatal("case-insensitive duplicate name was not detected")
	}
	if s.nameTaken("Ahmed", nil) {
		t.Fatal("an available name was incorrectly rejected")
	}
	if s.nameTaken("ALI", ali) {
		t.Fatal("client should be able to keep or change capitalization of own name")
	}
}

func TestRenameWithoutOtherClients(t *testing.T) {
	s := newServer()
	ali := &Client{name: "Ali"}
	s.clients[ali] = true

	response := s.rename(ali, "Ahmed")
	if ali.name != "Ahmed" || response != "Your name is now Ahmed.\n" {
		t.Fatalf("unexpected rename result: name=%q response=%q", ali.name, response)
	}

	if s.rename(ali, "") != "Usage: /name NewName\n" {
		t.Fatal("missing usage message for blank new name")
	}
}

func TestConcurrentConnectionLimit(t *testing.T) {
	s := newServer()
	var wg sync.WaitGroup
	var countMu sync.Mutex
	accepted := 0

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if s.reserveConnection() {
				countMu.Lock()
				accepted++
				countMu.Unlock()
			}
		}()
	}
	wg.Wait()

	if accepted != maxConnections {
		t.Fatalf("accepted %d connections; expected %d", accepted, maxConnections)
	}
}

func TestNameAvailableAfterLeaving(t *testing.T) {
	s := newServer()
	ali := &Client{name: "Ali"}
	s.clients[ali] = true

	s.leave(ali)

	if s.nameTaken("ali", nil) {
		t.Fatal("name should become available after disconnect")
	}
}
