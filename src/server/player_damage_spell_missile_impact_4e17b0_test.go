package server

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestPlayerDamageSpellMissileImpact4E17B0(t *testing.T) {
	for _, owner := range []string{"self", "player", "NPC"} {
		for _, tc := range []struct {
			name               string
			raw, effective     int32
			armor, carry, next float32
		}{
			{"raw-8", 8, 8, 0, 0, 0},
			{"armored-8", 8, 6, .25, .25, .25},
			{"positive-minimum", 1, 1, .5, 0, .5},
			{"tie-even-down", 5, 2, .5, 0, .5},
			{"tie-even-up", 3, 2, .5, 0, -.5},
			{"raw-zero", 0, 0, .25, .25, .25},
			{"negative", -3, -2, .25, .25, 0},
			{"negative-tie", -5, -2, .5, 0, -.5},
		} {
			for _, tailResult := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/default-%t", owner, tc.name, tailResult), func(t *testing.T) {
					v, a, w := spellMissileImpactFixture4E17B0(t, owner)
					ud := v.UpdateDataPlayer()
					ud.Field57, ud.Field21 = math.Float32bits(tc.armor), math.Float32bits(tc.carry)
					ud.Field76, ud.Field75 = 88, 77
					r := damageMeleeRuntimeFixture4E17B0(t)
					var events []string
					r.Frame = func() uint32 { t.Fatal("unbuffed entry read frame"); return 0 }
					r.CoopMode = func() bool { events = append(events, "Coop"); return false }
					r.BlockSourceExcluded = func(attack *Object) bool {
						if attack != w || ud.Field76 != 0 || ud.Field75 != 77 {
							t.Fatal("exclusion preceded marker clear or changed identity")
						}
						events = append(events, "exclude")
						return false
					}
					r.BlockDirection = func(target *Object, pos types.Pointf) bool {
						if target != v || pos != w.PrevPos || ud.Field76 != map[bool]uint32{false: 1, true: 0}[a == w] {
							t.Fatal("facing identity/position/marker")
						}
						events = append(events, "direction")
						return false
					}
					r.GodMode = func() bool { events = append(events, "GodMode"); return false }
					r.QuestMode = func() bool { events = append(events, "Quest"); return false }
					r.FireProtection = func(*Object) float64 { t.Fatal("IMPACT fire protection"); return 0 }
					r.ElectricArmorScale = func(*Object) float32 { t.Fatal("IMPACT electric protection"); return 0 }
					r.DefaultDamage = func(target, source, weapon *Object, amount int32, typ object.DamageType) bool {
						if target != v || source != a || weapon != w || amount != tc.effective || typ != object.DamageImpact ||
							ud.Field21 != math.Float32bits(tc.next) {
							t.Fatalf("switch tail amount=%d want=%d carry=%g want=%g", amount, tc.effective, math.Float32frombits(ud.Field21), tc.next)
						}
						if a == w {
							if ud.Field76 != 2 || ud.Field75 != 11 {
								t.Fatal("self-weapon damage-type marker")
							}
						} else if ud.Field76 != 1 || ud.Field75 != uint32(w.TypeInd) {
							t.Fatal("distinct missile marker")
						}
						events = append(events, "default")
						return tailResult
					}
					beforeMissile := *w
					h, result := PlayerDamageNative4E17B0(v, a, w, tc.raw, object.DamageImpact, r)
					if !h || result != tailResult || *w != beforeMissile || !slices.Equal(events, []string{"Coop", "exclude", "direction", "GodMode", "Quest", "default"}) {
						t.Fatalf("result=%t/%t events=%v", h, result, events)
					}
				})
			}
		}
	}
}

func TestPlayerDamageSpellMissileImpactEarly4E17B0(t *testing.T) {
	for _, owner := range []string{"self", "player", "NPC"} {
		for _, gate := range []string{"no-update", "dead", "invulnerable-audio", "invulnerable-silent", "observer", "Coop-self"} {
			t.Run(owner+"/"+gate, func(t *testing.T) {
				v, a, w := spellMissileImpactFixture4E17B0(t, owner)
				ud := v.UpdateDataPlayer()
				ud.Field76, ud.Field75 = 88, 77
				ud.Player.Field3680, ud.Player.CameraFollowObj = 2, w
				r := damageMeleeRuntimeFixture4E17B0(t)
				switch gate {
				case "no-update":
					v.ObjFlags |= object.FlagNoUpdate
				case "dead":
					v.ObjFlags |= object.FlagDead
				case "invulnerable-audio", "invulnerable-silent":
					v.Buffs |= 1 << 23
					r.Frame = func() uint32 {
						if gate == "invulnerable-silent" {
							return 1401
						}
						return 1400
					}
				case "observer":
					ud.Player.Field3680 |= 1
				case "Coop-self":
					w.ObjOwner, a = v, v
					r.CoopMode = func() bool { return true }
				}
				audio := 0
				r.Audio = func(id int, target *Object) {
					if gate != "invulnerable-audio" || id != 71 || target != v {
						t.Fatal("early audio identity")
					}
					audio++
				}
				r.ObserveClear = func(*Object) { t.Fatal("early gate cleared possession") }
				r.BlockSourceExcluded = func(*Object) bool { t.Fatal("early gate reached exclusion"); return false }
				r.DefaultDamage = func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
					t.Fatal("early gate reached HP")
					return false
				}
				before, beforePlayer, beforeMissile := *ud, *ud.Player, *w
				h, result := PlayerDamageNative4E17B0(v, a, w, 8, object.DamageImpact, r)
				want := gate == "invulnerable-audio" || gate == "invulnerable-silent"
				if !h || result != want || audio != map[bool]int{false: 0, true: 1}[gate == "invulnerable-audio"] ||
					*ud != before || *ud.Player != beforePlayer || *w != beforeMissile || v.HealthData.Cur != 200 {
					t.Fatalf("early=%t/%t audio=%d", h, result, audio)
				}
			})
		}
	}
}

func TestPlayerDamageSpellMissileImpactPossession4E17B0(t *testing.T) {
	for _, owner := range []string{"self", "player", "NPC"} {
		for _, defense := range []string{"none", "Reflect", "shield", "GreatSword", "excluded", "rear-Reflect"} {
			t.Run(owner+"/"+defense, func(t *testing.T) {
				v, a, w := spellMissileImpactFixture4E17B0(t, owner)
				cached := v.UpdateDataPlayer()
				cached.Player.Field3680, cached.Player.CameraFollowObj = 2, w
				cached.Field57, cached.Field21 = math.Float32bits(.25), math.Float32bits(.125)
				cached.Field76, cached.Field75 = 88, 77
				if defense == "shield" || defense == "excluded" {
					cached.State, cached.Player.ArmorEquip = PlayerState16, 0x1000000
				}
				if defense == "GreatSword" {
					cached.Player.WeaponEquip = 0x400
				}
				live := &PlayerUpdateData{Player: &Player{Field3680: 2, CameraFollowObj: a}, State: PlayerState13,
					Field57: math.Float32bits(.5), Field21: math.Float32bits(.5), Field76: 31, Field75: 33}
				item := &Object{ObjClass: object.ClassArmor, ObjSubClass: 2, ObjFlags: object.FlagEquipped, HealthData: &HealthData{Cur: 30}}
				if defense == "GreatSword" {
					item.ObjClass, item.ObjSubClass = object.ClassWeapon, 0x400
				}
				if defense == "shield" || defense == "GreatSword" {
					v.InvFirstItem = item
				}
				r := damageMeleeRuntimeFixture4E17B0(t)
				var events []string
				previousPosition := w.PrevPos
				r.ObserveClear = func(target *Object) {
					if target != v || cached.Field76 != 0 || cached.Field75 != 77 || len(events) != 0 {
						t.Fatal("ObserveClear/marker ordering")
					}
					events = append(events, "observe")
					cached.Field57 = math.Float32bits(.9)
					cached.Player.WeaponEquip, cached.Player.ArmorEquip = 0, 0
					v.UpdateData = unsafe.Pointer(live)
					if defense == "Reflect" || defense == "rear-Reflect" {
						v.Buffs |= 1 << 27
					}
				}
				r.BlockSourceExcluded = func(attack *Object) bool {
					if attack != w || cached.Field76 != 0 || cached.Field75 != 77 {
						t.Fatal("exclusion preceded prefix or attribution")
					}
					events = append(events, "exclude")
					w.TypeInd, w.PrevPos = 530, types.Ptf(-72, 87)
					return defense == "excluded"
				}
				r.BlockDirection = func(target *Object, pos types.Pointf) bool {
					if target != v || v.UpdateDataPlayer() != live {
						t.Fatal("direction identity/live update")
					}
					if len(events) == 1 && (defense == "Reflect" || defense == "rear-Reflect") {
						if pos != w.PosVec || cached.Field76 != 0 {
							t.Fatal("Reflect uses current position before exclusion")
						}
					} else if pos != previousPosition || cached.Field76 != map[bool]uint32{false: 1, true: 0}[a == w] {
						t.Fatal("ordinary facing lost previous-position snapshot/cached marker")
					}
					events = append(events, "direction")
					return defense == "Reflect" || defense == "shield" || defense == "GreatSword"
				}
				r.ProjectileReflect = func(attack, target *Object) {
					if attack != w || target != v {
						t.Fatal("reflection identity")
					}
					events = append(events, "reflect")
				}
				r.ClearOwner = func(attack *Object) {
					if attack != w {
						t.Fatal("owner identity")
					}
					events = append(events, "clear")
				}
				r.SetOwner = func(target, attack *Object) {
					if target != v || attack != w {
						t.Fatal("owner identity")
					}
					events = append(events, "owner")
				}
				r.ChangeOwner = func(attack, target *Object) {
					if attack != w || target != v || defense != "Reflect" {
						t.Fatal("ChangeOwner branch")
					}
					events = append(events, "change")
				}
				r.Audio = func(id int, target *Object) {
					if target != v || id != map[string]int{"Reflect": 122, "shield": 878, "GreatSword": 890}[defense] {
						t.Fatal("defense audio")
					}
					events = append(events, "audio")
				}
				r.BlockDamagePercent = func() float64 { events = append(events, "balance"); return .5 }
				r.CanDamageBlockItem = func(got *Object) bool { return got == item }
				r.Melee.CanDamageBlockWeapon = r.CanDamageBlockItem
				wear := func(got, target, source, attack *Object, amount float32, typ object.DamageType) bool {
					if got != item || target != v || source != a || attack != w || amount != 2.5 || typ != object.DamageImpact {
						t.Fatal("block wear identity")
					}
					events = append(events, "wear")
					return true
				}
				r.DamageBlockItem, r.Melee.DamageBlockWeapon = wear, wear
				r.Melee.RandomInt = func(min, max int) int {
					if min != 18 || max != 20 {
						t.Fatal("RNG bounds")
					}
					events = append(events, "rng")
					return 19
				}
				r.PlayerSetState = func(target *Object, state PlayerState) bool {
					if target != v || state != 19 {
						t.Fatal("block stance")
					}
					events = append(events, "state")
					live.State = state
					return true
				}
				tail := r.DefaultDamage
				r.DefaultDamage = func(target, source, attack *Object, amount int32, typ object.DamageType) bool {
					if target != v || source != a || attack != w || amount != 4 || typ != object.DamageImpact {
						t.Fatal("cached absorption/live carry tail")
					}
					events = append(events, "default")
					return tail(target, source, attack, amount, typ)
				}
				h, result := PlayerDamageNative4E17B0(v, a, w, 5, object.DamageImpact, r)
				want := []string{"observe", "exclude", "direction", "default"}
				switch defense {
				case "Reflect":
					want = []string{"observe", "direction", "reflect", "clear", "owner", "change", "audio"}
				case "shield":
					want = []string{"observe", "exclude", "direction", "audio", "reflect", "balance", "wear"}
				case "GreatSword":
					want = []string{"observe", "exclude", "direction", "reflect", "audio", "rng", "state", "balance", "wear"}
				case "excluded":
					want = []string{"observe", "exclude", "default"}
				case "rear-Reflect":
					want = []string{"observe", "direction", "exclude", "direction", "default"}
				}
				damaged := defense == "none" || defense == "excluded" || defense == "rear-Reflect"
				if !h || result != damaged || !slices.Equal(events, want) || live.Field76 != 31 || live.Field75 != 33 || cached.Field21 != math.Float32bits(.125) {
					t.Fatalf("possession=%t/%t events=%v want=%v", h, result, events, want)
				}
				if damaged {
					if v.HealthData.Cur != 196 || live.Field21 != math.Float32bits(.25) {
						t.Fatal("live HP/carry")
					}
					if a == w {
						if cached.Field76 != 2 || cached.Field75 != 11 {
							t.Fatal("self marker")
						}
					} else if cached.Field76 != 1 || cached.Field75 != 530 {
						t.Fatal("distinct marker")
					}
				} else if v.HealthData.Cur != 200 || live.Field21 != math.Float32bits(.5) {
					t.Fatal("blocked damage reached HP/carry")
				}
			})
		}
	}
}

func TestPlayerDamageSpellMissileImpactArmorQuest4E17B0(t *testing.T) {
	for _, owner := range []string{"self", "player", "NPC"} {
		for _, mode := range []string{"ordinary", "Quest", "Quest-minimum", "GodMode"} {
			t.Run(owner+"/"+mode, func(t *testing.T) {
				v, a, w := spellMissileImpactFixture4E17B0(t, owner)
				armor := damageMeleeArmorFixture4E17B0(v, .5, .25)
				ud := v.UpdateDataPlayer()
				ud.Field76, ud.Field75 = 88, 77
				r := damageMeleeRuntimeFixture4E17B0(t)
				damageMeleeArmorRuntime4E17B0(&r, armor, .5)
				var events []string
				lookup, damageArmor := r.ItemArmorValue, r.DamageArmor
				r.ItemArmorValue = func(item *Object) float32 { events = append(events, "lookup"); return lookup(item) }
				r.DamageArmor = func(item, source, attack *Object, amount int32, typ object.DamageType) bool {
					if item != armor || source != a || attack != w || amount != 4 || typ != object.DamageImpact || ud.Field21 != math.Float32bits(.25) {
						t.Fatal("armor use before live carry or wrong amount")
					}
					events = append(events, "wear")
					return damageArmor(item, source, attack, amount, typ)
				}
				r.GodMode = func() bool {
					if armor.HealthData.Cur != 21 {
						t.Fatal("GodMode preceded armor")
					}
					events = append(events, "god")
					return mode == "GodMode"
				}
				r.QuestMode = func() bool { events = append(events, "quest"); return mode == "Quest" || mode == "Quest-minimum" }
				r.QuestDamageScale = func() float32 {
					events = append(events, "scale")
					if mode == "Quest-minimum" {
						return .01
					}
					return .5
				}
				amount := int32(4)
				if mode == "Quest" {
					amount = 2
				}
				if mode == "Quest-minimum" {
					amount = 1
				}
				tail := r.DefaultDamage
				r.DefaultDamage = func(target, source, attack *Object, got int32, typ object.DamageType) bool {
					if target != v || source != a || attack != w || got != amount || typ != object.DamageImpact {
						t.Fatal("late Quest/identity tail")
					}
					events = append(events, "default")
					return tail(target, source, attack, got, typ)
				}
				h, result := PlayerDamageNative4E17B0(v, a, w, 8, object.DamageImpact, r)
				want := []string{"lookup", "wear", "god"}
				wantHP := uint16(200)
				if mode != "GodMode" {
					want = append(want, "quest")
					if mode != "ordinary" {
						want = append(want, "scale")
					}
					want = append(want, "default")
					wantHP -= uint16(amount)
				}
				if !h || !result || !slices.Equal(events, want) || v.HealthData.Cur != wantHP || armor.HealthData.Cur != 21 ||
					ud.Field21 != math.Float32bits(.25) || math.Abs(float64(*(*float32)(armor.UpdateData)-.4)) > 1e-6 {
					t.Fatalf("armor/Quest=%t/%t events=%v want=%v HP=%d", h, result, events, want, v.HealthData.Cur)
				}
			})
		}
	}
}

func TestPlayerDamageSpellMissileImpactUseBoundaries4E17B0(t *testing.T) {
	for _, boundary := range []string{"player-info", "observe", "exclude", "direction", "Reflect-effect", "Reflect-owner", "Reflect-change", "Reflect-audio", "armor", "Quest-scale", "default"} {
		t.Run(boundary, func(t *testing.T) {
			v, a, w := spellMissileImpactFixture4E17B0(t, "self")
			ud := v.UpdateDataPlayer()
			ud.Field76, ud.Field75, ud.Field57, ud.Field21 = 88, 77, math.Float32bits(.25), math.Float32bits(.25)
			r := damageMeleeRuntimeFixture4E17B0(t)
			var events []string
			r.ObserveClear = func(*Object) { events = append(events, "observe") }
			r.BlockSourceExcluded = func(*Object) bool { events = append(events, "exclude"); return false }
			r.BlockDirection = func(*Object, types.Pointf) bool { events = append(events, "direction"); return false }
			r.ProjectileReflect = func(*Object, *Object) { events = append(events, "reflect") }
			r.ClearOwner = func(*Object) { events = append(events, "clear") }
			r.SetOwner = func(*Object, *Object) { events = append(events, "owner") }
			r.ChangeOwner = func(*Object, *Object) { events = append(events, "change") }
			r.Audio = func(int, *Object) { events = append(events, "audio") }
			r.GodMode = func() bool { events = append(events, "god"); return false }
			r.QuestMode = func() bool { events = append(events, "quest"); return boundary == "Quest-scale" }
			r.DefaultDamage = func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
				t.Fatal("unsupported use reached default")
				return false
			}
			want := []string{}
			marker, markerType, carry := uint32(0), uint32(77), float32(.25)
			switch boundary {
			case "player-info":
				ud.Player = nil
				marker = 88
			case "observe":
				ud.Player.Field3680 = 2
				ud.Player.CameraFollowObj = w
				r.ObserveClear = nil
			case "exclude":
				r.BlockSourceExcluded = nil
			case "direction":
				r.BlockDirection = nil
				want = []string{"exclude"}
			case "Reflect-effect", "Reflect-owner", "Reflect-change", "Reflect-audio":
				v.Buffs |= 1 << 27
				r.BlockDirection = func(*Object, types.Pointf) bool { events = append(events, "direction"); return true }
				want = []string{"direction"}
				switch boundary {
				case "Reflect-effect":
					r.ProjectileReflect = nil
				case "Reflect-owner":
					r.ClearOwner = nil
					want = append(want, "reflect")
				case "Reflect-change":
					r.ChangeOwner = nil
					want = append(want, "reflect", "clear", "owner")
				case "Reflect-audio":
					r.Audio = nil
					want = append(want, "reflect", "clear", "owner", "change")
				}
			case "armor":
				damageMeleeArmorFixture4E17B0(v, .25, .25)
				want = []string{"exclude", "direction"}
			case "Quest-scale", "default":
				want = []string{"exclude", "direction", "god", "quest"}
				marker, markerType = 2, 11
				r.DefaultDamage = nil
			}
			reason := ""
			r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
			h, result := PlayerDamageNative4E17B0(v, a, w, 8, object.DamageImpact, r)
			if h || result || reason == "" || !slices.Equal(events, want) || ud.Field76 != marker || ud.Field75 != markerType ||
				ud.Field21 != math.Float32bits(carry) || v.HealthData.Cur != 200 {
				t.Fatalf("use=%t/%t reason=%q events=%v want=%v marker=%d/%d", h, result, reason, events, want, ud.Field76, ud.Field75)
			}
		})
	}
}

func TestPlayerDamageSpellMissileImpactReflectLive4E17B0(t *testing.T) {
	for _, tc := range []struct {
		name                        string
		afterReflect                object.SubClass
		nonmissile, clearAfterOwner bool
		want                        []string
	}{
		{"Pixie", 3, false, false, []string{"direction", "reflect", "clear", "owner", "change", "audio"}},
		{"retain-bit40", 0x42, false, false, []string{"direction", "reflect", "change", "audio"}},
		{"nonmissile-after-reflect", 3, true, false, []string{"direction", "reflect", "clear", "owner", "audio"}},
		{"clear-bit2-after-owner", 3, false, true, []string{"direction", "reflect", "clear", "owner", "audio"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v, a, w := spellMissileImpactFixture4E17B0(t, "self")
			v.Buffs |= 1 << 27
			ud := v.UpdateDataPlayer()
			ud.Field76, ud.Field75 = 88, 77
			r := damageMeleeRuntimeFixture4E17B0(t)
			var events []string
			r.BlockDirection = func(target *Object, pos types.Pointf) bool {
				if target != v || pos != w.PosVec || ud.Field76 != 0 || ud.Field75 != 77 {
					t.Fatal("Reflect prefix")
				}
				events = append(events, "direction")
				return true
			}
			r.ProjectileReflect = func(attack, target *Object) {
				if attack != w || target != v {
					t.Fatal("Reflect identity")
				}
				events = append(events, "reflect")
				w.ObjSubClass = tc.afterReflect
				if tc.nonmissile {
					w.ObjClass = object.ClassSimple
				}
			}
			r.ClearOwner = func(*Object) { events = append(events, "clear") }
			r.SetOwner = func(target, attack *Object) {
				if target != v || attack != w {
					t.Fatal("ownership identity")
				}
				events = append(events, "owner")
				if tc.clearAfterOwner {
					w.ObjSubClass = 0
				}
			}
			r.ChangeOwner = func(*Object, *Object) { events = append(events, "change") }
			r.Audio = func(id int, target *Object) {
				if id != 122 || target != v {
					t.Fatal("Reflect audio")
				}
				events = append(events, "audio")
			}
			r.BlockSourceExcluded = func(*Object) bool { t.Fatal("Reflect reached exclusion"); return false }
			r.DefaultDamage = func(_, _, _ *Object, _ int32, _ object.DamageType) bool { t.Fatal("Reflect reached HP"); return false }
			h, result := PlayerDamageNative4E17B0(v, a, w, 8, object.DamageImpact, r)
			if !h || result || !slices.Equal(events, tc.want) || ud.Field76 != 0 || ud.Field75 != 77 || v.HealthData.Cur != 200 {
				t.Fatalf("live Reflect=%t/%t events=%v want=%v", h, result, events, tc.want)
			}
		})
	}
}
