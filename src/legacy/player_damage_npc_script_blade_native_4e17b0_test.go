package legacy

import (
	"math"
	"runtime"
	"testing"

	"github.com/opennox/libs/object"
)

func TestPlayerDamageNPCScriptBladeNativeCallback4E17B0HP(t *testing.T) {
	npcReflectLegacyServer4E17B0(t)
	var pin runtime.Pinner
	defer pin.Unpin()
	npc, _ := npcReflectLegacyObjects4E17B0(t, &pin)
	ud := npc.UpdateDataMonster()
	ud.Field518, ud.Field1 = math.Float32bits(.5), math.Float32bits(.4)
	for i, want := range []uint16{196, 191, 187} {
		// Real C dispatcher -> registered PlayerDamage -> DefaultDamage ->
		// UnitDamageClear/UnitSetHP. No replacement damage or HP callback.
		if !objectDamageDispatchCallNative(npc, nil, nil, 9, object.DamageBlade) || npc.HealthData.Cur != want || ud.Field547 != 2 || ud.Field546 != 0 || npc.Obj130 != nil || npc.Field131 != 0 || npc.Frame134 != 1400 {
			t.Fatalf("native nil-source hit=%d HP=%d/%d marker=%d/%d", i, npc.HealthData.Cur, want, ud.Field547, ud.Field546)
		}
		t.Logf("C->PlayerDamage scripted BLADE: target=%p source=nil weapon=nil raw=9 hit=%d HP=%d carry=%g", npc, i, npc.HealthData.Cur, math.Float32frombits(ud.Field1))
	}
	runtime.KeepAlive(npc)
}
