package legacy

import (
	"fmt"
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func TestDefaultDamageMissileExplosionNativeCallback4E0B30(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	oldState := Nox_xxx_playerSetState_4FA020
	Nox_xxx_playerSetState_4FA020 = func(*server.Object, server.PlayerState) bool {
		t.Fatal("small EXPLOSION requested hurt state")
		return false
	}
	t.Cleanup(func() { Nox_xxx_playerSetState_4FA020 = oldState })
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	if err := srv.Types.ReadObjectType(&things.Thing{Name: "NativeMissileExplosionDefault", OnDamage: &things.ProcFunc{Name: "DefaultDamage"}}); err != nil {
		t.Fatal(err)
	}
	for _, toPlayer := range []bool{false, true} {
		for _, splash := range []bool{false, true} {
			t.Run(fmt.Sprintf("player-target-%t/splash-%t", toPlayer, splash), func(t *testing.T) {
				var pin runtime.Pinner
				defer pin.Unpin()
				target, source := npcReflectLegacyObjects4E17B0(t, &pin)
				target.Buffs = 0
				if toPlayer {
					target, source = source, target
				}
				target.Damage = srv.Types.ByID("NativeMissileExplosionDefault").Damage
				missile := &server.Object{TypeInd: 697, ObjClass: object.ClassMissile, ObjOwner: source, PrevPos: types.Ptf(23, 9)}
				pin.Pin(missile)
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(missile)) <= math.MaxUint32 {
					t.Fatal("missile below 4 GiB")
				}
				weapon := missile
				if splash {
					source, weapon = missile, nil
				}
				if !objectDamageDispatchCallNative(target, source, weapon, 9, object.DamageExplosion) || target.HealthData.Cur != 191 ||
					target.Field38 != math.MaxUint32 || target.Obj130 != missile || target.Pos132 != missile.PrevPos || target.Frame134 != 1400 || target.Field131 != 7 {
					t.Fatalf("C->DefaultDamage EXPLOSION HP=%d sync=%x attribution=%p", target.HealthData.Cur, target.Field38, target.Obj130)
				}
				if !toPlayer && !target.UpdateDataMonster().StatusFlags.Has(object.MonStatusInjured|object.MonStatusOnFire) {
					t.Fatal("NPC fire/injured status missing")
				}
				t.Logf("C->DefaultDamage->UnitSetHP: target=%p source=%p missile=%p HP=200->191", target, source, missile)
				runtime.KeepAlive(target)
				runtime.KeepAlive(source)
				runtime.KeepAlive(missile)
			})
		}
	}
}
