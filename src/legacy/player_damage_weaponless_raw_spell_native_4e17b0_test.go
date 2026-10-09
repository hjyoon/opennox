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

func TestPlayerDamageWeaponlessRawSpellNativeCallback4E17B0Pairs(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	oldState := Nox_xxx_playerSetState_4FA020
	Nox_xxx_playerSetState_4FA020 = func(*server.Object, server.PlayerState) bool {
		t.Fatal("small raw hit entered hurt state")
		return false
	}
	t.Cleanup(func() { Nox_xxx_playerSetState_4FA020 = oldState })
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	const id = "NativeRawSpellMonster"
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
						target.Damage = playerDamageMeleeCallbackNative4E17B0()
					} else if to == "monster" {
						target.ObjSubClass, target.Damage = 0x202, srv.Types.ByID(id).Damage
					}
					if from == "player" {
						source = playerSource
					} else if from == "monster" {
						source.ObjSubClass = 0x202
					}
					target.Buffs, source.Buffs = 0, 0
					source.PosVec, source.PrevPos = types.Ptf(20, 0), types.Ptf(-20, 7)
					for hit, want := range []uint16{191, 182, 173} {
						if !objectDamageDispatchCallNative(target, source, nil, 9, typ) || target.HealthData.Cur != want || source.HealthData.Cur != 200 || target.Field38 != math.MaxUint32 ||
							target.Obj130 != source || target.Pos132 != source.PrevPos || target.Field131 != uint32(typ) || target.Frame134 != 1400 {
							t.Fatalf("hit=%d HP=%d/%d source/type=%p/%d", hit, target.HealthData.Cur, want, target.Obj130, target.Field131)
						}
						if to == "player" {
							ud := target.UpdateDataPlayer()
							if ud.Field76 != 2 || ud.Field75 != uint32(typ) {
								t.Fatal("player raw marker missing")
							}
						} else {
							ud := target.UpdateDataMonster()
							if ud.Field547 != 2 || ud.Field546 != uint32(typ) {
								t.Fatal("monster/NPC raw marker missing")
							}
						}
					}
					t.Logf("C->%sDamage->UnitSetHP %s->%s type=%d target=%p source=%p HP=200->191->182->173", map[bool]string{true: "Default", false: "Player"}[to == "monster"], from, to, typ, target, source)
					runtime.KeepAlive(target)
					runtime.KeepAlive(source)
				})
			}
		}
	}
}

func TestPlayerDamageWeaponlessRawSpellNativeCallback4E17B0Reflect(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	oldState := Nox_xxx_playerSetState_4FA020
	Nox_xxx_playerSetState_4FA020 = func(*server.Object, server.PlayerState) bool { t.Fatal("small ray entered hurt state"); return false }
	t.Cleanup(func() { Nox_xxx_playerSetState_4FA020 = oldState })
	for _, player := range []bool{false, true} {
		for _, front := range []bool{false, true} {
			for _, typ := range []object.DamageType{object.DamageManaBomb, object.DamageZapRay} {
				t.Run(fmt.Sprintf("player-%t/front-%t/type-%d", player, front, typ), func(t *testing.T) {
					var pin runtime.Pinner
					defer pin.Unpin()
					npc, pc := npcReflectLegacyObjects4E17B0(t, &pin)
					target, source := npc, pc
					if player {
						target, source = pc, npc
					}
					target.Buffs = 1 << server.ENCHANT_REFLECTIVE_SHIELD
					target.Damage = playerDamageMeleeCallbackNative4E17B0()
					source.PosVec = types.Ptf(20, 0)
					if !front {
						source.PosVec.X = -20
					}
					source.PrevPos = types.Ptf(-source.PosVec.X, 7)
					if got := Nox_server_testTwoPointsAndDirection_4E6E50(target.PosVec, int16(target.Direction1), source.PosVec)&1 != 0; got != front {
						t.Fatal("native reflection facing fixture")
					}
					blocked := front && typ == object.DamageZapRay
					result := objectDamageDispatchCallNative(target, source, nil, 9, typ)
					want := uint16(191)
					if blocked {
						want = 200
					}
					if result == blocked || target.HealthData.Cur != want || target.HasEnchant(server.ENCHANT_REFLECTIVE_SHIELD) != true {
						t.Fatalf("native Reflect result/HP=%t/%d", result, target.HealthData.Cur)
					}
					if !blocked && target.Obj130 != source {
						t.Fatal("raw attribution missing")
					}
					_ = srv // Production FX/audio services use the real isolated server.
					runtime.KeepAlive(target)
					runtime.KeepAlive(source)
				})
			}
		}
	}
}
