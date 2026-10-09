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

func TestDefaultDamageWorldExplosionNativeCallback4E0B30(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	oldState := Nox_xxx_playerSetState_4FA020
	Nox_xxx_playerSetState_4FA020 = func(v *server.Object, state server.PlayerState) bool {
		if !v.Class().Has(object.ClassPlayer) || state != server.PlayerState30 {
			t.Fatal("wrong barrel hurt-state dispatch")
		}
		v.UpdateDataPlayer().State = state
		return true
	}
	t.Cleanup(func() { Nox_xxx_playerSetState_4FA020 = oldState })
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	const id = "NativeWorldExplosionDefault"
	if err := srv.Types.ReadObjectType(&things.Thing{Name: id, OnDamage: &things.ProcFunc{Name: "DefaultDamage"}}); err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"player", "monster", "npc"} {
		for variant := 0; variant < 2; variant++ {
			t.Run(fmt.Sprintf("%s/barrel-%d", kind, variant+1), func(t *testing.T) {
				var pin runtime.Pinner
				defer pin.Unpin()
				npc, player := npcReflectLegacyObjects4E17B0(t, &pin)
				target := npc
				if kind == "player" {
					target = player
				}
				if kind == "monster" {
					target.ObjSubClass = 0x202
				}
				target.Buffs = 0
				target.Damage = srv.Types.ByID(id).Damage
				source := &server.Object{TypeInd: uint16(773 + variant), ObjClass: object.ClassSimple | object.ClassLight, ObjFlags: object.FlagNoCollide, PrevPos: types.Ptf(23, -9)}
				pin.Pin(source)
				if source.UpdateData != nil || (unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(source)) <= math.MaxUint32) {
					t.Fatal("world source acquired AI data or truncated address")
				}
				if kind == "player" {
					ud := target.UpdateDataPlayer()
					ud.Field76, ud.Field75, ud.Field21, ud.Field57 = 99, 77, math.Float32bits(.25), math.Float32bits(.5)
				}
				for hit, want := range []uint16{170, 140, 110} {
					if !objectDamageDispatchCallNative(target, source, nil, 30, object.DamageExplosion) || target.HealthData.Cur != want || target.Field38 != math.MaxUint32 ||
						target.Obj130 != source || target.Pos132 != source.PrevPos || target.Field131 != 7 || target.Frame134 != 1400 {
						t.Fatalf("hit=%d HP=%d/%d sync=%x source=%p/%p", hit, target.HealthData.Cur, want, target.Field38, target.Obj130, source)
					}
					if kind == "player" {
						ud := target.UpdateDataPlayer()
						if ud.Field76 != 99 || ud.Field75 != 77 || ud.Field21 != math.Float32bits(.25) || ud.Field57 != math.Float32bits(.5) {
							t.Fatal("DefaultDamage replayed PlayerDamage armor/carry prefix")
						}
					} else if ud := target.UpdateDataMonster(); ud.Field547 != 2 || ud.Field546 != 7 {
						t.Fatal("native monster explosion marker missing")
					}
				}
				t.Logf("C->DefaultDamage->UnitSetHP: %s source=%p no UpdateData; HP=200->170->140->110", kind, source)
				runtime.KeepAlive(source)
				runtime.KeepAlive(npc)
				runtime.KeepAlive(player)
			})
		}
	}
}
