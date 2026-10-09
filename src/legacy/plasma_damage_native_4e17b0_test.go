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

func TestPlasmaDamageNativeCallback4E17B0UnitPairs(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	oldState := Nox_xxx_playerSetState_4FA020
	Nox_xxx_playerSetState_4FA020 = func(*server.Object, server.PlayerState) bool {
		t.Fatal("small Plasma hit entered hurt state")
		return false
	}
	t.Cleanup(func() { Nox_xxx_playerSetState_4FA020 = oldState })
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	const id = "NativePlasmaDefaultDamage"
	if err := srv.Types.ReadObjectType(&things.Thing{Name: id, OnDamage: &things.ProcFunc{Name: "DefaultDamage"}}); err != nil {
		t.Fatal(err)
	}
	for _, from := range []string{"player", "monster", "NPC"} {
		for _, to := range []string{"player", "monster", "NPC"} {
			t.Run(from+"-to-"+to, func(t *testing.T) {
				var pin runtime.Pinner
				defer pin.Unpin()
				npcTarget, playerTarget := npcReflectLegacyObjects4E17B0(t, &pin)
				npcSource, playerSource := npcReflectLegacyObjects4E17B0(t, &pin)
				target, source := npcTarget, npcSource
				if to == "player" {
					target = playerTarget
					target.Damage = playerDamageMeleeCallbackNative4E17B0()
				}
				if to == "monster" {
					target.ObjSubClass, target.Damage = 0x10202, srv.Types.ByID(id).Damage
				}
				if from == "player" {
					source = playerSource
				}
				if from == "monster" {
					source.ObjSubClass = 0x10202
				}
				target.Buffs, source.Buffs = 0, 0
				source.PosVec, source.PrevPos = types.Ptf(20, 0), types.Ptf(-20, 7)
				for hit, want := range []uint16{191, 182, 173} {
					if !objectDamageDispatchCallNative(target, source, nil, 9, object.DamagePlasma) || target.HealthData.Cur != want || source.HealthData.Cur != 200 || target.Field38 != math.MaxUint32 || target.Obj130 != source || target.Pos132 != source.PrevPos || target.Field131 != 14 || target.Frame134 != 1400 {
						t.Fatalf("hit=%d native Plasma HP=%d/%d source=%p type=%d", hit, target.HealthData.Cur, want, target.Obj130, target.Field131)
					}
					if to == "player" {
						if ud := target.UpdateDataPlayer(); ud.Field76 != 2 || ud.Field75 != 14 {
							t.Fatal("native player raw Plasma marker")
						}
					} else if ud := target.UpdateDataMonster(); ud.Field547 != 2 || ud.Field546 != 14 || !ud.StatusFlags.Has(object.MonStatusOnFire) || !ud.StatusFlags.Has(object.MonStatusInjured) {
						t.Fatal("native monster/NPC Plasma injured/on-fire markers")
					}
				}
				t.Logf("actual C->registered damage->UnitSetHP Plasma %s->%s target=%p source=%p HP=200->191->182->173", from, to, target, source)
				runtime.KeepAlive(target)
				runtime.KeepAlive(source)
			})
		}
	}
}
