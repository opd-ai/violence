package dht

import (
	"testing"
)

// Unit tests that exercise pure logic without starting a live libp2p host.
// Integration tests requiring a real network are in node_test.go (//go:build integration).

func TestMakeKey(t *testing.T) {
	tests := []struct {
		namespace string
		value     string
		want      string
	}{
		{"server", "myserver", "/violence/server/myserver"},
		{"genre", "fantasy", "/violence/genre/fantasy"},
		{"server", "", "/violence/server/"},
	}
	for _, tt := range tests {
		got := makeKey(tt.namespace, tt.value)
		if got != tt.want {
			t.Errorf("makeKey(%q, %q) = %q, want %q", tt.namespace, tt.value, got, tt.want)
		}
	}
}

func TestUpdateServerList(t *testing.T) {
	t.Run("add new server", func(t *testing.T) {
		out := updateServerList(nil, "sv1", true)
		if len(out) != 1 || out[0] != "sv1" {
			t.Errorf("unexpected list: %v", out)
		}
	})
	t.Run("add duplicate server", func(t *testing.T) {
		out := updateServerList([]string{"sv1"}, "sv1", true)
		if len(out) != 1 {
			t.Errorf("duplicate entry added: %v", out)
		}
	})
	t.Run("remove existing server", func(t *testing.T) {
		out := updateServerList([]string{"sv1", "sv2"}, "sv1", false)
		if len(out) != 1 || out[0] != "sv2" {
			t.Errorf("unexpected list after remove: %v", out)
		}
	})
	t.Run("remove absent server is no-op", func(t *testing.T) {
		in := []string{"sv1", "sv2"}
		out := updateServerList(in, "sv3", false)
		if len(out) != 2 {
			t.Errorf("unexpected list after noop remove: %v", out)
		}
	})
}

func TestViolenceValidator_Validate(t *testing.T) {
	v := ViolenceValidator{}

	if err := v.Validate("/violence/server/x", []byte("data")); err != nil {
		t.Errorf("unexpected error for valid record: %v", err)
	}
	if err := v.Validate("/violence/server/x", nil); err == nil {
		t.Error("expected error for nil value")
	}
	if err := v.Validate("/violence/server/x", []byte{}); err == nil {
		t.Error("expected error for empty value")
	}
}

func TestViolenceValidator_Select(t *testing.T) {
	v := ViolenceValidator{}

	idx, err := v.Select("k", [][]byte{[]byte("a"), []byte("b")})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if idx != 0 {
		t.Errorf("Select returned %d, want 0", idx)
	}

	_, err = v.Select("k", nil)
	if err == nil {
		t.Error("expected error for empty value list")
	}
}

func TestNewValidator(t *testing.T) {
	nv := NewValidator()
	if nv == nil {
		t.Fatal("NewValidator returned nil")
	}
	if _, ok := nv[ViolenceNamespace]; !ok {
		t.Errorf("namespace %q missing from validator", ViolenceNamespace)
	}
	if _, ok := nv["pk"]; !ok {
		t.Error("pk validator missing")
	}
}

func TestConfig_Fields(t *testing.T) {
	cfg := Config{
		ListenAddrs:    []string{"/ip4/127.0.0.1/tcp/0"},
		BootstrapPeers: []string{"peer1"},
		Mode:           "server",
	}
	if len(cfg.ListenAddrs) != 1 {
		t.Error("ListenAddrs not set correctly")
	}
	if cfg.Mode != "server" {
		t.Error("Mode not set correctly")
	}
}

func TestServerRecord_Fields(t *testing.T) {
	rec := ServerRecord{
		Name:       "test-server",
		Address:    "127.0.0.1:8080",
		Genre:      "fantasy",
		MaxPlayers: 16,
	}
	if rec.Name != "test-server" {
		t.Error("Name field not set correctly")
	}
	if rec.MaxPlayers != 16 {
		t.Error("MaxPlayers field not set correctly")
	}
}
