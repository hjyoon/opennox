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

// Registered C dispatch and Object.CallDamage retain initialized weapon-class missiles,
// terminal owner identity and native HP. The root wrapper is not called.
func TestPlayerDamageInitializedWeaponNativeDamage4E17B0(t *testing.T) {
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
		if err := srv.Types.ReadObjectType(&things.Thing{Name: "NativeInitializedWeapon" + name, OnDamage: &things.ProcFunc{Name: name}}); err != nil {
			t.Fatal(err)
		}
	}
	oldLog := srv.Log
	t.Cleanup(func() { srv.Log = oldLog })
	for _, sourceKind := range []string{"player", "monster", "world", "zero", "terminal"} {
		for _, weaponSubclass := range []object.SubClass{0x1000000, 0x4000, 0x10} {
			weaponKind := fmt.Sprintf("missile-%x", uint32(weaponSubclass))
			for _, typ := range []object.DamageType{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17} {
				for _, kind := range []string{"player", "npc", "ordinary-monster"} {
					for _, entry := range []struct{ caller, nilHP bool }{{}, {caller: true}, {nilHP: true}, {caller: true, nilHP: true}} {
						caller, nilHP := entry.caller, entry.nilHP
						for _, god := range []bool{false, true} {
							t.Run(fmt.Sprintf("source-%s/weapon-%s/type-%d/target-%s/caller-%t/god-%t/nil-hp-%t", sourceKind, weaponKind, int32(typ), kind, caller, god, nilHP), func(t *testing.T) {
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
								weapon.ObjClass, weapon.ObjSubClass, weapon.UpdateData, weapon.HealthData = object.ClassWeapon, weaponSubclass, nil, nil
								modifiers, free := alloc.New(server.ModifierInitData{})
								*modifiers = server.ModifierInitData{}
								t.Cleanup(free)
								weapon.InitData = unsafe.Pointer(modifiers)
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
								if sourceKind == "terminal" {
									attacker = weapon
									weapon.ObjOwner = nil
								}
								target.Damage = srv.Types.ByID("NativeInitializedWeapon" + callback).Damage
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
								wantMarker, wantKind := uint32(1), uint32(73)
								if sourceKind == "terminal" {
									wantMarker, wantKind = 2, uint32(typ)
									if kind == "player" && typ == object.DamageDeathMagic {
										wantMarker, wantKind = 0, 77
									}
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
									result = target.CallDamage(weapon.FindOwnerChainPlayer(), weapon, 9, typ)
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
										t.Fatal("original electric lowword/adjacent canary")
									}
								}
								if blocked {
									if target.Field38 != 0 || target.Frame134 != 77 || target.Obj130 != nil || target.Pos132 != types.Ptf(29, 31) {
										t.Fatal("GodMode crossed HP/attribution boundary")
									}
								} else if target.Field38 != wantSync || target.Frame134 != 1400 || target.Obj130 != weapon || target.Pos132 != attacker.PrevPos || target.Field131 != uint32(typ) {
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

func TestPlayerDamageInitializedWeaponImmunityNative4E17B0(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	const id = "NativeInitializedWeaponImmunity"
	if err := srv.Types.ReadObjectType(&things.Thing{Name: id, OnDamage: &things.ProcFunc{Name: "PlayerDamage"}}); err != nil {
		t.Fatal(err)
	}
	callback := srv.Types.ByID(id).Damage
	oldLog := srv.Log
	t.Cleanup(func() { srv.Log = oldLog })
	blob, err := readOriginalNonItemStartupFixture4E17B0("../server/testdata/player_damage_initialized_weapon_integrated_4e17b0.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		OriginalSHA string `json:"original_sha256"`
		Rows        []struct {
			Type           object.DamageType    `json:"damage_type"`
			Subclass       object.SubClass      `json:"subclass"`
			Forward        []uint32             `json:"forward_damage"`
			Carry          uint32               `json:"carry_bits"`
			Marker         uint32               `json:"marker"`
			Kind           uint32               `json:"marker_type"`
			Status         object.MonsterStatus `json:"status_flags"`
			Electric       []byte               `json:"electric_state_bytes"`
			SourceLatch    uint32               `json:"source_monster_latch"`
			SourceKind     string               `json:"source_kind"`
			WeaponKind     string               `json:"weapon_kind"`
			WeaponSubclass object.SubClass      `json:"weapon_subclass"`
			Position       []float32            `json:"position"`
			WeaponLatch    uint32               `json:"weapon_monster_latch"`
			Return         int                  `json:"return"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(fixture.Rows) != 1620 {
		t.Fatal("unexpected original integrated NPC fixture")
	}
	for _, row := range fixture.Rows {
		for _, mode := range []struct{ god, nilHP bool }{{}, {god: true}, {nilHP: true}} {
			t.Run(fmt.Sprintf("source-%s/weapon-%s/weapon-subclass-%x/type-%d/subclass-%x/god-%t/nil-hp-%t", row.SourceKind, row.WeaponKind, uint32(row.WeaponSubclass), int32(row.Type), uint32(row.Subclass), mode.god, mode.nilHP), func(t *testing.T) {
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
				weapon.ObjClass, weapon.ObjSubClass, weapon.UpdateData, weapon.HealthData = object.ClassWeapon, row.WeaponSubclass, nil, nil
				modifiers, free := alloc.New(server.ModifierInitData{})
				*modifiers = server.ModifierInitData{}
				t.Cleanup(free)
				weapon.InitData = unsafe.Pointer(modifiers)
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
				player.PrevPos = types.Ptf(-33, -44)
				weapon.PrevPos = types.Ptf(32.5, -18.25)
				target.TeamVal.ID, player.TeamVal.ID = 2, 1
				if row.SourceKind == "terminal" {
					player = weapon
					weapon.ObjOwner = nil
				}
				if row.SourceKind == "none" {
					player = nil
					weapon.ObjOwner = nil
				}
				target.Obj130, target.Pos132, target.Frame134 = player, types.Ptf(29, 31), 77
				ud := target.UpdateDataMonster()
				ud.Field518, ud.Field1 = math.Float32bits(.5), math.Float32bits(.25)
				ud.Field523_2, ud.Field523_3 = 0x11, 0xab
				if len(row.Forward) > 1 || len(row.Electric) != 2 || len(row.Position) != 2 {
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
					if target.Field38 != wantSync || target.Obj130 != weapon || target.Pos132 != types.Ptf(row.Position[0], row.Position[1]) || target.Frame134 != 1400 || target.Field131 != uint32(row.Type) {
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

func TestDefaultDamageInitializedWeaponPoisonShockNativeHP4E0B30(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.Audio.Init(srv)
	t.Cleanup(srv.Audio.Free)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	if !srv.Objs.Init(48) {
		t.Fatal("native object allocator")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	const id = "NativeInitializedWeaponShockAttacker"
	if err := srv.Types.ReadObjectType(&things.Thing{Name: id, OnDamage: &things.ProcFunc{Name: "PlayerDamage"}}); err != nil {
		t.Fatal(err)
	}
	oldLog := srv.Log
	t.Cleanup(func() { srv.Log = oldLog })
	oldState := Nox_xxx_playerSetState_4FA020
	t.Cleanup(func() { Nox_xxx_playerSetState_4FA020 = oldState })
	for _, weaponSubclass := range []object.SubClass{0x1000000, 0x4000, 0x10} {
		for _, sourcePlayer := range []bool{false, true} {
			for _, shock := range []bool{false, true} {
				t.Run(fmt.Sprintf("weapon-subclass-%x/source-player-%t/shock-%t", uint32(weaponSubclass), sourcePlayer, shock), func(t *testing.T) {
					noxflags.UnsetEngine(noxflags.EngineGodMode)
					var pin runtime.Pinner
					defer pin.Unpin()
					target, _ := npcReflectLegacyObjects4E17B0(t, &pin)
					actor, hero := npcReflectLegacyObjects4E17B0(t, &pin)
					// Buff removal uses object synchronization through the owning server.
					// Keep the real allocator's private server handle and reuse pinned data.
					bind := func(template *server.Object) *server.Object {
						obj := srv.Objs.NewObject(&server.ObjectType{})
						obj.TypeInd, obj.ObjClass, obj.ObjSubClass, obj.ObjFlags = template.TypeInd, template.ObjClass, template.ObjSubClass, template.ObjFlags
						obj.UpdateData, obj.HealthData = template.UpdateData, template.HealthData
						obj.PosVec, obj.PrevPos = template.PosVec, template.PrevPos
						return obj
					}
					target, actor, hero = bind(target), bind(actor), bind(hero)
					mods, free := alloc.New(server.ModifierInitData{})
					*mods = server.ModifierInitData{}
					t.Cleanup(free)
					weapon := &server.Object{ObjClass: object.ClassWeapon, ObjSubClass: weaponSubclass, TypeInd: 73, InitData: unsafe.Pointer(mods), ObjOwner: actor}
					if sourcePlayer {
						actor = hero
						actor.UpdateDataPlayer().Player.PlayerUnit = actor
						weapon.ObjOwner = actor
					}
					actor.Damage = srv.Types.ByID(id).Damage
					target.Buffs, actor.Buffs, weapon.Buffs = 0, 0, 0
					actor.TeamVal.ID, target.TeamVal.ID, weapon.TeamVal.ID = 1, 2, 1
					target.ObjSubClass = 0x210
					target.Field38, target.Frame134, target.Obj130, target.Pos132 = 0, 77, nil, types.Ptf(29, 31)
					u := target.UpdateDataMonster()
					u.Field547, u.Field546 = 99, 77
					if shock {
						target.Buffs |= 1 << server.ENCHANT_SHOCK
					}
					states := 0
					Nox_xxx_playerSetState_4FA020 = func(unit *server.Object, state server.PlayerState) bool {
						if !sourcePlayer || unit != actor || state != server.PlayerState23 {
							t.Fatal("unexpected Shock/hurt state")
						}
						states++
						unit.UpdateDataPlayer().State = state
						return true
					}
					var diagnostics bytes.Buffer
					srv.Log = slog.New(slog.NewTextHandler(&diagnostics, nil))
					r := defaultDamageWorldRuntime4E0B30(srv)
					r.GameplayFlag1 = func() bool { return true }
					r.BalanceFloatInd = func(key string, index int) float64 {
						if key != "ShockDamage" || index != 4 {
							t.Fatal("Shock balance boundary")
						}
						return 9
					}
					ok := server.DefaultDamageWorld4E0B30(target, actor, weapon, 9, object.DamagePoison, r)
					wantHP := uint16(200)
					wantStates := 0
					wantSync := uint32(0)
					if shock && weaponSubclass != 0x10 {
						wantHP = 191
						// Actual BuffOff calls SetBuffFlags/NeedSync, independent of HP.
						wantSync = math.MaxUint32
						if sourcePlayer {
							wantStates = 1
						}
					}
					if !ok || strings.Contains(diagnostics.String(), "not ported") || target.HealthData.Cur != 200 || actor.HealthData.Cur != wantHP || states != wantStates || target.HasEnchant(server.ENCHANT_SHOCK) != (shock && weaponSubclass == 0x10) || u.Field547 != 0 || u.Field546 != 77 || u.StatusFlags != 0 || target.Field38 != wantSync || target.Frame134 != 77 || target.Obj130 != nil || target.Pos132 != types.Ptf(29, 31) {
						t.Fatalf("native Shock/immunity result=%t source HP=%d/%d victim HP=%d marker=%d/%d states=%d/%d sync=%x/%x log=%s", ok, actor.HealthData.Cur, wantHP, target.HealthData.Cur, u.Field547, u.Field546, states, wantStates, target.Field38, wantSync, diagnostics.String())
					}
					runtime.KeepAlive(target)
					runtime.KeepAlive(actor)
					runtime.KeepAlive(weapon)
				})
			}
		}
	}
}
