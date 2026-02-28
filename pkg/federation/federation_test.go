package federation

import (
	"testing"
)

func TestNewFederation(t *testing.T) {
	fed := NewFederation()
	if fed == nil {
		t.Fatal("NewFederation returned nil")
	}
	if fed.servers == nil {
		t.Fatal("servers map not initialized")
	}
}

func TestFederation_Register(t *testing.T) {
	fed := NewFederation()
	fed.Register("server1", "localhost:8000")

	addr, ok := fed.Lookup("server1")
	if !ok {
		t.Fatal("server not found after registration")
	}
	if addr != "localhost:8000" {
		t.Fatalf("wrong address: got %q, want %q", addr, "localhost:8000")
	}
}

func TestFederation_RegisterWithInfo(t *testing.T) {
	fed := NewFederation()
	info := &ServerInfo{
		Name:       "server2",
		Address:    "localhost:8001",
		Players:    5,
		MaxPlayers: 16,
		Genre:      "scifi",
	}
	fed.RegisterWithInfo(info)

	addr, ok := fed.Lookup("server2")
	if !ok {
		t.Fatal("server not found after RegisterWithInfo")
	}
	if addr != "localhost:8001" {
		t.Fatalf("wrong address: got %q, want %q", addr, "localhost:8001")
	}
}

func TestFederation_LookupNotFound(t *testing.T) {
	fed := NewFederation()
	_, ok := fed.Lookup("nonexistent")
	if ok {
		t.Fatal("expected lookup to fail for nonexistent server")
	}
}

func TestFederation_Match(t *testing.T) {
	fed := NewFederation()

	// No servers - should fail
	_, err := fed.Match()
	if err == nil {
		t.Fatal("expected error when no servers available")
	}

	// Add available server
	fed.RegisterWithInfo(&ServerInfo{
		Name:       "server1",
		Address:    "localhost:8000",
		Players:    5,
		MaxPlayers: 16,
	})

	addr, err := fed.Match()
	if err != nil {
		t.Fatalf("match failed: %v", err)
	}
	if addr != "localhost:8000" {
		t.Fatalf("wrong server matched: got %q", addr)
	}
}

func TestFederation_MatchFullServers(t *testing.T) {
	fed := NewFederation()

	// Add full server
	fed.RegisterWithInfo(&ServerInfo{
		Name:       "full",
		Address:    "localhost:8000",
		Players:    16,
		MaxPlayers: 16,
	})

	_, err := fed.Match()
	if err == nil {
		t.Fatal("expected error when all servers full")
	}
}

func TestFederation_MatchGenre(t *testing.T) {
	fed := NewFederation()

	fed.RegisterWithInfo(&ServerInfo{
		Name:       "fantasy1",
		Address:    "localhost:8000",
		Players:    0,
		MaxPlayers: 16,
		Genre:      "fantasy",
	})
	fed.RegisterWithInfo(&ServerInfo{
		Name:       "scifi1",
		Address:    "localhost:8001",
		Players:    0,
		MaxPlayers: 16,
		Genre:      "scifi",
	})

	// Match fantasy
	addr, err := fed.MatchGenre("fantasy")
	if err != nil {
		t.Fatalf("MatchGenre failed: %v", err)
	}
	if addr != "localhost:8000" {
		t.Fatalf("wrong server: got %q, want localhost:8000", addr)
	}

	// Match scifi
	addr, err = fed.MatchGenre("scifi")
	if err != nil {
		t.Fatalf("MatchGenre failed: %v", err)
	}
	if addr != "localhost:8001" {
		t.Fatalf("wrong server: got %q, want localhost:8001", addr)
	}
}

func TestFederation_MatchGenreNotFound(t *testing.T) {
	fed := NewFederation()

	fed.RegisterWithInfo(&ServerInfo{
		Name:       "fantasy1",
		Address:    "localhost:8000",
		Players:    0,
		MaxPlayers: 16,
		Genre:      "fantasy",
	})

	_, err := fed.MatchGenre("horror")
	if err == nil {
		t.Fatal("expected error for unavailable genre")
	}
}

func TestFederation_List(t *testing.T) {
	fed := NewFederation()

	// Empty list
	list := fed.List()
	if len(list) != 0 {
		t.Fatalf("expected empty list, got %d servers", len(list))
	}

	// Add servers
	fed.Register("s1", "localhost:8000")
	fed.Register("s2", "localhost:8001")

	list = fed.List()
	if len(list) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(list))
	}
}

func TestFederation_Unregister(t *testing.T) {
	fed := NewFederation()
	fed.Register("server1", "localhost:8000")

	// Verify it exists
	_, ok := fed.Lookup("server1")
	if !ok {
		t.Fatal("server not found before unregister")
	}

	// Unregister
	fed.Unregister("server1")

	// Verify it's gone
	_, ok = fed.Lookup("server1")
	if ok {
		t.Fatal("server still exists after unregister")
	}
}

func TestFederation_UnregisterNonexistent(t *testing.T) {
	fed := NewFederation()
	// Should not panic
	fed.Unregister("nonexistent")
}

func TestSetGenre(t *testing.T) {
	// Should not panic
	SetGenre("fantasy")
	SetGenre("scifi")
	SetGenre("")
}
