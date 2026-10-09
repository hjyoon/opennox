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

func TestPlayerDamageWorldExplosionNPCNativeCallback4E17B0(t *testing.T) {
	_ = npcReflectLegacyServer4E17B0(t)
	for variant := 0; variant < 2; variant++ {
		for _, immune := range []bool{false, true} {
			t.Run(fmt.Sprintf("barrel-%d/immune-%t", variant+1, immune), func(t *testing.T) {
				var pin runtime.Pinner
				defer pin.Unpin()
				target, player := npcReflectLegacyObjects4E17B0(t, &pin)
				target.Buffs = 0
				if immune {
					target.ObjSubClass |= 0x400
				}
				source := &server.Object{TypeInd: uint16(773 + variant), ObjClass: object.ClassSimple | object.ClassLight, ObjFlags: object.FlagNoCollide, PrevPos: types.Ptf(23, -9)}
				pin.Pin(source)
				if source.UpdateData != nil || (unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(source)) <= math.MaxUint32) {
					t.Fatal("NPC barrel source lost native width or gained AI data")
				}
				ud := target.UpdateDataMonster()
				ud.Field1, ud.Field518 = math.Float32bits(.25), math.Float32bits(.5)
				want := []uint16{195, 191, 186}
				if immune {
					want = []uint16{198, 196, 194}
				}
				for hit, wantHP := range want {
					wantCarry := float32(-.25)
					if hit == 1 {
						wantCarry = .25
					}
					if !objectDamageDispatchCallNative(target, source, nil, 9, object.DamageExplosion) || target.HealthData.Cur != wantHP || target.Field38 != math.MaxUint32 ||
						target.Obj130 != source || target.Pos132 != source.PrevPos || target.Field131 != 7 || target.Frame134 != 1400 || ud.Field547 != 2 || ud.Field546 != 7 || ud.Field1 != math.Float32bits(wantCarry) {
						t.Fatalf("C->PlayerDamage NPC barrel hit=%d HP=%d/%d marker=%d/%d carry=%x/%x source=%p/%p", hit, target.HealthData.Cur, wantHP, ud.Field547, ud.Field546, ud.Field1, math.Float32bits(wantCarry), target.Obj130, source)
					}
				}
				t.Logf("C->PlayerDamage NPC->DefaultDamage->UnitSetHP: world source=%p immune=%t HP=%v carry=verified", source, immune, want)
				runtime.KeepAlive(source)
				runtime.KeepAlive(target)
				runtime.KeepAlive(player)
			})
		}
	}
}
