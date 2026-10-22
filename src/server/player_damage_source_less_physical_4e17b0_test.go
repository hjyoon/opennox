package server

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestSourceLessPhysicalPrefixMatchesGAMEEXE4E17B0(t *testing.T) {
	data, err := os.ReadFile("testdata/player_damage_source_less_physical_4e17b0.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		OriginalSHA256 string `json:"original_sha256"`
		Rows           []struct {
			Type    object.DamageType `json:"damage_type"`
			Player  bool              `json:"player"`
			Raw     int32             `json:"raw"`
			Forward uint32            `json:"forward_damage"`
			Marker  uint32            `json:"marker"`
			Kind    uint32            `json:"marker_type"`
			Carry   uint32            `json:"carry_bits"`
			Wear    uint32            `json:"wear"`
			Return  int               `json:"return"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA256 != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(fixture.Rows) != 48 {
		t.Fatal("unexpected original source-less physical fixture")
	}
	for _, row := range fixture.Rows {
		t.Run(fmt.Sprintf("player-%t/type-%d/raw-%d", row.Player, row.Type, row.Raw), func(t *testing.T) {
			v := damageMeleeUnitFixture4E17B0(t, row.Player)
			armor := damageMeleeArmorFixture4E17B0(v, .5, .25)
			r := damageMeleeRuntimeFixture4E17B0(t)
			damageMeleeArmorRuntime4E17B0(&r, armor, .5)
			wear, forwarded, calls := int32(0), int32(123456), 0
			r.DamageArmor = func(item, a, w *Object, d int32, typ object.DamageType) bool {
				if item != armor || a != nil || w != nil || typ != row.Type {
					t.Fatal("source-less armor boundary changed")
				}
				wear += d
				return true
			}
			r.DefaultDamage = func(target, a, w *Object, d int32, typ object.DamageType) bool {
				if target != v || a != nil || w != nil || typ != row.Type {
					t.Fatal("source-less default boundary changed")
				}
				calls++
				forwarded = d
				return true
			}
			r.CoopMode = func() bool { return false }
			r.GodMode = func() bool { return false }
			r.QuestMode = func() bool { return false }
			r.BlockDirection, r.BlockSourceExcluded, r.BlockSourceOnlyExcluded = nil, nil, nil
			h, ok := PlayerDamageNative4E17B0(v, nil, nil, row.Raw, row.Type, r)
			marker, kind, carry := damageMeleeMarker4E17B0(v)
			wantWear := max(int32(row.Wear), 0) // The item helper returns before callbacks for nonpositive wear.
			if !h || ok != (row.Return != 0) || calls != 1 || uint32(forwarded) != row.Forward ||
				marker != row.Marker || kind != row.Kind || math.Float32bits(carry) != row.Carry || wear != wantWear {
				t.Fatalf("prefix handled/result=%t/%t forward=%d/%d calls=%d marker=%d/%d carry=%x wear=%d/%d", h, ok, forwarded, int32(row.Forward), calls, marker, kind, math.Float32bits(carry), wear, wantWear)
			}
		})
	}
}

// The original source-less prefix skips facing and weapon defenses. BLADE
// and BITE enter full armor; CRUSH enters half armor. DefaultDamage must
// then accept the player without running that armor calculation again.
func TestPlayerDamageSourceLessPhysical4E17B0ArmorAndHP(t *testing.T) {
	for _, tc := range []struct {
		typ       object.DamageType
		hp, armor uint16
		carry     float32
	}{
		{object.DamageBlade, 195, 21, -.25},
		{object.DamageCrush, 193, 23, 0},
		{object.DamageBite, 195, 21, -.25},
	} {
		t.Run(fmt.Sprint(tc.typ), func(t *testing.T) {
			v := damageMeleeUnitFixture4E17B0(t, true)
			armor := damageMeleeArmorFixture4E17B0(v, .5, .25)
			r := damageMeleeRuntimeFixture4E17B0(t)
			damageMeleeArmorRuntime4E17B0(&r, armor, .5)
			r.BlockDirection = nil
			r.BlockSourceExcluded = nil
			r.BlockSourceOnlyExcluded = nil
			if h, ok := PlayerDamageNative4E17B0(v, nil, nil, 9, tc.typ, r); !h || !ok || v.HealthData.Cur != tc.hp {
				t.Fatalf("source-less physical handled/result=%t/%t HP=%d want=%d", h, ok, v.HealthData.Cur, tc.hp)
			}
			marker, kind, carry := damageMeleeMarker4E17B0(v)
			if marker != 2 || kind != uint32(tc.typ) || carry != tc.carry || armor.HealthData.Cur != tc.armor ||
				v.Obj130 != nil || v.Field131 != uint32(tc.typ) || v.Frame134 != 1400 {
				t.Fatalf("marker=%d/%d carry=%g armor=%d attribution=%p", marker, kind, carry, armor.HealthData.Cur, v.Obj130)
			}
		})
	}
}

func TestPlayerDamageSourceLessPhysical4E17B0Gates(t *testing.T) {
	for _, typ := range []object.DamageType{object.DamageBlade, object.DamageCrush, object.DamageBite} {
		for _, gate := range []string{"dead", "disabled", "invulnerable", "observer", "coop", "reflect", "god", "quest"} {
			t.Run(fmt.Sprintf("type-%d/%s", typ, gate), func(t *testing.T) {
				v := damageMeleeUnitFixture4E17B0(t, true)
				armor := damageMeleeArmorFixture4E17B0(v, .5, .25)
				r := damageMeleeRuntimeFixture4E17B0(t)
				damageMeleeArmorRuntime4E17B0(&r, armor, .5)
				var events []string
				r.Frame = func() uint32 { return 1400 }
				r.Audio = func(id int, target *Object) {
					if id != 71 || target != v {
						t.Fatal("unexpected source-less audio")
					}
					events = append(events, "audio")
				}
				r.CoopMode = func() bool { events = append(events, "coop"); return gate == "coop" }
				r.GodMode = func() bool { events = append(events, "god"); return gate == "god" }
				r.QuestMode = func() bool { events = append(events, "quest"); return gate == "quest" }
				r.QuestDamageScale = func() float32 { events = append(events, "scale"); return .5 }
				r.BlockDirection = nil
				r.BlockSourceExcluded = nil
				r.BlockSourceOnlyExcluded = nil
				r.DamageArmor = func(_, a, w *Object, _ int32, k object.DamageType) bool {
					if a != nil || w != nil || k != typ {
						t.Fatal("source-less wear acquired an attacker")
					}
					events = append(events, "wear")
					return true
				}
				forward := int32(0)
				r.DefaultDamage = func(_, a, w *Object, d int32, k object.DamageType) bool {
					if a != nil || w != nil || k != typ {
						t.Fatal("source-less HP acquired an attacker")
					}
					forward = d
					events = append(events, "default")
					return true
				}
				switch gate {
				case "dead":
					v.ObjFlags |= object.FlagDead
				case "disabled":
					v.ObjFlags |= object.FlagNoUpdate
				case "invulnerable":
					v.Buffs = 1 << playerDamageInvulnerableEnchant4E17B0
				case "observer":
					v.UpdateDataPlayer().Player.Field3680 = 1
				case "reflect":
					v.Buffs = 1 << playerDamageReflectEnchant4E17B0
				}
				h, ok := PlayerDamageNative4E17B0(v, nil, nil, 9, typ, r)
				wantEvents := []string{"coop", "wear", "god", "quest", "default"}
				wantForward, wantOK := int32(5), true
				if typ == object.DamageCrush {
					wantForward = 7
				}
				switch gate {
				case "dead", "disabled":
					wantEvents, wantForward, wantOK = nil, 0, false
				case "invulnerable":
					wantEvents, wantForward = []string{"audio"}, 0
				case "observer":
					wantEvents, wantForward, wantOK = []string{"coop"}, 0, false
				case "god":
					wantEvents, wantForward = []string{"coop", "wear", "god"}, 0
				case "quest":
					wantEvents = []string{"coop", "wear", "god", "quest", "scale", "default"}
					wantForward = 2
					if typ == object.DamageCrush {
						wantForward = 4
					}
				}
				if !h || ok != wantOK || forward != wantForward || fmt.Sprint(events) != fmt.Sprint(wantEvents) {
					t.Fatalf("gates handled/result=%t/%t forward=%d/%d events=%v/%v", h, ok, forward, wantForward, events, wantEvents)
				}
			})
		}
	}
}

func TestPlayerDamageSourceLessPhysical4E17B0CachedArmorAndLiveCarry(t *testing.T) {
	v := damageMeleeUnitFixture4E17B0(t, true)
	armor := damageMeleeArmorFixture4E17B0(v, .5, .125)
	cached := v.UpdateDataPlayer()
	cached.Player.Field3680, cached.Player.CameraFollowObj = 2, v
	cached.Field76, cached.Field75 = 99, 77
	live := &PlayerUpdateData{Player: &Player{}, Field57: math.Float32bits(.9), Field21: math.Float32bits(.25), Field76: 71, Field75: 73}
	r := damageMeleeRuntimeFixture4E17B0(t)
	damageMeleeArmorRuntime4E17B0(&r, armor, .5)
	observations := 0
	r.ObserveClear = func(target *Object) {
		if target != v || cached.Field76 != 0 {
			t.Fatal("cached marker was not cleared before ObserveClear")
		}
		observations++
		v.UpdateData = unsafe.Pointer(live)
	}
	r.DamageArmor = func(_, a, w *Object, d int32, typ object.DamageType) bool {
		// The outer wear amount is4. The item helper distributes it using
		// live absorption .9 and item value .5, then its existing .4 carry:
		// round(4*.5/.9+.4) ==3. HP absorption still uses cached .5.
		if a != nil || w != nil || d != 3 || typ != object.DamageBlade || live.Field21 != math.Float32bits(-.25) {
			t.Fatal("wear did not follow cached absorption and live carry")
		}
		return true
	}
	r.DefaultDamage = func(*Object, *Object, *Object, int32, object.DamageType) bool { return true }
	if h, ok := PlayerDamageNative4E17B0(v, nil, nil, 9, object.DamageBlade, r); !h || !ok || observations != 1 ||
		cached.Field76 != 2 || cached.Field75 != 0 || cached.Field21 != math.Float32bits(.125) ||
		live.Field76 != 71 || live.Field75 != 73 || live.Field21 != math.Float32bits(-.25) {
		t.Fatal("source-less cached marker/live carry contract changed")
	}
}

func TestPlayerDamageSourceLessPhysical4E17B0LateDefendAndShield(t *testing.T) {
	for _, typ := range []object.DamageType{object.DamageBlade, object.DamageCrush, object.DamageBite} {
		t.Run(fmt.Sprint(typ), func(t *testing.T) {
			v := damageMeleeUnitFixture4E17B0(t, true)
			armor := damageMeleeArmorFixture4E17B0(v, .5, .25)
			effect := &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
			armor.InitDataModifier().Modifiers[2] = effect
			v.Buffs = 1 << defaultDamageShieldEnchant4E0B30
			r := damageMeleeRuntimeFixture4E17B0(t)
			damageMeleeArmorRuntime4E17B0(&r, armor, .5)
			world := damageMeleeWorldRuntime4E0B30(t)
			var events []string
			world.CanApplyLateDefend = func(m *ModifierEff) bool { return m == effect }
			world.ApplyLateDefend = func(m *ModifierEff, item, target, weapon, source *Object, d int32, k object.DamageType) int32 {
				if m != effect || item != armor || target != v || weapon != nil || source != nil || k != typ {
					t.Fatal("source-less late Defend arguments changed")
				}
				events = append(events, "defend")
				return d * 2
			}
			world.ShieldReduce = func(target *Object, d *int32, k object.DamageType, source *Object) {
				if target != v || k != typ || source != nil {
					t.Fatal("source-less Shield arguments changed")
				}
				events = append(events, "shield")
				*d -= 2
			}
			r.DefaultDamage = func(v, a, w *Object, d int32, k object.DamageType) bool {
				return DefaultDamageWorld4E0B30(v, a, w, d, k, world)
			}
			wantHP := uint16(192) // Full armor:5, late Defend10, Shield8.
			if typ == object.DamageCrush {
				wantHP = 188 // Half armor:7, late Defend14, Shield12.
			}
			if h, ok := PlayerDamageNative4E17B0(v, nil, nil, 9, typ, r); !h || !ok || v.HealthData.Cur != wantHP || fmt.Sprint(events) != "[defend shield]" {
				t.Fatalf("late source-less tail handled/result=%t/%t HP=%d/%d events=%v", h, ok, v.HealthData.Cur, wantHP, events)
			}
		})
	}
}
