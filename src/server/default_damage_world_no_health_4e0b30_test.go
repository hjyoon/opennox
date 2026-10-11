package server

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type worldNoHealthOriginalRow4E0B30 struct {
	Name         string    `json:"name"`
	Type         uint16    `json:"type"`
	Class        uint32    `json:"class_bits"`
	Flags        uint32    `json:"flags"`
	Health       string    `json:"health"`
	Weapon       bool      `json:"weapon"`
	Damage       int32     `json:"damage"`
	InitialBuffs uint32    `json:"initial_buffs"`
	Result       int       `json:"result"`
	Buffs        uint32    `json:"buffs"`
	Attributed   bool      `json:"attributed"`
	DamageType   uint32    `json:"damage_type"`
	Frame        uint32    `json:"frame"`
	Position     [2]uint32 `json:"position_bits"`
	Events       []string  `json:"events"`
	HealthSafe   bool      `json:"health_unchanged"`
	GuardsSafe   bool      `json:"source_and_guards_unchanged"`
}

func TestDefaultDamageWorldNoHealthMatchesGAMEEXE4E0B30(t *testing.T) {
	data, err := os.ReadFile("testdata/default_damage_world_no_health_4e0b30.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle struct {
		Original string                           `json:"original_sha256"`
		Function string                           `json:"function_sha256"`
		Health   string                           `json:"health_gate_sha256"`
		Safe     bool                             `json:"executable_bytes_unchanged"`
		Faults   []json.RawMessage                `json:"memory_faults"`
		Rows     []worldNoHealthOriginalRow4E0B30 `json:"rows"`
	}
	if err := json.Unmarshal(data, &oracle); err != nil {
		t.Fatal(err)
	}
	if oracle.Original != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" ||
		oracle.Function != "6f045c2910bfb5e4a1100b5daaed3aeb5695bb401d3b447c63245c3543e0b871" ||
		oracle.Health != "2aa2fdff219fdcd15ffdaaf41783e5b00e412ad76fa4a7fdf74baf4449f8fbdf" ||
		!oracle.Safe || len(oracle.Faults) != 0 || len(oracle.Rows) != 160 {
		t.Fatal("unchanged PE32/no-health oracle metadata")
	}
	for _, row := range oracle.Rows {
		t.Run(row.Name, func(t *testing.T) {
			v, freeV := alloc.New(Object{})
			a, freeA := alloc.New(Object{})
			h, freeH := alloc.New(HealthData{})
			defer freeV()
			defer freeA()
			defer freeH()
			*v = Object{TypeInd: row.Type, ObjClass: object.Class(row.Class), ObjFlags: object.Flags(row.Flags),
				Buffs: row.InitialBuffs, Field131: 19, Frame134: 77, Pos132: types.Ptf(31, 47)}
			*a = Object{TypeInd: 706, ObjClass: 1572865, ObjSubClass: 1, ObjFlags: 553665028, PrevPos: types.Ptf(17, 19)}
			*h = HealthData{Cur: 120, Field2: 119}
			if row.Health == "zero-max" {
				v.HealthData = h
			} else if row.Health != "nil" {
				t.Fatal("unsealed health shape")
			}
			for _, p := range []unsafe.Pointer{unsafe.Pointer(v), unsafe.Pointer(a), unsafe.Pointer(h)} {
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
					t.Fatal("C-owned record is not above 4 GiB")
				}
			}
			beforeV, beforeA, beforeH := *v, *a, *h
			w := (*Object)(nil)
			if row.Weapon {
				w = a
			}
			events := []string{}
			reason := ""
			r := DefaultDamageWorldRuntime4E0B30{
				Frame:         func() uint32 { return 1400 },
				GameplayFlag1: func() bool { return true },
				IsZombie: func(got *Object) bool {
					if got != v {
						t.Fatal("wrong zombie target")
					}
					return false
				},
				IsEnemy: func(got, attacker *Object) bool {
					if got != v || attacker != a {
						t.Fatal("wrong enemy query")
					}
					events = append(events, "enemy")
					return false
				},
				Audio: func(id int, got *Object) {
					if id != 71 || got != v {
						t.Fatal("wrong invulnerability audio")
					}
					events = append(events, "audio")
				},
				BuffOff: func(got *Object, enchant EnchantID) {
					if got != v || enchant != 0 {
						t.Fatal("wrong invisibility removal")
					}
					v.Buffs &^= 1
					events = append(events, "buff")
				},
				DefaultDamageSound: func(got, attacker *Object) {
					if got != v || attacker != a {
						t.Fatal("wrong damage sound")
					}
					events = append(events, "sound")
				},
				ShieldReduce: func(got *Object, d *int32, typ object.DamageType, attacker *Object) {
					if got != v || attacker != a || typ != object.DamageCrush || *d != row.Damage {
						t.Fatal("wrong Shield inputs")
					}
					*d = 0
					events = append(events, "shield")
				},
				DamageClear: func(got *Object, damage int32) {
					if got != v || damage != row.Damage {
						t.Fatal("wrong signed clear inputs")
					}
					events = append(events, "clear")
					// Execute the production health gates, not a supplied no-op.
					unitDamageClearNative4EE5E0(got, damage, unitDamageClearNativeDeps4EE5E0{
						engineFlag: func(uint32) int32 { t.Fatal("nil/zero-Max reached engine flags"); return 0 },
						setHP:      func(*Object, uint16) { t.Fatal("nil/zero-Max wrote HP") },
					})
				},
				Unsupported: func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why },
			}
			got := DefaultDamageWorld4E0B30(v, a, w, row.Damage, object.DamageCrush, r)
			want := beforeV
			want.Buffs, want.Field131, want.Frame134 = row.Buffs, row.DamageType, row.Frame
			want.Pos132 = types.Ptf(math.Float32frombits(row.Position[0]), math.Float32frombits(row.Position[1]))
			if row.Attributed {
				want.Obj130 = a
			}
			if reason != "" || got != (row.Result != 0) || *v != want || !reflect.DeepEqual(events, row.Events) {
				t.Errorf("world no-health: unsupported=%q result=%t/%d attribution=%p/%p frame=%d/%d events=%v/%v", reason, got, row.Result, v.Obj130, want.Obj130, v.Frame134, want.Frame134, events, row.Events)
			}
			if *a != beforeA || *h != beforeH || !row.HealthSafe || !row.GuardsSafe {
				t.Fatal("source/health/guard bytes changed")
			}
		})
	}
	t.Log("WORLD NO-HEALTH: 160 unchanged PE32 cases; original DefaultDamage and UnitDamageClear nil/zero-Max gates; C-owned >4 GiB native records")
}

func TestDefaultDamageWorldNoHealthKeepsRequiredServices4E0B30(t *testing.T) {
	for _, service := range []string{"buff", "clear", "enemy"} {
		for _, weapon := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/weapon-%t", service, weapon), func(t *testing.T) {
				v := &Object{TypeInd: 714, Frame134: 77}
				a := &Object{ObjClass: 1572865}
				w := (*Object)(nil)
				if weapon {
					w = a
				}
				reason := ""
				r := DefaultDamageWorldRuntime4E0B30{
					GameplayFlag1: func() bool { return true },
					IsEnemy:       func(*Object, *Object) bool { return false },
					BuffOff:       func(*Object, EnchantID) { t.Fatal("missing service admitted mutations") },
					DamageClear:   func(*Object, int32) { t.Fatal("missing service reached clear") },
					Unsupported:   func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why },
				}
				want := "missing world projectile tail service"
				switch service {
				case "buff":
					r.BuffOff = nil
				case "clear":
					r.DamageClear = nil
				case "enemy":
					if !weapon {
						return
					} // Original nil weapon skips the enemy query.
					r.IsEnemy = nil
					want = "missing world projectile enemy service"
				}
				before := *v
				if !DefaultDamageWorld4E0B30(v, a, w, 1, object.DamageCrush, r) || reason != want || *v != before {
					t.Fatalf("required service gate=%q want=%q", reason, want)
				}
			})
		}
	}
}
