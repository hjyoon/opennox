package legacy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

// Direct callbacks and Object.CallDamage can retain a source and nil weapon.
// These tests execute the registered native C damage dispatch and real HP.
func TestPlayerDamageWeaponlessNonItemNativeDamage4E17B0(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	// Bind the production preflight service. Damage9 or absorbed5 never
	// reaches its hurt-state threshold; HP and damage remain native.
	oldState := Nox_xxx_playerSetState_4FA020
	Nox_xxx_playerSetState_4FA020 = func(*server.Object, server.PlayerState) bool {
		t.Fatal("below-threshold raw/fire damage requested hurt state")
		return false
	}
	t.Cleanup(func() { Nox_xxx_playerSetState_4FA020 = oldState })
	srv.Audio.Init(srv)
	t.Cleanup(srv.Audio.Free)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	for _, name := range []string{"PlayerDamage", "DefaultDamage"} {
		if err := srv.Types.ReadObjectType(&things.Thing{Name: "NativeWeaponlessNonItem" + name, OnDamage: &things.ProcFunc{Name: name}}); err != nil {
			t.Fatal(err)
		}
	}
	oldLog := srv.Log
	t.Cleanup(func() { srv.Log = oldLog })
	for _, sourceKind := range []string{"player", "monster", "world", "zero"} {
		for _, weaponKind := range []string{"none"} {
			for typ := object.DamageType(0); typ < 18; typ++ {
				for _, kind := range []string{"player", "npc", "ordinary-monster"} {
					for _, caller := range []bool{false, true} {
						for _, god := range []bool{false, true} {
							t.Run(fmt.Sprintf("source-%s/weapon-%s/type-%d/target-%s/caller-%t/god-%t", sourceKind, weaponKind, int32(typ), kind, caller, god), func(t *testing.T) {
								if god {
									noxflags.SetEngine(noxflags.EngineGodMode)
								} else {
									noxflags.UnsetEngine(noxflags.EngineGodMode)
								}
								var pin runtime.Pinner
								defer pin.Unpin()
								npc, _ := npcReflectLegacyObjects4E17B0(t, &pin)
								_, attacker := npcReflectLegacyObjects4E17B0(t, &pin)
								weapon, _ := npcReflectLegacyObjects4E17B0(t, &pin)
								attacker.UpdateDataPlayer().Player.PlayerUnit = attacker
								attacker.UpdateDataPlayer().Player.PlayerInd = 9
								if sourceKind != "player" {
									attacker, _ = npcReflectLegacyObjects4E17B0(t, &pin)
								}
								if sourceKind == "world" || sourceKind == "zero" {
									attacker.ObjClass, attacker.ObjSubClass, attacker.UpdateData, attacker.HealthData = object.ClassDoor, 0, nil, nil
									if sourceKind == "zero" {
										attacker.ObjClass = 0
									}
								}
								if weaponKind == "world" {
									weapon.ObjClass, weapon.ObjSubClass, weapon.UpdateData, weapon.HealthData = object.ClassArmor, 0, nil, nil
								}
								weapon.TypeInd, weapon.ObjOwner = 73, attacker
								attacker.Buffs, weapon.Buffs = 0, 0
								_, hero := npcReflectLegacyObjects4E17B0(t, &pin)
								target := npc
								callback := "PlayerDamage"
								if kind == "player" {
									target = hero
								} else if kind == "ordinary-monster" {
									target.ObjSubClass = 0
									callback = "DefaultDamage"
								}
								target.Buffs = 0
								attacker.PrevPos = types.Ptf(-25, 4)
								weapon.PrevPos = types.Ptf(32.5, -18.25)
								attacker.TeamVal.ID, target.TeamVal.ID, weapon.TeamVal.ID = 1, 2, 1
								attacker.TypeInd = 71
								weapon = nil
								target.Damage = srv.Types.ByID("NativeWeaponlessNonItem" + callback).Damage
								target.Field38, target.Frame134, target.Obj130, target.Pos132 = 0, 77, nil, types.Ptf(29, 31)
								var carry, marker, kindField *uint32
								if kind == "player" {
									u := target.UpdateDataPlayer()
									u.Player.PlayerUnit = target
									u.Player.PlayerInd = 8
									u.Field57, u.Field21, u.Field76, u.Field75 = math.Float32bits(.5), math.Float32bits(.25), 99, 77
									carry, marker, kindField = &u.Field21, &u.Field76, &u.Field75
								} else {
									u := target.UpdateDataMonster()
									u.Field518, u.Field1, u.Field547, u.Field546 = math.Float32bits(.5), math.Float32bits(.25), 99, 77
									carry, marker, kindField = &u.Field1, &u.Field547, &u.Field546
								}
								wantHP, wantCarry := uint16(195), float32(-.25)
								physical := typ == object.DamageBlade || typ == object.DamageImpale || typ == object.DamageExplosion || typ == object.DamageBite || typ == object.DamageClaw || typ == object.DamageImpact || typ == object.DamageCrush
								if !physical || kind == "ordinary-monster" {
									wantHP, wantCarry = 191, .25
								}
								if typ == object.DamageCrush && kind != "ordinary-monster" {
									wantHP, wantCarry = 193, 0
								}
								if typ == object.DamageExplosion && kind != "ordinary-monster" {
									wantHP, wantCarry = 196, .75
								}
								blocked := god && kind == "player"
								if blocked {
									wantHP = 200
								}
								wantMarker, wantKind := uint32(2), uint32(typ)
								if typ == object.DamageCrush || typ == object.DamageClaw {
									wantMarker, wantKind = 1, 71
								}
								if kind == "player" && typ == object.DamageDeathMagic {
									wantMarker, wantKind = 0, 77
								}
								var diagnostics bytes.Buffer
								srv.Log = slog.New(slog.NewTextHandler(&diagnostics, nil))
								var result bool
								if caller {
									result = target.CallDamage(attacker, nil, 9, typ)
								} else {
									result = objectDamageDispatchCallNative(target, attacker, weapon, 9, typ)
								}
								if !result || strings.Contains(diagnostics.String(), "not ported") || target.HealthData.Cur != wantHP || *carry != math.Float32bits(wantCarry) || *marker != wantMarker || *kindField != wantKind {
									t.Fatalf("weaponless source result=%t HP=%d/%d carry=%x/%x marker=%d/%d log=%s", result, target.HealthData.Cur, wantHP, *carry, math.Float32bits(wantCarry), *marker, *kindField, diagnostics.String())
								}
								if blocked {
									if target.Field38 != 0 || target.Frame134 != 77 || target.Obj130 != nil || target.Pos132 != types.Ptf(29, 31) {
										t.Fatal("GodMode crossed HP/attribution boundary")
									}
								} else if target.Field38 != math.MaxUint32 || target.Frame134 != 1400 || target.Obj130 != attacker || target.Pos132 != attacker.PrevPos || target.Field131 != uint32(typ) {
									t.Fatal("native weaponless-source attribution/HP sync")
								}
								runtime.KeepAlive(target)
								runtime.KeepAlive(attacker)
								runtime.KeepAlive(weapon)
							})
						}
					}
				}
			}
		}
	}
}

func TestPlayerDamageWeaponlessNonItemImmunityNative4E17B0(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	const id = "NativeWeaponlessNonItemImmunity"
	if err := srv.Types.ReadObjectType(&things.Thing{Name: id, OnDamage: &things.ProcFunc{Name: "PlayerDamage"}}); err != nil {
		t.Fatal(err)
	}
	callback := srv.Types.ByID(id).Damage
	oldLog := srv.Log
	t.Cleanup(func() { srv.Log = oldLog })
	blob, err := os.ReadFile("../server/testdata/player_damage_weaponless_nonitem_integrated_4e17b0.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		OriginalSHA string `json:"original_sha256"`
		Rows        []struct {
			Type        object.DamageType    `json:"damage_type"`
			Subclass    object.SubClass      `json:"subclass"`
			Forward     []uint32             `json:"forward_damage"`
			Carry       uint32               `json:"carry_bits"`
			Marker      uint32               `json:"marker"`
			Kind        uint32               `json:"marker_type"`
			Status      object.MonsterStatus `json:"status_flags"`
			Electric    []byte               `json:"electric_state_bytes"`
			SourceLatch uint32               `json:"source_monster_latch"`
			SourceKind  string               `json:"source_kind"`
			WeaponKind  string               `json:"weapon_kind"`
			WeaponLatch uint32               `json:"weapon_monster_latch"`
			Return      int                  `json:"return"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(fixture.Rows) != 360 {
		t.Fatal("unexpected original integrated NPC fixture")
	}
	for _, row := range fixture.Rows {
		for _, mode := range []struct{ god, nilHP bool }{{}, {god: true}, {nilHP: true}} {
			t.Run(fmt.Sprintf("source-%s/weapon-%s/type-%d/subclass-%x/god-%t/nil-hp-%t", row.SourceKind, row.WeaponKind, int32(row.Type), uint32(row.Subclass), mode.god, mode.nilHP), func(t *testing.T) {
				if mode.god {
					noxflags.SetEngine(noxflags.EngineGodMode)
				} else {
					noxflags.UnsetEngine(noxflags.EngineGodMode)
				}
				var pin runtime.Pinner
				defer pin.Unpin()
				var diagnostics bytes.Buffer
				srv.Log = slog.New(slog.NewTextHandler(&diagnostics, nil))
				target, _ := npcReflectLegacyObjects4E17B0(t, &pin)
				_, player := npcReflectLegacyObjects4E17B0(t, &pin)
				weapon, _ := npcReflectLegacyObjects4E17B0(t, &pin)
				player.UpdateDataPlayer().Player.PlayerUnit = player
				if row.SourceKind != "player" {
					player, _ = npcReflectLegacyObjects4E17B0(t, &pin)
				}
				if row.SourceKind == "world" || row.SourceKind == "zero" {
					player.ObjClass, player.ObjSubClass, player.UpdateData, player.HealthData = object.ClassDoor, 0, nil, nil
					if row.SourceKind == "zero" {
						player.ObjClass = 0
					}
				}
				if row.WeaponKind == "world" {
					weapon.ObjClass, weapon.ObjSubClass, weapon.UpdateData, weapon.HealthData = object.ClassArmor, 0, nil, nil
				}
				weapon.TypeInd, weapon.ObjOwner, weapon.TeamVal.ID = 73, player, 1
				player.Buffs, weapon.Buffs = 0, 0
				if mode.nilHP {
					target.HealthData = nil
				}
				target.ObjSubClass = row.Subclass
				target.Damage = callback
				target.Buffs = 0
				if row.SourceKind == "monster" {
					player.UpdateDataMonster().Field130 = 0
				}
				player.PrevPos = types.Ptf(-25, 4)
				weapon.PrevPos = types.Ptf(32.5, -18.25)
				target.TeamVal.ID, player.TeamVal.ID = 2, 1
				player.TypeInd = 71
				weapon = nil
				target.Obj130, target.Pos132, target.Frame134 = player, types.Ptf(29, 31), 77
				ud := target.UpdateDataMonster()
				ud.Field518, ud.Field1 = math.Float32bits(.5), math.Float32bits(.25)
				ud.Field523_2, ud.Field523_3 = 0x11, 0xab
				if len(row.Forward) > 1 || len(row.Electric) != 2 {
					t.Fatal("invalid original boundary/state vector")
				}
				wantHP := uint16(200)
				if len(row.Forward) != 0 {
					wantHP -= uint16(row.Forward[0])
				}
				result := objectDamageDispatchCallNative(target, player, weapon, 9, row.Type)
				if strings.Contains(diagnostics.String(), "not ported") {
					t.Fatal(diagnostics.String())
				}
				hpOK := target.HealthData == nil
				if !mode.nilHP {
					hpOK = target.HealthData != nil && target.HealthData.Cur == wantHP
				}
				if result != (row.Return != 0) || !hpOK || ud.Field1 != row.Carry || ud.Field547 != row.Marker || ud.Field546 != row.Kind || ud.StatusFlags != row.Status || ud.Field523_2 != row.Electric[0] || ud.Field523_3 != row.Electric[1] {
					t.Fatalf("C NPC result=%t HP-valid=%t want=%d carry=%x/%x marker=%d/%d kind=%d/%d status=%x/%x electric=%x,%x", result, hpOK, wantHP, ud.Field1, row.Carry, ud.Field547, row.Marker, ud.Field546, row.Kind, ud.StatusFlags, row.Status, ud.Field523_2, ud.Field523_3)
				}
				if row.SourceKind == "monster" && player.UpdateDataMonster().Field130 != row.SourceLatch {
					t.Fatal("source monster hit timestamp")
				}
				if row.WeaponKind == "monster" && weapon.UpdateDataMonster().Field130 != row.WeaponLatch {
					t.Fatal("distinct weapon timestamp changed")
				}
				if len(row.Forward) == 0 {
					if target.Field38 != 0 || target.Obj130 != player || target.Pos132 != types.Ptf(29, 31) || target.Frame134 != 77 {
						t.Fatal("immune NPC crossed attribution/HP boundary")
					}
				} else {
					wantSync := uint32(math.MaxUint32)
					if mode.nilHP {
						wantSync = 0
					}
					if target.Field38 != wantSync || target.Obj130 != player || target.Pos132 != player.PrevPos || target.Frame134 != 1400 || target.Field131 != uint32(row.Type) {
						t.Fatal("weaponless-source NPC attribution/HP sync")
					}
				}
				runtime.KeepAlive(target)
				runtime.KeepAlive(player)
				runtime.KeepAlive(weapon)
			})
		}
	}
}
