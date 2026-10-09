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

func TestPlayerDamageNPCEnvironmentNativeCallback(t *testing.T) {
	_ = npcReflectLegacyServer4E17B0(t)
	for _, hazard := range []string{"lava", "spike", "spike-block", "immobile-spike-block"} {
		for _, immune := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/fire-immune-%t", hazard, immune), func(t *testing.T) {
				var pin runtime.Pinner
				defer pin.Unpin()
				target, owner := npcReflectLegacyObjects4E17B0(t, &pin)
				target.Buffs = 0
				if immune {
					target.ObjSubClass |= 0x400
				}
				ud := target.UpdateDataMonster()
				ud.Field1, ud.Field518 = math.Float32bits(.25), math.Float32bits(.5)
				typ := object.DamageImpale
				var source, weapon *server.Object
				want := []uint16{195, 191, 186}
				if hazard == "lava" {
					typ, want = object.DamageLava, []uint16{191, 182, 173}
					if immune {
						want = []uint16{200, 200, 200}
					}
				} else {
					class := object.ClassDangerous | object.ClassVisibleEnable | object.ClassImmobile
					if hazard == "spike-block" {
						class = object.ClassDangerous | object.ClassVisibleEnable | object.ClassSimple
					}
					weapon = &server.Object{TypeInd: 211, ObjClass: class, PrevPos: types.Ptf(23, -9)}
					pin.Pin(weapon)
					source = weapon
				}
				for hit, wantHP := range want {
					wantCarry := float32(-.25)
					if hit == 1 || hazard == "lava" {
						wantCarry = .25
					}
					if !objectDamageDispatchCallNative(target, source, weapon, 9, typ) || target.HealthData.Cur != wantHP || ud.Field1 != math.Float32bits(wantCarry) {
						t.Fatalf("C NPC environment hit=%d HP=%d/%d carry=%x/%x", hit, target.HealthData.Cur, wantHP, ud.Field1, math.Float32bits(wantCarry))
					}
					if !(hazard == "lava" && immune) && (target.Obj130 != weapon || target.Field131 != uint32(typ) || target.Frame134 != 1400 || target.Field38 != math.MaxUint32 || ud.Field547 != 2 || ud.Field546 != uint32(typ)) {
						t.Fatal("NPC environment attribution")
					}
					if hazard == "lava" && ud.StatusFlags&object.MonStatusOnFire == 0 {
						t.Fatal("NPC lava OnFire latch")
					}
				}
				runtime.KeepAlive(target)
				runtime.KeepAlive(owner)
				runtime.KeepAlive(weapon)
			})
		}
	}
}
