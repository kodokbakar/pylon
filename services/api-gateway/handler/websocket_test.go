package handler

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/coder/websocket"
)

type fakeConnection struct {
	mu     sync.Mutex
	closed bool
	code   websocket.StatusCode
	reason string
}

func (c *fakeConnection) Close(code websocket.StatusCode, reason string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.closed = true
	c.code = code
	c.reason = reason

	return nil
}

func (c *fakeConnection) isClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.closed
}

func TestConnectionManagerTryAddRemoveAndCount(t *testing.T) {
	manager := NewConnectionManager(10)

	if manager.Count() != 0 {
		t.Fatalf("expected initial count 0, got %d", manager.Count())
	}

	firstConn := &fakeConnection{}
	secondConn := &fakeConnection{}

	if !manager.TryAdd("user-1", firstConn) {
		t.Fatal("expected first connection to be added")
	}

	if !manager.TryAdd("user-1", secondConn) {
		t.Fatal("expected second connection to be added")
	}

	if manager.Count() != 2 {
		t.Fatalf("expected count 2, got %d", manager.Count())
	}

	manager.Remove("user-1", firstConn)

	if manager.Count() != 1 {
		t.Fatalf("expected count 1 after removing first connection, got %d", manager.Count())
	}

	manager.Remove("user-1", secondConn)

	if manager.Count() != 0 {
		t.Fatalf("expected count 0 after removing all connections, got %d", manager.Count())
	}
}

func TestConnectionManagerTryAddRejectsOverLimit(t *testing.T) {
	manager := NewConnectionManager(1)

	if !manager.TryAdd("user-1", &fakeConnection{}) {
		t.Fatal("expected first connection to be added")
	}

	if manager.TryAdd("user-2", &fakeConnection{}) {
		t.Fatal("expected second connection to be rejected")
	}

	if manager.Count() != 1 {
		t.Fatalf("expected count 1, got %d", manager.Count())
	}
}

func TestConnectionManagerTryAddIsConcurrencySafe(t *testing.T) {
	manager := NewConnectionManager(100)

	var accepted int32
	var wg sync.WaitGroup

	for i := range 250 {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			if manager.TryAdd("user-1", &fakeConnection{}) {
				atomic.AddInt32(&accepted, 1)
			}
		}(i)
	}

	wg.Wait()

	if accepted != 100 {
		t.Fatalf("expected exactly 100 accepted connections, got %d", accepted)
	}

	if manager.Count() != 100 {
		t.Fatalf("expected manager count 100, got %d", manager.Count())
	}
}

func TestConnectionManagerShutdownClosesConnections(t *testing.T) {
	manager := NewConnectionManager(10)

	connections := make([]*fakeConnection, 0, 3)
	for i := range 3 {
		conn := &fakeConnection{}
		connections = append(connections, conn)

		if !manager.TryAdd(fmt.Sprintf("user-%d", i), conn) {
			t.Fatalf("expected connection %d to be added", i)
		}
	}

	manager.Shutdown()

	if manager.Count() != 0 {
		t.Fatalf("expected count 0 after shutdown, got %d", manager.Count())
	}

	for i, conn := range connections {
		if !conn.isClosed() {
			t.Fatalf("expected connection %d to be closed", i)
		}

		if conn.code != websocket.StatusGoingAway {
			t.Fatalf("expected connection %d close code going away, got %d", i, conn.code)
		}
	}
}

func TestNormalizeOriginPatternsTrimsAndAddsHostPatterns(t *testing.T) {
	got := normalizeOriginPatterns([]string{
		" http://localhost:3000 ",
		"",
		" http://localhost:5173 ",
	})

	expected := []string{
		"http://localhost:3000",
		"localhost:3000",
		"http://localhost:5173",
		"localhost:5173",
	}

	if len(got) != len(expected) {
		t.Fatalf("expected %d origin patterns, got %#v", len(expected), got)
	}

	for i, pattern := range expected {
		if got[i] != pattern {
			t.Fatalf("expected origin pattern %d to be %q, got %q", i, pattern, got[i])
		}
	}
}
