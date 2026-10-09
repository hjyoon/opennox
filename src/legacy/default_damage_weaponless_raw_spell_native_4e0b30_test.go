package legacy

import (
	"fmt"
	"math"
	"runtime"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func TestDefaultDamageWorldWeaponlessRawSpellNativeCallback4E0B30(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	oldState := Nox_xxx_playerSetState_4FA020
	Nox_xxx_playerSetState_4FA020 = func(*server.Object, server.PlayerState) bool {
		t.Fatal("small raw hit entered hurt state")
		return false
	}
	t.Cleanup(func() { Nox_xxx_playerSetState_4FA020 = oldState })
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	const id = "NativeWorldRawSpellDefault"
	if err := srv.Types.ReadObjectType(&things.Thing{Name: id, OnDamage: &things.ProcFunc{Name: "DefaultDamage"}}); err != nil {
		t.Fatal(err)
	}
	for _, from := range []string{"player", "monster", "NPC"} {
		for _, to := range []string{"player", "monster", "NPC"} {
			if from == to {
				continue
			}
			for _, typ := range []object.DamageType{object.DamageManaBomb, object.DamageZapRay} {
				t.Run(fmt.Sprintf("%s-to-%s/type-%d", from, to, typ), func(t *testing.T) {
					var pin runtime.Pinner
					defer pin.Unpin()
					npcTarget, playerTarget := npcReflectLegacyObjects4E17B0(t, &pin)
					npcSource, playerSource := npcReflectLegacyObjects4E17B0(t, &pin)
					target, source := npcTarget, npcSource
					if to == "player" {
						target = playerTarget
					} else if to == "monster" {
						target.ObjSubClass = 0x202
					}
					if from == "player" {
						source = playerSource
					} else if from == "monster" {
						source.ObjSubClass = 0x202
					}
					target.Buffs, source.Buffs = 0, 0
					source.PrevPos = types.Ptf(-17, 23)
					target.Damage = srv.Types.ByID(id).Damage
					for hit, want := range []uint16{191, 182, 173} {
						if !objectDamageDispatchCallNative(target, source, nil, 9, typ) || target.HealthData.Cur != want || source.HealthData.Cur != 200 ||
							target.Field38 != math.MaxUint32 || target.Obj130 != source || target.Pos132 != source.PrevPos || target.Field131 != uint32(typ) || target.Frame134 != 1400 {
							t.Fatalf("hit=%d HP=%d/%d type=%d source=%p", hit, target.HealthData.Cur, want, target.Field131, target.Obj130)
						}
						if to != "player" && (target.UpdateDataMonster().Field547 != 2 || target.UpdateDataMonster().Field546 != uint32(typ)) {
							t.Fatal("native raw hit marker missing")
						}
						wantHitFrame := uint32(0)
						if srv.IsEnemyTo(target, source) {
							wantHitFrame = 1400
						}
						if from != "player" && source.UpdateDataMonster().Field130 != wantHitFrame {
							t.Fatalf("native hostile-hit timestamp=%d want %d", source.UpdateDataMonster().Field130, wantHitFrame)
						}
					}
					t.Logf("C->DefaultDamage->UnitSetHP %s->%s type=%d target=%p source=%p HP=200->191->182->173", from, to, typ, target, source)
					runtime.KeepAlive(target)
					runtime.KeepAlive(source)
				})
			}
		}
	}
}
