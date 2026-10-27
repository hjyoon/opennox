package legacy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"runtime"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Registered C dispatch and Object.CallDamage retain pure spell missiles,
// terminal owner identity and native HP. The root wrapper is not called.
func TestPlayerDamageSourceLessMixedMissileNativeDamage4E17B0(t *testing.T) {
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
		if err := srv.Types.ReadObjectType(&things.Thing{Name: "NativeSourceLessMixedMissile" + name, OnDamage: &things.ProcFunc{Name: name}}); err != nil {
			t.Fatal(err)
		}
	}
	oldLog := srv.Log
	t.Cleanup(func() { srv.Log = oldLog })
	for _, weaponClass := range []object.Class{0x1000001, 0x05200001, 0x01280001} {
		for _, weaponKind := range []string{"missile"} {
			for _, typ := range []object.DamageType{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17} {
				for _, kind := range []string{"player", "npc", "ordinary-monster"} {
					for _, entry := range []struct{ caller, nilHP bool }{{}, {caller: true}, {nilHP: true}, {caller: true, nilHP: true}} {
						caller, nilHP := entry.caller, entry.nilHP
						for _, god := range []bool{false, true} {
							t.Run(fmt.Sprintf("weapon-%s/class-%x/type-%d/target-%s/caller-%t/god-%t/nil-hp-%t", weaponKind, uint32(weaponClass), int32(typ), kind, caller, god, nilHP), func(t *testing.T) {
								if god {
									noxflags.SetEngine(noxflags.EngineGodMode)
								} else {
									noxflags.UnsetEngine(noxflags.EngineGodMode)
								}
								var pin runtime.Pinner
								defer pin.Unpin()
								npc, _ := npcReflectLegacyObjects4E17B0(t, &pin)
								var attacker *server.Object
								weapon, _ := npcReflectLegacyObjects4E17B0(t, &pin)
								weapon.ObjClass, weapon.ObjSubClass, weapon.UpdateData, weapon.HealthData = weaponClass, 0x10, nil, nil
								modifiers, free := alloc.New(server.ModifierInitData{})
								*modifiers = server.ModifierInitData{}
								t.Cleanup(free)
								weapon.InitData = unsafe.Pointer(modifiers)
								weapon.TypeInd, weapon.ObjOwner, weapon.Buffs = 73, nil, 0
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

								weapon.PrevPos = types.Ptf(32.5, -18.25)
								target.TeamVal.ID, weapon.TeamVal.ID = 2, 1

								target.Damage = srv.Types.ByID("NativeSourceLessMixedMissile" + callback).Damage
								target.Field38, target.Frame134, target.Obj130, target.Pos132 = 0, 77, nil, types.Ptf(29, 31)
								var carry, marker, kindField *uint32
								if kind == "player" {
									u := target.UpdateDataPlayer()
									u.Player.PlayerUnit = target
									u.Player.PlayerInd = 8
									u.Field40_0, u.Field40_1 = 0xab11, 0xefcd
									u.Field57, u.Field21, u.Field76, u.Field75 = math.Float32bits(.5), math.Float32bits(.25), 99, 77
									carry, marker, kindField = &u.Field21, &u.Field76, &u.Field75
								} else {
									u := target.UpdateDataMonster()
									u.Field518, u.Field1, u.Field547, u.Field546 = math.Float32bits(.5), math.Float32bits(.25), 99, 77
									carry, marker, kindField = &u.Field1, &u.Field547, &u.Field546
								}
								// Original startup: raw9, armor.5, carry.25 gives4 HP and carry.75.
								wantHP, wantCarry := uint16(196), float32(.75)
								physical := typ == object.DamageBlade || typ == object.DamageImpale || typ == object.DamageExplosion || typ == object.DamageBite || typ == object.DamageClaw || typ == object.DamageImpact || typ == object.DamageCrush
								if !physical || kind == "ordinary-monster" {
									wantHP, wantCarry = 191, .25
								}
								if typ == object.DamageCrush && kind != "ordinary-monster" {
									wantHP, wantCarry = 193, 0
								}
								blocked := god && kind == "player"
								if blocked {
									wantHP = 200
								}
								wantMarker, wantKind := uint32(2), uint32(typ)
								if kind == "player" && typ == object.DamageDeathMagic {
									wantMarker, wantKind = 0, 77
								}
								var diagnostics bytes.Buffer
								srv.Log = slog.New(slog.NewTextHandler(&diagnostics, nil))
								if nilHP {
									target.HealthData = nil
								}
								wantSync := uint32(math.MaxUint32)
								if nilHP {
									wantSync = 0
								}
								var result bool
								if caller {
									result = target.CallDamage(nil, weapon, 9, typ)
								} else {
									result = objectDamageDispatchCallNative(target, attacker, weapon, 9, typ)
								}
								hpRead := -1
								hpOK := nilHP && target.HealthData == nil
								if target.HealthData != nil {
									hpRead = int(target.HealthData.Cur)
									hpOK = !nilHP && target.HealthData.Cur == wantHP
								}
								if !result || strings.Contains(diagnostics.String(), "not ported") || !hpOK || *carry != math.Float32bits(wantCarry) || *marker != wantMarker || *kindField != wantKind {
									t.Fatalf("pure missile result=%t HP=%d/%d carry=%x/%x marker=%d/%d log=%s", result, hpRead, wantHP, *carry, math.Float32bits(wantCarry), *marker, *kindField, diagnostics.String())
								}
								if kind == "player" {
									u := target.UpdateDataPlayer()
									wantElectric := uint16(0xab11)
									if !blocked && (typ == object.DamageElectric || typ == object.DamageAirborneElectric) {
										wantElectric = 2
									}
									if u.Field40_0 != wantElectric || u.Field40_1 != 0xefcd {
										t.Fatal("original electric word or adjacent canary changed")
									}
								}
								if blocked {
									if target.Field38 != 0 || target.Frame134 != 77 || target.Obj130 != nil || target.Pos132 != types.Ptf(29, 31) {
										t.Fatal("GodMode crossed HP/attribution boundary")
									}
								} else if target.Field38 != wantSync || target.Frame134 != 1400 || target.Obj130 != weapon || target.Pos132 != (types.Pointf{}) || target.Field131 != uint32(typ) {
									t.Fatal("native player-source attribution/HP sync")
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

func TestPlayerDamageSourceLessMixedMissileImmunityNative4E17B0(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	const id = "NativeSourceLessMixedMissileImmunity"
	if err := srv.Types.ReadObjectType(&things.Thing{Name: id, OnDamage: &things.ProcFunc{Name: "PlayerDamage"}}); err != nil {
		t.Fatal(err)
	}
	callback := srv.Types.ByID(id).Damage
	oldLog := srv.Log
	t.Cleanup(func() { srv.Log = oldLog })
	blob, err := readOriginalNonItemStartupFixture4E17B0("../server/testdata/player_damage_source_less_mixed_missile_integrated_4e17b0.json")
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
			WeaponClass object.Class         `json:"weapon_class"`
			WeaponLatch uint32               `json:"weapon_monster_latch"`
			Return      int                  `json:"return"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(fixture.Rows) != 270 {
		t.Fatal("unexpected original integrated NPC fixture")
	}
	for _, row := range fixture.Rows {
		for _, mode := range []struct{ god, nilHP bool }{{}, {god: true}, {nilHP: true}} {
			t.Run(fmt.Sprintf("class-%x/type-%d/subclass-%x/god-%t/nil-hp-%t", uint32(row.WeaponClass), int32(row.Type), uint32(row.Subclass), mode.god, mode.nilHP), func(t *testing.T) {
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
				var player *server.Object
				weapon, _ := npcReflectLegacyObjects4E17B0(t, &pin)
				weapon.ObjClass, weapon.ObjSubClass, weapon.UpdateData, weapon.HealthData = row.WeaponClass, 0x10, nil, nil
				modifiers, free := alloc.New(server.ModifierInitData{})
				*modifiers = server.ModifierInitData{}
				t.Cleanup(free)
				weapon.InitData = unsafe.Pointer(modifiers)
				weapon.TypeInd, weapon.ObjOwner, weapon.TeamVal.ID, weapon.Buffs = 73, nil, 1, 0
				if mode.nilHP {
					target.HealthData = nil
				}
				target.ObjSubClass = row.Subclass
				target.Damage = callback
				target.Buffs = 0
				if row.SourceKind == "monster" {
					player.UpdateDataMonster().Field130 = 0
				}

				weapon.PrevPos = types.Ptf(32.5, -18.25)
				target.TeamVal.ID = 2

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
					if target.Field38 != wantSync || target.Obj130 != weapon || target.Pos132 != (types.Pointf{}) || target.Frame134 != 1400 || target.Field131 != uint32(row.Type) {
						t.Fatal("monster-source NPC attribution/HP sync")
					}
				}
				runtime.KeepAlive(target)
				runtime.KeepAlive(player)
				runtime.KeepAlive(weapon)
			})
		}
	}
}

// Original DefaultDamage calls BuffOff(INVISIBLE) for LAVA even with a nil
// source and a nonnil missile. Exercise the registered C adapter and real
// buff synchronization, including the independent nil-health path.
func TestPlayerDamageSourceLessMixedMissileLavaInvisibleNative4E17B0(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.Audio.Init(srv)
	t.Cleanup(srv.Audio.Free)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	if !srv.Objs.Init(16) {
		t.Fatal("native object allocator")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	for _, callback := range []string{"PlayerDamage", "DefaultDamage"} {
		if err := srv.Types.ReadObjectType(&things.Thing{Name: "SourceLessMixedLava" + callback, OnDamage: &things.ProcFunc{Name: callback}}); err != nil {
			t.Fatal(err)
		}
	}
	oldLog := srv.Log
	t.Cleanup(func() { srv.Log = oldLog })
	for _, class := range []object.Class{0x1000001, 0x05200001, 0x01280001} {
		for _, npc := range []bool{false, true} {
			for _, nilHP := range []bool{false, true} {
				t.Run(fmt.Sprintf("class-%x/npc-%t/nil-hp-%t", uint32(class), npc, nilHP), func(t *testing.T) {
					var pin runtime.Pinner
					defer pin.Unpin()
					template, _ := npcReflectLegacyObjects4E17B0(t, &pin)
					// Preserve the allocator's private server handle for BuffOff.
					target := srv.Objs.NewObject(&server.ObjectType{})
					target.TypeInd, target.ObjClass, target.ObjSubClass = template.TypeInd, template.ObjClass, 0
					target.UpdateData, target.HealthData = template.UpdateData, template.HealthData
					callback := "DefaultDamage"
					if npc {
						target.ObjSubClass, callback = 0x11012, "PlayerDamage"
					}
					target.Damage = srv.Types.ByID("SourceLessMixedLava" + callback).Damage
					if nilHP {
						target.HealthData = nil
					}
					target.Buffs = 1 << server.ENCHANT_INVISIBLE
					target.BuffsDur[server.ENCHANT_INVISIBLE], target.BuffsPower[server.ENCHANT_INVISIBLE] = 321, 7
					target.Field38, target.Frame134, target.Obj130, target.Pos132 = 0, 77, nil, types.Ptf(29, 31)
					u := target.UpdateDataMonster()
					u.Field547, u.Field546 = 99, 77
					weapon, _ := npcReflectLegacyObjects4E17B0(t, &pin)
					weapon.ObjClass, weapon.ObjSubClass, weapon.UpdateData, weapon.HealthData = class, 0x10, nil, nil
					modifiers, free := alloc.New(server.ModifierInitData{})
					*modifiers = server.ModifierInitData{}
					t.Cleanup(free)
					weapon.InitData = unsafe.Pointer(modifiers)
					weapon.TypeInd, weapon.ObjOwner, weapon.Buffs = 73, nil, 0
					weapon.PrevPos = types.Ptf(32.5, -18.25)
					var diagnostics bytes.Buffer
					srv.Log = slog.New(slog.NewTextHandler(&diagnostics, nil))
					ok := objectDamageDispatchCallNative(target, nil, weapon, 9, object.DamageLava)
					hpOK := target.HealthData == nil
					if !nilHP {
						hpOK = target.HealthData != nil && target.HealthData.Cur == 191
					}
					if !ok || !hpOK || strings.Contains(diagnostics.String(), "not ported") || target.Buffs != 0 || target.BuffsDur[server.ENCHANT_INVISIBLE] != 0 || target.BuffsPower[server.ENCHANT_INVISIBLE] != 0 || target.Field38 != math.MaxUint32 || target.Frame134 != 1400 || target.Obj130 != weapon || target.Pos132 != (types.Pointf{}) || target.Field131 != uint32(object.DamageLava) || u.Field547 != 2 || u.Field546 != uint32(object.DamageLava) {
						t.Fatalf("LAVA result=%t HP-valid=%t buffs=%x sync=%x marker=%d/%d log=%s", ok, hpOK, target.Buffs, target.Field38, u.Field547, u.Field546, diagnostics.String())
					}
					runtime.KeepAlive(target)
					runtime.KeepAlive(weapon)
				})
			}
		}
	}
}
