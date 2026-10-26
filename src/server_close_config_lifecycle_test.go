package opennox

import (
	"testing"

	"github.com/opennox/libs/strman"
)

func TestServerCloseReleasesConfigurationBindings(t *testing.T) {
	before := append([]func(){}, onConfigRead...)
	t.Cleanup(func() { onConfigRead = before })
	t.Setenv("NOX_NET_XOR", "true")
	t.Setenv("NOX_NET_NAT", "false")
	t.Setenv("NOX_DEBUG_NET", "true")
	closed := NewServer(nil, nil, strman.New())
	live := NewServer(nil, nil, strman.New())
	t.Cleanup(closed.Close)
	t.Cleanup(live.Close)
	closed.Close()
	newer := NewServer(nil, nil, strman.New())
	t.Cleanup(newer.Close)
	closed.Close() // A reused subscription slot still belongs to the new server.
	for _, s := range []*Server{closed, live, newer} {
		s.NetStr.Xor, s.UseNAT, s.NetStr.Debug = false, true, false
	}
	bindings := 0
	for i, fn := range onConfigRead {
		if fn == nil || i < len(before) && before[i] != nil {
			continue
		}
		bindings++
		fn()
	}
	if bindings != 6 {
		t.Errorf("live server bindings=%d, want 6 for two live servers", bindings)
	}
	if closed.NetStr.Xor || !closed.UseNAT || closed.NetStr.Debug {
		t.Error("configuration reload still updates the closed server")
	}
	for _, s := range []*Server{live, newer} {
		if !s.NetStr.Xor || s.UseNAT || !s.NetStr.Debug {
			t.Error("configuration reload failed to update a live server")
		}
	}
	for _, s := range []*Server{live, newer} {
		s.Close()
	}
	for i, fn := range onConfigRead {
		if fn != nil && (i >= len(before) || before[i] == nil) {
			t.Error("closed server retains a configuration binding")
		}
	}
}
