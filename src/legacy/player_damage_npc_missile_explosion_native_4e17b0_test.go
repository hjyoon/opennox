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

func TestPlayerDamageNPCMissileExplosionNativeCallback4E17B0(t *testing.T) {
	_ = npcReflectLegacyServer4E17B0(t)
	for _, splash := range []bool{false, true} {
		t.Run(fmt.Sprintf("splash-%t", splash), func(t *testing.T) {
			var pin runtime.Pinner
			defer pin.Unpin()
			target, source := npcReflectLegacyObjects4E17B0(t, &pin)
			target.Buffs = 0
			ud := target.UpdateDataMonster()
			ud.Field518, ud.Field1 = math.Float32bits(0.5), math.Float32bits(0.25)
			missile := &server.Object{TypeInd: 697, ObjClass: object.ClassMissile, ObjOwner: source, PrevPos: types.Ptf(23, 9)}
			pin.Pin(missile)
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(missile)) <= math.MaxUint32 {
				t.Fatal("missile below 4 GiB")
			}
			weapon := missile
			if splash {
				source, weapon = missile, nil
			}
			wantMarker, wantType := uint32(1), uint32(missile.TypeInd)
			if splash {
				wantMarker, wantType = 2, 7
			}
			if !objectDamageDispatchCallNative(target, source, weapon, 8, object.DamageExplosion) || target.HealthData.Cur != 196 ||
				target.Field38 != math.MaxUint32 || target.Obj130 != missile || target.Pos132 != missile.PrevPos || target.Field131 != 7 || target.Frame134 != 1400 ||
				ud.Field547 != wantMarker || ud.Field546 != wantType || ud.Field1 != math.Float32bits(0.25) {
				t.Fatalf("C->PlayerDamage->DefaultDamage EXPLOSION HP=%d sync=%x source=%p marker=%d/%d carry=%#x", target.HealthData.Cur, target.Field38, target.Obj130, ud.Field547, ud.Field546, ud.Field1)
			}
			t.Logf("C->PlayerDamage->DefaultDamage->UnitSetHP: target=%p missile=%p HP=200->196", target, missile)
			runtime.KeepAlive(target)
			runtime.KeepAlive(source)
			runtime.KeepAlive(missile)
		})
	}
}
