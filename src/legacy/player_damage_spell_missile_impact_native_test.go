package legacy

import (
	"math"
	"testing"

	"github.com/opennox/libs/object"
)

// Actual C->PlayerDamage->DefaultDamage->UnitSetHP, with C-owned native-width
// records and no replaced damage/HP services. Entry absorption/carry are
// controlled scalar fixture values, not a claim about stock armor equipment.
func TestPlayerDamageSpellMissileImpactCRecords4E17B0(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	if !srv.Objs.Init(16) {
		t.Fatal("native object allocator")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	for _, owner := range []string{"self", "player", "NPC"} {
		t.Run(owner, func(t *testing.T) {
			v, a, w := spellMissileImpactCRecords4E17B0(t, srv, owner)
			v.Damage = playerDamageMeleeCallbackNative4E17B0()
			ud := v.UpdateDataPlayer()
			ud.Field57, ud.Field21 = math.Float32bits(.25), math.Float32bits(.25)
			ud.Field76, ud.Field75 = 88, 77
			beforeMissile := *w
			for hit, hp := range []uint16{54, 48, 42} {
				if !objectDamageDispatchCallNative(v, a, w, 8, object.DamageImpact) || v.HealthData.Cur != hp ||
					v.Obj130 != w || v.Field131 != 11 || v.Frame134 != 1400 || v.Pos132 != w.PrevPos ||
					v.Field38 != math.MaxUint32 || ud.Field21 != math.Float32bits(.25) || *w != beforeMissile {
					t.Fatalf("C entry hit=%d HP=%d want=%d source=%p weapon=%p", hit, v.HealthData.Cur, hp, a, w)
				}
				if a == w {
					if ud.Field76 != 2 || ud.Field75 != 11 {
						t.Fatal("self missile marker")
					}
				} else if ud.Field76 != 1 || ud.Field75 != uint32(w.TypeInd) {
					t.Fatal("distinct missile marker")
				}
			}
			t.Logf("C-owned >4 GiB C->PlayerDamage->DefaultDamage: player=%p terminal-parent=%p Pixie=%p owner=%s raw IMPACT=8 effective=6 HP=60->54->48->42", v, a, w, owner)
		})
	}
}
