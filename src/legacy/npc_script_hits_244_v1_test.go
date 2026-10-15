package legacy

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

func TestNPCScriptHits244V1NativeCallbackHP(t *testing.T) {
	for _, tc := range []struct {
		typ  object.DamageType
		raw  int32
		want [3]uint16
	}{{object.DamageCrush, 200, [3]uint16{800, 600, 400}}, {object.DamageBite, 50, [3]uint16{950, 900, 850}}} {
		t.Run(fmt.Sprint(tc.typ), func(t *testing.T) {
			npcReflectLegacyServer4E17B0(t)
			var pin runtime.Pinner
			defer pin.Unpin()
			npc, _ := npcReflectLegacyObjects4E17B0(t, &pin)
			npc.TypeInd, npc.ObjClass, npc.ObjSubClass, npc.ObjFlags = 1343, 2097154, 69650, 16777732
			*npc.HealthData = server.HealthData{Cur: 1000, Field2: 1000, Max: 1000}
			ud := npc.UpdateDataMonster()
			for i, want := range tc.want {
				// Actual C dispatcher -> registered PlayerDamage -> DefaultDamage
				// -> UnitDamageClear/UnitSetHP. No replacement HP callback.
				if !objectDamageDispatchCallNative(npc, nil, nil, tc.raw, tc.typ) || npc.HealthData.Cur != want ||
					ud.Field547 != 2 || ud.Field546 != uint32(tc.typ) || npc.Obj130 != nil || npc.Field131 != uint32(tc.typ) || npc.Frame134 != 1400 ||
					npc.TypeInd != 1343 || uint32(npc.Class()) != 2097154 || uint32(npc.SubClass()) != 69650 || uint32(npc.Flags()) != 16777732 {
					t.Fatalf("nil NPC raw=%d type=%d hit=%d HP=%d/%d marker=%d/%d", tc.raw, tc.typ, i, npc.HealthData.Cur, want, ud.Field547, ud.Field546)
				}
				t.Logf("R244 C->PlayerDamage->DefaultDamage->UnitSetHP: target=%p source=nil weapon=nil raw=%d type=%d hit=%d HP=%d", npc, tc.raw, tc.typ, i, npc.HealthData.Cur)
			}
			runtime.KeepAlive(npc)
		})
	}
}
