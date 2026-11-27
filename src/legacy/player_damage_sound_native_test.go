package legacy

import (
	"testing"

	"github.com/opennox/opennox/v1/server"
)

func TestPlayerDamageSoundRuntimeKeepsGoOwnedObjects(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	item := &server.Object{HealthData: &server.HealthData{Cur: 25, Max: 25}}
	target := &server.Object{InvFirstItem: item}
	source := &server.Object{HealthData: &server.HealthData{Cur: 70, Max: 70}}
	old := Nox_xxx_soundPlayerDamageSound_5328B0
	t.Cleanup(func() { Nox_xxx_soundPlayerDamageSound_5328B0 = old })

	for _, path := range []struct {
		name string
		play func(*server.Object, *server.Object)
	}{
		{"DefaultDamage", defaultDamageWorldRuntime4E0B30(srv).PlayerDamageSound},
		{"PlayerDamage", playerDamageRuntime4E17B0(srv).PlayerDamageSound},
	} {
		t.Run(path.name, func(t *testing.T) {
			calls := 0
			Nox_xxx_soundPlayerDamageSound_5328B0 = func(receiver, attacker *server.Object) int {
				calls++
				if receiver != target || receiver.InvFirstItem != item || attacker != source {
					t.Fatalf("sound objects changed: receiver=%p attacker=%p", receiver, attacker)
				}
				return 1
			}
			path.play(target, source)
			if calls != 1 {
				t.Fatalf("sound callback ran %d times, want 1", calls)
			}
		})
	}
}
