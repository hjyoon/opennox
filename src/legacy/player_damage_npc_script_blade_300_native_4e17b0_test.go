package legacy

import (
	"runtime"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

func TestPlayerDamageNPCScriptBlade300NativeCallback4E17B0HP(t *testing.T) {
	npcReflectLegacyServer4E17B0(t)
	var pin runtime.Pinner
	defer pin.Unpin()
	npc, _ := npcReflectLegacyObjects4E17B0(t, &pin)
	// Exact class/subclass/flags/type from the reported raw-300 NPC hit.
	// Keep HP nonlethal here to exercise the actual C/Go/native HP chain;
	// the stock Wiz01a observer separately proves the real death tail.
	npc.TypeInd, npc.ObjClass, npc.ObjSubClass, npc.ObjFlags = 1343, object.Class(2097154), 69650, object.Flags(150995460)
	*npc.HealthData = server.HealthData{Cur: 1000, Field2: 1000, Max: 1000}
	ud := npc.UpdateDataMonster()
	for i, want := range []uint16{700, 400, 100} {
		if !objectDamageDispatchCallNative(npc, nil, nil, 300, object.DamageBlade) || npc.HealthData.Cur != want ||
			ud.Field547 != 2 || ud.Field546 != 0 || npc.Obj130 != nil || npc.Field131 != 0 || npc.Frame134 != 1400 ||
			npc.TypeInd != 1343 || uint32(npc.Class()) != 2097154 || uint32(npc.SubClass()) != 69650 || uint32(npc.Flags()) != 150995460 {
			t.Fatalf("reported raw-300 NPC hit=%d HP=%d/%d marker=%d/%d flags=%x", i, npc.HealthData.Cur, want, ud.Field547, ud.Field546, uint32(npc.Flags()))
		}
		t.Logf("C->PlayerDamage reported NPC shape: target=%p type=1343 class=2097154 subclass=69650 flags=150995460 source=nil weapon=nil raw=300 type=BLADE hit=%d HP=%d", npc, i, npc.HealthData.Cur)
	}
	runtime.KeepAlive(npc)
}
