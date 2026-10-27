package legacy

import (
	"bytes"
	"fmt"
	"log/slog"
	"math"
	"runtime"
	"strings"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
)

// nsObj.Damage delegates to Object.DoDamage, which resolves the initialized
// player as terminal owner and retains that same object as weapon. Exercise
// the same owner-resolution/CallDamage boundary and C dispatcher separately.
// The root DoDamage wrapper itself is not called by this legacy-package test.
func TestPlayerDamagePlayerSelfSourceNativeDamage4E17B0(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.Audio.Init(srv)
	t.Cleanup(srv.Audio.Free)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	for _, name := range []string{"PlayerDamage", "DefaultDamage"} {
		if err := srv.Types.ReadObjectType(&things.Thing{Name: "NativePlayerSelfSource" + name, OnDamage: &things.ProcFunc{Name: name}}); err != nil {
			t.Fatal(err)
		}
	}
	oldLog := srv.Log
	t.Cleanup(func() { srv.Log = oldLog })
	for _, typ := range []object.DamageType{object.DamageBlade, object.DamageImpale, object.DamageDrain, object.DamageBite, object.DamageClaw} {
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
						target.Damage = srv.Types.ByID("NativePlayerSelfSource" + callback).Damage
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
						wantHP, wantCarry := uint16(191), float32(.25)
						if kind != "ordinary-monster" {
							damage, carryBits, _ := nativeSelfStrikeOriginal4E17B0(t, true, kind == "player", typ, 9, .5, .25)
							wantHP, wantCarry = uint16(200-damage), math.Float32frombits(carryBits)
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
