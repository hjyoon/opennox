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

// Actual C dispatcher -> registered DefaultDamage -> UnitSetHP. The three
// stock spike class combinations cover Spike/PeriodicSpike, both blocks and
// both rotating variants. There is no artificial friendly IsEnemy override.
func TestDefaultDamageEnvironmentNativeCallback(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	const id = "NativeEnvironmentDefault"
	if err := srv.Types.ReadObjectType(&things.Thing{Name: id, OnDamage: &things.ProcFunc{Name: "DefaultDamage"}}); err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"monster", "npc"} {
		for _, hazard := range []string{"lava", "spike", "spike-block", "immobile-spike-block"} {
			for _, immune := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/fire-immune-%t", kind, hazard, immune), func(t *testing.T) {
					var pin runtime.Pinner
					defer pin.Unpin()
					target, owner := npcReflectLegacyObjects4E17B0(t, &pin)
					target.Buffs = 0
					if kind == "monster" {
						target.ObjSubClass = 0x202
					}
					if immune {
						target.ObjSubClass |= 0x400
					}
					target.Damage = srv.Types.ByID(id).Damage
					ud := target.UpdateDataMonster()
					ud.Field1, ud.Field518 = math.Float32bits(.25), math.Float32bits(.5)
					typ := object.DamageImpale
					var source, weapon *server.Object
					if hazard == "lava" {
						typ = object.DamageLava
					} else {
						class := object.ClassDangerous | object.ClassVisibleEnable
						if hazard == "spike-block" {
							class |= object.ClassSimple
						} else {
							class |= object.ClassImmobile
						}
						weapon = &server.Object{TypeInd: 211, ObjClass: class, PrevPos: types.Ptf(23, -9)}
						pin.Pin(weapon)
						source = weapon
						if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(weapon)) <= math.MaxUint32 {
							t.Fatal("hazard pointer is not native width")
						}
					}
					for hit := 1; hit <= 3; hit++ {
						wantHP := uint16(200 - hit*3)
						blocked := hazard == "lava" && immune
						if blocked {
							wantHP = 200
						}
						if !objectDamageDispatchCallNative(target, source, weapon, 3, typ) || target.HealthData.Cur != wantHP {
							t.Fatalf("C environment hit=%d HP=%d want=%d", hit, target.HealthData.Cur, wantHP)
						}
						if ud.Field1 != math.Float32bits(.25) || ud.Field518 != math.Float32bits(.5) {
							t.Fatal("DefaultDamage replayed NPC armor/carry")
						}
						if !blocked && (target.Obj130 != weapon || target.Field131 != uint32(typ) || target.Frame134 != 1400 || target.Field38 != math.MaxUint32 || ud.Field547 != 2 || ud.Field546 != uint32(typ) || ud.StatusFlags&object.MonStatusInjured == 0) {
							t.Fatal("environment hit lost native attribution/health synchronization")
						}
						if hazard == "lava" && ud.StatusFlags&object.MonStatusOnFire == 0 {
							t.Fatal("original entry-time OnFire latch missing")
						}
					}
					runtime.KeepAlive(owner)
					runtime.KeepAlive(target)
					runtime.KeepAlive(weapon)
				})
			}
		}
	}
}
