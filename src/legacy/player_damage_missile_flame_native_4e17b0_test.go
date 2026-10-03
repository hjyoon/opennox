package legacy

import (
	"fmt"
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func TestPlayerDamageNPCMissileFlameNativeCallback4E17B0(t *testing.T) {
	_ = npcReflectLegacyServer4E17B0(t)
	for _, splash := range []bool{false, true} {
		t.Run(fmt.Sprintf("splash-%t", splash), func(t *testing.T) {
			var pin runtime.Pinner
			defer pin.Unpin()
			target, source := npcReflectLegacyObjects4E17B0(t, &pin)
			target.Buffs = 0
			missile := &server.Object{TypeInd: 695, ObjClass: object.ClassMissile, ObjOwner: source, PrevPos: types.Ptf(23, 9)}
			pin.Pin(missile)
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(missile)) <= math.MaxUint32 {
				t.Fatal("missile below 4 GiB")
			}
			weapon := missile
			if splash {
				source, weapon = missile, nil
			}
			if !objectDamageDispatchCallNative(target, source, weapon, 9, object.DamageFlame) || target.HealthData.Cur != 191 ||
				target.Field38 != math.MaxUint32 || target.Obj130 != missile || target.Pos132 != missile.PrevPos || target.Field131 != 1 || target.Frame134 != 1400 {
				t.Fatalf("C->PlayerDamage->DefaultDamage FLAME HP=%d sync=%x source=%p", target.HealthData.Cur, target.Field38, target.Obj130)
			}
			t.Logf("C->PlayerDamage->DefaultDamage->UnitSetHP: target=%p missile=%p HP=200->191", target, missile)
			runtime.KeepAlive(target)
			runtime.KeepAlive(source)
			runtime.KeepAlive(missile)
		})
	}
}

func TestPlayerDamagePlayerMissileFlameNativeCallback4E17B0(t *testing.T) {
	_ = npcReflectLegacyServer4E17B0(t)
	oldSetState := Nox_xxx_playerSetState_4FA020
	Nox_xxx_playerSetState_4FA020 = func(*server.Object, server.PlayerState) bool {
		t.Fatal("below-threshold FLAME requested hurt state")
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
			missile := &server.Object{TypeInd: 695, ObjClass: object.ClassMissile, ObjOwner: npc, PrevPos: types.Ptf(23, 9)}
			pin.Pin(missile)
			source, weapon := npc, missile
			if splash {
				source, weapon = missile, nil
			}
			if !objectDamageDispatchCallNative(target, source, weapon, 9, object.DamageFlame) || target.HealthData.Cur != 191 ||
				target.Field38 != math.MaxUint32 || target.Obj130 != missile || target.Pos132 != missile.PrevPos || target.Field131 != 1 || target.Frame134 != 1400 ||
				target.UpdateDataPlayer().Field76 != map[bool]uint32{false: 1, true: 2}[splash] {
				t.Fatalf("C->PlayerDamage->DefaultDamage player FLAME HP=%d sync=%x source=%p", target.HealthData.Cur, target.Field38, target.Obj130)
			}
			t.Logf("C->PlayerDamage->DefaultDamage->UnitSetHP: NPC=%p target=%p missile=%p HP=200->191", npc, target, missile)
			runtime.KeepAlive(target)
			runtime.KeepAlive(npc)
			runtime.KeepAlive(missile)
		})
	}
}
