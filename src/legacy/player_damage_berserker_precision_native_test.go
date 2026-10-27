package legacy

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

// The original observations stop at armor wear and DefaultDamage. This test
// separately executes registered C PlayerDamage, real armor lookup/wear,
// ArmorDamage, DefaultDamage and HP updates without replacing damage or HP.
func TestPlayerDamageBerserkerNative4E17B0ProductionHP(t *testing.T) {
	testPhysicalDamageNative4E17B0ProductionHP(t, false)
}

func TestPlayerDamageTrollImpactNative4E17B0ProductionHP(t *testing.T) {
	testPhysicalDamageNative4E17B0ProductionHP(t, true)
}

func testPhysicalDamageNative4E17B0ProductionHP(t *testing.T, troll bool) {
	path := "../server/testdata/berserker_crush_startup_original.json"
	typ := object.DamageCrush
	if troll {
		path = "../server/testdata/troll_impact_startup_original.json"
		typ = object.DamageImpact
	}
	blob, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SHA  string `json:"original_sha256"`
		Rows []struct {
			Input struct {
				Player bool   `json:"player"`
				Raw    int32  `json:"raw"`
				Armor  uint32 `json:"absorption_bits32"`
				Carry  uint32 `json:"carry_bits32"`
				Quest  bool   `json:"quest"`
			} `json:"input"`
			Events []struct {
				Service string `json:"service"`
				Damage  uint32 `json:"damage_bits32"`
			} `json:"events"`
			Carry uint32 `json:"carry_bits32"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(fixture.Rows) != 60 {
		t.Fatal("original charge fixture unavailable")
	}
	srv := npcReflectLegacyServer4E17B0(t)
	srv.Audio.Init(srv)
	t.Cleanup(srv.Audio.Free)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	for _, proc := range []string{"PlayerDamage", "ArmorDamage"} {
		if err := srv.Types.ReadObjectType(&things.Thing{Name: "NativeCharge" + proc, OnDamage: &things.ProcFunc{Name: proc}}); err != nil {
			t.Fatal(err)
		}
	}
	srv.NetSendPacketXxx = func(int, []byte, *server.Object, int, int) int { return 1 }
	// The client-owned sound handler is installed by the root package in a
	// real game. This legacy-only fixture observes that non-damage boundary.
	oldSound := Nox_xxx_soundPlayerDamageSound_5328B0
	var soundTarget, soundSource *server.Object
	Nox_xxx_soundPlayerDamageSound_5328B0 = func(target, source *server.Object) int {
		if target != soundTarget || source != soundSource {
			t.Fatal("native physical sound identity")
		}
		return 1
	}
	t.Cleanup(func() { Nox_xxx_soundPlayerDamageSound_5328B0 = oldSound })
	oldSetState := Nox_xxx_playerSetState_4FA020
	Nox_xxx_playerSetState_4FA020 = func(*server.Object, server.PlayerState) bool {
		t.Fatal("below-threshold physical hit requested hurt state")
		return false
	}
	t.Cleanup(func() { Nox_xxx_playerSetState_4FA020 = oldSetState })
	definition := selfWeaponElectricNativeAlloc(t, server.Modifier{TypeInd: 301, DamageCoeffOrArmor64: .5})
	old := srv.Modif.Dword_5d4594_251608
	srv.Modif.Dword_5d4594_251608 = definition
	t.Cleanup(func() { srv.Modif.Dword_5d4594_251608 = old })
	for _, playerTarget := range []bool{false, true} {
		t.Run(fmt.Sprintf("player-%t", playerTarget), func(t *testing.T) {
			makeUnit := func(player bool, index byte) *server.Object {
				hp := selfWeaponElectricNativeAlloc(t, server.HealthData{Cur: 200, Max: 200, Field2: 200})
				u := selfWeaponElectricNativeAlloc(t, server.Object{TypeInd: 1343, ObjClass: object.ClassMonster, ObjSubClass: 0x11012, HealthData: hp, Material: 0x4000})
				if player {
					info := selfWeaponElectricNativeAlloc(t, server.Player{PlayerUnit: u, PlayerInd: index})
					ud := selfWeaponElectricNativeAlloc(t, server.PlayerUpdateData{Player: info, State: server.PlayerState13})
					u.TypeInd, u.ObjClass, u.ObjSubClass, u.UpdateData = 713, object.ClassPlayer, 0, unsafe.Pointer(ud)
				} else {
					u.UpdateData = unsafe.Pointer(selfWeaponElectricNativeAlloc(t, server.MonsterUpdateData{}))
				}
				return u
			}
			target, source := makeUnit(playerTarget, 8), makeUnit(!troll, 9)
			soundTarget, soundSource = target, source
			target.TeamVal.ID, source.TeamVal.ID = 2, 1
			source.PrevPos = types.Ptf(12.5, -40.25)
			target.Damage = srv.Types.ByID("NativeChargePlayerDamage").Damage
			var carry *uint32
			if playerTarget {
				ud := target.UpdateDataPlayer()
				ud.Field21, ud.Field57 = math.Float32bits(.25), math.Float32bits(.5)
				carry = &ud.Field21
			} else {
				ud := target.UpdateDataMonster()
				ud.Field1, ud.Field518 = math.Float32bits(.25), math.Float32bits(.5)
				carry = &ud.Field1
			}
			itemHP := selfWeaponElectricNativeAlloc(t, server.HealthData{Cur: 100, Max: 100, Field2: 100})
			itemCarry := selfWeaponElectricNativeAlloc(t, server.WeaponArmorUpdateData{})
			init := selfWeaponElectricNativeAlloc(t, server.ModifierInitData{})
			armor := selfWeaponElectricNativeAlloc(t, server.Object{TypeInd: 301, ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped | object.FlagEnabled, InvHolder: target, HealthData: itemHP, UpdateData: unsafe.Pointer(itemCarry), InitData: unsafe.Pointer(init), Damage: srv.Types.ByID("NativeChargeArmorDamage").Damage})
			target.InvFirstItem = armor
			if unsafe.Sizeof(uintptr(0)) > 4 && (uintptr(target.CObj()) <= math.MaxUint32 || uintptr(source.CObj()) <= math.MaxUint32 || uintptr(armor.CObj()) <= math.MaxUint32) {
				t.Fatal("expected C-owned records above 4 GiB")
			}
			for hit := 0; hit < 3; hit++ {
				var expectedDamage, expectedArmor int32
				var nextCarry uint32
				found := false
				for _, row := range fixture.Rows {
					if row.Input.Player == playerTarget && row.Input.Raw == 3 && row.Input.Armor == math.Float32bits(.5) && row.Input.Carry == *carry && !row.Input.Quest {
						if len(row.Events) != 2 || row.Events[0].Service != "armor" || row.Events[1].Service != "default" {
							t.Fatal("original services missing")
						}
						expectedArmor, expectedDamage, nextCarry = int32(row.Events[0].Damage), int32(row.Events[1].Damage), row.Carry
						found = true
						break
					}
				}
				if !found {
					t.Fatal("missing original repeated-hit input")
				}
				beforeHP, beforeItem := target.HealthData.Cur, itemHP.Cur
				ok := objectDamageDispatchCallNative(target, source, source, 3, typ)
				if !ok || target.HealthData.Cur != beforeHP-uint16(expectedDamage) || itemHP.Cur != beforeItem-uint16(expectedArmor) || *carry != nextCarry || itemCarry.Field0 != 0 {
					t.Fatalf("native charge hit=%d result=%t HP=%d want=%d item HP=%d want=%d carry=%08x want=%08x", hit, ok, target.HealthData.Cur, beforeHP-uint16(expectedDamage), itemHP.Cur, beforeItem-uint16(expectedArmor), *carry, nextCarry)
				}
				t.Logf("actual C physical hit=%d HP=%d armor=%d carry=%08x", hit, target.HealthData.Cur, itemHP.Cur, *carry)
			}
		})
	}
}
