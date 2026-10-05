package legacy

import (
	"math"
	"runtime"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func TestPlayerDamagePlayerWeaponlessExplosionNativeCallback4E17B0(t *testing.T) {
	_ = npcReflectLegacyServer4E17B0(t)
	oldState := Nox_xxx_playerSetState_4FA020
	Nox_xxx_playerSetState_4FA020 = func(*server.Object, server.PlayerState) bool {
		t.Fatal("small weapon-less EXPLOSION requested hurt state")
		return false
	}
	t.Cleanup(func() { Nox_xxx_playerSetState_4FA020 = oldState })
	for _, kind := range []string{"none", "npc", "self"} {
		t.Run(kind, func(t *testing.T) {
			var pin runtime.Pinner
			defer pin.Unpin()
			npc, target := npcReflectLegacyObjects4E17B0(t, &pin)
			npc.Buffs = 0
			target.Damage = playerDamageMeleeCallbackNative4E17B0()
			ud := target.UpdateDataPlayer()
			ud.Field76, ud.Field75, ud.Field21, ud.Field57 = 99, 77, math.Float32bits(0.25), math.Float32bits(0.5)
			var source *server.Object
			if kind == "npc" {
				source = npc
			}
			if kind == "self" {
				source = target
			}
			wantPos := types.Pointf{}
			if source != nil {
				source.PrevPos = types.Ptf(23, 9)
				wantPos = source.PrevPos
			}
			for hit, wantHP := range []uint16{195, 191, 186} {
				wantCarry := float32(-0.25)
				if hit == 1 {
					wantCarry = 0.25
				}
				if !objectDamageDispatchCallNative(target, source, nil, 9, object.DamageExplosion) || target.HealthData.Cur != wantHP ||
					target.Field38 != math.MaxUint32 || target.Obj130 != source || target.Pos132 != wantPos || target.Field131 != 7 || target.Frame134 != 1400 ||
					ud.Field76 != 2 || ud.Field75 != 7 || ud.Field21 != math.Float32bits(wantCarry) || ud.Field57 != math.Float32bits(0.5) {
					t.Fatalf("C->PlayerDamage weapon-less EXPLOSION hit=%d HP=%d sync=%x source=%p marker=%d/%d carry=%#x",
						hit, target.HealthData.Cur, target.Field38, target.Obj130, ud.Field76, ud.Field75, ud.Field21)
				}
			}
			t.Logf("C->PlayerDamage->DefaultDamage->UnitSetHP: target=%p source=%p HP=200->195->191->186; live fractional carry", target, source)
			runtime.KeepAlive(target)
			runtime.KeepAlive(npc)
		})
	}
}
