package legacy

import (
	"math"
	"runtime"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func TestDefaultDamagePlayerWeaponlessExplosionNativeCallback4E0B30(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	oldState := Nox_xxx_playerSetState_4FA020
	Nox_xxx_playerSetState_4FA020 = func(*server.Object, server.PlayerState) bool {
		t.Fatal("small weapon-less EXPLOSION requested hurt state")
		return false
	}
	t.Cleanup(func() { Nox_xxx_playerSetState_4FA020 = oldState })
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	if err := srv.Types.ReadObjectType(&things.Thing{Name: "NativePlayerWeaponlessExplosionDefault", OnDamage: &things.ProcFunc{Name: "DefaultDamage"}}); err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"none", "npc", "self"} {
		t.Run(kind, func(t *testing.T) {
			var pin runtime.Pinner
			defer pin.Unpin()
			npc, target := npcReflectLegacyObjects4E17B0(t, &pin)
			npc.Buffs = 0
			target.Damage = srv.Types.ByID("NativePlayerWeaponlessExplosionDefault").Damage
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
			for hit, wantHP := range []uint16{191, 182, 173} {
				if !objectDamageDispatchCallNative(target, source, nil, 9, object.DamageExplosion) || target.HealthData.Cur != wantHP ||
					target.Field38 != math.MaxUint32 || target.Obj130 != source || target.Pos132 != wantPos || target.Field131 != 7 || target.Frame134 != 1400 ||
					ud.Field76 != 99 || ud.Field75 != 77 || ud.Field21 != math.Float32bits(0.25) || ud.Field57 != math.Float32bits(0.5) {
					t.Fatalf("C->DefaultDamage weapon-less EXPLOSION hit=%d HP=%d sync=%x source=%p marker=%d/%d carry=%#x", hit, target.HealthData.Cur, target.Field38, target.Obj130, ud.Field76, ud.Field75, ud.Field21)
				}
			}
			t.Logf("C->DefaultDamage->UnitSetHP: target=%p source=%p HP=200->191->182->173; prefix untouched", target, source)
			runtime.KeepAlive(target)
			runtime.KeepAlive(npc)
		})
	}
}
