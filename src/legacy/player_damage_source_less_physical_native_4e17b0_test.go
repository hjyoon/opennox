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

func TestPlayerDamageSourceLessPhysicalNativeCallback4E17B0(t *testing.T) {
	_ = npcReflectLegacyServer4E17B0(t)
	oldState := Nox_xxx_playerSetState_4FA020
	Nox_xxx_playerSetState_4FA020 = func(*server.Object, server.PlayerState) bool {
		t.Fatal("below-threshold source-less hit requested a hurt state")
		return false
	}
	t.Cleanup(func() { Nox_xxx_playerSetState_4FA020 = oldState })
	for _, typ := range []object.DamageType{object.DamageBlade, object.DamageCrush, object.DamageBite} {
		t.Run(fmt.Sprint(typ), func(t *testing.T) {
			var pin runtime.Pinner
			defer pin.Unpin()
			npc, target := npcReflectLegacyObjects4E17B0(t, &pin)
			target.Damage = playerDamageMeleeCallbackNative4E17B0()
			ud := target.UpdateDataPlayer()
			ud.Field57, ud.Field21 = math.Float32bits(.5), math.Float32bits(.25)
			ud.Field76, ud.Field75 = 99, 77
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(target)) <= math.MaxUint32 {
				t.Fatal("source-less player fixture must be above 4 GiB")
			}
			wantHP, wantCarry := uint16(195), float32(-.25)
			if typ == object.DamageCrush {
				wantHP, wantCarry = 193, 0
			}
			if !objectDamageDispatchCallNative(target, nil, nil, 9, typ) || target.HealthData.Cur != wantHP ||
				target.Field38 != math.MaxUint32 || target.Obj130 != nil || target.Pos132 != (types.Pointf{}) || target.Field131 != uint32(typ) || target.Frame134 != 1400 ||
				ud.Field76 != 2 || ud.Field75 != uint32(typ) || ud.Field21 != math.Float32bits(wantCarry) {
				t.Fatalf("C->PlayerDamage->DefaultDamage->UnitSetHP type=%d HP=%d/%d marker=%d/%d carry=%x attribution=%p", typ, target.HealthData.Cur, wantHP, ud.Field76, ud.Field75, ud.Field21, target.Obj130)
			}
			runtime.KeepAlive(npc)
			runtime.KeepAlive(target)
		})
	}
}
