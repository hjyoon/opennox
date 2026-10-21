package server

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type selfMissileFlameOriginalRow4E17B0 struct {
	Name                     string   `json:"name"`
	Target                   string   `json:"target"`
	Type                     uint16   `json:"type"`
	ClassBits                uint32   `json:"class_bits"`
	SubClass                 uint32   `json:"subclass"`
	Gate                     string   `json:"gate"`
	Damage                   int32    `json:"damage"`
	Scale                    *float32 `json:"scale"`
	Flags                    uint32   `json:"flags"`
	Buffs                    uint32   `json:"buffs"`
	Frame                    uint32   `json:"frame"`
	Result                   int      `json:"result"`
	Marker                   uint32   `json:"marker"`
	MarkerType               uint32   `json:"marker_type"`
	Carry                    uint32   `json:"carry"`
	Armor                    uint32   `json:"armor"`
	State                    uint8    `json:"state"`
	ForwardDamage            uint32   `json:"forward_damage"`
	ForwardCalls             int      `json:"forward_calls"`
	WearBits                 uint32   `json:"wear_bits"`
	BlockCalls               int      `json:"block_calls"`
	Events                   []string `json:"events"`
	SourceAndGuardsUnchanged bool     `json:"source_and_guards_unchanged"`
}

// The declared DefaultDamage boundary records its input, not HP behavior.
// The legacy test exercises C dispatch and production HP without substitution.
func TestSelfMissileFlamePlayerPrefixMatchesGAMEEXE4E17B0(t *testing.T) {
	data, err := os.ReadFile("testdata/player_damage_self_missile_flame_4e17b0.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		OriginalSHA256 string `json:"original_sha256"`
		FunctionSHA256 string `json:"function_sha256"`
		Source         struct {
			Type      uint16 `json:"type"`
			ClassBits uint32 `json:"class_bits"`
			SubClass  uint32 `json:"subclass"`
			Flags     uint32 `json:"flags"`
		} `json:"source"`
		Rows                     []selfMissileFlameOriginalRow4E17B0 `json:"rows"`
		Entries                  map[string]int                      `json:"unhooked_entries"`
		Ranges                   []json.RawMessage                   `json:"original_ranges"`
		MemoryFaults             []json.RawMessage                   `json:"memory_faults"`
		ExecutableBytesUnchanged bool                                `json:"executable_bytes_unchanged"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA256 != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" ||
		fixture.FunctionSHA256 != "c3e71619fd8d5e8c0aff27b5d098db02ee5bed0c6827f495ce6cf54326062ed9" ||
		len(fixture.Rows) != 228 || len(fixture.Ranges) != 6 || len(fixture.MemoryFaults) != 0 ||
		fixture.Entries["0x4e17b0"] != 228 || !fixture.ExecutableBytesUnchanged ||
		fixture.Source.Type != 694 || fixture.Source.ClassBits != 2621441 || fixture.Source.SubClass != 1 || fixture.Source.Flags != 553665028 {
		t.Fatal("original self-sourced missile FLAME provenance/count changed")
	}
	for _, row := range fixture.Rows {
		t.Run(row.Name, func(t *testing.T) {
			player := row.Target == "NewPlayer"
			victim := "NPC"
			if player {
				victim = "player"
			}
			v, a, w := worldFlameFixture4E17B0(t, "self", victim)
			if a != w {
				t.Fatal("reported source and weapon must be identical")
			}
			v.TypeInd, v.ObjClass, v.ObjSubClass, v.ObjFlags = row.Type, object.Class(row.ClassBits), object.SubClass(row.SubClass), object.Flags(row.Flags)
			v.Buffs = row.Buffs
			w.TypeInd, w.ObjClass, w.ObjSubClass, w.ObjFlags = fixture.Source.Type, object.Class(fixture.Source.ClassBits), object.SubClass(fixture.Source.SubClass), object.Flags(fixture.Source.Flags)
			w.PrevPos, w.PosVec = types.Ptf(17, 19), types.Ptf(23, 29)
			var marker, kind, carry, armor *uint32
			shieldGate := row.Gate == "shield-front" || row.Gate == "shield-back"
			if player {
				u := v.UpdateDataPlayer()
				u.Field76, u.Field75, u.Field21, u.Field57 = 99, 91, math.Float32bits(.25), math.Float32bits(.5)
				marker, kind, carry, armor = &u.Field76, &u.Field75, &u.Field21, &u.Field57
				if shieldGate {
					u.State, u.Player.ArmorEquip = PlayerState16, 0x1000000
				}
				if row.Gate == "observer-disabled" {
					u.Player.Field3680 = 1
				}
			} else {
				u := v.UpdateDataMonster()
				u.Field547, u.Field546, u.Field1, u.Field518 = 99, 91, math.Float32bits(.25), math.Float32bits(.5)
				marker, kind, carry, armor = &u.Field547, &u.Field546, &u.Field1, &u.Field518
				if shieldGate {
					u.ArmorEquipFlags, u.AIStack[0].Action = 0x1000000, 21
				}
			}
			var shield *Object
			if shieldGate {
				item, free := alloc.New(Object{})
				t.Cleanup(free)
				*item = Object{ObjClass: object.ClassArmor, ObjSubClass: 2, ObjFlags: object.FlagEquipped}
				v.InvFirstItem, shield = item, item
			}
			beforeV, beforeW, beforeH := *v, *w, *v.HealthData
			var beforeP PlayerUpdateData
			var beforeInfo Player
			var beforeM MonsterUpdateData
			if player {
				beforeP, beforeInfo = *v.UpdateDataPlayer(), *v.UpdateDataPlayer().Player
			} else {
				beforeM = *v.UpdateDataMonster()
			}
			got := row
			got.Result, got.ForwardDamage, got.ForwardCalls, got.WearBits, got.BlockCalls = 0, 0xfeedface, 0, 0, 0
			got.Events = []string{}
			r := PlayerDamageRuntime4E17B0{
				Frame:    func() uint32 { return row.Frame },
				CoopMode: func() bool { got.Events = append(got.Events, "coop"); return false },
				Audio: func(id int, target *Object) {
					if target != v || (id != 71 && id != 122 && id != 878) {
						t.Fatal("original audio identity")
					}
					got.Events = append(got.Events, fmt.Sprintf("audio-%d", id))
				},
				BlockSourceExcluded: func(effective *Object) bool {
					if effective != w {
						t.Fatal("exclusion identity")
					}
					got.Events = append(got.Events, "exclude")
					return false
				},
				BlockDirection: func(target *Object, pos types.Pointf) bool {
					wantPos := w.PrevPos
					if row.Gate == "reflect" {
						wantPos = w.PosVec
					}
					if target != v || pos != wantPos {
						t.Fatal("original facing identity")
					}
					got.Events = append(got.Events, "face")
					return row.Gate == "shield-front" || row.Gate == "reflect"
				},
				BlockDamagePercent: func() float64 { got.Events = append(got.Events, "balance"); return .5 },
				CanDamageBlockItem: func(item *Object) bool { return item == shield },
				DamageBlockItem: func(item, target, source, effective *Object, amount float32, typ object.DamageType) bool {
					if item != shield || item == nil || target != v || source != w || effective != w || typ != object.DamageFlame {
						t.Fatal("equipped original shield identity")
					}
					got.Events = append(got.Events, "shield")
					got.BlockCalls++
					got.WearBits = math.Float32bits(amount)
					return true
				},
				// As in the PE record, the shield has no health. ItemArmorValue
				// is a declared zero input; EquipDamage owns the nil-health gate.
				ItemArmorValue: func(item *Object) float32 {
					if item != shield {
						t.Fatal("unexpected armor lookup")
					}
					return 0
				},
				GodMode:   func() bool { got.Events = append(got.Events, "god"); return row.Gate == "god" },
				QuestMode: func() bool { got.Events = append(got.Events, "quest"); return row.Scale != nil },
				QuestDamageScale: func() float32 {
					if row.Scale == nil {
						t.Fatal("unexpected Quest scale")
					}
					got.Events = append(got.Events, "scale")
					return *row.Scale
				},
				DefaultDamage: func(target, source, effective *Object, damage int32, typ object.DamageType) bool {
					if target != v || source != w || effective != w || typ != object.DamageFlame {
						t.Fatal("declared DefaultDamage identity")
					}
					got.Events = append(got.Events, "default")
					got.ForwardDamage, got.ForwardCalls = uint32(damage), got.ForwardCalls+1
					return true
				},
				ProjectileReflect: func(effective, target *Object) {
					if effective != w || target != v {
						t.Fatal("reflection identity")
					}
					got.Events = append(got.Events, "reflect")
				},
				ClearOwner: func(effective *Object) {
					if effective != w {
						t.Fatal("clear-owner identity")
					}
					got.Events = append(got.Events, "clear-owner")
				},
				SetOwner: func(target, effective *Object) {
					if target != v || effective != w {
						t.Fatal("set-owner identity")
					}
					got.Events = append(got.Events, "set-owner")
				},
				ElectricArmorScale: func(*Object) float32 { t.Fatal("FLAME entered electric absorption"); return 0 },
				Unsupported: func(why string, _, _, _ *Object, _ int32, _ object.DamageType) {
					t.Fatal("self-sourced missile FLAME: " + why)
				},
			}
			handled, result := PlayerDamageNative4E17B0(v, w, w, row.Damage, object.DamageFlame, r)
			if result {
				got.Result = 1
			}
			got.Marker, got.MarkerType, got.Carry, got.Armor = *marker, *kind, *carry, *armor
			got.State = 0
			if player {
				got.State = uint8(v.UpdateDataPlayer().State)
			}
			got.SourceAndGuardsUnchanged = *v == beforeV && *w == beforeW && *v.HealthData == beforeH
			if player {
				after := *v.UpdateDataPlayer()
				after.Field76, after.Field75 = beforeP.Field76, beforeP.Field75
				got.SourceAndGuardsUnchanged = got.SourceAndGuardsUnchanged && reflect.DeepEqual(after, beforeP) && reflect.DeepEqual(*after.Player, beforeInfo)
			} else {
				after := *v.UpdateDataMonster()
				after.Field547, after.Field546 = beforeM.Field547, beforeM.Field546
				got.SourceAndGuardsUnchanged = got.SourceAndGuardsUnchanged && reflect.DeepEqual(after, beforeM)
			}
			if !handled || !reflect.DeepEqual(got, row) {
				t.Fatalf("self-sourced missile FLAME prefix differs:\n got=%+v\nwant=%+v handled=%t", got, row, handled)
			}
		})
	}
	t.Log("SELF-SOURCED MISSILE FLAME PREFIX: 228 unchanged PE32 inputs including reported96; class2621441 source=weapon; signed FLAME/Quest/reflect/shield/immunity gates; C-owned >4 GiB native records; HP boundary is argument-only")
}
