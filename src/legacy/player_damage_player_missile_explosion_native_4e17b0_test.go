package legacy

import (
	"fmt"
	"math"
	"runtime"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func TestPlayerDamagePlayerMissileExplosionNativeCallback4E17B0(t *testing.T) {
	_ = npcReflectLegacyServer4E17B0(t)
	oldSetState := Nox_xxx_playerSetState_4FA020
	Nox_xxx_playerSetState_4FA020 = func(*server.Object, server.PlayerState) bool {
		t.Fatal("below-threshold EXPLOSION requested hurt state")
		return false
	}
	t.Cleanup(func() { Nox_xxx_playerSetState_4FA020 = oldSetState })
	for _, splash := range []bool{false, true} {
		t.Run(fmt.Sprintf("splash-%t", splash), func(t *testing.T) {
			var pin runtime.Pinner
			defer pin.Unpin()
			npc, target := npcReflectLegacyObjects4E17B0(t, &pin)
			npc.Buffs = 0
			target.Damage = playerDamageMeleeCallbackNative4E17B0()
			ud := target.UpdateDataPlayer()
			ud.Field57, ud.Field21 = math.Float32bits(0.5), math.Float32bits(0.25)
			missile := &server.Object{TypeInd: 697, ObjClass: object.ClassMissile, ObjOwner: npc, PrevPos: types.Ptf(23, 9)}
			pin.Pin(missile)
			source, weapon := npc, missile
			if splash {
				source, weapon = missile, nil
			}
			wantMarker, wantType := uint32(1), uint32(missile.TypeInd)
			if splash {
				wantMarker, wantType = 2, 7
			}
			if !objectDamageDispatchCallNative(target, source, weapon, 8, object.DamageExplosion) || target.HealthData.Cur != 196 ||
				target.Field38 != math.MaxUint32 || target.Obj130 != missile || target.Pos132 != missile.PrevPos || target.Field131 != 7 || target.Frame134 != 1400 ||
				ud.Field76 != wantMarker || ud.Field75 != wantType || ud.Field21 != math.Float32bits(0.25) {
				t.Fatalf("C->PlayerDamage->DefaultDamage EXPLOSION HP=%d sync=%x source=%p marker=%d/%d carry=%#x", target.HealthData.Cur, target.Field38, target.Obj130, ud.Field76, ud.Field75, ud.Field21)
			}
			t.Logf("C->PlayerDamage->DefaultDamage->UnitSetHP: NPC=%p target=%p missile=%p HP=200->196", npc, target, missile)
			runtime.KeepAlive(target)
			runtime.KeepAlive(npc)
			runtime.KeepAlive(missile)
		})
	}
}
