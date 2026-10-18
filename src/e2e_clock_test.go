package opennox

import (
	"testing"
	"time"

	"github.com/opennox/libs/platform"
)

func TestE2EGameClockSelection(t *testing.T) {
	previous := platform.Get()
	t.Cleanup(func() { platform.Set(previous) })
	for _, tc := range []struct {
		name     string
		mode     string
		ordinary bool
		invalid  bool
	}{
		{name: "default"},
		{name: "virtual", mode: "virtual"},
		{name: "ordinary", mode: "ordinary", ordinary: true},
		{name: "invalid", mode: "real", invalid: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ordinary := platform.New()
			platform.Set(ordinary)
			p := &platformE2E{}
			if err := setE2EGameClock(p, tc.mode); (err != nil) != tc.invalid {
				t.Fatalf("clock %q error = %v, invalid = %t", tc.mode, err, tc.invalid)
			}
			if tc.ordinary || tc.invalid {
				if platform.Get() != ordinary {
					t.Fatal("production clock and random source were replaced")
				}
				before := platform.Ticks()
				p.tick(24 * time.Hour)
				if delta := platform.Ticks() - before; delta < 0 || delta >= time.Hour {
					t.Fatalf("scenario clock changed game time by %v", delta)
				}
				return
			}
			if platform.Get() != p {
				t.Fatal("default/virtual mode did not install the E2E platform")
			}
			p.tick(2 * time.Second)
			if got := platform.Ticks(); got != 2*time.Second {
				t.Fatalf("virtual game ticks = %v", got)
			}
		})
	}
}
