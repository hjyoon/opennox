package server

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"slices"
	"testing"

	"github.com/opennox/libs/object"
)

// The new shape admits stationary dangerous world fire, not new weapon,
// missile or unit damage. Type/name IDs do not determine admission.
func TestWorldFlameFountainShape4E17B0(t *testing.T) {
	bits := []object.Class{object.ClassFire, object.ClassSimple, object.ClassDangerous, object.ClassImmobile}
	for mask := range 16 {
		t.Run(fmt.Sprintf("bits-%x", mask), func(t *testing.T) {
			var class object.Class
			for i, bit := range bits {
				if mask&(1<<i) != 0 {
					class |= bit
				}
			}
			fire := object.ClassFire | object.ClassSimple | object.ClassDangerous
			fountain := object.ClassDangerous | object.ClassImmobile
			want := class&fire == fire || class&fountain == fountain
			w := &Object{ObjClass: class}
			if got := playerDamageWorldFlameShape4E17B0(w, object.DamageFlame); got != want {
				t.Fatalf("class=%x flame admission=%t want=%t", uint32(class), got, want)
			}
		})
	}
	for _, extra := range []object.Class{object.ClassPlayer, object.ClassMonster, object.ClassWeapon, object.ClassWand, object.ClassMissile} {
		t.Run(fmt.Sprintf("reject-%x", uint32(extra)), func(t *testing.T) {
			w := &Object{ObjClass: 4259840 | extra}
			if playerDamageWorldFlameShape4E17B0(w, object.DamageFlame) {
				t.Fatal("fountain admission captured a separately ported damage shape")
			}
		})
	}
	for typ := object.DamageType(0); typ <= object.DamageAirborneElectric; typ++ {
		if typ == object.DamageFlame {
			continue
		}
		t.Run(fmt.Sprintf("reject-type-%d", typ), func(t *testing.T) {
			if playerDamageWorldFlameShape4E17B0(&Object{ObjClass: 4259840}, typ) {
				t.Fatal("fountain admission reinterpreted a non-FLAME type")
			}
		})
	}
	t.Run("nil", func(t *testing.T) {
		if playerDamageWorldFlameShape4E17B0(nil, object.DamageFlame) {
			t.Fatal("nil flame admitted")
		}
	})
	t.Run("name-independent", func(t *testing.T) {
		w := &Object{ObjClass: 4259840 | object.ClassLight | object.ClassVisibleEnable}
		for _, id := range []uint16{0, 1202, 65535} {
			w.TypeInd = id
			if !playerDamageWorldFlameShape4E17B0(w, object.DamageFlame) {
				t.Fatal("stationary fire depended on a name/type ID")
			}
		}
	})
}

type fountainOriginalRow4E0B30 struct {
	Name             string    `json:"name"`
	Type             uint16    `json:"type"`
	Gate             string    `json:"gate"`
	Damage           int32     `json:"damage"`
	Protection       float64   `json:"protection"`
	SubClass         uint32    `json:"subclass"`
	Flags            uint32    `json:"flags"`
	Result           int       `json:"result"`
	Marker           uint32    `json:"marker"`
	MarkerType       uint32    `json:"marker_type"`
	Status           uint32    `json:"status"`
	Buffs            uint32    `json:"buffs"`
	WeaponAttributed int       `json:"weapon_attributed"`
	Position         [2]uint32 `json:"position"`
	DamageType       uint32    `json:"damage_type"`
	Frame            uint32    `json:"frame"`
	State            uint8     `json:"state"`
	Carry            uint32    `json:"carry"`
	HPArgument       uint32    `json:"hp_argument"`
	HPCalls          int       `json:"hp_calls"`
	Events           []string  `json:"events"`
}

// This compares unhooked original field stores/return and declared service
// arguments/order. The HP service here records an argument; actual C-owned
// dispatcher -> registered callback -> UnitSetHP is checked in legacy.
func TestDefaultDamageFountain4E0B30OriginalTranscript(t *testing.T) {
	data, err := os.ReadFile("testdata/default_damage_fountain_4e0b30.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		OriginalSHA256           string                      `json:"original_sha256"`
		FunctionSHA256           string                      `json:"function_sha256"`
		Rows                     []fountainOriginalRow4E0B30 `json:"rows"`
		ExecutableBytesUnchanged bool                        `json:"executable_bytes_unchanged"`
		MemoryFaults             []json.RawMessage           `json:"memory_faults"`
		Entries                  map[string]int              `json:"unhooked_entries"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA256 != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" ||
		fixture.FunctionSHA256 != "6f045c2910bfb5e4a1100b5daaed3aeb5695bb401d3b447c63245c3543e0b871" ||
		len(fixture.Rows) != 144 || len(fixture.MemoryFaults) != 0 || !fixture.ExecutableBytesUnchanged || fixture.Entries["0x4e0b30"] != 144 {
		t.Fatal("original evidence provenance/count changed")
	}
	for _, row := range fixture.Rows {
		// Retain the nearest-control record and project only its HP input
		// through the separately qualified original startup arithmetic.
		if row.HPCalls == 1 && row.Gate != "shield" {
			row.HPArgument = uint32(defaultDamageOriginalProtected4E0B30(t, row.Damage, row.Protection))
		}
		t.Run(row.Name, func(t *testing.T) {
			v, a, w := worldFlameFixture4E17B0(t, "self", "monster")
			if a != w {
				t.Fatal("fountain source/weapon identity")
			}
			v.TypeInd, v.ObjClass, v.ObjSubClass, v.ObjFlags = row.Type, 2621442, object.SubClass(row.SubClass), object.Flags(row.Flags)
			ud := v.UpdateDataMonster()
			ud.Field523_2 = 88
			v.Buffs = 1 | 1<<22
			if row.Gate == "invulnerable" {
				v.Buffs |= 1 << 23
			}
			if row.Gate == "shield" {
				v.Buffs |= 1 << 26
			}
			w.TypeInd, w.ObjClass, w.ObjFlags = 1202, 4259840, 16777748
			beforeW := *w
			got := row
			got.Events, got.HPCalls, got.HPArgument = []string{}, 0, 0xfeedface
			r := DefaultDamageWorldRuntime4E0B30{
				Frame:         func() uint32 { return 1400 },
				GameplayFlag1: func() bool { return true },
				IsEnemy: func(target, source *Object) bool {
					if target != v || source != w {
						t.Fatal("original enemy query identity")
					}
					got.Events = append(got.Events, "enemy")
					if row.Gate == "live-grant" {
						v.ObjSubClass |= 0x400
					} else if row.Gate == "live-revoke" {
						v.ObjSubClass &^= 0x400
					}
					return false
				},
				FireProtection: func(target *Object) float64 {
					if target != v {
						t.Fatal("protection identity")
					}
					got.Events = append(got.Events, "fire")
					return row.Protection
				},
				Audio: func(id int, target *Object) {
					if target != v {
						t.Fatal("audio identity")
					}
					got.Events = append(got.Events, fmt.Sprintf("audio-%d", id))
				},
				BuffOff: func(target *Object, enchant EnchantID) {
					if target != v || enchant != 0 {
						t.Fatal("visibility identity")
					}
					v.Buffs &^= 1
					got.Events = append(got.Events, "buff")
				},
				DefaultDamageSound: func(target, source *Object) {
					if target != v || source != w {
						t.Fatal("sound attribution")
					}
					got.Events = append(got.Events, "sound")
				},
				ShieldReduce: func(target *Object, damage *int32, typ object.DamageType, source *Object) {
					if target != v || source != w || typ != object.DamageFlame {
						t.Fatal("Shield identity/type")
					}
					*damage = 0
					got.Events = append(got.Events, "shield")
				},
				DamageClear: func(target *Object, damage int32) {
					if target != v {
						t.Fatal("HP identity")
					}
					got.HPCalls++
					got.HPArgument = uint32(damage)
					got.Events = append(got.Events, "hp")
				},
				Unsupported: func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(why) },
			}
			got.Result = 0
			if DefaultDamageWorld4E0B30(v, w, w, row.Damage, object.DamageFlame, r) {
				got.Result = 1
			}
			got.Marker, got.MarkerType, got.Status = ud.Field547, ud.Field546, uint32(ud.StatusFlags)
			got.Buffs = v.Buffs
			got.WeaponAttributed = 0
			if v.Obj130 == w {
				got.WeaponAttributed = 1
			}
			got.Position = [2]uint32{math.Float32bits(v.Pos132.X), math.Float32bits(v.Pos132.Y)}
			got.DamageType, got.Frame, got.State, got.Carry = v.Field131, v.Frame134, ud.Field523_2, ud.Field1
			if !reflect.DeepEqual(got, row) || *w != beforeW {
				t.Fatalf("original transcript differs:\n got=%+v\nwant=%+v", got, row)
			}
		})
	}
}

func TestDefaultDamageFountain4E0B30ImmuneNeedsNoHPTail(t *testing.T) {
	for _, live := range []bool{false, true} {
		t.Run(fmt.Sprintf("live-%t", live), func(t *testing.T) {
			v, _, w := worldFlameFixture4E17B0(t, "self", "monster")
			v.ObjClass, v.ObjSubClass, v.ObjFlags = 2621442, 332802, 554172932
			v.HealthData = nil
			w.TypeInd, w.ObjClass, w.ObjFlags = 1202, 4259840, 16777748
			if live {
				v.ObjSubClass &^= 0x400
			}
			var events []string
			r := DefaultDamageWorldRuntime4E0B30{
				IsEnemy: func(*Object, *Object) bool {
					events = append(events, "enemy")
					v.ObjSubClass |= 0x400
					return false
				},
				Unsupported: func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(why) },
			}
			if !DefaultDamageWorld4E0B30(v, w, w, 3, object.DamageFlame, r) ||
				!slices.Equal(events, []string{"enemy"}) || v.UpdateDataMonster().Field547 != 0 || v.Obj130 != nil || v.Frame134 != 77 {
				t.Fatal("immunity must precede health/protection/tail-service admission")
			}
		})
	}
}
