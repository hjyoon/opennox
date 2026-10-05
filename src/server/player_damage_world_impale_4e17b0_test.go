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

func TestPlayerDamageWorldImpale4E17B0(t *testing.T) {
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
					v, a, w := worldImpaleFixture4E17B0(t, owner)
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
					r.FireProtection = func(*Object) float64 { t.Fatal("IMPALE fire protection"); return 0 }
					r.ElectricArmorScale = func(*Object) float32 { t.Fatal("IMPALE electric protection"); return 0 }
					r.DefaultDamage = func(target, source, weapon *Object, amount int32, typ object.DamageType) bool {
						if target != v || source != a || weapon != w || amount != tc.effective || typ != object.DamageImpale ||
							ud.Field21 != math.Float32bits(tc.next) {
							t.Fatalf("switch tail amount=%d want=%d carry=%g want=%g", amount, tc.effective, math.Float32frombits(ud.Field21), tc.next)
						}
						if a == w {
							if ud.Field76 != 2 || ud.Field75 != 3 {
								t.Fatal("self-weapon damage-type marker")
							}
						} else if ud.Field76 != 1 || ud.Field75 != uint32(w.TypeInd) {
							t.Fatal("distinct missile marker")
						}
						events = append(events, "default")
						return tailResult
					}
					beforeMissile := *w
					h, result := PlayerDamageNative4E17B0(v, a, w, tc.raw, object.DamageImpale, r)
					if !h || result != tailResult || *w != beforeMissile || !slices.Equal(events, []string{"Coop", "exclude", "direction", "GodMode", "Quest", "default"}) {
						t.Fatalf("result=%t/%t events=%v", h, result, events)
					}
				})
			}
		}
	}
}

func TestPlayerDamageWorldImpaleEarly4E17B0(t *testing.T) {
	for _, owner := range []string{"self", "player", "NPC"} {
		for _, gate := range []string{"no-update", "dead", "invulnerable-audio", "invulnerable-silent", "observer", "Coop-self"} {
			t.Run(owner+"/"+gate, func(t *testing.T) {
				v, a, w := worldImpaleFixture4E17B0(t, owner)
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
				h, result := PlayerDamageNative4E17B0(v, a, w, 8, object.DamageImpale, r)
				want := gate == "invulnerable-audio" || gate == "invulnerable-silent"
				if !h || result != want || audio != map[bool]int{false: 0, true: 1}[gate == "invulnerable-audio"] ||
					*ud != before || *ud.Player != beforePlayer || *w != beforeMissile || v.HealthData.Cur != 200 {
					t.Fatalf("early=%t/%t audio=%d", h, result, audio)
				}
			})
		}
	}
}

func TestPlayerDamageWorldImpaleArmorQuest4E17B0(t *testing.T) {
	for _, owner := range []string{"self", "player", "NPC"} {
		for _, mode := range []string{"ordinary", "Quest", "Quest-minimum", "GodMode"} {
			t.Run(owner+"/"+mode, func(t *testing.T) {
				v, a, w := worldImpaleFixture4E17B0(t, owner)
				armor := damageMeleeArmorFixture4E17B0(v, .5, .25)
				ud := v.UpdateDataPlayer()
				ud.Field76, ud.Field75 = 88, 77
				r := damageMeleeRuntimeFixture4E17B0(t)
				damageMeleeArmorRuntime4E17B0(&r, armor, .5)
				var events []string
				lookup, damageArmor := r.ItemArmorValue, r.DamageArmor
				r.ItemArmorValue = func(item *Object) float32 { events = append(events, "lookup"); return lookup(item) }
				r.DamageArmor = func(item, source, attack *Object, amount int32, typ object.DamageType) bool {
					if item != armor || source != a || attack != w || amount != 4 || typ != object.DamageImpale || ud.Field21 != math.Float32bits(.25) {
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
					if target != v || source != a || attack != w || got != amount || typ != object.DamageImpale {
						t.Fatal("late Quest/identity tail")
					}
					events = append(events, "default")
					return tail(target, source, attack, got, typ)
				}
				h, result := PlayerDamageNative4E17B0(v, a, w, 8, object.DamageImpale, r)
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

func TestPlayerDamageWorldImpalePossession4E17B0(t *testing.T) {
	for _, owner := range []string{"self", "player", "NPC"} {
		for _, defense := range []string{"none", "Reflect", "shield", "GreatSword", "staff", "excluded", "berserk", "rear-shield"} {
			t.Run(owner+"/"+defense, func(t *testing.T) {
				v, a, w := worldImpaleFixture4E17B0(t, owner)
				cached := v.UpdateDataPlayer()
				cached.Player.Field3680, cached.Player.CameraFollowObj = 2, w
				cached.Field57, cached.Field21 = math.Float32bits(.25), math.Float32bits(.125)
				cached.Field76, cached.Field75 = 88, 77
				if defense == "shield" || defense == "excluded" || defense == "rear-shield" {
					cached.State, cached.Player.ArmorEquip = PlayerState16, 0x1000000
				}
				if defense == "berserk" {
					cached.State, cached.Player.ArmorEquip = PlayerState1, 0x1000000
				}
				if defense == "GreatSword" {
					cached.Player.WeaponEquip = 0x400
				}
				if defense == "staff" {
					cached.Player.WeaponEquip = 0x8000
				}
				live := &PlayerUpdateData{Player: &Player{}, State: PlayerState13,
					Field57: math.Float32bits(.5), Field21: math.Float32bits(.5), Field76: 31, Field75: 33}
				item := &Object{ObjClass: object.ClassArmor, ObjSubClass: 2, ObjFlags: object.FlagEquipped}
				if defense == "shield" || defense == "berserk" {
					v.InvFirstItem = item
				}
				r := damageMeleeRuntimeFixture4E17B0(t)
				var events []string
				previous := w.PrevPos
				r.ObserveClear = func(target *Object) {
					if target != v || cached.Field76 != 0 || cached.Field75 != 77 || len(events) != 0 {
						t.Fatal("ObserveClear/marker order")
					}
					events = append(events, "observe")
					cached.Field57 = math.Float32bits(.9)
					cached.Player.WeaponEquip, cached.Player.ArmorEquip = 0, 0
					v.UpdateData = unsafe.Pointer(live)
					if defense == "Reflect" {
						v.Buffs |= 1 << 27
					}
				}
				r.BlockSourceExcluded = func(attack *Object) bool {
					if attack != w || cached.Field76 != 0 || cached.Field75 != 77 {
						t.Fatal("exclusion before prefix/after marker")
					}
					events = append(events, "exclude")
					w.TypeInd, w.PrevPos = 530, types.Ptf(-72, 87)
					return defense == "excluded"
				}
				r.BlockDirection = func(target *Object, pos types.Pointf) bool {
					if target != v || pos != previous || v.UpdateDataPlayer() != live ||
						cached.Field76 != map[bool]uint32{false: 1, true: 0}[a == w] {
						t.Fatal("facing lost live update/previous snapshot/cached marker")
					}
					events = append(events, "direction")
					return defense != "none" && defense != "rear-shield"
				}
				r.ProjectileReflect = func(*Object, *Object) { t.Fatal("world spike reflected as missile") }
				r.ClearOwner = func(*Object) { t.Fatal("world spike ownership cleared") }
				r.BerserkShieldBlock = func(target *Object) bool {
					if defense != "berserk" || target != v || live.State != PlayerState13 {
						t.Fatal("berserker cached/live record")
					}
					events = append(events, "berserk")
					return true
				}
				r.Audio = func(id int, target *Object) {
					if id != 878 || target != v {
						t.Fatal("shield audio identity")
					}
					events = append(events, "audio")
				}
				r.BlockDamagePercent = func() float64 { events = append(events, "balance"); return .5 }
				r.CanDamageBlockItem = func(got *Object) bool { return got == item }
				r.DamageBlockItem = func(got, target, source, hazard *Object, amount float32, typ object.DamageType) bool {
					if got != item || target != v || source != a || hazard != w || amount != 2.5 || typ != object.DamageImpale {
						t.Fatal("shield wear identity")
					}
					events = append(events, "wear")
					return true
				}
				tail := r.DefaultDamage
				r.DefaultDamage = func(target, source, hazard *Object, amount int32, typ object.DamageType) bool {
					if target != v || source != a || hazard != w || amount != 4 || typ != object.DamageImpale {
						t.Fatal("cached absorption/live carry tail")
					}
					events = append(events, "default")
					return tail(target, source, hazard, amount, typ)
				}
				h, result := PlayerDamageNative4E17B0(v, a, w, 5, object.DamageImpale, r)
				want := []string{"observe", "exclude", "direction", "default"}
				blocked := defense == "shield" || defense == "berserk"
				if blocked {
					want = []string{"observe", "exclude", "direction"}
					if defense == "berserk" {
						want = append(want, "berserk")
					}
					want = append(want, "audio", "balance", "wear")
				}
				if defense == "excluded" {
					want = []string{"observe", "exclude", "default"}
				}
				if !h || result == blocked || !slices.Equal(events, want) || live.Field76 != 31 || live.Field75 != 33 ||
					cached.Field21 != math.Float32bits(.125) || v.HealthData.Cur != map[bool]uint16{false: 196, true: 200}[blocked] {
					t.Fatalf("possession=%t/%t events=%v want=%v HP=%d", h, result, events, want, v.HealthData.Cur)
				}
				if live.Field21 != math.Float32bits(map[bool]float32{false: .25, true: .5}[blocked]) {
					t.Fatal("live carry")
				}
				if a != w && (cached.Field76 != 1 || cached.Field75 != 530) ||
					a == w && !blocked && (cached.Field76 != 2 || cached.Field75 != 3) ||
					a == w && blocked && (cached.Field76 != 0 || cached.Field75 != 77) {
					t.Fatal("cached attribution")
				}
			})
		}
	}
}

func TestPlayerDamageWorldImpaleShieldLive4E17B0(t *testing.T) {
	for _, kind := range []string{"replacement", "broken", "nil"} {
		t.Run(kind, func(t *testing.T) {
			v, a, w := worldImpaleFixture4E17B0(t, "self")
			ud := v.UpdateDataPlayer()
			ud.State, ud.Player.ArmorEquip = PlayerState16, 0x1000000
			old := &Object{ObjClass: object.ClassArmor, ObjSubClass: 2, ObjFlags: object.FlagEquipped}
			live := &Object{ObjClass: object.ClassArmor, ObjSubClass: 2, ObjFlags: object.FlagEquipped}
			v.InvFirstItem = old
			r := damageMeleeRuntimeFixture4E17B0(t)
			r.BlockDirection = func(*Object, types.Pointf) bool { return true }
			var events []string
			r.Audio = func(id int, target *Object) {
				if id != 878 || target != v {
					t.Fatal("audio")
				}
				events = append(events, "audio")
			}
			r.BlockDamagePercent = func() float64 {
				events = append(events, "balance")
				v.InvFirstItem = live
				if kind == "nil" {
					v.InvFirstItem, live = nil, nil
				}
				return .75
			}
			r.CanDamageBlockItem = func(item *Object) bool { return item == live }
			r.DamageBlockItem = func(item, target, source, hazard *Object, amount float32, typ object.DamageType) bool {
				if item != live || target != v || source != a || hazard != w || amount != 3.75 || typ != object.DamageImpale {
					t.Fatal("live shield selection")
				}
				events = append(events, "wear")
				if kind == "broken" {
					live.ObjFlags |= object.FlagDestroyed
				}
				return true
			}
			r.PlayerSetState = func(target *Object, state PlayerState) bool {
				if kind != "broken" || target != v || state != PlayerState13 {
					t.Fatal("broken shield")
				}
				events = append(events, "state")
				return true
			}
			r.DefaultDamage = func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
				t.Fatal("blocked world IMPALE reached HP")
				return false
			}
			h, result := PlayerDamageNative4E17B0(v, a, w, 5, object.DamageImpale, r)
			want := []string{"audio", "balance", "wear"}
			if kind == "broken" {
				want = append(want, "state")
			}
			if !h || result || !slices.Equal(events, want) || v.HealthData.Cur != 200 {
				t.Fatalf("shield=%t/%t events=%v", h, result, events)
			}
		})
	}
}

func TestPlayerDamageWorldImpaleUseBoundaries4E17B0(t *testing.T) {
	for _, boundary := range []string{"player-info", "observe", "exclude", "direction", "armor", "Quest-scale", "default", "live-projectile", "shield-audio", "berserk-service"} {
		t.Run(boundary, func(t *testing.T) {
			v, a, w := worldImpaleFixture4E17B0(t, "self")
			ud := v.UpdateDataPlayer()
			ud.Field76, ud.Field75, ud.Field57, ud.Field21 = 88, 77, math.Float32bits(.25), math.Float32bits(.25)
			r := damageMeleeRuntimeFixture4E17B0(t)
			var events []string
			r.ObserveClear = func(*Object) { events = append(events, "observe") }
			r.BlockSourceExcluded = func(*Object) bool { events = append(events, "exclude"); return false }
			r.BlockDirection = func(*Object, types.Pointf) bool {
				events = append(events, "direction")
				return boundary == "shield-audio" || boundary == "berserk-service"
			}
			r.GodMode = func() bool { events = append(events, "god"); return false }
			r.QuestMode = func() bool { events = append(events, "quest"); return boundary == "Quest-scale" }
			r.DefaultDamage = func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
				t.Fatal("unsupported reached HP")
				return false
			}
			want := []string{}
			marker, markerType := uint32(0), uint32(77)
			switch boundary {
			case "player-info":
				ud.Player = nil
				marker = 88
			case "observe":
				ud.Player.Field3680, ud.Player.CameraFollowObj = 2, w
				r.ObserveClear = nil
			case "exclude":
				r.BlockSourceExcluded = nil
			case "direction":
				r.BlockDirection = nil
				want = []string{"exclude"}
			case "armor":
				damageMeleeArmorFixture4E17B0(v, .25, .25)
				want = []string{"exclude", "direction"}
			case "Quest-scale", "default":
				want = []string{"exclude", "direction", "god", "quest"}
				marker, markerType = 2, 3
				r.DefaultDamage = nil
			case "live-projectile":
				ud.Player.Field3680, ud.Player.CameraFollowObj = 2, w
				r.ObserveClear = func(*Object) {
					events = append(events, "observe")
					v.Buffs |= 1 << 27
					w.ObjClass |= object.ClassMissile
				}
				want = []string{"observe"}
			case "shield-audio":
				ud.State, ud.Player.ArmorEquip = PlayerState16, 0x1000000
				want = []string{"exclude", "direction"}
			case "berserk-service":
				ud.State, ud.Player.ArmorEquip = PlayerState1, 0x1000000
				want = []string{"exclude", "direction"}
			}
			reason := ""
			r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
			h, result := PlayerDamageNative4E17B0(v, a, w, 8, object.DamageImpale, r)
			if h || result || reason == "" || !slices.Equal(events, want) || ud.Field76 != marker || ud.Field75 != markerType ||
				ud.Field21 != math.Float32bits(.25) || v.HealthData.Cur != 200 {
				t.Fatalf("boundary=%t/%t reason=%s events=%v want=%v marker=%d/%d", h, result, reason, events, want, ud.Field76, ud.Field75)
			}
		})
	}
}
