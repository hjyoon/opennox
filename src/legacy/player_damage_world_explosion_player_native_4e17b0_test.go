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

func TestPlayerDamageWorldExplosionPlayerNativeCallback4E17B0(t *testing.T) {
	_ = npcReflectLegacyServer4E17B0(t)
	for variant := 0; variant < 2; variant++ {
		t.Run(fmt.Sprintf("barrel-%d", variant+1), func(t *testing.T) {
			var pin runtime.Pinner
			defer pin.Unpin()
			npc, target := npcReflectLegacyObjects4E17B0(t, &pin)
			target.Damage = playerDamageMeleeCallbackNative4E17B0()
			source := &server.Object{TypeInd: uint16(773 + variant), ObjClass: object.ClassSimple | object.ClassLight, ObjFlags: object.FlagNoCollide, PrevPos: types.Ptf(23, -9)}
			pin.Pin(source)
			if source.UpdateData != nil || (unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(source)) <= math.MaxUint32) {
				t.Fatal("world source must retain high address without AI data")
			}
			ud := target.UpdateDataPlayer()
			ud.Field76, ud.Field75, ud.Field21, ud.Field57 = 99, 77, math.Float32bits(.25), math.Float32bits(.5)
			for hit, wantHP := range []uint16{195, 191, 186} {
				wantCarry := float32(-.25)
				if hit == 1 {
					wantCarry = .25
				}
				if !objectDamageDispatchCallNative(target, source, nil, 9, object.DamageExplosion) || target.HealthData.Cur != wantHP || target.Field38 != math.MaxUint32 ||
					target.Obj130 != source || target.Pos132 != source.PrevPos || target.Field131 != 7 || target.Frame134 != 1400 || ud.Field76 != 2 || ud.Field75 != 7 || ud.Field21 != math.Float32bits(wantCarry) {
					t.Fatalf("C->PlayerDamage barrel hit=%d HP=%d/%d marker=%d/%d carry=%x/%x source=%p/%p sync=%x", hit, target.HealthData.Cur, wantHP, ud.Field76, ud.Field75, ud.Field21, math.Float32bits(wantCarry), target.Obj130, source, target.Field38)
				}
			}
			t.Logf("C->PlayerDamage->DefaultDamage->UnitSetHP: world source=%p no UpdateData HP=200->195->191->186 carry=verified", source)
			runtime.KeepAlive(source)
			runtime.KeepAlive(target)
			runtime.KeepAlive(npc)
		})
	}
}
