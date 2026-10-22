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

// nsObj.Damage delegates to Object.DoDamage, which resolves the initialized
// player as terminal owner and retains that same object as weapon. Exercise
// the same owner-resolution/CallDamage boundary and C dispatcher separately.
// The root DoDamage wrapper itself is not called by this legacy-package test.
func TestPlayerDamagePlayerSelfElementalNativeDamage4E17B0(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	// Bind the production preflight service. Damage9 or absorbed5 never
	// reaches its hurt-state threshold; HP and damage remain native.
	oldState := Nox_xxx_playerSetState_4FA020
	Nox_xxx_playerSetState_4FA020 = func(*server.Object, server.PlayerState) bool {
		t.Fatal("below-threshold elemental damage requested hurt state")
		return false
	}
	t.Cleanup(func() { Nox_xxx_playerSetState_4FA020 = oldState })
	srv.Audio.Init(srv)
	t.Cleanup(srv.Audio.Free)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	for _, name := range []string{"PlayerDamage", "DefaultDamage"} {
		if err := srv.Types.ReadObjectType(&things.Thing{Name: "NativePlayerSelfElemental" + name, OnDamage: &things.ProcFunc{Name: name}}); err != nil {
			t.Fatal(err)
		}
	}
	oldLog := srv.Log
	t.Cleanup(func() { srv.Log = oldLog })
	for _, typ := range []object.DamageType{object.DamageFlame, object.DamageExplosion, object.DamageImpact, object.DamageLava} {
		for _, kind := range []string{"player", "npc", "ordinary-monster"} {
			for _, caller := range []bool{false, true} {
				for _, god := range []bool{false, true} {
					t.Run(fmt.Sprintf("type-%d/target-%s/caller-%t/god-%t", int32(typ), kind, caller, god), func(t *testing.T) {
						if god {
							noxflags.SetEngine(noxflags.EngineGodMode)
						} else {
							noxflags.UnsetEngine(noxflags.EngineGodMode)
						}
						var pin runtime.Pinner
						defer pin.Unpin()
						npc, attacker := npcReflectLegacyObjects4E17B0(t, &pin)
						_, hero := npcReflectLegacyObjects4E17B0(t, &pin)
						target := npc
						callback := "PlayerDamage"
						if kind == "player" {
							target = hero
						} else if kind == "ordinary-monster" {
							target.ObjSubClass = 0
							callback = "DefaultDamage"
						}
						attacker.UpdateDataPlayer().Player.PlayerUnit = attacker
						attacker.UpdateDataPlayer().Player.PlayerInd = 9
						attacker.PrevPos = types.Ptf(32.5, -18.25)
						attacker.TeamVal.ID, target.TeamVal.ID = 1, 2
						target.Damage = srv.Types.ByID("NativePlayerSelfElemental" + callback).Damage
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
						if typ == object.DamageFlame || typ == object.DamageLava || kind == "ordinary-monster" {
							wantHP, wantCarry = 191, .25
						}
						blocked := god && kind == "player"
						if blocked {
							wantHP = 200
						}
						var diagnostics bytes.Buffer
						srv.Log = slog.New(slog.NewTextHandler(&diagnostics, nil))
						var result bool
						if caller {
							result = target.CallDamage(attacker.FindOwnerChainPlayer(), attacker, 9, typ)
						} else {
							result = objectDamageDispatchCallNative(target, attacker, attacker, 9, typ)
						}
						if !result || strings.Contains(diagnostics.String(), "not ported") || target.HealthData.Cur != wantHP || *carry != math.Float32bits(wantCarry) || *marker != 2 || *kindField != uint32(typ) {
							t.Fatalf("initialized player source result=%t HP=%d/%d carry=%x/%x marker=%d/%d log=%s", result, target.HealthData.Cur, wantHP, *carry, math.Float32bits(wantCarry), *marker, *kindField, diagnostics.String())
						}
						if blocked {
							if target.Field38 != 0 || target.Frame134 != 77 || target.Obj130 != nil || target.Pos132 != types.Ptf(29, 31) {
								t.Fatal("GodMode crossed HP/attribution boundary")
							}
						} else if target.Field38 != math.MaxUint32 || target.Frame134 != 1400 || target.Obj130 != attacker || target.Pos132 != attacker.PrevPos || target.Field131 != uint32(typ) {
							t.Fatal("native player-source attribution/HP sync")
						}
						runtime.KeepAlive(target)
						runtime.KeepAlive(attacker)
					})
				}
			}
		}
	}
}

func TestPlayerDamagePlayerSelfElementalImmunityNative4E17B0(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	const id = "NativePlayerSelfElementalImmunity"
	if err := srv.Types.ReadObjectType(&things.Thing{Name: id, OnDamage: &things.ProcFunc{Name: "PlayerDamage"}}); err != nil {
		t.Fatal(err)
	}
	callback := srv.Types.ByID(id).Damage
	oldLog := srv.Log
	t.Cleanup(func() { srv.Log = oldLog })
	blob, err := os.ReadFile("../server/testdata/player_damage_player_self_elemental_integrated_4e17b0.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		OriginalSHA string `json:"original_sha256"`
		Rows        []struct {
			Type     object.DamageType    `json:"damage_type"`
			Subclass object.SubClass      `json:"subclass"`
			Forward  []uint32             `json:"forward_damage"`
			Carry    uint32               `json:"carry_bits"`
			Marker   uint32               `json:"marker"`
			Kind     uint32               `json:"marker_type"`
			Status   object.MonsterStatus `json:"status_flags"`
			Electric []byte               `json:"electric_state_bytes"`
			Return   int                  `json:"return"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(fixture.Rows) != 20 {
		t.Fatal("unexpected original integrated NPC fixture")
	}
	for _, row := range fixture.Rows {
		for _, mode := range []struct{ god, nilHP bool }{{}, {god: true}, {nilHP: true}} {
			t.Run(fmt.Sprintf("type-%d/subclass-%x/god-%t/nil-hp-%t", int32(row.Type), uint32(row.Subclass), mode.god, mode.nilHP), func(t *testing.T) {
				if mode.god {
					noxflags.SetEngine(noxflags.EngineGodMode)
				} else {
					noxflags.UnsetEngine(noxflags.EngineGodMode)
				}
				var pin runtime.Pinner
				defer pin.Unpin()
				var diagnostics bytes.Buffer
				srv.Log = slog.New(slog.NewTextHandler(&diagnostics, nil))
				target, player := npcReflectLegacyObjects4E17B0(t, &pin)
				if mode.nilHP {
					target.HealthData = nil
				}
				target.ObjSubClass = row.Subclass
				target.Damage = callback
				target.Buffs = 0
				player.UpdateDataPlayer().Player.PlayerUnit = player
				player.PrevPos = types.Ptf(32.5, -18.25)
				target.TeamVal.ID, player.TeamVal.ID = 2, 1
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
				result := objectDamageDispatchCallNative(target, player, player, 9, row.Type)
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
						t.Fatal("player-source NPC attribution/HP sync")
					}
				}
				runtime.KeepAlive(target)
				runtime.KeepAlive(player)
			})
		}
	}
}

// The unchanged original prefix/DefaultDamage reaches the HP boundary even
// with nil player health. UnitDamageClear owns the nil-health guard.
func TestPlayerDamagePlayerSelfElementalPlayerNilHealthNative4E17B0(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	oldState := Nox_xxx_playerSetState_4FA020
	Nox_xxx_playerSetState_4FA020 = func(*server.Object, server.PlayerState) bool {
		t.Fatal("unexpected nil-health hurt state")
		return false
	}
	t.Cleanup(func() { Nox_xxx_playerSetState_4FA020 = oldState })
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	const id = "NativePlayerSelfElementalNilHP"
	if err := srv.Types.ReadObjectType(&things.Thing{Name: id, OnDamage: &things.ProcFunc{Name: "PlayerDamage"}}); err != nil {
		t.Fatal(err)
	}
	blob, err := os.ReadFile("../server/testdata/player_damage_player_self_elemental_player_nil_health_4e17b0.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		OriginalSHA string `json:"original_sha256"`
		Rows        []struct {
			Type   object.DamageType `json:"damage_type"`
			Carry  uint32            `json:"carry_bits"`
			Marker uint32            `json:"marker"`
			Kind   uint32            `json:"marker_type"`
			Return int               `json:"return"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(fixture.Rows) != 4 {
		t.Fatal("unexpected original nil player health fixture")
	}
	noxflags.UnsetEngine(noxflags.EngineGodMode)
	oldLog := srv.Log
	t.Cleanup(func() { srv.Log = oldLog })
	for _, row := range fixture.Rows {
		for _, caller := range []bool{false, true} {
			t.Run(fmt.Sprintf("type-%d/caller-%t", row.Type, caller), func(t *testing.T) {
				var pin runtime.Pinner
				defer pin.Unpin()
				_, player := npcReflectLegacyObjects4E17B0(t, &pin)
				_, target := npcReflectLegacyObjects4E17B0(t, &pin)
				target.HealthData = nil
				target.Damage = srv.Types.ByID(id).Damage
				target.Buffs = 0
				target.UpdateDataPlayer().Player.PlayerUnit = target
				player.UpdateDataPlayer().Player.PlayerUnit = player
				player.PrevPos = types.Ptf(32.5, -18.25)
				player.TeamVal.ID, target.TeamVal.ID = 1, 2
				target.Field38, target.Frame134, target.Obj130, target.Pos132 = 0, 77, nil, types.Ptf(29, 31)
				ud := target.UpdateDataPlayer()
				ud.Field57, ud.Field21, ud.Field76, ud.Field75 = math.Float32bits(.5), math.Float32bits(.25), 99, 77
				var diagnostics bytes.Buffer
				srv.Log = slog.New(slog.NewTextHandler(&diagnostics, nil))
				var result bool
				if caller {
					result = target.CallDamage(player.FindOwnerChainPlayer(), player, 9, row.Type)
				} else {
					result = objectDamageDispatchCallNative(target, player, player, 9, row.Type)
				}
				if result != (row.Return != 0) || strings.Contains(diagnostics.String(), "not ported") || target.HealthData != nil || ud.Field21 != row.Carry || ud.Field76 != row.Marker || ud.Field75 != row.Kind {
					t.Fatalf("nil-health result=%t carry=%x/%x marker=%d/%d kind=%d/%d log=%s", result, ud.Field21, row.Carry, ud.Field76, row.Marker, ud.Field75, row.Kind, diagnostics.String())
				}
				if target.Field38 != 0 || target.Frame134 != 1400 || target.Obj130 != player || target.Pos132 != player.PrevPos || target.Field131 != uint32(row.Type) {
					t.Fatal("nil-health attribution/HP sync")
				}
				runtime.KeepAlive(target)
				runtime.KeepAlive(player)
			})
		}
	}
}
